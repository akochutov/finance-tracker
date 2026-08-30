package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/akochutov/finance-tracker/internal/company"
	"github.com/google/uuid"
)

type listCompaniesResponse struct {
	Companies []company.Company `json:"companies"`
}

type createCompanyRequest struct {
	Name    string  `json:"name"`
	Note    *string `json:"note"`
	TaxID   *string `json:"tax_id"`
	Address *string `json:"address"`
}

type updateCompanyRequest struct {
	Name    *string `json:"name"`
	Note    *string `json:"note"`
	TaxID   *string `json:"tax_id"`
	Address *string `json:"address"`
}

func (s *Server) handleListCompanies() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := s.companies.List(r.Context())
		if err != nil {
			log.Printf("list companies: %v", err)
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		writeJSON(w, http.StatusOK, listCompaniesResponse{Companies: list})
	}
}

func (s *Server) handleGetCompany() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}

		result, err := s.companies.GetByID(r.Context(), id)
		if err != nil {
			if errors.Is(err, company.ErrNotFound) {
				writeError(w, http.StatusNotFound, "company not found")
				return
			}
			log.Printf("internal error: %v", err)
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		writeJSON(w, http.StatusOK, result)
	}
}

func (s *Server) handleCreateCompany() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req createCompanyRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}

		comp := company.Company{
			Name:    req.Name,
			Note:    req.Note,
			TaxID:   req.TaxID,
			Address: req.Address,
		}

		created, err := s.companies.Create(r.Context(), comp)
		if err != nil {
			if errors.Is(err, company.ErrInvalidInput) {
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

func (s *Server) handleUpdateCompany() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}

		var req updateCompanyRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}

		if req.Name == nil {
			writeError(w, http.StatusBadRequest, "name must be filled")
			return
		}

		comp := company.Company{
			ID:      id,
			Name:    *req.Name,
			Note:    req.Note,
			TaxID:   req.TaxID,
			Address: req.Address,
		}

		updated, err := s.companies.Update(r.Context(), comp)
		if err != nil {
			if errors.Is(err, company.ErrInvalidInput) {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}
			if errors.Is(err, company.ErrNotFound) {
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

func (s *Server) handleDeactivateCompany() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}

		if err := s.companies.Deactivate(r.Context(), id); err != nil {
			if errors.Is(err, company.ErrNotFound) {
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
