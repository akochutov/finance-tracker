package ratesource

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
)

const MinPollIntervalSeconds = 900

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Get(ctx context.Context, kind string) (RateSource, error) {
	if kind != KindFiat && kind != KindCrypto {
		return RateSource{}, ErrInvalidKind
	}

	src, err := s.repo.Get(ctx, kind)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return defaultFor(kind), nil
		}
		return RateSource{}, err
	}

	return src, nil
}

func (s *Service) List(ctx context.Context) ([]RateSource, error) {
	stored, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	byKind := make(map[string]RateSource, len(stored))
	for _, src := range stored {
		byKind[src.Kind] = src
	}

	out := make([]RateSource, 0, 2)
	for _, kind := range []string{KindFiat, KindCrypto} {
		if src, ok := byKind[kind]; ok {
			out = append(out, src)
		} else {
			out = append(out, defaultFor(kind))
		}
	}

	return out, nil
}

func (s *Service) Save(ctx context.Context, src RateSource) (RateSource, error) {
	if src.Kind != KindFiat && src.Kind != KindCrypto {
		return RateSource{}, ErrInvalidKind
	}
	if src.Source == "" {
		return RateSource{}, fmt.Errorf("%w: source is required", ErrInvalidConfig)
	}
	if src.URLTemplate == "" {
		return RateSource{}, fmt.Errorf("%w: url_template is required", ErrInvalidConfig)
	}
	if err := validateURLTemplate(src.URLTemplate); err != nil {
		return RateSource{}, err
	}
	if src.PollInterval < MinPollIntervalSeconds {
		return RateSource{}, fmt.Errorf("%w: poll interval must be at least %d seconds", ErrInvalidConfig, MinPollIntervalSeconds)
	}
	if src.RequestTimeout <= 0 {
		return RateSource{}, fmt.Errorf("%w: request timeout must be positive", ErrInvalidConfig)
	}

	return s.repo.Upsert(ctx, src)
}

func defaultFor(kind string) RateSource {
	switch kind {
	case KindFiat:
		return RateSource{
			Kind:           KindFiat,
			Source:         "frankfurter",
			URLTemplate:    "https://api.frankfurter.dev/v2/rate/{base}/{quote}?date={date}",
			PollInterval:   86400,
			RequestTimeout: 10,
		}
	case KindCrypto:
		return RateSource{
			Kind:           KindCrypto,
			Source:         "",
			URLTemplate:    "",
			PollInterval:   900,
			RequestTimeout: 10,
		}
	default:
		return RateSource{Kind: kind}
	}
}

func validateURLTemplate(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("%w: url_template is not a valid URL", ErrInvalidConfig)
	}
	if u.Scheme != "https" {
		return fmt.Errorf("%w: url_template must use https", ErrInvalidConfig)
	}
	if u.Host == "" {
		return fmt.Errorf("%w: url_template must include a host", ErrInvalidConfig)
	}
	if !strings.Contains(raw, "{base}") || !strings.Contains(raw, "{quote}") {
		return fmt.Errorf("%w: url_template must contain {base} and {quote} placeholders", ErrInvalidConfig)
	}
	return nil
}
