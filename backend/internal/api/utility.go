package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/akochutov/finance-tracker/internal/utility"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

const monthLayout = "2006-01"

type listServiceTypesResponse struct {
	ServiceTypes []utility.ServiceType `json:"service_types"`
}

type listAddressesResponse struct {
	Addresses []utility.Address `json:"addresses"`
}

type listUtilityAccountsResponse struct {
	Accounts []utility.Account `json:"accounts"`
}

type listMetersResponse struct {
	Meters []utility.Meter `json:"meters"`
}

type listReadingsResponse struct {
	Readings []utility.Reading `json:"readings"`
}

type addressRequest struct {
	Address string `json:"address"`
}

type createUtilityAccountRequest struct {
	AddressID         uuid.UUID  `json:"address_id"`
	Service           string     `json:"service"`
	Number            string     `json:"number"`
	ExpenseCategoryID *uuid.UUID `json:"expense_category_id"`
}

type updateUtilityAccountRequest struct {
	Number            string     `json:"number"`
	ExpenseCategoryID *uuid.UUID `json:"expense_category_id"`
}

type createMeterRequest struct {
	AccountID     uuid.UUID                  `json:"account_id"`
	Serial        string                     `json:"serial"`
	Registers     string                     `json:"registers"`
	InstalledOn   string                     `json:"installed_on"`
	RemovedOn     *string                    `json:"removed_on"`
	InitialOn     *string                    `json:"initial_on"`
	InitialValues map[string]decimal.Decimal `json:"initial_values"`
}

type updateMeterRequest struct {
	Serial      string  `json:"serial"`
	InstalledOn string  `json:"installed_on"`
	RemovedOn   *string `json:"removed_on"`
}

type createReadingsRequest struct {
	TakenOn  string `json:"taken_on"`
	Readings []struct {
		MeterID uuid.UUID        `json:"meter_id"`
		Zone    string           `json:"zone"`
		Value   *decimal.Decimal `json:"value"`
	} `json:"readings"`
}

type updateReadingRequest struct {
	TakenOn string           `json:"taken_on"`
	Value   *decimal.Decimal `json:"value"`
}

func pathUUID(w http.ResponseWriter, r *http.Request, name string) (uuid.UUID, bool) {
	id, err := uuid.Parse(r.PathValue(name))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid "+name)
		return uuid.Nil, false
	}
	return id, true
}

func parseDay(value, field string) (time.Time, error) {
	t, err := time.Parse(expenseDateLayout, value)
	if err != nil {
		return time.Time{}, errors.New(field + " must be YYYY-MM-DD")
	}
	return t, nil
}

func parseMonth(value, field string) (time.Time, error) {
	t, err := time.Parse(monthLayout, value)
	if err != nil {
		return time.Time{}, errors.New(field + " must be YYYY-MM")
	}
	return t, nil
}

func writeUtilityError(w http.ResponseWriter, err error, op string) {
	switch {
	case errors.Is(err, utility.ErrInvalidInput),
		errors.Is(err, utility.ErrInitialReadingRequired):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, utility.ErrAddressNotFound),
		errors.Is(err, utility.ErrAccountNotFound),
		errors.Is(err, utility.ErrMeterNotFound),
		errors.Is(err, utility.ErrReadingNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, utility.ErrAddressTaken),
		errors.Is(err, utility.ErrAccountTaken),
		errors.Is(err, utility.ErrSerialTaken),
		errors.Is(err, utility.ErrReadingExists),
		errors.Is(err, utility.ErrAddressInactive),
		errors.Is(err, utility.ErrAddressHasActiveAccounts),
		errors.Is(err, utility.ErrAccountInactive),
		errors.Is(err, utility.ErrInitialReadingLocked):
		writeError(w, http.StatusConflict, err.Error())
	default:
		log.Printf("%s: %v", op, err)
		writeError(w, http.StatusInternalServerError, "internal error")
	}
}

func (s *Server) handleListServiceTypes() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := s.utilities.ListServiceTypes(r.Context())
		if err != nil {
			writeUtilityError(w, err, "list service types")
			return
		}
		writeJSON(w, http.StatusOK, listServiceTypesResponse{ServiceTypes: list})
	}
}

func (s *Server) handleListAddresses() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := s.utilities.ListAddresses(r.Context())
		if err != nil {
			writeUtilityError(w, err, "list addresses")
			return
		}
		writeJSON(w, http.StatusOK, listAddressesResponse{Addresses: list})
	}
}

func (s *Server) handleCreateAddress() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req addressRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}

		created, err := s.utilities.CreateAddress(r.Context(), req.Address)
		if err != nil {
			writeUtilityError(w, err, "create address")
			return
		}
		writeJSON(w, http.StatusCreated, created)
	}
}

func (s *Server) handleUpdateAddress() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := pathUUID(w, r, "id")
		if !ok {
			return
		}

		var req addressRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}

		updated, err := s.utilities.UpdateAddress(r.Context(), id, req.Address)
		if err != nil {
			writeUtilityError(w, err, "update address")
			return
		}
		writeJSON(w, http.StatusOK, updated)
	}
}

func (s *Server) handleDeactivateAddress() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := pathUUID(w, r, "id")
		if !ok {
			return
		}
		if err := s.utilities.SetAddressActive(r.Context(), id, false); err != nil {
			writeUtilityError(w, err, "deactivate address")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func (s *Server) handleActivateAddress() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := pathUUID(w, r, "id")
		if !ok {
			return
		}
		if err := s.utilities.SetAddressActive(r.Context(), id, true); err != nil {
			writeUtilityError(w, err, "activate address")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func (s *Server) handleListUtilityAccounts() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := s.utilities.ListAccounts(r.Context())
		if err != nil {
			writeUtilityError(w, err, "list utility accounts")
			return
		}
		writeJSON(w, http.StatusOK, listUtilityAccountsResponse{Accounts: list})
	}
}

func (s *Server) handleCreateUtilityAccount() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req createUtilityAccountRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		if req.AddressID == uuid.Nil {
			writeError(w, http.StatusBadRequest, "address_id is required")
			return
		}

		created, err := s.utilities.CreateAccount(r.Context(), req.AddressID, req.Service, req.Number, req.ExpenseCategoryID)
		if err != nil {
			writeUtilityError(w, err, "create utility account")
			return
		}
		writeJSON(w, http.StatusCreated, created)
	}
}

func (s *Server) handleUpdateUtilityAccount() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := pathUUID(w, r, "id")
		if !ok {
			return
		}

		var req updateUtilityAccountRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}

		updated, err := s.utilities.UpdateAccount(r.Context(), id, req.Number, req.ExpenseCategoryID)
		if err != nil {
			writeUtilityError(w, err, "update utility account")
			return
		}
		writeJSON(w, http.StatusOK, updated)
	}
}

func (s *Server) handleDeactivateUtilityAccount() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := pathUUID(w, r, "id")
		if !ok {
			return
		}
		if err := s.utilities.SetAccountActive(r.Context(), id, false); err != nil {
			writeUtilityError(w, err, "deactivate utility account")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func (s *Server) handleActivateUtilityAccount() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := pathUUID(w, r, "id")
		if !ok {
			return
		}
		if err := s.utilities.SetAccountActive(r.Context(), id, true); err != nil {
			writeUtilityError(w, err, "activate utility account")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func (s *Server) handleListMeters() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := s.utilities.ListMeters(r.Context())
		if err != nil {
			writeUtilityError(w, err, "list meters")
			return
		}
		writeJSON(w, http.StatusOK, listMetersResponse{Meters: list})
	}
}

func (s *Server) handleCreateMeter() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req createMeterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		if req.AccountID == uuid.Nil {
			writeError(w, http.StatusBadRequest, "account_id is required")
			return
		}
		installedOn, err := parseDay(req.InstalledOn, "installed_on")
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		removedOn, err := parseOptionalDate(req.RemovedOn)
		if err != nil {
			writeError(w, http.StatusBadRequest, "removed_on must be YYYY-MM-DD")
			return
		}
		initialOn, err := parseOptionalDate(req.InitialOn)
		if err != nil {
			writeError(w, http.StatusBadRequest, "initial_on must be YYYY-MM-DD")
			return
		}
		var initialDay time.Time
		if initialOn != nil {
			initialDay = *initialOn
		}

		created, err := s.utilities.CreateMeter(r.Context(), req.AccountID, req.Serial, req.Registers, installedOn, removedOn, initialDay, req.InitialValues)
		if err != nil {
			writeUtilityError(w, err, "create meter")
			return
		}
		writeJSON(w, http.StatusCreated, created)
	}
}

func (s *Server) handleUpdateMeter() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := pathUUID(w, r, "id")
		if !ok {
			return
		}

		var req updateMeterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		installedOn, err := parseDay(req.InstalledOn, "installed_on")
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		removedOn, err := parseOptionalDate(req.RemovedOn)
		if err != nil {
			writeError(w, http.StatusBadRequest, "removed_on must be YYYY-MM-DD")
			return
		}

		updated, err := s.utilities.UpdateMeter(r.Context(), id, req.Serial, installedOn, removedOn)
		if err != nil {
			writeUtilityError(w, err, "update meter")
			return
		}
		writeJSON(w, http.StatusOK, updated)
	}
}

func (s *Server) handleDeleteMeter() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := pathUUID(w, r, "id")
		if !ok {
			return
		}
		if err := s.utilities.DeleteMeter(r.Context(), id); err != nil {
			writeUtilityError(w, err, "delete meter")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func (s *Server) handleCreateReadings() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req createReadingsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		takenOn, err := parseDay(req.TakenOn, "taken_on")
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		values := make(map[uuid.UUID]map[string]decimal.Decimal, len(req.Readings))
		for _, rd := range req.Readings {
			if rd.MeterID == uuid.Nil || rd.Value == nil {
				writeError(w, http.StatusBadRequest, "each reading needs meter_id and value")
				return
			}
			zone := strings.ToLower(strings.TrimSpace(rd.Zone))
			if zone == "" {
				zone = utility.ZoneSingle
			}
			if values[rd.MeterID] == nil {
				values[rd.MeterID] = map[string]decimal.Decimal{}
			}
			if _, dup := values[rd.MeterID][zone]; dup {
				writeError(w, http.StatusBadRequest, "a meter register appears twice in one round")
				return
			}
			values[rd.MeterID][zone] = *rd.Value
		}

		created, err := s.utilities.CreateReadings(r.Context(), takenOn, values)
		if err != nil {
			writeUtilityError(w, err, "create readings")
			return
		}
		writeJSON(w, http.StatusCreated, listReadingsResponse{Readings: created})
	}
}

func (s *Server) handleLatestReadings() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := s.utilities.LatestReadings(r.Context())
		if err != nil {
			writeUtilityError(w, err, "latest readings")
			return
		}
		writeJSON(w, http.StatusOK, listReadingsResponse{Readings: list})
	}
}

func (s *Server) handleListReadings() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		meterID, ok := pathUUID(w, r, "id")
		if !ok {
			return
		}

		list, err := s.utilities.ListReadings(r.Context(), meterID)
		if err != nil {
			writeUtilityError(w, err, "list readings")
			return
		}
		writeJSON(w, http.StatusOK, listReadingsResponse{Readings: list})
	}
}

func (s *Server) handleUpdateReading() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		meterID, ok := pathUUID(w, r, "id")
		if !ok {
			return
		}
		readingID, ok := pathUUID(w, r, "rid")
		if !ok {
			return
		}

		var req updateReadingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		takenOn, err := parseDay(req.TakenOn, "taken_on")
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if req.Value == nil {
			writeError(w, http.StatusBadRequest, "value is required")
			return
		}

		updated, err := s.utilities.UpdateReading(r.Context(), meterID, readingID, takenOn, *req.Value)
		if err != nil {
			writeUtilityError(w, err, "update reading")
			return
		}
		writeJSON(w, http.StatusOK, updated)
	}
}

func (s *Server) handleDeleteReading() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		meterID, ok := pathUUID(w, r, "id")
		if !ok {
			return
		}
		readingID, ok := pathUUID(w, r, "rid")
		if !ok {
			return
		}
		if err := s.utilities.DeleteReading(r.Context(), meterID, readingID); err != nil {
			writeUtilityError(w, err, "delete reading")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
