package api

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/akochutov/finance-tracker/internal/currency"
)

type listCurrenciesResponse struct {
	Currencies []currency.Currency `json:"currencies"`
}

func (s *Server) handleListCurrencies() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := s.currencies.List(r.Context())
		if err != nil {
			log.Printf("list currencies: %v", err)
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}
		writeJSON(w, http.StatusOK, listCurrenciesResponse{Currencies: list})
	}
}

func (s *Server) handleGetCurrency() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := strings.ToUpper(strings.TrimSpace(r.PathValue("code")))
		if code == "" {
			log.Printf("code is empty")
			writeError(w, http.StatusBadRequest, "code is empty")
			return
		}

		cur, err := s.currencies.GetByCode(r.Context(), code)
		if err != nil {
			if errors.Is(err, currency.ErrNotFound) {
				writeError(w, http.StatusNotFound, "currency not found")
				return
			}
			log.Printf("internal error: %v", err)
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}
		writeJSON(w, http.StatusOK, cur)
	}
}
