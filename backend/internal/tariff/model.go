package tariff

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

const (
	ModeWhole       = "whole"
	ModeProgressive = "progressive"
)

type Tariff struct {
	ID        uuid.UUID `json:"id"`
	Service   string    `json:"service"`
	Zone      string    `json:"zone"`
	ValidFrom time.Time `json:"valid_from"`
	Currency  string    `json:"currency"`
	TierMode  string    `json:"tier_mode"`
	Tiers     []Tier    `json:"tiers"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Tier struct {
	UpTo  *decimal.Decimal `json:"up_to"`
	Price decimal.Decimal  `json:"price"`
}
