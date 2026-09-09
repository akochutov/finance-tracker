package api

import (
	"github.com/akochutov/finance-tracker/internal/company"
	"github.com/akochutov/finance-tracker/internal/currency"
	"github.com/akochutov/finance-tracker/internal/income"
	"github.com/akochutov/finance-tracker/internal/requisite"
)

type Services struct {
	Currency        *currency.Service
	Company         *company.Service
	BankRequisite   *requisite.BankService
	CryptoRequisite *requisite.CryptoService
	Income          *income.Service
}
