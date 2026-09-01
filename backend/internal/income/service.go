package income

import (
	"context"
	"errors"
	"fmt"

	"github.com/akochutov/finance-tracker/internal/company"
	"github.com/akochutov/finance-tracker/internal/currency"
	"github.com/akochutov/finance-tracker/internal/requisite"
	"github.com/google/uuid"
)

var ErrInvalidInput = errors.New("invalid income data")

type Service struct {
	repo       *Repository
	companies  *company.Service
	currencies *currency.Service
	bankReq    *requisite.BankService
	cryptoReq  *requisite.CryptoService
}

func NewService(
	repo *Repository,
	companies *company.Service,
	currencies *currency.Service,
	bankReq *requisite.BankService,
	cryptoReq *requisite.CryptoService,
) *Service {
	return &Service{
		repo:       repo,
		companies:  companies,
		currencies: currencies,
		bankReq:    bankReq,
		cryptoReq:  cryptoReq,
	}
}

func (s *Service) List(ctx context.Context) ([]Income, error) {
	return s.repo.List(ctx)
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (Income, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, in Income) (Income, error) {
	if in.OccurredAt.IsZero() {
		return Income{}, fmt.Errorf("%w: occurred_at is required", ErrInvalidInput)
	}
	in.OccurredAt = in.OccurredAt.UTC()

	if in.Amount.IsZero() || in.Amount.IsNegative() {
		return Income{}, fmt.Errorf("%w: amount must be positive", ErrInvalidInput)
	}

	if in.PaymentType != "bank" && in.PaymentType != "crypto" {
		return Income{}, fmt.Errorf("%w: payment type must be bank or crypto", ErrInvalidInput)
	}

	if _, err := s.companies.GetByID(ctx, in.PayerID); err != nil {
		if errors.Is(err, company.ErrNotFound) {
			return Income{}, fmt.Errorf("%w: payer not found", ErrInvalidInput)
		}
		return Income{}, err
	}
	if _, err := s.companies.GetByID(ctx, in.BeneficiaryID); err != nil {
		if errors.Is(err, company.ErrNotFound) {
			return Income{}, fmt.Errorf("%w: beneficiary not found", ErrInvalidInput)
		}
		return Income{}, err
	}

	if _, err := s.currencies.GetByCode(ctx, in.Currency); err != nil {
		if errors.Is(err, currency.ErrNotFound) {
			return Income{}, fmt.Errorf("%w: currency not found", ErrInvalidInput)
		}
		return Income{}, err
	}

	if err := s.validateRequisite(ctx, in.PaymentType, in.PayerRequisiteID, in.PayerID); err != nil {
		return Income{}, err
	}
	if err := s.validateRequisite(ctx, in.PaymentType, in.BeneficiaryRequisiteID, in.BeneficiaryID); err != nil {
		return Income{}, err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return Income{}, fmt.Errorf("generate uuid: %w", err)
	}
	in.ID = id
	in.IsActive = true

	return s.repo.Create(ctx, in)
}

func (s *Service) validateRequisite(ctx context.Context, paymentType string, requisiteID, companyID uuid.UUID) error {
	switch paymentType {
	case "bank":
		req, err := s.bankReq.GetByID(ctx, requisiteID)
		if err != nil {
			if errors.Is(err, requisite.ErrNotFound) {
				return fmt.Errorf("%w: bank requisite not found", ErrInvalidInput)
			}
			return err
		}
		if req.CompanyID != companyID {
			return fmt.Errorf("%w: requisite does not belong to the company", ErrInvalidInput)
		}
	case "crypto":
		req, err := s.cryptoReq.GetByID(ctx, requisiteID)
		if err != nil {
			if errors.Is(err, requisite.ErrNotFound) {
				return fmt.Errorf("%w: crypto requisite not found", ErrInvalidInput)
			}
			return err
		}
		if req.CompanyID != companyID {
			return fmt.Errorf("%w: requisite does not belong to the company", ErrInvalidInput)
		}
	}
	return nil
}
