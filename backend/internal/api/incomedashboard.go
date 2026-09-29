package api

import (
	"log"
	"net/http"
)

func (s *Server) handleGetIncomeDashboard() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := s.incomeDashboard.Get(r.Context())
		if err != nil {
			log.Printf("income dashboard: %v", err)
			writeError(w, http.StatusInternalServerError, "failed to build income dashboard")
			return
		}
		writeJSON(w, http.StatusOK, data)
	}
}
