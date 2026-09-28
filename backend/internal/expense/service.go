package expense

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/akochutov/finance-tracker/internal/expensecategory"
	"github.com/google/uuid"
)

var paymentTypes = map[string]bool{
	"bank":   true,
	"crypto": true,
	"cash":   true,
}

type Service struct {
	repo       *Repository
	categories *expensecategory.Service
	cache      SuggestionCache
}

func NewService(repo *Repository, categories *expensecategory.Service, cache SuggestionCache) *Service {
	return &Service{repo: repo, categories: categories, cache: cache}
}

func (s *Service) List(ctx context.Context, from, to *time.Time) ([]Expense, error) {
	if from != nil && to != nil && to.Before(*from) {
		return nil, fmt.Errorf("%w: to is before from", ErrInvalidInput)
	}
	return s.repo.List(ctx, from, to)
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (Expense, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	s.invalidateSuggestions(ctx)
	return nil
}

func (s *Service) Create(ctx context.Context, e Expense) (Expense, error) {
	if err := normalizeHeader(&e); err != nil {
		return Expense{}, err
	}

	if err := s.prepareItems(ctx, &e, nil); err != nil {
		return Expense{}, err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return Expense{}, fmt.Errorf("generate uuid: %w", err)
	}
	e.ID = id

	out, err := s.repo.Create(ctx, e)
	if err != nil {
		return Expense{}, err
	}
	s.invalidateSuggestions(ctx)
	return out, nil
}

func (s *Service) Update(ctx context.Context, e Expense) (Expense, error) {
	existing, err := s.repo.GetByID(ctx, e.ID)
	if err != nil {
		return Expense{}, err
	}

	if err := normalizeHeader(&e); err != nil {
		return Expense{}, err
	}

	keepInactive := make(map[uuid.UUID]bool, len(existing.Items))
	for _, it := range existing.Items {
		keepInactive[it.CategoryID] = true
	}

	if err := s.prepareItems(ctx, &e, keepInactive); err != nil {
		return Expense{}, err
	}

	out, err := s.repo.Update(ctx, e)
	if err != nil {
		return Expense{}, err
	}
	s.invalidateSuggestions(ctx)
	return out, nil
}

func (s *Service) Suggestions(ctx context.Context) ([]Suggestion, error) {
	list, ok, err := s.cache.Get(ctx)
	if err != nil {
		log.Printf("WARN suggestions cache get: %v", err)
	}
	if ok {
		return list, nil
	}

	list, err = s.repo.Suggestions(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.cache.Set(ctx, list); err != nil {
		log.Printf("WARN suggestions cache set: %v", err)
	}

	return list, nil
}

func normalizeHeader(e *Expense) error {
	if e.OccurredOn.IsZero() {
		return fmt.Errorf("%w: occurred_on is required", ErrInvalidInput)
	}
	e.OccurredOn = dateOnly(e.OccurredOn)

	e.Currency = strings.ToUpper(strings.TrimSpace(e.Currency))
	if e.Currency == "" {
		return fmt.Errorf("%w: currency is required", ErrInvalidInput)
	}

	e.PaymentType = trimOrNil(e.PaymentType)
	if e.PaymentType != nil && !paymentTypes[*e.PaymentType] {
		return fmt.Errorf("%w: unknown payment_type %q", ErrInvalidInput, *e.PaymentType)
	}

	e.Note = trimOrNil(e.Note)
	return nil
}

func (s *Service) prepareItems(ctx context.Context, e *Expense, keepInactive map[uuid.UUID]bool) error {
	if len(e.Items) == 0 {
		return fmt.Errorf("%w: at least one item is required", ErrInvalidInput)
	}

	checked := make(map[uuid.UUID]bool)
	for i := range e.Items {
		it := &e.Items[i]
		line := i + 1

		if err := normalizeItem(it, line); err != nil {
			return err
		}

		if !checked[it.CategoryID] {
			if err := s.checkCategory(ctx, it.CategoryID, line, keepInactive[it.CategoryID]); err != nil {
				return err
			}
			checked[it.CategoryID] = true
		}

		id, err := uuid.NewV7()
		if err != nil {
			return fmt.Errorf("generate uuid: %w", err)
		}
		it.ID = id
		it.LineNo = line
	}

	return nil
}

func normalizeItem(it *Item, line int) error {
	it.Description = strings.TrimSpace(it.Description)
	if it.Description == "" {
		return fmt.Errorf("%w: line %d: description is required", ErrInvalidInput, line)
	}
	if it.CategoryID == uuid.Nil {
		return fmt.Errorf("%w: line %d: category is required", ErrInvalidInput, line)
	}
	if it.Price.IsNegative() {
		return fmt.Errorf("%w: line %d: price must not be negative", ErrInvalidInput, line)
	}
	if !it.Quantity.IsPositive() {
		return fmt.Errorf("%w: line %d: quantity must be positive", ErrInvalidInput, line)
	}
	if it.Discount.IsNegative() {
		return fmt.Errorf("%w: line %d: discount must not be negative", ErrInvalidInput, line)
	}
	if it.Price.Mul(it.Quantity).LessThan(it.Discount) {
		return fmt.Errorf("%w: line %d: discount exceeds line total", ErrInvalidInput, line)
	}

	if (it.PeriodFrom == nil) != (it.PeriodTo == nil) {
		return fmt.Errorf("%w: line %d: period needs both start and end", ErrInvalidInput, line)
	}
	if it.PeriodFrom != nil {
		from := dateOnly(*it.PeriodFrom)
		to := dateOnly(*it.PeriodTo)
		if to.Before(from) {
			return fmt.Errorf("%w: line %d: period ends before it starts", ErrInvalidInput, line)
		}
		it.PeriodFrom = &from
		it.PeriodTo = &to
	}

	return nil
}

func (s *Service) checkCategory(ctx context.Context, id uuid.UUID, line int, allowInactive bool) error {
	cat, err := s.categories.GetCategory(ctx, id)
	if err != nil {
		if errors.Is(err, expensecategory.ErrCategoryNotFound) {
			return fmt.Errorf("%w: line %d: unknown category", ErrInvalidInput, line)
		}
		return fmt.Errorf("get category: %w", err)
	}
	if !cat.IsActive && !allowInactive {
		return fmt.Errorf("%w: line %d: category %q is inactive", ErrInvalidInput, line, cat.Name)
	}
	return nil
}

func dateOnly(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func trimOrNil(s *string) *string {
	if s == nil {
		return nil
	}
	v := strings.TrimSpace(*s)
	if v == "" {
		return nil
	}
	return &v
}

func (s *Service) invalidateSuggestions(ctx context.Context) {
	ctx = context.WithoutCancel(ctx)
	if err := s.cache.Invalidate(ctx); err != nil {
		log.Printf("WARN suggestions cache invalidate: %v", err)
	}
}
