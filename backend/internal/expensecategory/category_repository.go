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

type CategoryRepository struct {
	db *pgxpool.Pool
}

func NewCategoryRepository(db *pgxpool.Pool) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) Create(ctx context.Context, cat Category) (Category, error) {
	const q = `
		INSERT INTO expense_categories (id, group_id, name, is_active)
		VALUES ($1, $2, $3, $4)
		RETURNING id, group_id, name, is_active, created_at, updated_at`

	var out Category
	err := r.db.QueryRow(ctx, q, cat.ID, cat.GroupID, cat.Name, cat.IsActive).
		Scan(&out.ID, &out.GroupID, &out.Name, &out.IsActive, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case uniqueViolation:
				return Category{}, ErrCategoryNameTaken
			case foreignKeyViolation:
				return Category{}, ErrGroupNotFound
			}
		}
		return Category{}, fmt.Errorf("insert category: %w", err)
	}

	return out, nil
}

func (r *CategoryRepository) GetByID(ctx context.Context, id uuid.UUID) (Category, error) {
	const q = `
		SELECT id, group_id, name, is_active, created_at, updated_at
		FROM expense_categories
		WHERE id = $1`

	var out Category
	err := r.db.QueryRow(ctx, q, id).
		Scan(&out.ID, &out.GroupID, &out.Name, &out.IsActive, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Category{}, ErrCategoryNotFound
		}
		return Category{}, fmt.Errorf("get category by id: %w", err)
	}

	return out, nil
}

func (r *CategoryRepository) List(ctx context.Context) ([]Category, error) {
	const q = `
		SELECT id, group_id, name, is_active, created_at, updated_at
		FROM expense_categories
		ORDER BY name`

	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	defer rows.Close()

	categories := make([]Category, 0)
	for rows.Next() {
		var c Category
		err := rows.Scan(&c.ID, &c.GroupID, &c.Name, &c.IsActive, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan categories: %w", err)
		}
		categories = append(categories, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate categories: %w", err)
	}

	return categories, nil
}

func (r *CategoryRepository) HasActiveInGroup(ctx context.Context, groupID uuid.UUID) (bool, error) {
	const q = `
		SELECT EXISTS (
			SELECT 1
			FROM expense_categories
			WHERE group_id = $1 AND is_active
		)`

	var exists bool
	if err := r.db.QueryRow(ctx, q, groupID).Scan(&exists); err != nil {
		return false, fmt.Errorf("check active categories in group: %w", err)
	}

	return exists, nil
}

func (r *CategoryRepository) Update(ctx context.Context, id, groupID uuid.UUID, name string) (Category, error) {
	const q = `
		UPDATE expense_categories
		SET group_id = $1, name = $2
		WHERE id = $3
		RETURNING id, group_id, name, is_active, created_at, updated_at`

	var out Category
	err := r.db.QueryRow(ctx, q, groupID, name, id).
		Scan(&out.ID, &out.GroupID, &out.Name, &out.IsActive, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case uniqueViolation:
				return Category{}, ErrCategoryNameTaken
			case foreignKeyViolation:
				return Category{}, ErrGroupNotFound
			}
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return Category{}, ErrCategoryNotFound
		}
		return Category{}, fmt.Errorf("update category: %w", err)
	}

	return out, nil
}

func (r *CategoryRepository) SetActive(ctx context.Context, id uuid.UUID, active bool) error {
	const q = `UPDATE expense_categories SET is_active = $1 WHERE id = $2`

	tag, err := r.db.Exec(ctx, q, active, id)
	if err != nil {
		return fmt.Errorf("set category active: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrCategoryNotFound
	}

	return nil
}
