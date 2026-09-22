package dashboard

import "github.com/shopspring/decimal"

type Dashboard struct {
	Currency    string           `json:"currency"`
	TotalIncome decimal.Decimal  `json:"total_income"`
	Records     RecordStats      `json:"records"`
	Period      Period           `json:"period"`
	ByMonth     []MonthBucket    `json:"by_month"`
	ByCurrency  []CurrencyBucket `json:"by_currency"`
	ByCompany   []CompanyBucket  `json:"by_company"`
	Averages    Averages         `json:"averages"`
	Periods     PeriodTotals     `json:"periods"`
}

type RecordStats struct {
	Total       int `json:"total"`
	Converted   int `json:"converted"`
	Unconverted int `json:"unconverted"`
}

type Period struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type MonthBucket struct {
	Month string          `json:"month"`
	Total decimal.Decimal `json:"total"`
}

type CurrencyBucket struct {
	Currency string          `json:"currency"`
	Total    decimal.Decimal `json:"total"`
}

type CompanyBucket struct {
	CompanyID string          `json:"company_id"`
	Name      string          `json:"name"`
	Total     decimal.Decimal `json:"total"`
}

type Averages struct {
	MonthlyAllTime  decimal.Decimal `json:"monthly_all_time"`
	MonthlyThisYear decimal.Decimal `json:"monthly_this_year"`
}

type PeriodTotals struct {
	ThisYear  decimal.Decimal `json:"this_year"`
	LastMonth decimal.Decimal `json:"last_month"`
	ThisMonth decimal.Decimal `json:"this_month"`
}
