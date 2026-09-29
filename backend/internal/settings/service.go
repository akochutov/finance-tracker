package settings

import (
	"context"
	"errors"

	"github.com/akochutov/finance-tracker/internal/currency"
)

const DefaultBaseCurrency = "USD"

type Service struct {
	repo       *Repository
	currencies *currency.Service
}

func NewService(repo *Repository, currencies *currency.Service) *Service {
	return &Service{repo: repo, currencies: currencies}
}

func (s *Service) Get(ctx context.Context) (Settings, error) {
	settings, err := s.repo.Get(ctx)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return resolve(Settings{BaseCurrency: DefaultBaseCurrency}), nil
		}
		return Settings{}, err
	}

	return resolve(settings), nil
}

func (s *Service) SetBaseCurrency(ctx context.Context, code string) (Settings, error) {
	if err := s.requireFiat(ctx, code); err != nil {
		return Settings{}, err
	}

	out, err := s.repo.Upsert(ctx, code)
	if err != nil {
		return Settings{}, err
	}
	return resolve(out), nil
}

func (s *Service) SetExpenseBaseCurrency(ctx context.Context, code string) (Settings, error) {
	if err := s.requireFiat(ctx, code); err != nil {
		return Settings{}, err
	}

	current, err := s.Get(ctx)
	if err != nil {
		return Settings{}, err
	}

	out, err := s.repo.UpsertExpenseBaseCurrency(ctx, current.BaseCurrency, code)
	if err != nil {
		return Settings{}, err
	}
	return resolve(out), nil
}

func (s *Service) requireFiat(ctx context.Context, code string) error {
	cur, err := s.currencies.GetByCode(ctx, code)
	if err != nil {
		return err
	}
	if cur.Kind != "fiat" {
		return ErrNotFiat
	}
	return nil
}

func resolve(s Settings) Settings {
	if s.ExpenseBaseCurrency == "" {
		s.ExpenseBaseCurrency = s.BaseCurrency
	}
	return s
}
