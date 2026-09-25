package expensecategory

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GroupRepository struct {
	db *pgxpool.Pool
}

func NewGroupRepository(db *pgxpool.Pool) *GroupRepository {
	return &GroupRepository{db: db}
}

func (r *GroupRepository) Create(ctx context.Context, group Group) (Group, error) {
	const q = `
		INSERT INTO expense_groups (id, name, is_active)
		VALUES ($1, $2, $3)
		RETURNING id, name, is_active, created_at, updated_at`

	var out Group
	err := r.db.QueryRow(ctx, q, group.ID, group.Name, group.IsActive).
		Scan(&out.ID, &out.Name, &out.IsActive, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
			return Group{}, ErrGroupNameTaken
		}
		return Group{}, fmt.Errorf("insert group: %w", err)
	}

	return out, nil
}

func (r *GroupRepository) GetByID(ctx context.Context, id uuid.UUID) (Group, error) {
	const q = `
		SELECT id, name, is_active, created_at, updated_at
		FROM expense_groups
		WHERE id = $1`

	var out Group
	err := r.db.QueryRow(ctx, q, id).
		Scan(&out.ID, &out.Name, &out.IsActive, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Group{}, ErrGroupNotFound
		}
		return Group{}, fmt.Errorf("get group by id: %w", err)
	}

	return out, nil
}

func (r *GroupRepository) List(ctx context.Context) ([]Group, error) {
	const q = `
		SELECT id, name, is_active, created_at, updated_at
		FROM expense_groups
		ORDER BY name`

	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list groups: %w", err)
	}
	defer rows.Close()

	groups := make([]Group, 0)
	for rows.Next() {
		var g Group
		err := rows.Scan(&g.ID, &g.Name, &g.IsActive, &g.CreatedAt, &g.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan groups: %w", err)
		}
		groups = append(groups, g)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate groups: %w", err)
	}

	return groups, nil
}

func (r *GroupRepository) Update(ctx context.Context, id uuid.UUID, name string) (Group, error) {
	const q = `
		UPDATE expense_groups
		SET name = $1
		WHERE id = $2
		RETURNING id, name, is_active, created_at, updated_at`

	var out Group
	err := r.db.QueryRow(ctx, q, name, id).
		Scan(&out.ID, &out.Name, &out.IsActive, &out.CreatedAt, &out.UpdatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
			return Group{}, ErrGroupNameTaken
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return Group{}, ErrGroupNotFound
		}
		return Group{}, fmt.Errorf("update group: %w", err)
	}

	return out, nil
}

func (r *GroupRepository) SetActive(ctx context.Context, id uuid.UUID, active bool) error {
	const q = `UPDATE expense_groups SET is_active = $1 WHERE id = $2`

	tag, err := r.db.Exec(ctx, q, active, id)
	if err != nil {
		return fmt.Errorf("set group active: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrGroupNotFound
	}

	return nil
}
