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
	"github.com/shopspring/decimal"
)

const readingColumns = "id, meter_id, taken_on, value, is_initial, created_at, updated_at"

type ReadingRepository struct {
	db *pgxpool.Pool
}

func NewReadingRepository(db *pgxpool.Pool) *ReadingRepository {
	return &ReadingRepository{db: db}
}

func (r *ReadingRepository) CreateBatch(ctx context.Context, readings []Reading) ([]Reading, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	out := make([]Reading, 0, len(readings))
	for _, rd := range readings {
		saved, err := insertReading(ctx, tx, rd)
		if err != nil {
			return nil, err
		}
		out = append(out, saved)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return out, nil
}

func (r *ReadingRepository) GetByID(ctx context.Context, id uuid.UUID) (Reading, error) {
	const q = `SELECT ` + readingColumns + ` FROM readings WHERE id = $1`

	var out Reading
	if err := scanReading(r.db.QueryRow(ctx, q, id), &out); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Reading{}, ErrReadingNotFound
		}
		return Reading{}, fmt.Errorf("get reading by id: %w", err)
	}

	return out, nil
}

func (r *ReadingRepository) ListByMeter(ctx context.Context, meterID uuid.UUID) ([]Reading, error) {
	const q = `
		SELECT ` + readingColumns + `
		FROM readings
		WHERE meter_id = $1
		ORDER BY taken_on DESC`

	return r.list(ctx, q, meterID)
}

func (r *ReadingRepository) LatestPerMeter(ctx context.Context) ([]Reading, error) {
	const q = `
		SELECT DISTINCT ON (meter_id) ` + readingColumns + `
		FROM readings
		ORDER BY meter_id, taken_on DESC`

	return r.list(ctx, q)
}

func (r *ReadingRepository) Neighbors(ctx context.Context, meterID uuid.UUID, day time.Time, excludeID uuid.UUID) (*Reading, *Reading, error) {
	const prevQ = `
		SELECT ` + readingColumns + `
		FROM readings
		WHERE meter_id = $1 AND taken_on < $2 AND id <> $3
		ORDER BY taken_on DESC
		LIMIT 1`
	const nextQ = `
		SELECT ` + readingColumns + `
		FROM readings
		WHERE meter_id = $1 AND taken_on > $2 AND id <> $3
		ORDER BY taken_on
		LIMIT 1`

	prev, err := r.optionalReading(ctx, prevQ, meterID, day, excludeID)
	if err != nil {
		return nil, nil, fmt.Errorf("previous reading: %w", err)
	}
	next, err := r.optionalReading(ctx, nextQ, meterID, day, excludeID)
	if err != nil {
		return nil, nil, fmt.Errorf("next reading: %w", err)
	}

	return prev, next, nil
}

func (r *ReadingRepository) Update(ctx context.Context, id uuid.UUID, takenOn time.Time, value decimal.Decimal) (Reading, error) {
	const q = `
		UPDATE readings
		SET taken_on = $1, value = $2
		WHERE id = $3
		RETURNING ` + readingColumns

	var out Reading
	err := scanReading(r.db.QueryRow(ctx, q, takenOn, value, id), &out)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Reading{}, ErrReadingNotFound
		}
		if mapped := readingWriteError(err); mapped != nil {
			return Reading{}, mapped
		}
		return Reading{}, fmt.Errorf("update reading: %w", err)
	}

	return out, nil
}

func (r *ReadingRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM readings WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete reading: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrReadingNotFound
	}

	return nil
}

func insertReading(ctx context.Context, tx pgx.Tx, rd Reading) (Reading, error) {
	const q = `
		INSERT INTO readings (id, meter_id, taken_on, value, is_initial)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING ` + readingColumns

	var out Reading
	err := scanReading(tx.QueryRow(ctx, q, rd.ID, rd.MeterID, rd.TakenOn, rd.Value, rd.IsInitial), &out)
	if err != nil {
		if mapped := readingWriteError(err); mapped != nil {
			return Reading{}, mapped
		}
		return Reading{}, fmt.Errorf("insert reading: %w", err)
	}

	return out, nil
}

func (r *ReadingRepository) list(ctx context.Context, q string, args ...any) ([]Reading, error) {
	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list readings: %w", err)
	}
	defer rows.Close()

	readings := make([]Reading, 0)
	for rows.Next() {
		var rd Reading
		if err := scanReading(rows, &rd); err != nil {
			return nil, fmt.Errorf("scan readings: %w", err)
		}
		readings = append(readings, rd)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate readings: %w", err)
	}

	return readings, nil
}

func (r *ReadingRepository) optionalReading(ctx context.Context, q string, args ...any) (*Reading, error) {
	var rd Reading
	if err := scanReading(r.db.QueryRow(ctx, q, args...), &rd); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &rd, nil
}

func readingWriteError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return nil
	}
	switch pgErr.ConstraintName {
	case "uq_readings_meter_day":
		return ErrReadingExists
	case "uq_readings_initial":
		return fmt.Errorf("%w: the meter already has an initial reading", ErrInvalidInput)
	case "readings_meter_id_fkey":
		return ErrMeterNotFound
	case "readings_value_check":
		return fmt.Errorf("%w: reading must not be negative", ErrInvalidInput)
	}
	return nil
}

func scanReading(row pgx.Row, rd *Reading) error {
	return row.Scan(&rd.ID, &rd.MeterID, &rd.TakenOn, &rd.Value, &rd.IsInitial, &rd.CreatedAt, &rd.UpdatedAt)
}
