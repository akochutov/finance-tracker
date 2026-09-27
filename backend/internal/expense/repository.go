package expense

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	expenseColumns = `id, occurred_on, currency, payment_type, note, created_at, updated_at`
	itemColumns    = `id, expense_id, line_no, description, category_id, price, quantity, discount, amount, period_from, period_to`
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, e Expense) (Expense, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Expense{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	const q = `
		INSERT INTO expenses (id, occurred_on, currency, payment_type, note)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING ` + expenseColumns

	var out Expense
	err = scanExpense(tx.QueryRow(ctx, q, e.ID, e.OccurredOn, e.Currency, e.PaymentType, e.Note), &out)
	if err != nil {
		if msg, ok := inputError(err); ok {
			return Expense{}, fmt.Errorf("%w: %s", ErrInvalidInput, msg)
		}
		return Expense{}, fmt.Errorf("insert expense: %w", err)
	}

	items, err := insertItems(ctx, tx, out.ID, e.Items)
	if err != nil {
		return Expense{}, err
	}
	out.Items = items

	if err := tx.Commit(ctx); err != nil {
		return Expense{}, fmt.Errorf("commit tx: %w", err)
	}

	return out, nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (Expense, error) {
	const q = `SELECT ` + expenseColumns + ` FROM expenses WHERE id = $1`

	var out Expense
	if err := scanExpense(r.db.QueryRow(ctx, q, id), &out); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Expense{}, ErrNotFound
		}
		return Expense{}, fmt.Errorf("get expense by id: %w", err)
	}

	byExpense, err := r.itemsByExpense(ctx, []uuid.UUID{id})
	if err != nil {
		return Expense{}, err
	}
	out.Items = itemsOrEmpty(byExpense[id])

	return out, nil
}

func (r *Repository) List(ctx context.Context) ([]Expense, error) {
	const q = `
		SELECT ` + expenseColumns + `
		FROM expenses
		ORDER BY occurred_on DESC, created_at DESC`

	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list expenses: %w", err)
	}
	defer rows.Close()

	expenses := make([]Expense, 0)
	ids := make([]uuid.UUID, 0)
	for rows.Next() {
		var e Expense
		if err := scanExpense(rows, &e); err != nil {
			return nil, fmt.Errorf("scan expenses: %w", err)
		}
		expenses = append(expenses, e)
		ids = append(ids, e.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate expenses: %w", err)
	}

	if len(expenses) == 0 {
		return expenses, nil
	}

	byExpense, err := r.itemsByExpense(ctx, ids)
	if err != nil {
		return nil, err
	}
	for i := range expenses {
		expenses[i].Items = itemsOrEmpty(byExpense[expenses[i].ID])
	}

	return expenses, nil
}

func (r *Repository) Update(ctx context.Context, e Expense) (Expense, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Expense{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	const q = `
		UPDATE expenses
		SET occurred_on = $1, currency = $2, payment_type = $3, note = $4
		WHERE id = $5
		RETURNING ` + expenseColumns

	var out Expense
	err = scanExpense(tx.QueryRow(ctx, q, e.OccurredOn, e.Currency, e.PaymentType, e.Note, e.ID), &out)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Expense{}, ErrNotFound
		}
		if msg, ok := inputError(err); ok {
			return Expense{}, fmt.Errorf("%w: %s", ErrInvalidInput, msg)
		}
		return Expense{}, fmt.Errorf("update expense: %w", err)
	}

	if _, err := tx.Exec(ctx, `DELETE FROM expense_items WHERE expense_id = $1`, out.ID); err != nil {
		return Expense{}, fmt.Errorf("delete expense items: %w", err)
	}

	items, err := insertItems(ctx, tx, out.ID, e.Items)
	if err != nil {
		return Expense{}, err
	}
	out.Items = items

	if err := tx.Commit(ctx); err != nil {
		return Expense{}, fmt.Errorf("commit tx: %w", err)
	}

	return out, nil
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM expenses WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete expense: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func insertItems(ctx context.Context, tx pgx.Tx, expenseID uuid.UUID, items []Item) ([]Item, error) {
	const q = `
		INSERT INTO expense_items
			(id, expense_id, line_no, description, category_id,
			 price, quantity, discount, period_from, period_to)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING ` + itemColumns

	out := make([]Item, 0, len(items))
	for _, it := range items {
		var saved Item
		err := scanItem(tx.QueryRow(ctx, q,
			it.ID, expenseID, it.LineNo, it.Description, it.CategoryID,
			it.Price, it.Quantity, it.Discount, it.PeriodFrom, it.PeriodTo,
		), &saved)
		if err != nil {
			if msg, ok := inputError(err); ok {
				return nil, fmt.Errorf("%w: line %d: %s", ErrInvalidInput, it.LineNo, msg)
			}
			return nil, fmt.Errorf("insert expense item line %d: %w", it.LineNo, err)
		}
		out = append(out, saved)
	}

	return out, nil
}

func (r *Repository) itemsByExpense(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID][]Item, error) {
	const q = `
		SELECT ` + itemColumns + `
		FROM expense_items
		WHERE expense_id = ANY($1)
		ORDER BY expense_id, line_no`

	rows, err := r.db.Query(ctx, q, ids)
	if err != nil {
		return nil, fmt.Errorf("list expense items: %w", err)
	}
	defer rows.Close()

	byExpense := make(map[uuid.UUID][]Item, len(ids))
	for rows.Next() {
		var it Item
		if err := scanItem(rows, &it); err != nil {
			return nil, fmt.Errorf("scan expense items: %w", err)
		}
		byExpense[it.ExpenseID] = append(byExpense[it.ExpenseID], it)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate expense items: %w", err)
	}

	return byExpense, nil
}

func scanExpense(row pgx.Row, e *Expense) error {
	return row.Scan(&e.ID, &e.OccurredOn, &e.Currency, &e.PaymentType, &e.Note, &e.CreatedAt, &e.UpdatedAt)
}

func scanItem(row pgx.Row, it *Item) error {
	return row.Scan(
		&it.ID, &it.ExpenseID, &it.LineNo, &it.Description, &it.CategoryID,
		&it.Price, &it.Quantity, &it.Discount, &it.Amount, &it.PeriodFrom, &it.PeriodTo,
	)
}

func itemsOrEmpty(items []Item) []Item {
	if items == nil {
		return []Item{}
	}
	return items
}

func inputError(err error) (string, bool) {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return "", false
	}

	switch pgErr.Code {
	case foreignKeyViolation:
		switch pgErr.ConstraintName {
		case "expenses_currency_fkey":
			return "unknown currency", true
		case "expense_items_category_id_fkey":
			return "unknown category", true
		}
	case uniqueViolation:
		if pgErr.ConstraintName == "uq_expense_items_line" {
			return "duplicate line number", true
		}
	case checkViolation:
		return "violates " + pgErr.ConstraintName, true
	}

	return "", false
}
