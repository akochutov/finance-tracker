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

const (
	meterColumns    = "id, account_id, serial, installed_on, removed_on, created_at, updated_at"
	registerColumns = "id, meter_id, zone"
)

type MeterRepository struct {
	db *pgxpool.Pool
}

func NewMeterRepository(db *pgxpool.Pool) *MeterRepository {
	return &MeterRepository{db: db}
}

func (r *MeterRepository) Create(ctx context.Context, meter Meter) (Meter, error) {
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

	registers, err := insertRegisters(ctx, tx, out.ID, meter.Registers)
	if err != nil {
		return Meter{}, err
	}
	out.Registers = registers

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

	byMeter, err := r.registersByMeter(ctx, []uuid.UUID{id})
	if err != nil {
		return Meter{}, err
	}
	out.Registers = registersOrEmpty(byMeter[id])

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
	ids := make([]uuid.UUID, 0)
	for rows.Next() {
		var m Meter
		if err := scanMeter(rows, &m); err != nil {
			return nil, fmt.Errorf("scan meters: %w", err)
		}
		meters = append(meters, m)
		ids = append(ids, m.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate meters: %w", err)
	}

	if len(meters) == 0 {
		return meters, nil
	}

	byMeter, err := r.registersByMeter(ctx, ids)
	if err != nil {
		return nil, err
	}
	for i := range meters {
		meters[i].Registers = registersOrEmpty(byMeter[meters[i].ID])
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

	byMeter, err := r.registersByMeter(ctx, []uuid.UUID{id})
	if err != nil {
		return Meter{}, err
	}
	out.Registers = registersOrEmpty(byMeter[id])

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

func (r *MeterRepository) registersByMeter(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID][]Register, error) {
	const q = `
		SELECT ` + registerColumns + `
		FROM meter_registers
		WHERE meter_id = ANY($1)
		ORDER BY meter_id, zone`

	rows, err := r.db.Query(ctx, q, ids)
	if err != nil {
		return nil, fmt.Errorf("list meter registers: %w", err)
	}
	defer rows.Close()

	byMeter := make(map[uuid.UUID][]Register, len(ids))
	for rows.Next() {
		var reg Register
		if err := scanRegister(rows, &reg); err != nil {
			return nil, fmt.Errorf("scan meter registers: %w", err)
		}
		byMeter[reg.MeterID] = append(byMeter[reg.MeterID], reg)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate meter registers: %w", err)
	}

	return byMeter, nil
}

func insertRegisters(ctx context.Context, tx pgx.Tx, meterID uuid.UUID, registers []Register) ([]Register, error) {
	const q = `
		INSERT INTO meter_registers (id, meter_id, zone)
		VALUES ($1, $2, $3)
		RETURNING ` + registerColumns

	out := make([]Register, 0, len(registers))
	for _, reg := range registers {
		var saved Register
		if err := scanRegister(tx.QueryRow(ctx, q, reg.ID, meterID, reg.Zone), &saved); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				switch pgErr.Code {
				case uniqueViolation:
					return nil, fmt.Errorf("%w: duplicate zone %q", ErrInvalidInput, reg.Zone)
				case checkViolation:
					return nil, fmt.Errorf("%w: unknown zone %q", ErrInvalidInput, reg.Zone)
				}
			}
			return nil, fmt.Errorf("insert meter register: %w", err)
		}
		out = append(out, saved)
	}

	return out, nil
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

func scanRegister(row pgx.Row, reg *Register) error {
	return row.Scan(&reg.ID, &reg.MeterID, &reg.Zone)
}

func registersOrEmpty(registers []Register) []Register {
	if registers == nil {
		return []Register{}
	}
	return registers
}
