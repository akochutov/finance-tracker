package settings

import "time"

type Settings struct {
	BaseCurrency string    `json:"base_currency"`
	UpdatedAt    time.Time `json:"updated_at"`
}
