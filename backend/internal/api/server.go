package api

import (
	"context"
	"net/http"
	"time"

	"github.com/akochutov/finance-tracker/internal/currency"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	db         *pgxpool.Pool
	mux        *http.ServeMux
	currencies *currency.Service
}

func New(db *pgxpool.Pool, currencies *currency.Service) *Server {
	s := &Server{db: db, mux: http.NewServeMux(), currencies: currencies}
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
