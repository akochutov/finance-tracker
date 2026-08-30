package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/akochutov/finance-tracker/internal/currency"
)

type listCurrenciesResponse struct {
	Currencies []currency.Currency `json:"currencies"`
}

type createCurrencyRequest struct {
	Code          string `json:"code"`
	Name          string `json:"name"`
	Kind          string `json:"kind"`
	DecimalPlaces int    `json:"decimal_places"`
}

type updateCurrencyRequest struct {
	Name          *string `json:"name"`
	DecimalPlaces *int    `json:"decimal_places"`
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

		result, err := s.currencies.GetByCode(r.Context(), code)
		if err != nil {
			if errors.Is(err, currency.ErrNotFound) {
				writeError(w, http.StatusNotFound, "currency not found")
				return
			}
			log.Printf("internal error: %v", err)
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		writeJSON(w, http.StatusOK, result)
	}
}

func (s *Server) handleCreateCurrency() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req createCurrencyRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}

		cur := currency.Currency{
			Code:          req.Code,
			Name:          req.Name,
			Kind:          req.Kind,
			DecimalPlaces: req.DecimalPlaces,
		}

		created, err := s.currencies.Create(r.Context(), cur)
		if err != nil {
			if errors.Is(err, currency.ErrInvalidInput) {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}
			if errors.Is(err, currency.ErrAlreadyExists) {
				writeError(w, http.StatusConflict, err.Error())
				return
			}
			log.Printf("internal error: %v", err)
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		writeJSON(w, http.StatusCreated, created)
	}
}

func (s *Server) handleUpdateCurrency() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := strings.ToUpper(strings.TrimSpace(r.PathValue("code")))

		var req updateCurrencyRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}

		if req.Name == nil || req.DecimalPlaces == nil {
			writeError(w, http.StatusBadRequest, "name and decimal_places are required")
			return
		}

		updated, err := s.currencies.Update(r.Context(), code, *req.Name, *req.DecimalPlaces)
		if err != nil {
			if errors.Is(err, currency.ErrInvalidInput) {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}
			if errors.Is(err, currency.ErrNotFound) {
				writeError(w, http.StatusNotFound, err.Error())
				return
			}
			log.Printf("internal error: %v", err)
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		writeJSON(w, http.StatusOK, updated)
	}
}

func (s *Server) handleDeactivateCurrency() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := strings.ToUpper(strings.TrimSpace(r.PathValue("code")))

		err := s.currencies.Deactivate(r.Context(), code)
		if err != nil {
			if errors.Is(err, currency.ErrInvalidInput) {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}
			if errors.Is(err, currency.ErrNotFound) {
				writeError(w, http.StatusNotFound, err.Error())
				return
			}
			log.Printf("internal error: %v", err)
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
