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

const accountColumns = "id, address_id, service, number, is_active, created_at, updated_at"

type AccountRepository struct {
	db *pgxpool.Pool
}

func NewAccountRepository(db *pgxpool.Pool) *AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) Create(ctx context.Context, account Account) (Account, error) {
	const q = `
		INSERT INTO accounts (id, address_id, service, number, is_active)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING ` + accountColumns

	var out Account
	err := scanAccount(
		r.db.QueryRow(
			ctx, q, account.ID, account.AddressID, account.Service, account.Number, account.IsActive,
		), &out)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch {
			case pgErr.Code == uniqueViolation:
				return Account{}, ErrAccountTaken
			case pgErr.ConstraintName == "accounts_address_id_fkey":
				return Account{}, ErrAddressNotFound
			case pgErr.ConstraintName == "accounts_service_fkey":
				return Account{}, fmt.Errorf("%w: unknown service", ErrInvalidInput)
			}
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

func (r *AccountRepository) Update(ctx context.Context, id uuid.UUID, number string) (Account, error) {
	const q = `
		UPDATE accounts SET number = $1
		WHERE id = $2
		RETURNING ` + accountColumns

	var out Account
	err := scanAccount(r.db.QueryRow(ctx, q, number, id), &out)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
			return Account{}, ErrAccountTaken
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return Account{}, ErrAccountNotFound
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

func scanAccount(row pgx.Row, a *Account) error {
	return row.Scan(&a.ID, &a.AddressID, &a.Service, &a.Number, &a.IsActive, &a.CreatedAt, &a.UpdatedAt)
}
