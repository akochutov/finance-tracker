package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/akochutov/finance-tracker/internal/api"
	"github.com/akochutov/finance-tracker/internal/config"
	"github.com/akochutov/finance-tracker/internal/currency"
	"github.com/akochutov/finance-tracker/internal/platform/postgres"
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

	// --- Repositories ---
	currencyRepo := currency.NewRepository(db)

	srv := &http.Server{
		Addr:           cfg.HTTPAddr,
		Handler:        api.New(db, currencyRepo),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Fatal(srv.ListenAndServe())
}
