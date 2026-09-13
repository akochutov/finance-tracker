package exchangerate

import (
	"time"

	"github.com/shopspring/decimal"
)

const (
	Pivot        = "USD"
	SourceManual = "manual"
)

type Rate struct {
	Currency  string          `json:"currency"`
	Source    string          `json:"source"`
	RateAt    time.Time       `json:"rate_at"`
	Rate      decimal.Decimal `json:"rate"`
	FetchedAt time.Time       `json:"fetched_at"`
}
