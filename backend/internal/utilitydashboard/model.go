package utilitydashboard

import "github.com/shopspring/decimal"

type Dashboard struct {
	Currency string         `json:"currency"`
	Period   Period         `json:"period"`
	Previous Period         `json:"previous"`
	Months   []string       `json:"months"`
	Services []ServiceBlock `json:"services"`
	Coverage Coverage       `json:"coverage"`
}

type Period struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type ServiceBlock struct {
	Service   string             `json:"service"`
	Name      string             `json:"name"`
	Unit      string             `json:"unit"`
	Summary   Summary            `json:"summary"`
	Previous  Summary            `json:"previous"`
	Series    []Series           `json:"series"`
	Estimated []*decimal.Decimal `json:"estimated"`
	Paid      []decimal.Decimal  `json:"paid"`
}

type Summary struct {
	Consumption         decimal.Decimal `json:"consumption"`
	Zones               []ZoneAmount    `json:"zones"`
	Estimated           decimal.Decimal `json:"estimated"`
	Paid                decimal.Decimal `json:"paid"`
	MonthsWithoutTariff int             `json:"months_without_tariff"`
	TracksPayments      bool            `json:"tracks_payments"`
}

type ZoneAmount struct {
	Zone     string          `json:"zone"`
	Quantity decimal.Decimal `json:"quantity"`
}

type Series struct {
	Key         string             `json:"key"`
	Label       string             `json:"label"`
	Kind        string             `json:"kind"`
	Consumption []*decimal.Decimal `json:"consumption"`
	Cost        []*decimal.Decimal `json:"cost"`
}

type Coverage struct {
	MissingTariffs []string `json:"missing_tariffs"`
	MissingSplits  []string `json:"missing_splits"`
	MissingRates   []string `json:"missing_rates"`
}

const (
	KindMeter   = "meter"
	KindZone    = "zone"
	KindUnsplit = "unsplit"
)
