package api

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/akochutov/finance-tracker/internal/expensedashboard"
)

func (s *Server) handleGetExpenseDashboard() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		from, err := time.Parse(expenseDateLayout, q.Get("from"))
		if err != nil {
			writeError(w, http.StatusBadRequest, "from is required, YYYY-MM-DD")
			return
		}
		to, err := time.Parse(expenseDateLayout, q.Get("to"))
		if err != nil {
			writeError(w, http.StatusBadRequest, "to is required, YYYY-MM-DD")
			return
		}

		data, err := s.expenseDashboard.Get(r.Context(), from, to)
		if err != nil {
			if errors.Is(err, expensedashboard.ErrInvalidPeriod) {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}
			log.Printf("expense dashboard: %v", err)
			writeError(w, http.StatusInternalServerError, "failed to build expense dashboard")
			return
		}

		writeJSON(w, http.StatusOK, data)
	}
}
