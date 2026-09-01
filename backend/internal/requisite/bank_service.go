package requisite

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

var ErrInvalidInput = errors.New("invalid requisite data")

type BankService struct {
	repo *BankRepository
}

func NewBankService(repo *BankRepository) *BankService {
	return &BankService{repo: repo}
}

func (s *BankService) Create(ctx context.Context, br BankRequisite) (BankRequisite, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return BankRequisite{}, fmt.Errorf("generate uuid: %w", err)
	}
	br.ID = id

	if br.ValidFrom.IsZero() {
		br.ValidFrom = time.Now().UTC()
	}
	br.ValidTo = nil

	br.BeneficiaryName = strings.TrimSpace(br.BeneficiaryName)
	br.AccountNumber = strings.TrimSpace(br.AccountNumber)
	br.BankName = strings.TrimSpace(br.BankName)
	br.BankSwift = strings.TrimSpace(br.BankSwift)
	if br.BeneficiaryName == "" || br.AccountNumber == "" || br.BankName == "" || br.BankSwift == "" {
		return BankRequisite{}, fmt.Errorf("%w: required bank fields are empty", ErrInvalidInput)
	}

	return s.repo.Create(ctx, br)
}

func (s *BankService) ListByCompany(ctx context.Context, companyID uuid.UUID) ([]BankRequisite, error) {
	return s.repo.ListByCompany(ctx, companyID)
}

func (s *BankService) GetByID(ctx context.Context, id uuid.UUID) (BankRequisite, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *BankService) Close(ctx context.Context, id uuid.UUID, validTo time.Time) error {
	req, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if req.ValidTo != nil {
		return fmt.Errorf("%w: requisite already closed", ErrInvalidInput)
	}
	if !validTo.After(req.ValidFrom) {
		return fmt.Errorf("%w: valid_to must be after valid_from", ErrInvalidInput)
	}

	return s.repo.Close(ctx, id, validTo)
}
