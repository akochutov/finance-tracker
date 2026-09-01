package income

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("income not found")

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, in Income) (Income, error) {
	const q = `
		INSERT INTO incomes (
			id, payer_id, beneficiary_id, amount, currency, occurred_at, payment_type, 
			payer_requisite_id, beneficiary_requisite_id, note, is_active
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING 
			id, payer_id, beneficiary_id, amount, currency, occurred_at, payment_type,
    		payer_requisite_id, beneficiary_requisite_id, note, is_active, created_at, updated_at`

	var out Income
	err := r.db.QueryRow(
		ctx, q, in.ID, in.PayerID, in.BeneficiaryID, in.Amount, in.Currency, in.OccurredAt,
		in.PaymentType, in.PayerRequisiteID, in.BeneficiaryRequisiteID, in.Note, in.IsActive,
	).Scan(
		&out.ID, &out.PayerID, &out.BeneficiaryID, &out.Amount, &out.Currency, &out.OccurredAt,
		&out.PaymentType, &out.PayerRequisiteID, &out.BeneficiaryRequisiteID, &out.Note, &out.IsActive,
		&out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		return Income{}, fmt.Errorf("insert income: %w", err)
	}

	return out, nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (Income, error) {
	const q = `
		SELECT
			id, payer_id, beneficiary_id, amount, currency, occurred_at, payment_type,
    		payer_requisite_id, beneficiary_requisite_id, note, is_active, created_at, updated_at
		FROM incomes
		WHERE id = $1`

	var out Income
	err := r.db.QueryRow(ctx, q, id).Scan(
		&out.ID, &out.PayerID, &out.BeneficiaryID, &out.Amount, &out.Currency, &out.OccurredAt,
		&out.PaymentType, &out.PayerRequisiteID, &out.BeneficiaryRequisiteID, &out.Note, &out.IsActive,
		&out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Income{}, ErrNotFound
		}
		return Income{}, fmt.Errorf("get income by id: %w", err)
	}

	return out, nil
}

func (r *Repository) List(ctx context.Context) ([]Income, error) {
	const q = `
		SELECT
			id, payer_id, beneficiary_id, amount, currency, occurred_at, payment_type,
    		payer_requisite_id, beneficiary_requisite_id, note, is_active, created_at, updated_at
		FROM incomes
		ORDER BY occurred_at DESC`

	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("get incomes: %w", err)
	}
	defer rows.Close()

	incomes := make([]Income, 0)
	for rows.Next() {
		var in Income
		err := rows.Scan(
			&in.ID, &in.PayerID, &in.BeneficiaryID, &in.Amount, &in.Currency, &in.OccurredAt,
			&in.PaymentType, &in.PayerRequisiteID, &in.BeneficiaryRequisiteID, &in.Note, &in.IsActive,
			&in.CreatedAt, &in.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan income: %w", err)
		}
		incomes = append(incomes, in)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate incomes: %w", err)
	}

	return incomes, nil
}
