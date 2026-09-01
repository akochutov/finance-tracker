package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/akochutov/finance-tracker/internal/income"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type listIncomesResponse struct {
	Incomes []income.Income `json:"incomes"`
}

type createIncomeRequest struct {
	PayerID                uuid.UUID       `json:"payer_id"`
	BeneficiaryID          uuid.UUID       `json:"beneficiary_id"`
	Amount                 decimal.Decimal `json:"amount"`
	Currency               string          `json:"currency"`
	OccurredAt             *time.Time      `json:"occurred_at"`
	PaymentType            string          `json:"payment_type"`
	PayerRequisiteID       uuid.UUID       `json:"payer_requisite_id"`
	BeneficiaryRequisiteID uuid.UUID       `json:"beneficiary_requisite_id"`
	Note                   *string         `json:"note"`
}

func (s *Server) handleListIncomes() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := s.incomes.List(r.Context())
		if err != nil {
			log.Printf("internal error: %v", err)
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		writeJSON(w, http.StatusOK, listIncomesResponse{Incomes: list})
	}
}

func (s *Server) handleGetIncome() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}

		result, err := s.incomes.GetByID(r.Context(), id)
		if err != nil {
			if errors.Is(err, income.ErrNotFound) {
				writeError(w, http.StatusNotFound, "income not found")
				return
			}
			log.Printf("internal error: %v", err)
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		writeJSON(w, http.StatusOK, result)
	}
}

func (s *Server) handleCreateIncome() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req createIncomeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}

		if req.OccurredAt == nil {
			writeError(w, http.StatusBadRequest, "occurred_at is required")
			return
		}

		in := income.Income{
			PayerID:                req.PayerID,
			BeneficiaryID:          req.BeneficiaryID,
			Amount:                 req.Amount,
			Currency:               req.Currency,
			OccurredAt:             *req.OccurredAt,
			PaymentType:            req.PaymentType,
			PayerRequisiteID:       req.PayerRequisiteID,
			BeneficiaryRequisiteID: req.BeneficiaryRequisiteID,
			Note:                   req.Note,
		}

		created, err := s.incomes.Create(r.Context(), in)
		if err != nil {
			if errors.Is(err, income.ErrInvalidInput) {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}
			log.Printf("internal error: %v", err)
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		writeJSON(w, http.StatusCreated, created)
	}
}
