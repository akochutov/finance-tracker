package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/akochutov/finance-tracker/internal/requisite"
	"github.com/google/uuid"
)

type listBankRequisitesResponse struct {
	BankRequisites []requisite.BankRequisite `json:"bank_requisites"`
}

type createBankRequisiteRequest struct {
	BeneficiaryName        string     `json:"beneficiary_name"`
	AccountNumber          string     `json:"account_number"`
	BankName               string     `json:"bank_name"`
	BankSwift              string     `json:"bank_swift"`
	BankAddress            *string    `json:"bank_address"`
	CorrespondentBankName  *string    `json:"correspondent_bank_name"`
	CorrespondentBankSwift *string    `json:"correspondent_bank_swift"`
	IntermediaryBankName   *string    `json:"intermediary_bank_name"`
	IntermediaryBankSwift  *string    `json:"intermediary_bank_swift"`
	ValidFrom              *time.Time `json:"valid_from"`
}

type closeBankRequisiteRequest struct {
	ValidTo *time.Time `json:"valid_to"`
}

func (s *Server) handleListBankRequisites() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		companyID, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}

		list, err := s.bankRequisites.ListByCompany(r.Context(), companyID)
		if err != nil {
			log.Printf("list bank requisites: %v", err)
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		writeJSON(w, http.StatusOK, listBankRequisitesResponse{BankRequisites: list})
	}
}

func (s *Server) handleCreateBankRequisite() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		companyID, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}

		var req createBankRequisiteRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}

		br := requisite.BankRequisite{
			CompanyID:              companyID,
			BeneficiaryName:        req.BeneficiaryName,
			AccountNumber:          req.AccountNumber,
			BankName:               req.BankName,
			BankSwift:              req.BankSwift,
			BankAddress:            req.BankAddress,
			CorrespondentBankName:  req.CorrespondentBankName,
			CorrespondentBankSwift: req.CorrespondentBankSwift,
			IntermediaryBankName:   req.IntermediaryBankName,
			IntermediaryBankSwift:  req.IntermediaryBankSwift,
		}
		if req.ValidFrom != nil {
			br.ValidFrom = (*req.ValidFrom).UTC()
		}

		created, err := s.bankRequisites.Create(r.Context(), br)
		if err != nil {
			if errors.Is(err, requisite.ErrInvalidInput) {
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

func (s *Server) handleCloseBankRequisite() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(r.PathValue("rid"))
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid requisite_id")
			return
		}

		var req closeBankRequisiteRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}

		validTo := time.Now().UTC()
		if req.ValidTo != nil {
			validTo = (*req.ValidTo).UTC()
		}

		if err := s.bankRequisites.Close(r.Context(), id, validTo); err != nil {
			if errors.Is(err, requisite.ErrInvalidInput) {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}
			if errors.Is(err, requisite.ErrNotFound) {
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
