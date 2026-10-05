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

const addressColumns = "id, address, is_active, created_at, updated_at"

type AddressRepository struct {
	db *pgxpool.Pool
}

func NewAddressRepository(db *pgxpool.Pool) *AddressRepository {
	return &AddressRepository{db: db}
}

func (r *AddressRepository) Create(ctx context.Context, address Address) (Address, error) {
	const q = `
		INSERT INTO addresses (id, address, is_active)
		VALUES ($1, $2, $3)
		RETURNING ` + addressColumns

	var out Address
	err := scanAddress(r.db.QueryRow(ctx, q, address.ID, address.Address, address.IsActive), &out)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
			return Address{}, ErrAddressTaken
		}
		return Address{}, fmt.Errorf("insert address: %w", err)
	}

	return out, nil
}

func (r *AddressRepository) GetByID(ctx context.Context, id uuid.UUID) (Address, error) {
	const q = `SELECT ` + addressColumns + ` FROM addresses WHERE id = $1`

	var out Address
	if err := scanAddress(r.db.QueryRow(ctx, q, id), &out); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Address{}, ErrAddressNotFound
		}
		return Address{}, fmt.Errorf("get address by id: %w", err)
	}

	return out, nil
}

func (r *AddressRepository) List(ctx context.Context) ([]Address, error) {
	const q = `SELECT ` + addressColumns + ` FROM addresses ORDER BY address`

	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list addresses: %w", err)
	}
	defer rows.Close()

	addresses := make([]Address, 0)
	for rows.Next() {
		var a Address
		if err := scanAddress(rows, &a); err != nil {
			return nil, fmt.Errorf("scan addresses: %w", err)
		}
		addresses = append(addresses, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate addresses: %w", err)
	}

	return addresses, nil
}

func (r *AddressRepository) Update(ctx context.Context, id uuid.UUID, address string) (Address, error) {
	const q = `
		UPDATE addresses SET address = $1
		WHERE id = $2
		RETURNING ` + addressColumns

	var out Address
	err := scanAddress(r.db.QueryRow(ctx, q, address, id), &out)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
			return Address{}, ErrAddressTaken
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return Address{}, ErrAddressNotFound
		}
		return Address{}, fmt.Errorf("update address: %w", err)
	}

	return out, nil
}

func (r *AddressRepository) SetActive(ctx context.Context, id uuid.UUID, active bool) error {
	const q = `UPDATE addresses SET is_active = $1 WHERE id = $2`

	tag, err := r.db.Exec(ctx, q, active, id)
	if err != nil {
		return fmt.Errorf("set address active: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrAddressNotFound
	}

	return nil
}

func scanAddress(row pgx.Row, a *Address) error {
	return row.Scan(&a.ID, &a.Address, &a.IsActive, &a.CreatedAt, &a.UpdatedAt)
}
