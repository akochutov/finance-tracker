package utility

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const accountColumns = "id, address_id, service, number, expense_category_id, is_active, created_at, updated_at"

type AccountRepository struct {
	db *pgxpool.Pool
}

func NewAccountRepository(db *pgxpool.Pool) *AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) Create(ctx context.Context, account Account) (Account, error) {
	const q = `
		INSERT INTO accounts (id, address_id, service, number, expense_category_id, is_active)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING ` + accountColumns

	var out Account
	err := scanAccount(r.db.QueryRow(ctx, q,
		account.ID, account.AddressID, account.Service, account.Number,
		account.ExpenseCategoryID, account.IsActive,
	), &out)
	if err != nil {
		if mapped := accountWriteError(err); mapped != nil {
			return Account{}, mapped
		}
		return Account{}, fmt.Errorf("insert account: %w", err)
	}

	return out, nil
}

func (r *AccountRepository) GetByID(ctx context.Context, id uuid.UUID) (Account, error) {
	const q = `SELECT ` + accountColumns + ` FROM accounts WHERE id = $1`

	var out Account
	if err := scanAccount(r.db.QueryRow(ctx, q, id), &out); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Account{}, ErrAccountNotFound
		}
		return Account{}, fmt.Errorf("get account by id: %w", err)
	}

	return out, nil
}

func (r *AccountRepository) List(ctx context.Context) ([]Account, error) {
	const q = `SELECT ` + accountColumns + ` FROM accounts ORDER BY service, number`

	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list accounts: %w", err)
	}
	defer rows.Close()

	accounts := make([]Account, 0)
	for rows.Next() {
		var a Account
		if err := scanAccount(rows, &a); err != nil {
			return nil, fmt.Errorf("scan accounts: %w", err)
		}
		accounts = append(accounts, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate accounts: %w", err)
	}

	return accounts, nil
}

func (r *AccountRepository) Update(ctx context.Context, id uuid.UUID, number string, categoryID *uuid.UUID) (Account, error) {
	const q = `
		UPDATE accounts SET number = $1, expense_category_id = $2
		WHERE id = $3
		RETURNING ` + accountColumns

	var out Account
	err := scanAccount(r.db.QueryRow(ctx, q, number, categoryID, id), &out)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Account{}, ErrAccountNotFound
		}
		if mapped := accountWriteError(err); mapped != nil {
			return Account{}, mapped
		}
		return Account{}, fmt.Errorf("update account: %w", err)
	}

	return out, nil
}

func (r *AccountRepository) SetActive(ctx context.Context, id uuid.UUID, active bool) error {
	const q = `UPDATE accounts SET is_active = $1 WHERE id = $2`

	tag, err := r.db.Exec(ctx, q, active, id)
	if err != nil {
		return fmt.Errorf("set account active: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrAccountNotFound
	}

	return nil
}

func accountWriteError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return nil
	}
	switch pgErr.ConstraintName {
	case "uq_accounts_address_service_number":
		return ErrAccountTaken
	case "accounts_address_id_fkey":
		return ErrAddressNotFound
	case "accounts_service_fkey":
		return fmt.Errorf("%w: unknown service", ErrInvalidInput)
	case "accounts_expense_category_id_fkey":
		return fmt.Errorf("%w: unknown expense category", ErrInvalidInput)
	}
	return nil
}

func scanAccount(row pgx.Row, a *Account) error {
	return row.Scan(
		&a.ID, &a.AddressID, &a.Service, &a.Number,
		&a.ExpenseCategoryID, &a.IsActive, &a.CreatedAt, &a.UpdatedAt,
	)
}
