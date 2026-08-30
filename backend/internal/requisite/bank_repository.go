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

var ErrNotFound = errors.New("requisite not found")

type BankRepository struct {
	db *pgxpool.Pool
}

func NewBankRepository(db *pgxpool.Pool) *BankRepository {
	return &BankRepository{db: db}
}

func (r *BankRepository) Create(ctx context.Context, br BankRequisite) (BankRequisite, error) {
	const q = `
		INSERT INTO bank_requisites (
			id, company_id, beneficiary_name, account_number, bank_name, bank_swift, bank_address,
			correspondent_bank_name, correspondent_bank_swift, intermediary_bank_name, intermediary_bank_swift,
			valid_from, valid_to
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING 
			id, company_id, beneficiary_name, account_number, bank_name, bank_swift, bank_address,
			correspondent_bank_name, correspondent_bank_swift, intermediary_bank_name, intermediary_bank_swift,
			valid_from, valid_to, created_at, updated_at`

	var out BankRequisite
	err := r.db.QueryRow(
		ctx, q, br.ID, br.CompanyID, br.BeneficiaryName, br.AccountNumber,
		br.BankName, br.BankSwift, br.BankAddress, br.CorrespondentBankName, br.CorrespondentBankSwift,
		br.IntermediaryBankName, br.IntermediaryBankSwift, br.ValidFrom, br.ValidTo,
	).Scan(
		&out.ID, &out.CompanyID, &out.BeneficiaryName, &out.AccountNumber,
		&out.BankName, &out.BankSwift, &out.BankAddress, &out.CorrespondentBankName, &out.CorrespondentBankSwift,
		&out.IntermediaryBankName, &out.IntermediaryBankSwift, &out.ValidFrom, &out.ValidTo,
		&out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		return BankRequisite{}, fmt.Errorf("insert bank requisite: %w", err)
	}

	return out, nil
}

func (r *BankRepository) GetByID(ctx context.Context, id uuid.UUID) (BankRequisite, error) {
	const q = `
		SELECT 
			id, company_id, beneficiary_name, account_number, bank_name, bank_swift, bank_address,
			correspondent_bank_name, correspondent_bank_swift, intermediary_bank_name, intermediary_bank_swift,
			valid_from, valid_to, created_at, updated_at
		FROM bank_requisites
		WHERE id = $1`

	var out BankRequisite
	err := r.db.QueryRow(ctx, q, id).Scan(
		&out.ID, &out.CompanyID, &out.BeneficiaryName, &out.AccountNumber,
		&out.BankName, &out.BankSwift, &out.BankAddress, &out.CorrespondentBankName, &out.CorrespondentBankSwift,
		&out.IntermediaryBankName, &out.IntermediaryBankSwift, &out.ValidFrom, &out.ValidTo,
		&out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return BankRequisite{}, ErrNotFound
		}
		return BankRequisite{}, fmt.Errorf("get bank requisite by id: %w", err)
	}

	return out, nil
}

func (r *BankRepository) ListByCompany(ctx context.Context, companyID uuid.UUID) ([]BankRequisite, error) {
	const q = `
		SELECT 
			id, company_id, beneficiary_name, account_number, bank_name, bank_swift, bank_address,
			correspondent_bank_name, correspondent_bank_swift, intermediary_bank_name, intermediary_bank_swift,
			valid_from, valid_to, created_at, updated_at
		FROM bank_requisites
		WHERE company_id = $1
		ORDER BY valid_from DESC`

	rows, err := r.db.Query(ctx, q, companyID)
	if err != nil {
		return nil, fmt.Errorf("get bank requisites: %w", err)
	}
	defer rows.Close()

	bankRequisites := make([]BankRequisite, 0)
	for rows.Next() {
		var br BankRequisite
		err := rows.Scan(
			&br.ID, &br.CompanyID, &br.BeneficiaryName, &br.AccountNumber,
			&br.BankName, &br.BankSwift, &br.BankAddress, &br.CorrespondentBankName, &br.CorrespondentBankSwift,
			&br.IntermediaryBankName, &br.IntermediaryBankSwift, &br.ValidFrom, &br.ValidTo,
			&br.CreatedAt, &br.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan bank requisite: %w", err)
		}
		bankRequisites = append(bankRequisites, br)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate bank requisites: %w", err)
	}

	return bankRequisites, nil
}

func (r *BankRepository) Close(ctx context.Context, id uuid.UUID, validTo time.Time) error {
	const q = `UPDATE bank_requisites SET valid_to = $1 WHERE id = $2 AND valid_to IS NULL`

	tag, err := r.db.Exec(ctx, q, validTo, id)
	if err != nil {
		return fmt.Errorf("close bank requisite: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
