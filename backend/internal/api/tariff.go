package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/akochutov/finance-tracker/internal/tariff"
	"github.com/shopspring/decimal"
)

type listTariffsResponse struct {
	Tariffs []tariff.Tariff `json:"tariffs"`
}

type tierRequest struct {
	UpTo  *decimal.Decimal `json:"up_to"`
	Price *decimal.Decimal `json:"price"`
}

type createTariffRequest struct {
	Service   string        `json:"service"`
	Zone      string        `json:"zone"`
	ValidFrom string        `json:"valid_from"`
	Currency  string        `json:"currency"`
	TierMode  string        `json:"tier_mode"`
	Tiers     []tierRequest `json:"tiers"`
}

type updateTariffRequest struct {
	ValidFrom string        `json:"valid_from"`
	Currency  string        `json:"currency"`
	TierMode  string        `json:"tier_mode"`
	Tiers     []tierRequest `json:"tiers"`
}

func toTiers(in []tierRequest) ([]tariff.Tier, error) {
	out := make([]tariff.Tier, 0, len(in))
	for _, t := range in {
		if t.Price == nil {
			return nil, errors.New("every tier needs a price")
		}
		out = append(out, tariff.Tier{UpTo: t.UpTo, Price: *t.Price})
	}
	return out, nil
}

func writeTariffError(w http.ResponseWriter, err error, op string) {
	switch {
	case errors.Is(err, tariff.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, tariff.ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, tariff.ErrAlreadyExists):
		writeError(w, http.StatusConflict, err.Error())
	default:
		log.Printf("%s: %v", op, err)
		writeError(w, http.StatusInternalServerError, "internal error")
	}
}

func (s *Server) handleListTariffs() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := s.tariffs.List(r.Context())
		if err != nil {
			writeTariffError(w, err, "list tariffs")
			return
		}
		writeJSON(w, http.StatusOK, listTariffsResponse{Tariffs: list})
	}
}

func (s *Server) handleCreateTariff() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req createTariffRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		validFrom, err := parseDay(req.ValidFrom, "valid_from")
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		tiers, err := toTiers(req.Tiers)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		created, err := s.tariffs.Create(r.Context(), req.Service, req.Zone, validFrom, req.Currency, req.TierMode, tiers)
		if err != nil {
			writeTariffError(w, err, "create tariff")
			return
		}
		writeJSON(w, http.StatusCreated, created)
	}
}

func (s *Server) handleUpdateTariff() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := pathUUID(w, r, "id")
		if !ok {
			return
		}

		var req updateTariffRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		validFrom, err := parseDay(req.ValidFrom, "valid_from")
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		tiers, err := toTiers(req.Tiers)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		updated, err := s.tariffs.Update(r.Context(), id, validFrom, req.Currency, req.TierMode, tiers)
		if err != nil {
			writeTariffError(w, err, "update tariff")
			return
		}
		writeJSON(w, http.StatusOK, updated)
	}
}

func (s *Server) handleDeleteTariff() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := pathUUID(w, r, "id")
		if !ok {
			return
		}
		if err := s.tariffs.Delete(r.Context(), id); err != nil {
			writeTariffError(w, err, "delete tariff")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
