package api

import (
	"errors"
	"log"
	"net/http"

	"github.com/akochutov/finance-tracker/internal/utilitydashboard"
)

func (s *Server) handleGetUtilityDashboard() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		from, err := parseMonth(r.URL.Query().Get("from"), "from")
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		to, err := parseMonth(r.URL.Query().Get("to"), "to")
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		data, err := s.utilityDashboard.Get(r.Context(), from, to)
		if err != nil {
			if errors.Is(err, utilitydashboard.ErrInvalidInput) {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}
			log.Printf("utility dashboard: %v", err)
			writeError(w, http.StatusInternalServerError, "failed to build utility dashboard")
			return
		}
		writeJSON(w, http.StatusOK, data)
	}
}
