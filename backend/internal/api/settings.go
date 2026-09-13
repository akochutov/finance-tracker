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
	BaseCurrency string `json:"base_currency"`
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

		code := strings.ToUpper(strings.TrimSpace(req.BaseCurrency))
		if code == "" {
			writeError(w, http.StatusBadRequest, "base_currency is required")
			return
		}

		out, err := s.settings.SetBaseCurrency(r.Context(), code)
		if err != nil {
			switch {
			case errors.Is(err, settings.ErrNotFiat):
				writeError(w, http.StatusBadRequest, "base currency must be a fiat currency")
			case errors.Is(err, currency.ErrNotFound):
				writeError(w, http.StatusBadRequest, "currency not found")
			default:
				log.Printf("update settings: %v", err)
				writeError(w, http.StatusInternalServerError, "failed to update settings")
			}
			return
		}

		writeJSON(w, http.StatusOK, out)
	}
}
