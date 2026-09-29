package expensedashboard

import "github.com/shopspring/decimal"

type Dashboard struct {
	Currency string       `json:"currency"`
	Period   Period       `json:"period"`
	Previous Period       `json:"previous"`
	Summary  Summary      `json:"summary"`
	Groups   []GroupTotal `json:"groups"`
	Trend    []MonthTrend `json:"trend"`
	Largest  []Receipt    `json:"largest"`
	Coverage Coverage     `json:"coverage"`
}

type Period struct {
	From string `json:"from"`
	To   string `json:"to"`
	Days int    `json:"days"`
}

type Summary struct {
	Total                decimal.Decimal  `json:"total"`
	PreviousTotal        decimal.Decimal  `json:"previous_total"`
	DailyAverage         decimal.Decimal  `json:"daily_average"`
	PreviousDailyAverage decimal.Decimal  `json:"previous_daily_average"`
	Receipts             int              `json:"receipts"`
	Fixed                decimal.Decimal  `json:"fixed"`
	Variable             decimal.Decimal  `json:"variable"`
	Income               decimal.Decimal  `json:"income"`
	SavingsRate          *decimal.Decimal `json:"savings_rate"`
	PreviousSavingsRate  *decimal.Decimal `json:"previous_savings_rate"`
}

type GroupTotal struct {
	GroupID    string          `json:"group_id"`
	Name       string          `json:"name"`
	Total      decimal.Decimal `json:"total"`
	Previous   decimal.Decimal `json:"previous"`
	Categories []CategoryTotal `json:"categories"`
}

type CategoryTotal struct {
	CategoryID string          `json:"category_id"`
	Name       string          `json:"name"`
	Total      decimal.Decimal `json:"total"`
}

type MonthTrend struct {
	Month    string          `json:"month"`
	Total    decimal.Decimal `json:"total"`
	Fixed    decimal.Decimal `json:"fixed"`
	Variable decimal.Decimal `json:"variable"`
	Income   decimal.Decimal `json:"income"`
	ByGroup  []GroupAmount   `json:"by_group"`
}

type GroupAmount struct {
	GroupID string          `json:"group_id"`
	Total   decimal.Decimal `json:"total"`
}

type Receipt struct {
	ExpenseID  string          `json:"expense_id"`
	OccurredOn string          `json:"occurred_on"`
	Total      decimal.Decimal `json:"total"`
	Group      string          `json:"group"`
	Summary    string          `json:"summary"`
	Lines      int             `json:"lines"`
}

type Coverage struct {
	UnconvertedDays    int `json:"unconverted_days"`
	UnconvertedIncomes int `json:"unconverted_incomes"`
}
