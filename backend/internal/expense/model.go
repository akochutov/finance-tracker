package expense

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Item struct {
	ID          uuid.UUID       `json:"id"`
	ExpenseID   uuid.UUID       `json:"expense_id"`
	LineNo      int             `json:"line_no"`
	Description string          `json:"description"`
	CategoryID  uuid.UUID       `json:"category_id"`
	Price       decimal.Decimal `json:"price"`
	Quantity    decimal.Decimal `json:"quantity"`
	Discount    decimal.Decimal `json:"discount"`
	Amount      decimal.Decimal `json:"amount"`
	PeriodFrom  *time.Time      `json:"period_from"`
	PeriodTo    *time.Time      `json:"period_to"`
}

type Expense struct {
	ID          uuid.UUID `json:"id"`
	OccurredOn  time.Time `json:"occurred_on"`
	Currency    string    `json:"currency"`
	PaymentType *string   `json:"payment_type"`
	Note        *string   `json:"note"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Items       []Item    `json:"items"`
}

type Suggestion struct {
	Description string     `json:"description"`
	CategoryID  *uuid.UUID `json:"category_id"`
	Uses        int        `json:"uses"`
}
