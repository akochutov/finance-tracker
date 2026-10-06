package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/akochutov/finance-tracker/internal/api"
	"github.com/akochutov/finance-tracker/internal/company"
	"github.com/akochutov/finance-tracker/internal/config"
	"github.com/akochutov/finance-tracker/internal/currency"
	"github.com/akochutov/finance-tracker/internal/exchangerate"
	"github.com/akochutov/finance-tracker/internal/expense"
	"github.com/akochutov/finance-tracker/internal/expensecategory"
	"github.com/akochutov/finance-tracker/internal/expensedashboard"
	"github.com/akochutov/finance-tracker/internal/income"
	"github.com/akochutov/finance-tracker/internal/incomedashboard"
	"github.com/akochutov/finance-tracker/internal/platform/postgres"
	"github.com/akochutov/finance-tracker/internal/platform/redis"
	"github.com/akochutov/finance-tracker/internal/ratefetch"
	"github.com/akochutov/finance-tracker/internal/ratesource"
	"github.com/akochutov/finance-tracker/internal/requisite"
	"github.com/akochutov/finance-tracker/internal/scheduler"
	"github.com/akochutov/finance-tracker/internal/settings"
	"github.com/akochutov/finance-tracker/internal/tariff"
	"github.com/akochutov/finance-tracker/internal/utility"
	"github.com/akochutov/finance-tracker/internal/utilitydashboard"
)

func main() {
	// --- Config ---
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("loading config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// --- Cache connect ---
	rdb, err := redis.New(cfg.RedisURL)
	if err != nil {
		log.Fatalf("redis: %v", err)
	}
	defer rdb.Close()

	pingCtx, pingCancel := context.WithTimeout(ctx, 2*time.Second)
	err = rdb.Ping(pingCtx).Err()
	pingCancel()
	if err != nil {
		log.Printf("WARN redis unavailable, running without cache: %v", err)
	} else {
		log.Println("connected to redis")
	}

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

	incomeDashboardService := incomedashboard.NewService(
		incomeService, companyService, exchangeRateService, settingsService,
	)

	backfiller := ratefetch.NewBackfiller(
		currencyService, rateSourceService, rateRegistry, exchangeRateService,
	)

	expenseGroupRepo := expensecategory.NewGroupRepository(db)
	expenseCategoryRepo := expensecategory.NewCategoryRepository(db)
	expenseCategoryService := expensecategory.NewService(expenseGroupRepo, expenseCategoryRepo)

	expenseRepo := expense.NewRepository(db)
	suggestionCache := expense.NewRedisSuggestionCache(rdb, time.Hour)
	expenseService := expense.NewService(expenseRepo, expenseCategoryService, suggestionCache)

	expenseDashboardRepo := expensedashboard.NewRepository(db)
	expenseDashboardService := expensedashboard.NewService(
		expenseDashboardRepo, exchangeRateService, settingsService, incomeService,
	)

	serviceTypeRepo := utility.NewServiceTypeRepository(db)
	addressRepo := utility.NewAddressRepository(db)
	accountRepo := utility.NewAccountRepository(db)
	meterRepo := utility.NewMeterRepository(db)
	readingRepo := utility.NewReadingRepository(db)
	zoneUsageRepo := utility.NewZoneUsageRepository(db)
	utilityService := utility.NewService(serviceTypeRepo, addressRepo, accountRepo, meterRepo, readingRepo, zoneUsageRepo)

	tariffRepo := tariff.NewRepository(db)
	tariffService := tariff.NewService(tariffRepo)

	utilityDashboardRepo := utilitydashboard.NewRepository(db)
	utilityDashboardService := utilitydashboard.NewService(utilityDashboardRepo, tariffService, exchangeRateService, settingsService)

	sched := scheduler.New(rateFetchService, rateSourceService)
	sched.Start(ctx)

	srv := &http.Server{
		Addr: cfg.HTTPAddr,
		Handler: api.New(db, api.Services{
			Currency:         currencyService,
			Company:          companyService,
			BankRequisite:    bankRequisiteService,
			CryptoRequisite:  cryptoRequisiteService,
			Income:           incomeService,
			Settings:         settingsService,
			ExchangeRate:     exchangeRateService,
			RateSource:       rateSourceService,
			RateFetch:        rateFetchService,
			IncomeDashboard:  incomeDashboardService,
			Backfiller:       backfiller,
			ExpenseCategory:  expenseCategoryService,
			Expense:          expenseService,
			ExpenseDashboard: expenseDashboardService,
			Utility:          utilityService,
			Tariff:           tariffService,
			UtilityDashboard: utilityDashboardService,
		}),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		log.Printf("listening on %s", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("http server: %v", err)
		}
	}()

	<-ctx.Done()
	stop()
	log.Println("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("http shutdown: %v", err)
	}

	sched.Wait()
	log.Println("stopped")
}
