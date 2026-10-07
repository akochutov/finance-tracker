package api

import (
	"github.com/akochutov/finance-tracker/internal/company"
	"github.com/akochutov/finance-tracker/internal/currency"
	"github.com/akochutov/finance-tracker/internal/exchangerate"
	"github.com/akochutov/finance-tracker/internal/expense"
	"github.com/akochutov/finance-tracker/internal/expensecategory"
	"github.com/akochutov/finance-tracker/internal/expensedashboard"
	"github.com/akochutov/finance-tracker/internal/income"
	"github.com/akochutov/finance-tracker/internal/incomedashboard"
	"github.com/akochutov/finance-tracker/internal/ratefetch"
	"github.com/akochutov/finance-tracker/internal/ratesource"
	"github.com/akochutov/finance-tracker/internal/requisite"
	"github.com/akochutov/finance-tracker/internal/settings"
	"github.com/akochutov/finance-tracker/internal/tariff"
	"github.com/akochutov/finance-tracker/internal/utility"
	"github.com/akochutov/finance-tracker/internal/utilitydashboard"
)

type Services struct {
	Currency         *currency.Service
	Company          *company.Service
	BankRequisite    *requisite.BankService
	CryptoRequisite  *requisite.CryptoService
	Income           *income.Service
	Settings         *settings.Service
	ExchangeRate     *exchangerate.Service
	RateSource       *ratesource.Service
	RateFetch        *ratefetch.Service
	IncomeDashboard  *incomedashboard.Service
	Backfiller       *ratefetch.Backfiller
	ExpenseCategory  *expensecategory.Service
	Expense          *expense.Service
	ExpenseDashboard *expensedashboard.Service
	Utility          *utility.Service
	Tariff           *tariff.Service
	UtilityDashboard *utilitydashboard.Service
}
