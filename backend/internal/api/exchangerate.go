package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/akochutov/finance-tracker/internal/exchangerate"
	"github.com/shopspring/decimal"
)

type listExchangeRatesResponse struct {
	ExchangeRates []exchangerate.Rate `json:"exchange_rates"`
}

type createExchangeRateRequest struct {
	Currency string          `json:"currency"`
	RateAt   time.Time       `json:"rate_at"`
	Rate     decimal.Decimal `json:"rate"`
}

type convertResponse struct {
	Amount decimal.Decimal `json:"amount"`
	From   string          `json:"from"`
	To     string          `json:"to"`
	At     time.Time       `json:"at"`
	Result decimal.Decimal `json:"result"`
}

func (s *Server) handleListExchangeRates() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("currency")))

		rates, err := s.exchangeRates.List(r.Context(), code)
		if err != nil {
			log.Printf("list exchange rates: %v", err)
			writeError(w, http.StatusInternalServerError, "failed to list exchange rates")
			return
		}

		writeJSON(w, http.StatusOK, listExchangeRatesResponse{ExchangeRates: rates})
	}
}

func (s *Server) handleCreateExchangeRate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req createExchangeRateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}

		code := strings.ToUpper(strings.TrimSpace(req.Currency))
		if code == "" {
			writeError(w, http.StatusBadRequest, "currency is required")
			return
		}
		if req.RateAt.IsZero() {
			writeError(w, http.StatusBadRequest, "rate_at is required")
			return
		}
		if req.Rate.IsZero() || req.Rate.IsNegative() {
			writeError(w, http.StatusBadRequest, "rate must be positive")
			return
		}

		out, err := s.exchangeRates.RecordManual(r.Context(), code, req.RateAt, req.Rate)
		if err != nil {
			switch {
			case errors.Is(err, exchangerate.ErrPivotRate):
				writeError(w, http.StatusBadRequest, "cannot store a rate for the pivot currency (USD)")
			default:
				log.Printf("record manual rate: %v", err)
				writeError(w, http.StatusInternalServerError, "failed to create exchange rate")
			}
			return
		}

		writeJSON(w, http.StatusCreated, out)
	}
}

func (s *Server) handleConvert() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		amount, err := decimal.NewFromString(q.Get("amount"))
		if err != nil {
			writeError(w, http.StatusBadRequest, "amount must be a valid number")
			return
		}

		from := strings.ToUpper(strings.TrimSpace(q.Get("from")))
		to := strings.ToUpper(strings.TrimSpace(q.Get("to")))
		if from == "" || to == "" {
			writeError(w, http.StatusBadRequest, "from and to are required")
			return
		}

		at := time.Now()
		if raw := q.Get("at"); raw != "" {
			at, err = time.Parse(time.RFC3339, raw)
			if err != nil {
				writeError(w, http.StatusBadRequest, "at must be RFC3339")
				return
			}
		}

		result, err := s.exchangeRates.Convert(r.Context(), amount, from, to, at)
		if err != nil {
			switch {
			case errors.Is(err, exchangerate.ErrRateNotFound):
				writeError(w, http.StatusNotFound, "no rate available for the given date")
			default:
				log.Printf("convert: %v", err)
				writeError(w, http.StatusInternalServerError, "failed to convert")
			}
			return
		}

		writeJSON(w, http.StatusOK, convertResponse{
			Amount: amount,
			From:   from,
			To:     to,
			At:     at,
			Result: result,
		})
	}
}
