package income

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Income struct {
	ID                     uuid.UUID       `json:"id"`
	PayerID                uuid.UUID       `json:"payer_id"`
	BeneficiaryID          uuid.UUID       `json:"beneficiary_id"`
	Amount                 decimal.Decimal `json:"amount"`
	Currency               string          `json:"currency"`
	OccurredAt             time.Time       `json:"occurred_at"`
	PaymentType            string          `json:"payment_type"`
	PayerRequisiteID       uuid.UUID       `json:"payer_requisite_id"`
	BeneficiaryRequisiteID uuid.UUID       `json:"beneficiary_requisite_id"`
	Note                   *string         `json:"note"`
	IsActive               bool            `json:"is_active"`
	CreatedAt              time.Time       `json:"created_at"`
	UpdatedAt              time.Time       `json:"updated_at"`
}
