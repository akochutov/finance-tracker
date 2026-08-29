package currency

import "time"

type Currency struct {
	Code          string    `json:"code"`
	Name          string    `json:"name"`
	Kind          string    `json:"kind"`
	DecimalPlaces int       `json:"decimal_places"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
