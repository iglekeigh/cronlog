package api

import (
	"encoding/json"
	"net/http"

	"github.com/user/cronlog/internal/store"
)

// JobStats holds aggregated statistics for a single job.
type JobStats struct {
	Job        string `json:"job"`
	Total      int    `json:"total"`
	Successes  int    `json:"successes"`
	Failures   int    `json:"failures"`
	LastStatus string `json:"last_status,omitempty"`
}

// StatsResponse is the response body for the stats endpoint.
type StatsResponse struct {
	Jobs []JobStats `json:"jobs"`
}

// handleStats returns aggregated run statistics grouped by job name.
func (h *Handler) handleStats(w http.ResponseWriter, r *http.Request) {
	entries, err := h.store.ListAll(r.Context())
	if err != nil {
		h.log.Error("failed to list entries for stats", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	agg := make(map[string]*JobStats)
	for _, e := range entries {
		s, ok := agg[e.Job]
		if !ok {
			s = &JobStats{Job: e.Job}
			agg[e.Job] = s
		}
		s.Total++
		if e.ExitCode == 0 {
			s.Successes++
		} else {
			s.Failures++
		}
		s.LastStatus = e.Status
	}

	resp := StatsResponse{Jobs: make([]JobStats, 0, len(agg))}
	for _, s := range agg {
		resp.Jobs = append(resp.Jobs, *s)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
