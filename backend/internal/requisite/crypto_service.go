package requisite

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type CryptoService struct {
	repo *CryptoRepository
}

func NewCryptoService(repo *CryptoRepository) *CryptoService {
	return &CryptoService{repo: repo}
}

func (s *CryptoService) Create(ctx context.Context, cr CryptoRequisite) (CryptoRequisite, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return CryptoRequisite{}, fmt.Errorf("generate uuid: %w", err)
	}
	cr.ID = id

	if cr.ValidFrom.IsZero() {
		cr.ValidFrom = time.Now().UTC()
	}
	cr.ValidTo = nil

	cr.Network = strings.TrimSpace(cr.Network)
	cr.WalletAddress = strings.TrimSpace(cr.WalletAddress)
	if cr.Network == "" || cr.WalletAddress == "" {
		return CryptoRequisite{}, fmt.Errorf("%w: network and wallet_address are required", ErrInvalidInput)
	}

	return s.repo.Create(ctx, cr)
}

func (s *CryptoService) ListByCompany(ctx context.Context, companyID uuid.UUID) ([]CryptoRequisite, error) {
	return s.repo.ListByCompany(ctx, companyID)
}

func (s *CryptoService) Close(ctx context.Context, id uuid.UUID, validTo time.Time) error {
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
