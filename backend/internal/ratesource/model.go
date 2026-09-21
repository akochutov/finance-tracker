package ratesource

import "time"

const (
	KindFiat   = "fiat"
	KindCrypto = "crypto"
)

type RateSource struct {
	Kind           string     `json:"kind"`
	Source         string     `json:"source"`
	URLTemplate    string     `json:"url_template"`
	PollInterval   int        `json:"poll_interval_seconds"`
	RequestTimeout int        `json:"request_timeout_seconds"`
	BackfillStart  *time.Time `json:"backfill_start"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func (s RateSource) PollIntervalDuration() time.Duration {
	return time.Duration(s.PollInterval) * time.Second
}

func (s RateSource) RequestTimeoutDuration() time.Duration {
	return time.Duration(s.RequestTimeout) * time.Second
}
