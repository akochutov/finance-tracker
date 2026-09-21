package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/akochutov/finance-tracker/internal/exchangerate"
	"github.com/akochutov/finance-tracker/internal/ratefetch"
	"github.com/akochutov/finance-tracker/internal/ratesource"
)

type rateSourcesResponse struct {
	RateSources []ratesource.RateSource `json:"rate_sources"`
}

type saveRateSourceRequest struct {
	Source         string `json:"source"`
	URLTemplate    string `json:"url_template"`
	PollInterval   int    `json:"poll_interval_seconds"`
	RequestTimeout int    `json:"request_timeout_seconds"`
	BackfillStart  string `json:"backfill_start"`
}

type rateProvidersResponse struct {
	Providers []string `json:"providers"`
}

type fetchResultResponse struct {
	Kind    string            `json:"kind"`
	Source  string            `json:"source"`
	Fetched int               `json:"fetched"`
	Skipped int               `json:"skipped"`
	Errors  map[string]string `json:"errors"`
}

func (s *Server) handleListRateSources() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sources, err := s.rateSources.List(r.Context())
		if err != nil {
			log.Printf("list rate sources: %v", err)
			writeError(w, http.StatusInternalServerError, "failed to list rate sources")
			return
		}
		writeJSON(w, http.StatusOK, rateSourcesResponse{RateSources: sources})
	}
}

func (s *Server) handleListRateProviders() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, rateProvidersResponse{Providers: s.rateFetch.ProviderNames()})
	}
}

// PUT /api/rate-sources/{kind}
func (s *Server) handleSaveRateSource() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		kind := strings.ToLower(strings.TrimSpace(r.PathValue("kind")))

		var req saveRateSourceRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}

		source := strings.TrimSpace(req.Source)

		if source != "" && !s.rateFetch.HasProvider(source) {
			writeError(w, http.StatusBadRequest, "unknown provider: "+source)
			return
		}

		var backfillStart *time.Time
		if raw := strings.TrimSpace(req.BackfillStart); raw != "" {
			t, err := time.Parse("2006-01-02", raw)
			if err != nil {
				writeError(w, http.StatusBadRequest, "backfill_start must be YYYY-MM-DD")
				return
			}
			backfillStart = &t
		}

		out, err := s.rateSources.Save(r.Context(), ratesource.RateSource{
			Kind:           kind,
			Source:         source,
			URLTemplate:    strings.TrimSpace(req.URLTemplate),
			PollInterval:   req.PollInterval,
			RequestTimeout: req.RequestTimeout,
			BackfillStart:  backfillStart,
		})
		if err != nil {
			switch {
			case errors.Is(err, ratesource.ErrInvalidKind),
				errors.Is(err, ratesource.ErrInvalidConfig):
				writeError(w, http.StatusBadRequest, err.Error())
			default:
				log.Printf("save rate source: %v", err)
				writeError(w, http.StatusInternalServerError, "failed to save rate source")
			}
			return
		}

		writeJSON(w, http.StatusOK, out)
	}
}

func (s *Server) handleFetchRates() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		kind := strings.ToLower(strings.TrimSpace(r.PathValue("kind")))

		result, err := s.rateFetch.FetchClass(r.Context(), kind, time.Now())
		if err != nil {
			switch {
			case errors.Is(err, ratesource.ErrInvalidKind),
				errors.Is(err, ratefetch.ErrNoProvider),
				errors.Is(err, exchangerate.ErrUnknownProvider):
				writeError(w, http.StatusBadRequest, err.Error())
			default:
				log.Printf("fetch rates: %v", err)
				writeError(w, http.StatusInternalServerError, "failed to fetch rates")
			}
			return
		}

		errsOut := make(map[string]string, len(result.Errors))
		for code, e := range result.Errors {
			errsOut[code] = e.Error()
		}

		writeJSON(w, http.StatusOK, fetchResultResponse{
			Kind:    result.Kind,
			Source:  result.Source,
			Fetched: result.Fetched,
			Skipped: result.Skipped,
			Errors:  errsOut,
		})
	}
}
