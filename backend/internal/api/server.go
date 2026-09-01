package api

import (
	"context"
	"net/http"
	"time"

	"github.com/akochutov/finance-tracker/internal/company"
	"github.com/akochutov/finance-tracker/internal/currency"
	"github.com/akochutov/finance-tracker/internal/income"
	"github.com/akochutov/finance-tracker/internal/requisite"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	db               *pgxpool.Pool
	mux              *http.ServeMux
	currencies       *currency.Service
	companies        *company.Service
	bankRequisites   *requisite.BankService
	cryptoRequisites *requisite.CryptoService
	incomes          *income.Service
}

func New(
	db *pgxpool.Pool,
	currencies *currency.Service,
	companies *company.Service,
	bankRequisites *requisite.BankService,
	cryptoRequisites *requisite.CryptoService,
	incomes *income.Service,
) *Server {
	s := &Server{
		db:               db,
		mux:              http.NewServeMux(),
		currencies:       currencies,
		companies:        companies,
		bankRequisites:   bankRequisites,
		cryptoRequisites: cryptoRequisites,
		incomes:          incomes,
	}
	s.routes()
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
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
