package requisite

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CryptoRepository struct {
	db *pgxpool.Pool
}

func NewCryptoRepository(db *pgxpool.Pool) *CryptoRepository {
	return &CryptoRepository{db: db}
}

func (r *CryptoRepository) Create(ctx context.Context, cr CryptoRequisite) (CryptoRequisite, error) {
	const q = `
		INSERT INTO crypto_requisites (
			id, company_id, network, wallet_address, valid_from, valid_to
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING
			id, company_id, network, wallet_address, 
			valid_from, valid_to, created_at, updated_at`

	var out CryptoRequisite
	err := r.db.QueryRow(
		ctx, q, cr.ID, cr.CompanyID, cr.Network, cr.WalletAddress,
		cr.ValidFrom, cr.ValidTo,
	).Scan(
		&out.ID, &out.CompanyID, &out.Network, &out.WalletAddress,
		&out.ValidFrom, &out.ValidTo, &out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		return CryptoRequisite{}, fmt.Errorf("insert crypto requisite: %w", err)
	}

	return out, nil
}

func (r *CryptoRepository) GetByID(ctx context.Context, id uuid.UUID) (CryptoRequisite, error) {
	const q = `
		SELECT id, company_id, network, wallet_address, valid_from, valid_to, created_at, updated_at
		FROM crypto_requisites
		WHERE id = $1`

	var out CryptoRequisite
	err := r.db.QueryRow(ctx, q, id).Scan(
		&out.ID, &out.CompanyID, &out.Network, &out.WalletAddress,
		&out.ValidFrom, &out.ValidTo, &out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return CryptoRequisite{}, ErrNotFound
		}
		return CryptoRequisite{}, fmt.Errorf("get crypto requisite by id: %w", err)
	}

	return out, nil
}

func (r *CryptoRepository) ListByCompany(ctx context.Context, companyID uuid.UUID) ([]CryptoRequisite, error) {
	const q = `
		SELECT id, company_id, network, wallet_address, valid_from, valid_to, created_at, updated_at
		FROM crypto_requisites
		WHERE company_id = $1
		ORDER BY valid_from DESC`

	rows, err := r.db.Query(ctx, q, companyID)
	if err != nil {
		return nil, fmt.Errorf("get crypto requisites: %w", err)
	}
	defer rows.Close()

	cryptoRequisites := make([]CryptoRequisite, 0)
	for rows.Next() {
		var cr CryptoRequisite
		err := rows.Scan(
			&cr.ID, &cr.CompanyID, &cr.Network, &cr.WalletAddress,
			&cr.ValidFrom, &cr.ValidTo, &cr.CreatedAt, &cr.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan crypto requisite: %w", err)
		}
		cryptoRequisites = append(cryptoRequisites, cr)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate crypto requisites: %w", err)
	}

	return cryptoRequisites, nil
}

func (r *CryptoRepository) Close(ctx context.Context, id uuid.UUID, validTo time.Time) error {
	const q = `UPDATE crypto_requisites SET valid_to = $1 WHERE id = $2 AND valid_to IS NULL`

	tag, err := r.db.Exec(ctx, q, validTo, id)
	if err != nil {
		return fmt.Errorf("close crypto requisite: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
