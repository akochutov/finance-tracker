package settings

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Get(ctx context.Context) (Settings, error) {
	const q = `SELECT base_currency, updated_at FROM settings;`

	var out Settings
	err := r.db.QueryRow(ctx, q).Scan(&out.BaseCurrency, &out.UpdatedAt)
	if err != nil {
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
		RETURNING base_currency, updated_at;`

	var out Settings
	err := r.db.QueryRow(ctx, q, baseCurrency).Scan(&out.BaseCurrency, &out.UpdatedAt)
	if err != nil {
		return Settings{}, fmt.Errorf("upsert settings: %w", err)
	}

	return out, nil
}
