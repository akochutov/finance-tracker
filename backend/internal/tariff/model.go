package tariff

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Tariff struct {
	ID        uuid.UUID       `json:"id"`
	Service   string          `json:"service"`
	Zone      string          `json:"zone"`
	ValidFrom time.Time       `json:"valid_from"`
	Price     decimal.Decimal `json:"price"`
	Currency  string          `json:"currency"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}
