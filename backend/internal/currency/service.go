package currency

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

var ErrInvalidInput = errors.New("invalid currency data")

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context) ([]Currency, error) {
	return s.repo.List(ctx)
}

func (s *Service) GetByCode(ctx context.Context, code string) (Currency, error) {
	return s.repo.GetByCode(ctx, code)
}

func (s *Service) Create(ctx context.Context, c Currency) (Currency, error) {
	c.Code = strings.ToUpper(strings.TrimSpace(c.Code))
	c.Name = strings.TrimSpace(c.Name)
	c.Kind = strings.ToLower(strings.TrimSpace(c.Kind))
	c.IsActive = true

	err := validate(c)
	if err != nil {
		return Currency{}, err
	}

	out, err := s.repo.Create(ctx, c)
	if err != nil {
		return Currency{}, err
	}

	return out, nil
}

func (s *Service) Deactivate(ctx context.Context, code string) error {
	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" {
		return fmt.Errorf("%w: code is empty", ErrInvalidInput)
	}
	return s.repo.Deactivate(ctx, code)
}

func validate(c Currency) error {
	if c.Code == "" {
		return fmt.Errorf("%w: code is empty", ErrInvalidInput)
	}
	if len(c.Code) > 10 {
		return fmt.Errorf("%w: code must be shorter than 10 letters", ErrInvalidInput)
	}

	if c.Name == "" {
		return fmt.Errorf("%w: name is empty", ErrInvalidInput)
	}

	if c.Kind != "fiat" && c.Kind != "crypto" {
		return fmt.Errorf("%w: kind must be fiat or crypto", ErrInvalidInput)
	}

	if c.DecimalPlaces < 0 {
		return fmt.Errorf("%w: decimal_places must not be negative", ErrInvalidInput)
	}
	if c.DecimalPlaces > 2 && c.Kind == "fiat" {
		return fmt.Errorf("%w: decimal_places must not be greater than 2 for fiat currencies", ErrInvalidInput)
	}
	if c.DecimalPlaces > 18 && c.Kind == "crypto" {
		return fmt.Errorf("%w: decimal_places must not be greater than 18 for crypto currencies", ErrInvalidInput)
	}

	return nil
}
