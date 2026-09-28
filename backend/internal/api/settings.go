package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/akochutov/finance-tracker/internal/currency"
	"github.com/akochutov/finance-tracker/internal/settings"
)

type updateSettingsRequest struct {
	BaseCurrency        *string `json:"base_currency"`
	ExpenseBaseCurrency *string `json:"expense_base_currency"`
}

func (s *Server) handleGetSettings() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		out, err := s.settings.Get(r.Context())
		if err != nil {
			log.Printf("get settings: %v", err)
			writeError(w, http.StatusInternalServerError, "failed to get setting")
			return
		}

		writeJSON(w, http.StatusOK, out)
	}
}

func (s *Server) handleUpdateSettings() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req updateSettingsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		if req.BaseCurrency == nil && req.ExpenseBaseCurrency == nil {
			writeError(w, http.StatusBadRequest, "base_currency or expense_base_currency is required")
			return
		}

		var out settings.Settings

		if req.BaseCurrency != nil {
			code := normalizeCode(*req.BaseCurrency)
			if code == "" {
				writeError(w, http.StatusBadRequest, "base_currency must not be empty")
				return
			}
			res, err := s.settings.SetBaseCurrency(r.Context(), code)
			if err != nil {
				writeSettingsError(w, err)
				return
			}
			out = res
		}

		if req.ExpenseBaseCurrency != nil {
			code := normalizeCode(*req.ExpenseBaseCurrency)
			if code == "" {
				writeError(w, http.StatusBadRequest, "expense_base_currency must not be empty")
				return
			}
			res, err := s.settings.SetExpenseBaseCurrency(r.Context(), code)
			if err != nil {
				writeSettingsError(w, err)
				return
			}
			out = res
		}

		writeJSON(w, http.StatusOK, out)
	}
}

func normalizeCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}

func writeSettingsError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, settings.ErrNotFiat):
		writeError(w, http.StatusBadRequest, "dashboard currency must be a fiat currency")
	case errors.Is(err, currency.ErrNotFound):
		writeError(w, http.StatusBadRequest, "currency not found")
	default:
		log.Printf("update settings: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to update settings")
	}
}
