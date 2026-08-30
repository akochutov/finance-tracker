package requisite

import (
	"time"

	"github.com/google/uuid"
)

type BankRequisite struct {
	ID                     uuid.UUID  `json:"id"`
	CompanyID              uuid.UUID  `json:"company_id"`
	BeneficiaryName        string     `json:"beneficiary_name"`
	AccountNumber          string     `json:"account_number"`
	BankName               string     `json:"bank_name"`
	BankSwift              string     `json:"bank_swift"`
	BankAddress            *string    `json:"bank_address"`
	CorrespondentBankName  *string    `json:"correspondent_bank_name"`
	CorrespondentBankSwift *string    `json:"correspondent_bank_swift"`
	IntermediaryBankName   *string    `json:"intermediary_bank_name"`
	IntermediaryBankSwift  *string    `json:"intermediary_bank_swift"`
	ValidFrom              time.Time  `json:"valid_from"`
	ValidTo                *time.Time `json:"valid_to"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
}

type CryptoRequisite struct {
	ID            uuid.UUID  `json:"id"`
	CompanyID     uuid.UUID  `json:"company_id"`
	Network       string     `json:"network"`
	WalletAddress string     `json:"wallet_address"`
	ValidFrom     time.Time  `json:"valid_from"`
	ValidTo       *time.Time `json:"valid_to"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}
