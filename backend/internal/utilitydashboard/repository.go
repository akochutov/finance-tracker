package utilitydashboard

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type ServiceRow struct {
	Code string
	Name string
	Unit string
}

type AccountRow struct {
	ID                uuid.UUID
	Service           string
	Number            string
	ExpenseCategoryID *uuid.UUID
}

type MeterRow struct {
	ID        uuid.UUID
	AccountID uuid.UUID
	Serial    string
	Registers string
}

type ReadingRow struct {
	MeterID uuid.UUID
	Zone    string
	TakenOn time.Time
	Value   decimal.Decimal
}

type PaymentRow struct {
	CategoryID uuid.UUID
	Month      time.Time
	Amount     decimal.Decimal
	Currency   string
}

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Services(ctx context.Context) ([]ServiceRow, error) {
	const q = `SELECT code, name, unit FROM services ORDER BY name`
	return collect(ctx, r.db, q, nil, func(row pgx.Rows) (ServiceRow, error) {
		var s ServiceRow
		err := row.Scan(&s.Code, &s.Name, &s.Unit)
		return s, err
	})
}

func (r *Repository) Accounts(ctx context.Context) ([]AccountRow, error) {
	const q = `
		SELECT id, service, number, expense_category_id
		FROM accounts
		ORDER BY service, number`
	return collect(ctx, r.db, q, nil, func(row pgx.Rows) (AccountRow, error) {
		var a AccountRow
		err := row.Scan(&a.ID, &a.Service, &a.Number, &a.ExpenseCategoryID)
		return a, err
	})
}

func (r *Repository) Meters(ctx context.Context) ([]MeterRow, error) {
	const q = `
		SELECT id, account_id, serial, registers
		FROM meters
		ORDER BY account_id, installed_on, serial`
	return collect(ctx, r.db, q, nil, func(row pgx.Rows) (MeterRow, error) {
		var m MeterRow
		err := row.Scan(&m.ID, &m.AccountID, &m.Serial, &m.Registers)
		return m, err
	})
}

func (r *Repository) Readings(ctx context.Context) ([]ReadingRow, error) {
	const q = `
		SELECT meter_id, zone, taken_on, value
		FROM readings
		ORDER BY meter_id, zone, taken_on`
	return collect(ctx, r.db, q, nil, func(row pgx.Rows) (ReadingRow, error) {
		var rd ReadingRow
		err := row.Scan(&rd.MeterID, &rd.Zone, &rd.TakenOn, &rd.Value)
		return rd, err
	})
}

func (r *Repository) Payments(ctx context.Context, categories []uuid.UUID, from, to time.Time) ([]PaymentRow, error) {
	if len(categories) == 0 {
		return []PaymentRow{}, nil
	}
	const q = `
		SELECT i.category_id,
		       date_trunc('month', i.period_from)::date,
		       i.amount,
		       e.currency
		FROM expense_items i
		JOIN expenses e ON e.id = i.expense_id
		WHERE i.category_id = ANY($1)
		  AND i.period_from IS NOT NULL
		  AND i.period_from >= $2 AND i.period_from < $3`
	return collect(ctx, r.db, q, []any{categories, from, to}, func(row pgx.Rows) (PaymentRow, error) {
		var p PaymentRow
		err := row.Scan(&p.CategoryID, &p.Month, &p.Amount, &p.Currency)
		return p, err
	})
}

func collect[T any](ctx context.Context, db *pgxpool.Pool, q string, args []any, scan func(pgx.Rows) (T, error)) ([]T, error) {
	rows, err := db.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("utility dashboard query: %w", err)
	}
	defer rows.Close()

	out := make([]T, 0)
	for rows.Next() {
		item, err := scan(rows)
		if err != nil {
			return nil, fmt.Errorf("utility dashboard scan: %w", err)
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("utility dashboard iterate: %w", err)
	}
	return out, nil
}
