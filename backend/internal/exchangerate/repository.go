package exchangerate

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Upsert(ctx context.Context, rate Rate) (Rate, error) {
	const q = `
		INSERT INTO exchange_rates (currency, source, rate_at, rate)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (currency, source, rate_at) DO UPDATE
			SET rate = EXCLUDED.rate, fetched_at = now()
		RETURNING currency, source, rate_at, rate, fetched_at;`

	var out Rate
	err := r.db.QueryRow(ctx, q, rate.Currency, rate.Source, rate.RateAt, rate.Rate).
		Scan(&out.Currency, &out.Source, &out.RateAt, &out.Rate, &out.FetchedAt)
	if err != nil {
		return Rate{}, fmt.Errorf("upsert exchange rate: %w", err)
	}

	return out, nil
}

func (r *Repository) LatestAsOf(ctx context.Context, currency string, at time.Time) (Rate, error) {
	const q = `
		SELECT currency, source, rate_at, rate, fetched_at
		  FROM exchange_rates
		 WHERE currency = $1
		   AND rate_at <= $2
		 ORDER BY rate_at DESC
		 LIMIT 1;`

	var out Rate
	err := r.db.QueryRow(ctx, q, currency, at).
		Scan(&out.Currency, &out.Source, &out.RateAt, &out.Rate, &out.FetchedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Rate{}, ErrRateNotFound
		}
		return Rate{}, fmt.Errorf("latest rate as of: %w", err)
	}

	return out, nil
}

func (r *Repository) List(ctx context.Context, currency string) ([]Rate, error) {
	const q = `
		SELECT currency, source, rate_at, rate, fetched_at
		  FROM exchange_rates
		 WHERE ($1 = '' OR currency = $1)
		 ORDER BY currency, rate_at DESC;`

	rows, err := r.db.Query(ctx, q, currency)
	if err != nil {
		return nil, fmt.Errorf("list exchange rates: %w", err)
	}
	defer rows.Close()

	rates := make([]Rate, 0)
	for rows.Next() {
		var rate Rate
		err := rows.Scan(&rate.Currency, &rate.Source, &rate.RateAt, &rate.Rate, &rate.FetchedAt)
		if err != nil {
			return nil, fmt.Errorf("scan exchange rate: %w", err)
		}
		rates = append(rates, rate)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate exchange rates: %w", err)
	}

	return rates, nil
}
