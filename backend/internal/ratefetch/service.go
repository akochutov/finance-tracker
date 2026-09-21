package ratefetch

import (
	"context"
	"fmt"
	"time"

	"github.com/akochutov/finance-tracker/internal/currency"
	"github.com/akochutov/finance-tracker/internal/exchangerate"
	"github.com/akochutov/finance-tracker/internal/ratesource"
)

type Result struct {
	Kind    string
	Source  string
	Fetched int
	Skipped int
	Errors  map[string]error
}

type Service struct {
	currencies *currency.Service
	sources    *ratesource.Service
	registry   *exchangerate.Registry
	rates      *exchangerate.Service
}

func NewService(
	currencies *currency.Service,
	sources *ratesource.Service,
	registry *exchangerate.Registry,
	rates *exchangerate.Service,
) *Service {
	return &Service{
		currencies: currencies,
		sources:    sources,
		registry:   registry,
		rates:      rates,
	}
}

func (s *Service) FetchClass(ctx context.Context, kind string, at time.Time) (Result, error) {
	cfg, err := s.sources.Get(ctx, kind) // validates kind, falls back to defaults
	if err != nil {
		return Result{}, err
	}
	if cfg.Source == "" {
		return Result{}, fmt.Errorf("%w: %s", ErrNoProvider, kind)
	}

	provider, err := s.registry.Build(cfg.Source, cfg.URLTemplate, cfg.RequestTimeoutDuration())
	if err != nil {
		return Result{}, err
	}

	all, err := s.currencies.List(ctx)
	if err != nil {
		return Result{}, err
	}

	result := Result{Kind: kind, Source: cfg.Source, Errors: map[string]error{}}

	for _, c := range all {
		if ctx.Err() != nil {
			break
		}
		if c.Kind != kind || !c.IsActive || c.Code == exchangerate.Pivot {
			continue
		}

		fetched, err := provider.FetchRate(ctx, c.Code, at)
		if err != nil {
			result.Errors[c.Code] = err
			result.Skipped++
			continue
		}

		if _, err := s.rates.Record(ctx, c.Code, cfg.Source, fetched.Rate, fetched.RateAt); err != nil {
			result.Errors[c.Code] = err
			result.Skipped++
			continue
		}
		result.Fetched++
	}

	return result, nil
}

func (s *Service) HasProvider(source string) bool {
	return s.registry.Has(source)
}

func (s *Service) ProviderNames() []string {
	return s.registry.Names()
}
