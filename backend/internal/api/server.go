package api

import (
	"context"
	"net/http"
	"time"

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
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	db                *pgxpool.Pool
	mux               *http.ServeMux
	handler           http.Handler
	currencies        *currency.Service
	companies         *company.Service
	bankRequisites    *requisite.BankService
	cryptoRequisites  *requisite.CryptoService
	incomes           *income.Service
	settings          *settings.Service
	exchangeRates     *exchangerate.Service
	rateSources       *ratesource.Service
	rateFetch         *ratefetch.Service
	incomeDashboard   *incomedashboard.Service
	backfiller        *ratefetch.Backfiller
	expenseCategories *expensecategory.Service
	expenses          *expense.Service
	expenseDashboard  *expensedashboard.Service
	utilities         *utility.Service
	tariffs           *tariff.Service
	utilityDashboard  *utilitydashboard.Service
}

func New(db *pgxpool.Pool, services Services) *Server {
	s := &Server{
		db:                db,
		mux:               http.NewServeMux(),
		currencies:        services.Currency,
		companies:         services.Company,
		bankRequisites:    services.BankRequisite,
		cryptoRequisites:  services.CryptoRequisite,
		incomes:           services.Income,
		settings:          services.Settings,
		exchangeRates:     services.ExchangeRate,
		rateSources:       services.RateSource,
		rateFetch:         services.RateFetch,
		incomeDashboard:   services.IncomeDashboard,
		backfiller:        services.Backfiller,
		expenseCategories: services.ExpenseCategory,
		expenses:          services.Expense,
		expenseDashboard:  services.ExpenseDashboard,
		utilities:         services.Utility,
		tariffs:           services.Tariff,
		utilityDashboard:  services.UtilityDashboard,
	}
	s.routes()
	s.handler = corsMiddleware(s.mux)
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.handler.ServeHTTP(w, r)
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /healthz", s.handleHealthz())

	s.mux.HandleFunc("GET /api/currencies", s.handleListCurrencies())
	s.mux.HandleFunc("GET /api/currencies/{code}", s.handleGetCurrency())
	s.mux.HandleFunc("POST /api/currencies", s.handleCreateCurrency())
	s.mux.HandleFunc("PUT /api/currencies/{code}", s.handleUpdateCurrency())
	s.mux.HandleFunc("DELETE /api/currencies/{code}", s.handleDeactivateCurrency())

	s.mux.HandleFunc("GET /api/companies", s.handleListCompanies())
	s.mux.HandleFunc("GET /api/companies/{id}", s.handleGetCompany())
	s.mux.HandleFunc("POST /api/companies", s.handleCreateCompany())
	s.mux.HandleFunc("PUT /api/companies/{id}", s.handleUpdateCompany())
	s.mux.HandleFunc("DELETE /api/companies/{id}", s.handleDeactivateCompany())

	s.mux.HandleFunc("GET /api/companies/{id}/bank-requisites", s.handleListBankRequisites())
	s.mux.HandleFunc("POST /api/companies/{id}/bank-requisites", s.handleCreateBankRequisite())
	s.mux.HandleFunc("POST /api/companies/{id}/bank-requisites/{rid}/close", s.handleCloseBankRequisite())

	s.mux.HandleFunc("GET /api/companies/{id}/crypto-requisites", s.handleListCryptoRequisites())
	s.mux.HandleFunc("POST /api/companies/{id}/crypto-requisites", s.handleCreateCryptoRequisite())
	s.mux.HandleFunc("POST /api/companies/{id}/crypto-requisites/{rid}/close", s.handleCloseCryptoRequisite())

	s.mux.HandleFunc("GET /api/incomes", s.handleListIncomes())
	s.mux.HandleFunc("GET /api/incomes/{id}", s.handleGetIncome())
	s.mux.HandleFunc("POST /api/incomes", s.handleCreateIncome())

	s.mux.HandleFunc("GET /api/settings", s.handleGetSettings())
	s.mux.HandleFunc("PUT /api/settings", s.handleUpdateSettings())

	s.mux.HandleFunc("GET /api/exchange-rates", s.handleListExchangeRates())
	s.mux.HandleFunc("POST /api/exchange-rates", s.handleCreateExchangeRate())
	s.mux.HandleFunc("GET /api/convert", s.handleConvert())

	s.mux.HandleFunc("GET /api/rate-sources", s.handleListRateSources())
	s.mux.HandleFunc("PUT /api/rate-sources/{kind}", s.handleSaveRateSource())
	s.mux.HandleFunc("POST /api/rate-sources/{kind}/fetch", s.handleFetchRates())
	s.mux.HandleFunc("GET /api/rate-providers", s.handleListRateProviders())

	s.mux.HandleFunc("GET /api/dashboard/incomes", s.handleGetIncomeDashboard())
	s.mux.HandleFunc("GET /api/dashboard/expenses", s.handleGetExpenseDashboard())
	s.mux.HandleFunc("GET /api/dashboard/utilities", s.handleGetUtilityDashboard())

	s.mux.HandleFunc("POST /api/rate-sources/{kind}/backfill", s.handleStartBackfill())
	s.mux.HandleFunc("GET /api/rate-sources/{kind}/backfill", s.handleBackfillStatus())

	s.mux.HandleFunc("GET /api/expense-groups", s.handleListExpenseGroups())
	s.mux.HandleFunc("POST /api/expense-groups", s.handleCreateExpenseGroup())
	s.mux.HandleFunc("PUT /api/expense-groups/{id}", s.handleUpdateExpenseGroup())
	s.mux.HandleFunc("DELETE /api/expense-groups/{id}", s.handleDeactivateExpenseGroup())
	s.mux.HandleFunc("POST /api/expense-groups/{id}/activate", s.handleActivateExpenseGroup())

	s.mux.HandleFunc("POST /api/expense-categories", s.handleCreateExpenseCategory())
	s.mux.HandleFunc("PUT /api/expense-categories/{id}", s.handleUpdateExpenseCategory())
	s.mux.HandleFunc("DELETE /api/expense-categories/{id}", s.handleDeactivateExpenseCategory())
	s.mux.HandleFunc("POST /api/expense-categories/{id}/activate", s.handleActivateExpenseCategory())
	s.mux.HandleFunc("PUT /api/expense-categories/{id}/dashboard", s.handleSetExpenseCategoryDashboard())

	s.mux.HandleFunc("GET /api/expenses", s.handleListExpenses())
	s.mux.HandleFunc("GET /api/expenses/{id}", s.handleGetExpense())
	s.mux.HandleFunc("POST /api/expenses", s.handleCreateExpense())
	s.mux.HandleFunc("PUT /api/expenses/{id}", s.handleUpdateExpense())
	s.mux.HandleFunc("DELETE /api/expenses/{id}", s.handleDeleteExpense())

	s.mux.HandleFunc("GET /api/expenses/suggestions", s.handleListExpenseSuggestions())

	s.mux.HandleFunc("GET /api/service-types", s.handleListServiceTypes())

	s.mux.HandleFunc("GET /api/addresses", s.handleListAddresses())
	s.mux.HandleFunc("POST /api/addresses", s.handleCreateAddress())
	s.mux.HandleFunc("PUT /api/addresses/{id}", s.handleUpdateAddress())
	s.mux.HandleFunc("DELETE /api/addresses/{id}", s.handleDeactivateAddress())
	s.mux.HandleFunc("POST /api/addresses/{id}/activate", s.handleActivateAddress())

	s.mux.HandleFunc("GET /api/utility-accounts", s.handleListUtilityAccounts())
	s.mux.HandleFunc("POST /api/utility-accounts", s.handleCreateUtilityAccount())
	s.mux.HandleFunc("PUT /api/utility-accounts/{id}", s.handleUpdateUtilityAccount())
	s.mux.HandleFunc("DELETE /api/utility-accounts/{id}", s.handleDeactivateUtilityAccount())
	s.mux.HandleFunc("POST /api/utility-accounts/{id}/activate", s.handleActivateUtilityAccount())

	s.mux.HandleFunc("GET /api/utility-accounts/{id}/zone-usage", s.handleListZoneUsage())
	s.mux.HandleFunc("PUT /api/utility-accounts/{id}/zone-usage", s.handleSetZoneUsage())
	s.mux.HandleFunc("DELETE /api/utility-accounts/{id}/zone-usage/{month}", s.handleDeleteZoneUsage())

	s.mux.HandleFunc("GET /api/meters", s.handleListMeters())
	s.mux.HandleFunc("POST /api/meters", s.handleCreateMeter())
	s.mux.HandleFunc("PUT /api/meters/{id}", s.handleUpdateMeter())
	s.mux.HandleFunc("DELETE /api/meters/{id}", s.handleDeleteMeter())

	s.mux.HandleFunc("POST /api/readings", s.handleCreateReadings())
	s.mux.HandleFunc("GET /api/readings/latest", s.handleLatestReadings())
	s.mux.HandleFunc("GET /api/meters/{id}/readings", s.handleListReadings())
	s.mux.HandleFunc("PUT /api/meters/{id}/readings/{rid}", s.handleUpdateReading())
	s.mux.HandleFunc("DELETE /api/meters/{id}/readings/{rid}", s.handleDeleteReading())

	s.mux.HandleFunc("GET /api/tariffs", s.handleListTariffs())
	s.mux.HandleFunc("POST /api/tariffs", s.handleCreateTariff())
	s.mux.HandleFunc("PUT /api/tariffs/{id}", s.handleUpdateTariff())
	s.mux.HandleFunc("DELETE /api/tariffs/{id}", s.handleDeleteTariff())
}

func (s *Server) handleHealthz() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		w.Header().Set("Content-Type", "application/json")
		if err := s.db.Ping(ctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"status":"unavailable"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}
}
