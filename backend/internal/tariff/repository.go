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
	"github.com/shopspring/decimal"
)

const tariffColumns = "id, service, zone, valid_from, price, currency, created_at, updated_at"

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, t Tariff) (Tariff, error) {
	const q = `
		INSERT INTO tariffs (id, service, zone, valid_from, price, currency)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING ` + tariffColumns

	var out Tariff
	err := scanTariff(r.db.QueryRow(ctx, q, t.ID, t.Service, t.Zone, t.ValidFrom, t.Price, t.Currency), &out)
	if err != nil {
		if mapped := tariffWriteError(err); mapped != nil {
			return Tariff{}, mapped
		}
		return Tariff{}, fmt.Errorf("insert tariff: %w", err)
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

	return tariffs, nil
}

func (r *Repository) Update(ctx context.Context, id uuid.UUID, validFrom time.Time, price decimal.Decimal, currency string) (Tariff, error) {
	const q = `
		UPDATE tariffs
		SET valid_from = $1, price = $2, currency = $3
		WHERE id = $4
		RETURNING ` + tariffColumns

	var out Tariff
	err := scanTariff(r.db.QueryRow(ctx, q, validFrom, price, currency, id), &out)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Tariff{}, ErrNotFound
		}
		if mapped := tariffWriteError(err); mapped != nil {
			return Tariff{}, mapped
		}
		return Tariff{}, fmt.Errorf("update tariff: %w", err)
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

	return out, nil
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
	case "tariffs_price_check":
		return fmt.Errorf("%w: price must be positive", ErrInvalidInput)
	}
	return nil
}

func scanTariff(row pgx.Row, t *Tariff) error {
	return row.Scan(&t.ID, &t.Service, &t.Zone, &t.ValidFrom, &t.Price, &t.Currency, &t.CreatedAt, &t.UpdatedAt)
}
