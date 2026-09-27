package api

import (
	"github.com/akochutov/finance-tracker/internal/company"
	"github.com/akochutov/finance-tracker/internal/currency"
	"github.com/akochutov/finance-tracker/internal/dashboard"
	"github.com/akochutov/finance-tracker/internal/exchangerate"
	"github.com/akochutov/finance-tracker/internal/expense"
	"github.com/akochutov/finance-tracker/internal/expensecategory"
	"github.com/akochutov/finance-tracker/internal/income"
	"github.com/akochutov/finance-tracker/internal/ratefetch"
	"github.com/akochutov/finance-tracker/internal/ratesource"
	"github.com/akochutov/finance-tracker/internal/requisite"
	"github.com/akochutov/finance-tracker/internal/settings"
)

type Services struct {
	Currency        *currency.Service
	Company         *company.Service
	BankRequisite   *requisite.BankService
	CryptoRequisite *requisite.CryptoService
	Income          *income.Service
	Settings        *settings.Service
	ExchangeRate    *exchangerate.Service
	RateSource      *ratesource.Service
	RateFetch       *ratefetch.Service
	Dashboard       *dashboard.Service
	Backfiller      *ratefetch.Backfiller
	ExpenseCategory *expensecategory.Service
	Expense         *expense.Service
}
