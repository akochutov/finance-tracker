package settings

import "time"

type Settings struct {
	BaseCurrency        string    `json:"base_currency"`
	ExpenseBaseCurrency string    `json:"expense_base_currency"`
	UpdatedAt           time.Time `json:"updated_at"`
}
