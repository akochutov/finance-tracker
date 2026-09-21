package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/akochutov/finance-tracker/internal/api"
	"github.com/akochutov/finance-tracker/internal/company"
	"github.com/akochutov/finance-tracker/internal/config"
	"github.com/akochutov/finance-tracker/internal/currency"
	"github.com/akochutov/finance-tracker/internal/exchangerate"
	"github.com/akochutov/finance-tracker/internal/income"
	"github.com/akochutov/finance-tracker/internal/platform/postgres"
	"github.com/akochutov/finance-tracker/internal/ratefetch"
	"github.com/akochutov/finance-tracker/internal/ratesource"
	"github.com/akochutov/finance-tracker/internal/requisite"
	"github.com/akochutov/finance-tracker/internal/settings"
)

func main() {
	// --- Config ---
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("loading config: %v", err)
	}

	ctx := context.Background()

	// --- Database connect ---
	db, err := postgres.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connecting to postgres: %v", err)
	}
	defer db.Close()
	log.Println("connected to database")

	// --- Repositories & Services ---
	currencyRepo := currency.NewRepository(db)
	currencyService := currency.NewService(currencyRepo)

	companyRepo := company.NewRepository(db)
	companyService := company.NewService(companyRepo)

	bankRequisiteRepo := requisite.NewBankRepository(db)
	bankRequisiteService := requisite.NewBankService(bankRequisiteRepo)
	cryptoRequisiteRepo := requisite.NewCryptoRepository(db)
	cryptoRequisiteService := requisite.NewCryptoService(cryptoRequisiteRepo)

	incomeRepo := income.NewRepository(db)
	incomeService := income.NewService(incomeRepo, companyService, currencyService, bankRequisiteService, cryptoRequisiteService)

	settingsRepo := settings.NewRepository(db)
	settingsService := settings.NewService(settingsRepo, currencyService)

	exchangeRateRepo := exchangerate.NewRepository(db)
	exchangeRateService := exchangerate.NewService(exchangeRateRepo)

	rateSourceRepo := ratesource.NewRepository(db)
	rateSourceService := ratesource.NewService(rateSourceRepo)

	rateRegistry := exchangerate.NewRegistry()

	rateFetchService := ratefetch.NewService(
		currencyService, rateSourceService, rateRegistry, exchangeRateService,
	)

	srv := &http.Server{
		Addr: cfg.HTTPAddr,
		Handler: api.New(db, api.Services{
			Currency:        currencyService,
			Company:         companyService,
			BankRequisite:   bankRequisiteService,
			CryptoRequisite: cryptoRequisiteService,
			Income:          incomeService,
			Settings:        settingsService,
			ExchangeRate:    exchangeRateService,
			RateSource:      rateSourceService,
			RateFetch:       rateFetchService,
		}),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Fatal(srv.ListenAndServe())
}
