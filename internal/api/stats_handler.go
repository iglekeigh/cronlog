package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/cronlog/internal/store"
)

// statsStoreReader is the interface required by the stats handler.
type statsStoreReader interface {
	GetStats() (*store.StatsResult, error)
}

// StatsHandler handles requests for aggregated job statistics.
type StatsHandler struct {
	store  statsStoreReader
	logger *slog.Logger
}

// NewStatsHandler creates a new StatsHandler.
func NewStatsHandler(s statsStoreReader, logger *slog.Logger) *StatsHandler {
	return &StatsHandler{store: s, logger: logger}
}

// ServeHTTP handles GET /stats.
func (h *StatsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	result, err := h.store.GetStats()
	if err != nil {
		h.logger.Error("failed to get stats", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		h.logger.Error("failed to encode stats response", "error", err)
	}
}
