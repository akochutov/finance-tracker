package settings

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const settingsColumns = `base_currency, expense_base_currency, updated_at`

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Get(ctx context.Context) (Settings, error) {
	const q = `SELECT ` + settingsColumns + ` FROM settings`

	var out Settings
	if err := scanSettings(r.db.QueryRow(ctx, q), &out); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Settings{}, ErrNotFound
		}
		return Settings{}, fmt.Errorf("get settings: %w", err)
	}

	return out, nil
}

func (r *Repository) Upsert(ctx context.Context, baseCurrency string) (Settings, error) {
	const q = `
		INSERT INTO settings (id, base_currency) VALUES (true, $1)
		ON CONFLICT (id) DO UPDATE
			SET base_currency = EXCLUDED.base_currency, updated_at = now()
		RETURNING ` + settingsColumns

	var out Settings
	if err := scanSettings(r.db.QueryRow(ctx, q, baseCurrency), &out); err != nil {
		return Settings{}, fmt.Errorf("upsert settings: %w", err)
	}

	return out, nil
}

func (r *Repository) UpsertExpenseBaseCurrency(ctx context.Context, baseCurrency, expenseCurrency string) (Settings, error) {
	const q = `
		INSERT INTO settings (id, base_currency, expense_base_currency) VALUES (true, $1, $2)
		ON CONFLICT (id) DO UPDATE
			SET expense_base_currency = EXCLUDED.expense_base_currency, updated_at = now()
		RETURNING ` + settingsColumns

	var out Settings
	if err := scanSettings(r.db.QueryRow(ctx, q, baseCurrency, expenseCurrency), &out); err != nil {
		return Settings{}, fmt.Errorf("upsert expense base currency: %w", err)
	}

	return out, nil
}

func scanSettings(row pgx.Row, s *Settings) error {
	var expense *string
	if err := row.Scan(&s.BaseCurrency, &expense, &s.UpdatedAt); err != nil {
		return err
	}
	if expense != nil {
		s.ExpenseBaseCurrency = *expense
	}
	return nil
}
