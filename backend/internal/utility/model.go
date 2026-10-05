package utility

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

const (
	ZoneSingle = "single"
	ZoneDay    = "day"
	ZoneNight  = "night"
)

type ServiceType struct {
	Code string `json:"code"`
	Name string `json:"name"`
	Unit string `json:"unit"`
}

type Address struct {
	ID        uuid.UUID `json:"id"`
	Address   string    `json:"address"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Account struct {
	ID        uuid.UUID `json:"id"`
	AddressID uuid.UUID `json:"address_id"`
	Service   string    `json:"service"`
	Number    string    `json:"number"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Meter struct {
	ID          uuid.UUID  `json:"id"`
	AccountID   uuid.UUID  `json:"account_id"`
	Serial      string     `json:"serial"`
	InstalledOn time.Time  `json:"installed_on"`
	RemovedOn   *time.Time `json:"removed_on"`
	Registers   []Register `json:"registers"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type Register struct {
	ID      uuid.UUID `json:"id"`
	MeterID uuid.UUID `json:"meter_id"`
	Zone    string    `json:"zone"`
}

type Reading struct {
	ID         uuid.UUID       `json:"id"`
	RegisterID uuid.UUID       `json:"register_id"`
	TakenOn    time.Time       `json:"taken_on"`
	Value      decimal.Decimal `json:"value"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}
