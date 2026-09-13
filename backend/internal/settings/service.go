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
			return Settings{BaseCurrency: DefaultBaseCurrency}, nil
		}
		return Settings{}, err
	}

	return settings, nil
}

func (s *Service) SetBaseCurrency(ctx context.Context, code string) (Settings, error) {
	cur, err := s.currencies.GetByCode(ctx, code)
	if err != nil {
		return Settings{}, err
	}

	if cur.Kind != "fiat" {
		return Settings{}, ErrNotFiat
	}

	return s.repo.Upsert(ctx, code)
}
