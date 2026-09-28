package expensecategory

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type Service struct {
	groups     *GroupRepository
	categories *CategoryRepository
}

func NewService(groups *GroupRepository, categories *CategoryRepository) *Service {
	return &Service{groups: groups, categories: categories}
}

func (s *Service) CreateGroup(ctx context.Context, name string) (Group, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Group{}, ErrNameRequired
	}

	id, err := uuid.NewV7()
	if err != nil {
		return Group{}, fmt.Errorf("generate uuid: %w", err)
	}

	group := Group{
		ID:       id,
		Name:     name,
		IsActive: true,
	}

	return s.groups.Create(ctx, group)
}

func (s *Service) ListGroups(ctx context.Context) ([]Group, error) {
	return s.groups.List(ctx)
}

func (s *Service) UpdateGroup(ctx context.Context, id uuid.UUID, name string) (Group, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Group{}, ErrNameRequired
	}

	return s.groups.Update(ctx, id, name)
}

func (s *Service) SetGroupActive(ctx context.Context, id uuid.UUID, active bool) error {
	if !active {
		hasActive, err := s.categories.HasActiveInGroup(ctx, id)
		if err != nil {
			return err
		}
		if hasActive {
			return ErrGroupNotEmpty
		}
	}

	return s.groups.SetActive(ctx, id, active)
}

func (s *Service) CreateCategory(ctx context.Context, groupID uuid.UUID, name string) (Category, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Category{}, ErrNameRequired
	}

	group, err := s.groups.GetByID(ctx, groupID)
	if err != nil {
		return Category{}, err
	}
	if !group.IsActive {
		return Category{}, ErrGroupInactive
	}

	id, err := uuid.NewV7()
	if err != nil {
		return Category{}, fmt.Errorf("generate uuid: %w", err)
	}

	cat := Category{
		ID:                 id,
		GroupID:            groupID,
		Name:               name,
		IsActive:           true,
		IncludeInDashboard: true,
	}

	return s.categories.Create(ctx, cat)
}

func (s *Service) GetCategory(ctx context.Context, id uuid.UUID) (Category, error) {
	return s.categories.GetByID(ctx, id)
}

func (s *Service) ListCategories(ctx context.Context) ([]Category, error) {
	return s.categories.List(ctx)
}

func (s *Service) UpdateCategory(ctx context.Context, id, groupID uuid.UUID, name string) (Category, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Category{}, ErrNameRequired
	}

	saved, err := s.categories.GetByID(ctx, id)
	if err != nil {
		return Category{}, err
	}

	if saved.GroupID != groupID {
		g, err := s.groups.GetByID(ctx, groupID)
		if err != nil {
			return Category{}, err
		}
		if !g.IsActive {
			return Category{}, ErrGroupInactive
		}
	}

	return s.categories.Update(ctx, id, groupID, name)
}

func (s *Service) SetCategoryActive(ctx context.Context, id uuid.UUID, active bool) error {
	if active {
		cat, err := s.categories.GetByID(ctx, id)
		if err != nil {
			return err
		}

		group, err := s.groups.GetByID(ctx, cat.GroupID)
		if err != nil {
			return err
		}
		if !group.IsActive {
			return ErrGroupInactive
		}
	}

	return s.categories.SetActive(ctx, id, active)
}

func (s *Service) SetCategoryInDashboard(ctx context.Context, id uuid.UUID, include bool) error {
	return s.categories.SetIncludeInDashboard(ctx, id, include)
}
