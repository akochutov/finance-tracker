package company

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

var ErrInvalidInput = errors.New("invalid company data")

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context) ([]Company, error) {
	return s.repo.List(ctx)
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (Company, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, c Company) (Company, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return Company{}, fmt.Errorf("generate uuid: %w", err)
	}

	c.ID = id
	c.Name = strings.TrimSpace(c.Name)
	if c.Name == "" {
		return Company{}, fmt.Errorf("%w: name is empty", ErrInvalidInput)
	}
	c.IsActive = true

	created, err := s.repo.Create(ctx, c)
	if err != nil {
		return Company{}, err
	}

	return created, nil
}

func (s *Service) Update(ctx context.Context, c Company) (Company, error) {
	c.Name = strings.TrimSpace(c.Name)
	if c.Name == "" {
		return Company{}, fmt.Errorf("%w: name is empty", ErrInvalidInput)
	}
	return s.repo.Update(ctx, c)
}

func (s *Service) Deactivate(ctx context.Context, id uuid.UUID) error {
	return s.repo.Deactivate(ctx, id)
}
