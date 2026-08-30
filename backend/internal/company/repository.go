package company

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("company not found")

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context) ([]Company, error) {
	const q = `
		SELECT id, name, note, tax_id, address, is_active, created_at, updated_at
		FROM companies`

	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("query companies: %w", err)
	}
	defer rows.Close()

	companies := make([]Company, 0)
	for rows.Next() {
		var c Company
		err := rows.Scan(&c.ID, &c.Name, &c.Note, &c.TaxID, &c.Address, &c.IsActive, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan company: %w", err)
		}
		companies = append(companies, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate companies: %w", err)
	}

	return companies, nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (Company, error) {
	const q = `
		SELECT id, name, note, tax_id, address, is_active, created_at, updated_at
		FROM companies
		WHERE id = $1`

	var c Company
	err := r.db.QueryRow(ctx, q, id).
		Scan(&c.ID, &c.Name, &c.Note, &c.TaxID, &c.Address, &c.IsActive, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Company{}, ErrNotFound
		}
		return Company{}, fmt.Errorf("get company: %w", err)
	}

	return c, nil
}

func (r *Repository) Create(ctx context.Context, c Company) (Company, error) {
	const q = `
		INSERT INTO companies (id, name, note, tax_id, address, is_active)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, name, note, tax_id, address, is_active, created_at, updated_at`

	var out Company
	err := r.db.QueryRow(ctx, q, c.ID, c.Name, c.Note, c.TaxID, c.Address, c.IsActive).
		Scan(&out.ID, &out.Name, &out.Note, &out.TaxID, &out.Address, &out.IsActive, &out.CreatedAt, &out.UpdatedAt)

	if err != nil {
		return Company{}, fmt.Errorf("insert company: %w", err)
	}

	return out, nil
}

func (r *Repository) Update(ctx context.Context, c Company) (Company, error) {
	const q = `
		UPDATE companies
		SET name = $1, note = $2, tax_id = $3, address = $4
		WHERE id = $5
		RETURNING id, name, note, tax_id, address, is_active, created_at, updated_at`

	var out Company
	err := r.db.QueryRow(ctx, q, c.Name, c.Note, c.TaxID, c.Address, c.ID).
		Scan(&out.ID, &out.Name, &out.Note, &out.TaxID, &out.Address, &out.IsActive, &out.CreatedAt, &out.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Company{}, ErrNotFound
		}
		return Company{}, fmt.Errorf("update company: %w", err)
	}

	return out, nil
}

func (r *Repository) Deactivate(ctx context.Context, id uuid.UUID) error {
	const q = `UPDATE companies SET is_active = false WHERE id = $1`

	tag, err := r.db.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("deactivate company: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
