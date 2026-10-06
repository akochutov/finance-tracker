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

func (s *Service) Create(ctx context.Context, service, zone string, validFrom time.Time, currency, tierMode string, tiers []Tier) (Tariff, error) {
	service = strings.ToLower(strings.TrimSpace(service))
	zone = strings.ToLower(strings.TrimSpace(zone))
	if service == "" {
		return Tariff{}, fmt.Errorf("%w: service is required", ErrInvalidInput)
	}
	if err := checkZone(service, zone); err != nil {
		return Tariff{}, err
	}

	validFrom, currency, tierMode, err := normalize(validFrom, currency, tierMode, tiers)
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
		Currency:  currency,
		TierMode:  tierMode,
		Tiers:     tiers,
	})
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, validFrom time.Time, currency, tierMode string, tiers []Tier) (Tariff, error) {
	validFrom, currency, tierMode, err := normalize(validFrom, currency, tierMode, tiers)
	if err != nil {
		return Tariff{}, err
	}
	return s.repo.Update(ctx, id, validFrom, currency, tierMode, tiers)
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) ActiveAt(ctx context.Context, service, zone string, day time.Time) (Tariff, error) {
	return s.repo.ActiveAt(ctx, service, zone, dateOnly(day))
}

func Cost(t Tariff, volume, priced decimal.Decimal) decimal.Decimal {
	if !volume.IsPositive() || !priced.IsPositive() || len(t.Tiers) == 0 {
		return decimal.Zero
	}

	if t.TierMode != ModeProgressive {
		return priced.Mul(tierFor(t.Tiers, volume).Price)
	}

	total := decimal.Zero
	lower := decimal.Zero
	remaining := volume
	for _, tier := range t.Tiers {
		part := remaining
		if tier.UpTo != nil {
			band := tier.UpTo.Sub(lower)
			if part.GreaterThan(band) {
				part = band
			}
			lower = *tier.UpTo
		}
		total = total.Add(part.Mul(tier.Price))
		remaining = remaining.Sub(part)
		if !remaining.IsPositive() {
			break
		}
	}
	return total.Mul(priced).Div(volume)
}

func tierFor(tiers []Tier, volume decimal.Decimal) Tier {
	for _, tier := range tiers {
		if tier.UpTo == nil || volume.LessThanOrEqual(*tier.UpTo) {
			return tier
		}
	}
	return tiers[len(tiers)-1]
}

func checkZone(service, zone string) error {
	switch zone {
	case utility.ZoneSingle, utility.ZoneDay, utility.ZoneNight:
		return nil
	case "":
		return fmt.Errorf("%w: zone is required", ErrInvalidInput)
	default:
		return fmt.Errorf("%w: unknown zone %q", ErrInvalidInput, zone)
	}
}

func normalize(validFrom time.Time, currency, tierMode string, tiers []Tier) (time.Time, string, string, error) {
	if validFrom.IsZero() {
		return time.Time{}, "", "", fmt.Errorf("%w: start date is required", ErrInvalidInput)
	}
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if currency == "" {
		return time.Time{}, "", "", fmt.Errorf("%w: currency is required", ErrInvalidInput)
	}
	tierMode = strings.ToLower(strings.TrimSpace(tierMode))
	if tierMode == "" {
		tierMode = ModeWhole
	}
	if tierMode != ModeWhole && tierMode != ModeProgressive {
		return time.Time{}, "", "", fmt.Errorf("%w: tier mode must be %q or %q", ErrInvalidInput, ModeWhole, ModeProgressive)
	}
	if err := checkTiers(tiers); err != nil {
		return time.Time{}, "", "", err
	}
	return dateOnly(validFrom), currency, tierMode, nil
}

func checkTiers(tiers []Tier) error {
	if len(tiers) == 0 {
		return fmt.Errorf("%w: at least one price is required", ErrInvalidInput)
	}
	last := len(tiers) - 1
	prev := decimal.Zero
	for i, tier := range tiers {
		if !tier.Price.IsPositive() {
			return fmt.Errorf("%w: tier %d: price must be positive", ErrInvalidInput, i+1)
		}
		if i == last {
			if tier.UpTo != nil {
				return fmt.Errorf("%w: the last tier must have no upper bound", ErrInvalidInput)
			}
			break
		}
		if tier.UpTo == nil {
			return fmt.Errorf("%w: tier %d: only the last tier can be open-ended", ErrInvalidInput, i+1)
		}
		if !tier.UpTo.GreaterThan(prev) {
			return fmt.Errorf("%w: tier %d: bounds must grow (%s after %s)", ErrInvalidInput, i+1, tier.UpTo, prev)
		}
		prev = *tier.UpTo
	}
	return nil
}

func dateOnly(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}
