package currency

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("currency not found")

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context) ([]Currency, error) {
	const q = `
		SELECT code, name, kind, decimal_places, is_active, created_at, updated_at
		FROM currencies`

	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("query currencies: %w", err)
	}
	defer rows.Close()

	currencies := make([]Currency, 0)
	for rows.Next() {
		var c Currency
		err := rows.Scan(&c.Code, &c.Name, &c.Kind, &c.DecimalPlaces, &c.IsActive, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan currency: %w", err)
		}
		currencies = append(currencies, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate currencies: %w", err)
	}

	return currencies, nil
}

func (r *Repository) GetByCode(ctx context.Context, code string) (Currency, error) {
	const q = `
		SELECT code, name, kind, decimal_places, is_active, created_at, updated_at
		FROM currencies
		WHERE code = $1`

	var c Currency
	err := r.db.QueryRow(ctx, q, code).Scan(&c.Code, &c.Name, &c.Kind, &c.DecimalPlaces, &c.IsActive, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Currency{}, ErrNotFound
		}
		return Currency{}, fmt.Errorf("get currency: %w", err)
	}

	return c, nil
}
