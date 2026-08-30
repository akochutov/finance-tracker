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
	"github.com/akochutov/finance-tracker/internal/platform/postgres"
	"github.com/akochutov/finance-tracker/internal/requisite"
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

	bankRequisiteRepo := requisite.NewRepository(db)
	bankRequisiteService := requisite.NewService(bankRequisiteRepo)

	srv := &http.Server{
		Addr:           cfg.HTTPAddr,
		Handler:        api.New(db, currencyService, companyService, bankRequisiteService),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Fatal(srv.ListenAndServe())
}
