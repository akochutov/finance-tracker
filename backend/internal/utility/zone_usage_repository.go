package utility

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

const zoneUsageColumns = "account_id, month, zone, quantity, updated_at"

type ZoneUsageRepository struct {
	db *pgxpool.Pool
}

func NewZoneUsageRepository(db *pgxpool.Pool) *ZoneUsageRepository {
	return &ZoneUsageRepository{db: db}
}

func (r *ZoneUsageRepository) SetMonth(ctx context.Context, usage []ZoneUsage) ([]ZoneUsage, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	const q = `
		INSERT INTO zone_usage (account_id, month, zone, quantity)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (account_id, month, zone) DO UPDATE
			SET quantity = EXCLUDED.quantity, updated_at = now()
		RETURNING ` + zoneUsageColumns

	out := make([]ZoneUsage, 0, len(usage))
	for _, u := range usage {
		var saved ZoneUsage
		err := scanZoneUsage(tx.QueryRow(ctx, q, u.AccountID, u.Month, u.Zone, u.Quantity), &saved)
		if err != nil {
			if mapped := zoneUsageWriteError(err); mapped != nil {
				return nil, mapped
			}
			return nil, fmt.Errorf("upsert zone usage: %w", err)
		}
		out = append(out, saved)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return out, nil
}

func (r *ZoneUsageRepository) ListByAccount(ctx context.Context, accountID uuid.UUID) ([]ZoneUsage, error) {
	const q = `
		SELECT ` + zoneUsageColumns + `
		FROM zone_usage
		WHERE account_id = $1
		ORDER BY month DESC, zone`

	rows, err := r.db.Query(ctx, q, accountID)
	if err != nil {
		return nil, fmt.Errorf("list zone usage: %w", err)
	}
	defer rows.Close()

	usage := make([]ZoneUsage, 0)
	for rows.Next() {
		var u ZoneUsage
		if err := scanZoneUsage(rows, &u); err != nil {
			return nil, fmt.Errorf("scan zone usage: %w", err)
		}
		usage = append(usage, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate zone usage: %w", err)
	}

	return usage, nil
}

func (r *ZoneUsageRepository) DeleteMonth(ctx context.Context, accountID uuid.UUID, month time.Time) error {
	const q = `DELETE FROM zone_usage WHERE account_id = $1 AND month = $2`

	if _, err := r.db.Exec(ctx, q, accountID, month); err != nil {
		return fmt.Errorf("delete zone usage: %w", err)
	}

	return nil
}

func zoneUsageWriteError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return nil
	}
	switch pgErr.ConstraintName {
	case "zone_usage_account_id_fkey":
		return ErrAccountNotFound
	case "zone_usage_zone_check":
		return fmt.Errorf("%w: unknown zone", ErrInvalidInput)
	case "zone_usage_quantity_check":
		return fmt.Errorf("%w: usage must not be negative", ErrInvalidInput)
	case "zone_usage_month_check":
		return fmt.Errorf("%w: month must be the first day of the month", ErrInvalidInput)
	}
	return nil
}

func scanZoneUsage(row pgx.Row, u *ZoneUsage) error {
	return row.Scan(&u.AccountID, &u.Month, &u.Zone, &u.Quantity, &u.UpdatedAt)
}
