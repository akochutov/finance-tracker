package ratesource

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

func (r *Repository) Get(ctx context.Context, kind string) (RateSource, error) {
	const q = `
		SELECT kind, source, url_template, poll_interval_seconds,
		       request_timeout_seconds, backfill_start, updated_at
		  FROM rate_sources
		 WHERE kind = $1;`

	var out RateSource
	err := r.db.QueryRow(ctx, q, kind).Scan(
		&out.Kind, &out.Source, &out.URLTemplate, &out.PollInterval,
		&out.RequestTimeout, &out.BackfillStart, &out.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return RateSource{}, ErrNotFound
		}
		return RateSource{}, fmt.Errorf("get rate source %q: %w", kind, err)
	}

	return out, nil
}

func (r *Repository) List(ctx context.Context) ([]RateSource, error) {
	const q = `
		SELECT kind, source, url_template, poll_interval_seconds,
		       request_timeout_seconds, backfill_start, updated_at
		  FROM rate_sources
		 ORDER BY kind;`

	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list rate sources: %w", err)
	}
	defer rows.Close()

	sources := make([]RateSource, 0)
	for rows.Next() {
		var s RateSource
		if err := rows.Scan(
			&s.Kind, &s.Source, &s.URLTemplate, &s.PollInterval,
			&s.RequestTimeout, &s.BackfillStart, &s.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan rate source: %w", err)
		}
		sources = append(sources, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rate sources: %w", err)
	}

	return sources, nil
}

func (r *Repository) Upsert(ctx context.Context, s RateSource) (RateSource, error) {
	const q = `
		INSERT INTO rate_sources
			(kind, source, url_template, poll_interval_seconds,
			 request_timeout_seconds, backfill_start)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (kind) DO UPDATE SET
			source                  = EXCLUDED.source,
			url_template            = EXCLUDED.url_template,
			poll_interval_seconds   = EXCLUDED.poll_interval_seconds,
			request_timeout_seconds = EXCLUDED.request_timeout_seconds,
			backfill_start          = EXCLUDED.backfill_start,
			updated_at              = now()
		RETURNING kind, source, url_template, poll_interval_seconds,
		          request_timeout_seconds, backfill_start, updated_at;`

	var out RateSource
	err := r.db.QueryRow(ctx, q, s.Kind,
		s.Source, s.URLTemplate, s.PollInterval,
		s.RequestTimeout, s.BackfillStart,
	).Scan(
		&out.Kind, &out.Source, &out.URLTemplate, &out.PollInterval,
		&out.RequestTimeout, &out.BackfillStart, &out.UpdatedAt,
	)
	if err != nil {
		return RateSource{}, fmt.Errorf("upsert rate source %q: %w", s.Kind, err)
	}

	return out, nil
}
