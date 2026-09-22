package api

import (
	"log"
	"net/http"
)

func (s *Server) handleGetDashboard() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := s.dashboard.Get(r.Context())
		if err != nil {
			log.Printf("dashboard: %v", err)
			writeError(w, http.StatusInternalServerError, "failed to build dashboard")
			return
		}
		writeJSON(w, http.StatusOK, data)
	}
}
