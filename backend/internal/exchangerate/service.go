package exchangerate

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

const divPrecision = 18

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Record(ctx context.Context, currency, source string, rate decimal.Decimal, rateAt time.Time) (Rate, error) {
	if currency == Pivot {
		return Rate{}, ErrPivotRate
	}
	return s.repo.Upsert(ctx, Rate{
		Currency: currency,
		Source:   source,
		RateAt:   rateAt,
		Rate:     rate,
	})
}

func (s *Service) RecordManual(ctx context.Context, currency string, rateAt time.Time, rate decimal.Decimal) (Rate, error) {
	return s.Record(ctx, currency, SourceManual, rate, rateAt)
}

func (s *Service) List(ctx context.Context, currency string) ([]Rate, error) {
	return s.repo.List(ctx, currency)
}

func (s *Service) Convert(ctx context.Context, amount decimal.Decimal, from, to string, at time.Time) (decimal.Decimal, error) {
	if from == to {
		return amount, nil
	}

	priceFrom, err := s.priceInUSD(ctx, from, at)
	if err != nil {
		return decimal.Zero, err
	}
	priceTo, err := s.priceInUSD(ctx, to, at)
	if err != nil {
		return decimal.Zero, err
	}

	return amount.Mul(priceFrom).DivRound(priceTo, divPrecision), nil
}

func (s *Service) priceInUSD(ctx context.Context, code string, at time.Time) (decimal.Decimal, error) {
	if code == Pivot {
		return decimal.NewFromInt(1), nil
	}

	rate, err := s.repo.LatestAsOf(ctx, code, at)
	if err != nil {
		return decimal.Zero, err
	}

	return rate.Rate, nil
}
