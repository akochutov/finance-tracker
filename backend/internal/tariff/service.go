package tariff

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/akochutov/finance-tracker/internal/utility"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context) ([]Tariff, error) {
	return s.repo.List(ctx)
}

func (s *Service) Create(ctx context.Context, service, zone string, validFrom time.Time, price decimal.Decimal, currency string) (Tariff, error) {
	service = strings.ToLower(strings.TrimSpace(service))
	zone = strings.ToLower(strings.TrimSpace(zone))
	if service == "" {
		return Tariff{}, fmt.Errorf("%w: service is required", ErrInvalidInput)
	}
	if err := checkZone(service, zone); err != nil {
		return Tariff{}, err
	}

	validFrom, currency, err := normalize(validFrom, price, currency)
	if err != nil {
		return Tariff{}, err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return Tariff{}, fmt.Errorf("generate uuid: %w", err)
	}

	return s.repo.Create(ctx, Tariff{
		ID:        id,
		Service:   service,
		Zone:      zone,
		ValidFrom: validFrom,
		Price:     price,
		Currency:  currency,
	})
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, validFrom time.Time, price decimal.Decimal, currency string) (Tariff, error) {
	validFrom, currency, err := normalize(validFrom, price, currency)
	if err != nil {
		return Tariff{}, err
	}
	return s.repo.Update(ctx, id, validFrom, price, currency)
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) ActiveAt(ctx context.Context, service, zone string, day time.Time) (Tariff, error) {
	return s.repo.ActiveAt(ctx, service, zone, dateOnly(day))
}

func checkZone(service, zone string) error {
	switch zone {
	case utility.ZoneSingle:
		return nil
	case utility.ZoneDay, utility.ZoneNight:
		if service == utility.ServiceElectricity {
			return nil
		}
		return fmt.Errorf("%w: only electricity has day and night tariffs", ErrInvalidInput)
	case "":
		return fmt.Errorf("%w: zone is required", ErrInvalidInput)
	default:
		return fmt.Errorf("%w: unknown zone %q", ErrInvalidInput, zone)
	}
}

func normalize(validFrom time.Time, price decimal.Decimal, currency string) (time.Time, string, error) {
	if validFrom.IsZero() {
		return time.Time{}, "", fmt.Errorf("%w: start date is required", ErrInvalidInput)
	}
	if !price.IsPositive() {
		return time.Time{}, "", fmt.Errorf("%w: price must be positive", ErrInvalidInput)
	}
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if currency == "" {
		return time.Time{}, "", fmt.Errorf("%w: currency is required", ErrInvalidInput)
	}
	return dateOnly(validFrom), currency, nil
}

func dateOnly(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}
