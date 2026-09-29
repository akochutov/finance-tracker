package expensedashboard

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

// AmountRow is the sum of expense lines for one day, currency, category
// and fixed/variable flag. Amount is in the expense's own currency.
type AmountRow struct {
	Day          time.Time
	Currency     string
	GroupID      uuid.UUID
	GroupName    string
	CategoryID   uuid.UUID
	CategoryName string
	Fixed        bool
	Amount       decimal.Decimal
}

// ReceiptRow is one expense with the total of its dashboard lines.
type ReceiptRow struct {
	ExpenseID    uuid.UUID
	Day          time.Time
	Currency     string
	Note         *string
	Lines        int
	Total        decimal.Decimal
	Group        string
	Descriptions string
}

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Amounts(ctx context.Context, from, to time.Time) ([]AmountRow, error) {
	const q = `
		SELECT e.occurred_on, e.currency,
		       g.id, g.name, c.id, c.name,
		       (i.period_from IS NOT NULL) AS fixed,
		       sum(i.amount)
		FROM expenses e
		JOIN expense_items i      ON i.expense_id = e.id
		JOIN expense_categories c ON c.id = i.category_id
		JOIN expense_groups g     ON g.id = c.group_id
		WHERE e.occurred_on BETWEEN $1 AND $2
		  AND c.include_in_dashboard
		GROUP BY e.occurred_on, e.currency, g.id, g.name, c.id, c.name, fixed`

	rows, err := r.db.Query(ctx, q, from, to)
	if err != nil {
		return nil, fmt.Errorf("dashboard amounts: %w", err)
	}
	defer rows.Close()

	out := make([]AmountRow, 0)
	for rows.Next() {
		var a AmountRow
		err := rows.Scan(&a.Day, &a.Currency, &a.GroupID, &a.GroupName,
			&a.CategoryID, &a.CategoryName, &a.Fixed, &a.Amount)
		if err != nil {
			return nil, fmt.Errorf("scan dashboard amounts: %w", err)
		}
		out = append(out, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate dashboard amounts: %w", err)
	}

	return out, nil
}

func (r *Repository) Receipts(ctx context.Context, from, to time.Time) ([]ReceiptRow, error) {
	const q = `
		WITH lines AS (
			SELECT e.id, e.occurred_on, e.currency, e.note,
			       i.line_no, i.description, i.amount, g.name AS group_name
			FROM expenses e
			JOIN expense_items i      ON i.expense_id = e.id
			JOIN expense_categories c ON c.id = i.category_id
			JOIN expense_groups g     ON g.id = c.group_id
			WHERE e.occurred_on BETWEEN $1 AND $2
			  AND c.include_in_dashboard
		),
		main_group AS (
			SELECT DISTINCT ON (id) id, group_name
			FROM (
				SELECT id, group_name, sum(amount) AS group_total
				FROM lines
				GROUP BY id, group_name
			) per_group
			ORDER BY id, group_total DESC, group_name
		)
		SELECT l.id, l.occurred_on, l.currency, l.note,
		       count(*),
		       sum(l.amount),
		       mg.group_name,
		       string_agg(l.description, ', ' ORDER BY l.line_no)
		FROM lines l
		JOIN main_group mg ON mg.id = l.id
		GROUP BY l.id, l.occurred_on, l.currency, l.note, mg.group_name`

	rows, err := r.db.Query(ctx, q, from, to)
	if err != nil {
		return nil, fmt.Errorf("dashboard receipts: %w", err)
	}
	defer rows.Close()

	out := make([]ReceiptRow, 0)
	for rows.Next() {
		var rc ReceiptRow
		err := rows.Scan(&rc.ExpenseID, &rc.Day, &rc.Currency, &rc.Note,
			&rc.Lines, &rc.Total, &rc.Group, &rc.Descriptions)
		if err != nil {
			return nil, fmt.Errorf("scan dashboard receipts: %w", err)
		}
		out = append(out, rc)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate dashboard receipts: %w", err)
	}

	return out, nil
}
