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

const meterColumns = "id, account_id, serial, installed_on, removed_on, created_at, updated_at"

type MeterRepository struct {
	db *pgxpool.Pool
}

func NewMeterRepository(db *pgxpool.Pool) *MeterRepository {
	return &MeterRepository{db: db}
}

func (r *MeterRepository) Create(ctx context.Context, meter Meter, initial Reading) (Meter, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Meter{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	const q = `
		INSERT INTO meters (id, account_id, serial, installed_on, removed_on)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING ` + meterColumns

	var out Meter
	err = scanMeter(tx.QueryRow(ctx, q,
		meter.ID, meter.AccountID, meter.Serial, meter.InstalledOn, meter.RemovedOn,
	), &out)
	if err != nil {
		if mapped := meterWriteError(err); mapped != nil {
			return Meter{}, mapped
		}
		return Meter{}, fmt.Errorf("insert meter: %w", err)
	}

	initial.MeterID = out.ID
	if _, err := insertReading(ctx, tx, initial); err != nil {
		return Meter{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return Meter{}, fmt.Errorf("commit tx: %w", err)
	}

	return out, nil
}

func (r *MeterRepository) GetByID(ctx context.Context, id uuid.UUID) (Meter, error) {
	const q = `SELECT ` + meterColumns + ` FROM meters WHERE id = $1`

	var out Meter
	if err := scanMeter(r.db.QueryRow(ctx, q, id), &out); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Meter{}, ErrMeterNotFound
		}
		return Meter{}, fmt.Errorf("get meter by id: %w", err)
	}

	return out, nil
}

func (r *MeterRepository) List(ctx context.Context) ([]Meter, error) {
	const q = `SELECT ` + meterColumns + ` FROM meters ORDER BY account_id, installed_on`

	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list meters: %w", err)
	}
	defer rows.Close()

	meters := make([]Meter, 0)
	for rows.Next() {
		var m Meter
		if err := scanMeter(rows, &m); err != nil {
			return nil, fmt.Errorf("scan meters: %w", err)
		}
		meters = append(meters, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate meters: %w", err)
	}

	return meters, nil
}

func (r *MeterRepository) Update(ctx context.Context, id uuid.UUID, serial string, installedOn time.Time, removedOn *time.Time) (Meter, error) {
	const q = `
		UPDATE meters
		SET serial = $1, installed_on = $2, removed_on = $3
		WHERE id = $4
		RETURNING ` + meterColumns

	var out Meter
	err := scanMeter(r.db.QueryRow(ctx, q, serial, installedOn, removedOn, id), &out)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Meter{}, ErrMeterNotFound
		}
		if mapped := meterWriteError(err); mapped != nil {
			return Meter{}, mapped
		}
		return Meter{}, fmt.Errorf("update meter: %w", err)
	}

	return out, nil
}

func (r *MeterRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM meters WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete meter: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrMeterNotFound
	}

	return nil
}

func meterWriteError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return nil
	}
	switch pgErr.ConstraintName {
	case "uq_meters_account_serial":
		return ErrSerialTaken
	case "meters_account_id_fkey":
		return ErrAccountNotFound
	case "chk_meters_dates":
		return fmt.Errorf("%w: removed before installed", ErrInvalidInput)
	}
	return nil
}

func scanMeter(row pgx.Row, m *Meter) error {
	return row.Scan(&m.ID, &m.AccountID, &m.Serial, &m.InstalledOn, &m.RemovedOn, &m.CreatedAt, &m.UpdatedAt)
}
