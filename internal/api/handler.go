package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/user/cronlog/internal/store"
)

// Handler holds dependencies for HTTP handlers.
type Handler struct {
	store  *store.Store
	logger *slog.Logger
}

// NewHandler creates a new Handler.
func NewHandler(s *store.Store, logger *slog.Logger) *Handler {
	return &Handler{store: s, logger: logger}
}

// ListEntries handles GET /entries?job=<name>&limit=<n>
func (h *Handler) ListEntries(w http.ResponseWriter, r *http.Request) {
	job := r.URL.Query().Get("job")
	limitStr := r.URL.Query().Get("limit")

	limit := 100
	if limitStr != "" {
		if n, err := strconv.Atoi(limitStr); err == nil && n > 0 {
			limit = n
		}
	}

	entries, err := h.store.ListByJob(r.Context(), job, limit)
	if err != nil {
		h.logger.Error("failed to list entries", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(entries); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

// GetStats handles GET /stats
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.store.Stats(r.Context())
	if err != nil {
		h.logger.Error("failed to get stats", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		h.logger.Error("failed to encode stats", "error", err)
	}
}
