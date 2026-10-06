package tariff

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const tariffColumns = "id, service, zone, valid_from, currency, tier_mode, created_at, updated_at"

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, t Tariff) (Tariff, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Tariff{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	const q = `
		INSERT INTO tariffs (id, service, zone, valid_from, currency, tier_mode)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING ` + tariffColumns

	var out Tariff
	err = scanTariff(tx.QueryRow(ctx, q, t.ID, t.Service, t.Zone, t.ValidFrom, t.Currency, t.TierMode), &out)
	if err != nil {
		if mapped := tariffWriteError(err); mapped != nil {
			return Tariff{}, mapped
		}
		return Tariff{}, fmt.Errorf("insert tariff: %w", err)
	}

	if err := insertTiers(ctx, tx, out.ID, t.Tiers); err != nil {
		return Tariff{}, err
	}
	out.Tiers = t.Tiers

	if err := tx.Commit(ctx); err != nil {
		return Tariff{}, fmt.Errorf("commit tx: %w", err)
	}

	return out, nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (Tariff, error) {
	const q = `SELECT ` + tariffColumns + ` FROM tariffs WHERE id = $1`

	var out Tariff
	if err := scanTariff(r.db.QueryRow(ctx, q, id), &out); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Tariff{}, ErrNotFound
		}
		return Tariff{}, fmt.Errorf("get tariff by id: %w", err)
	}

	if err := r.attachTiers(ctx, []*Tariff{&out}); err != nil {
		return Tariff{}, err
	}
	return out, nil
}

func (r *Repository) List(ctx context.Context) ([]Tariff, error) {
	const q = `
		SELECT ` + tariffColumns + `
		FROM tariffs
		ORDER BY service, zone, valid_from DESC`

	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list tariffs: %w", err)
	}
	defer rows.Close()

	tariffs := make([]Tariff, 0)
	for rows.Next() {
		var t Tariff
		if err := scanTariff(rows, &t); err != nil {
			return nil, fmt.Errorf("scan tariffs: %w", err)
		}
		tariffs = append(tariffs, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tariffs: %w", err)
	}

	ptrs := make([]*Tariff, len(tariffs))
	for i := range tariffs {
		ptrs[i] = &tariffs[i]
	}
	if err := r.attachTiers(ctx, ptrs); err != nil {
		return nil, err
	}
	return tariffs, nil
}

func (r *Repository) Update(ctx context.Context, id uuid.UUID, validFrom time.Time, currency, tierMode string, tiers []Tier) (Tariff, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Tariff{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	const q = `
		UPDATE tariffs
		SET valid_from = $1, currency = $2, tier_mode = $3
		WHERE id = $4
		RETURNING ` + tariffColumns

	var out Tariff
	err = scanTariff(tx.QueryRow(ctx, q, validFrom, currency, tierMode, id), &out)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Tariff{}, ErrNotFound
		}
		if mapped := tariffWriteError(err); mapped != nil {
			return Tariff{}, mapped
		}
		return Tariff{}, fmt.Errorf("update tariff: %w", err)
	}

	if _, err := tx.Exec(ctx, `DELETE FROM tariff_tiers WHERE tariff_id = $1`, id); err != nil {
		return Tariff{}, fmt.Errorf("delete tariff tiers: %w", err)
	}
	if err := insertTiers(ctx, tx, id, tiers); err != nil {
		return Tariff{}, err
	}
	out.Tiers = tiers

	if err := tx.Commit(ctx); err != nil {
		return Tariff{}, fmt.Errorf("commit tx: %w", err)
	}

	return out, nil
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM tariffs WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete tariff: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *Repository) ActiveAt(ctx context.Context, service, zone string, day time.Time) (Tariff, error) {
	const q = `
		SELECT ` + tariffColumns + `
		FROM tariffs
		WHERE service = $1 AND zone = $2 AND valid_from <= $3
		ORDER BY valid_from DESC
		LIMIT 1`

	var out Tariff
	if err := scanTariff(r.db.QueryRow(ctx, q, service, zone, day), &out); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Tariff{}, ErrNotFound
		}
		return Tariff{}, fmt.Errorf("active tariff: %w", err)
	}

	if err := r.attachTiers(ctx, []*Tariff{&out}); err != nil {
		return Tariff{}, err
	}
	return out, nil
}

func (r *Repository) attachTiers(ctx context.Context, tariffs []*Tariff) error {
	if len(tariffs) == 0 {
		return nil
	}
	ids := make([]uuid.UUID, len(tariffs))
	for i, t := range tariffs {
		ids[i] = t.ID
	}

	const q = `
		SELECT tariff_id, up_to, price
		FROM tariff_tiers
		WHERE tariff_id = ANY($1)
		ORDER BY tariff_id, position`

	rows, err := r.db.Query(ctx, q, ids)
	if err != nil {
		return fmt.Errorf("list tariff tiers: %w", err)
	}
	defer rows.Close()

	byTariff := make(map[uuid.UUID][]Tier, len(tariffs))
	for rows.Next() {
		var tariffID uuid.UUID
		var tier Tier
		if err := rows.Scan(&tariffID, &tier.UpTo, &tier.Price); err != nil {
			return fmt.Errorf("scan tariff tiers: %w", err)
		}
		byTariff[tariffID] = append(byTariff[tariffID], tier)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate tariff tiers: %w", err)
	}

	for _, t := range tariffs {
		t.Tiers = byTariff[t.ID]
		if t.Tiers == nil {
			t.Tiers = []Tier{}
		}
	}
	return nil
}

func insertTiers(ctx context.Context, tx pgx.Tx, tariffID uuid.UUID, tiers []Tier) error {
	const q = `
		INSERT INTO tariff_tiers (tariff_id, position, up_to, price)
		VALUES ($1, $2, $3, $4)`

	for i, tier := range tiers {
		if _, err := tx.Exec(ctx, q, tariffID, i+1, tier.UpTo, tier.Price); err != nil {
			if mapped := tariffWriteError(err); mapped != nil {
				return mapped
			}
			return fmt.Errorf("insert tariff tier: %w", err)
		}
	}
	return nil
}

func tariffWriteError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return nil
	}
	switch pgErr.ConstraintName {
	case "uq_tariffs_service_zone_from":
		return ErrAlreadyExists
	case "tariffs_service_fkey":
		return fmt.Errorf("%w: unknown service", ErrInvalidInput)
	case "tariffs_currency_fkey":
		return fmt.Errorf("%w: unknown currency", ErrInvalidInput)
	case "tariffs_zone_check":
		return fmt.Errorf("%w: unknown zone", ErrInvalidInput)
	case "tariffs_tier_mode_check":
		return fmt.Errorf("%w: unknown tier mode", ErrInvalidInput)
	case "tariff_tiers_price_check":
		return fmt.Errorf("%w: price must be positive", ErrInvalidInput)
	case "tariff_tiers_up_to_check":
		return fmt.Errorf("%w: a tier bound must be positive", ErrInvalidInput)
	case "uq_tariff_tiers_open":
		return fmt.Errorf("%w: only the last tier can be open-ended", ErrInvalidInput)
	}
	return nil
}

func scanTariff(row pgx.Row, t *Tariff) error {
	return row.Scan(&t.ID, &t.Service, &t.Zone, &t.ValidFrom, &t.Currency, &t.TierMode, &t.CreatedAt, &t.UpdatedAt)
}
