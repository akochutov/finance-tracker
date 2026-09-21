package exchangerate

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

type FetchedRate struct {
	Rate   decimal.Decimal
	RateAt time.Time
}

type RateProvider interface {
	FetchRate(ctx context.Context, currency string, at time.Time) (FetchedRate, error)
}
