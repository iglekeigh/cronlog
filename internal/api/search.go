package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/cronlog/internal/store"
)

// SearchHandler handles full-text search over log entries.
type SearchHandler struct {
	store  store.Store
	logger interface {
		Info(msg string, args ...any)
		Error(msg string, args ...any)
	}
}

// NewSearchHandler creates a new SearchHandler.
func NewSearchHandler(s store.Store, logger interface {
	Info(msg string, args ...any)
	Error(msg string, args ...any)
}) *SearchHandler {
	return &SearchHandler{store: s, logger: logger}
}

// ServeHTTP handles GET /search?q=<term>[&job=<job>]
func (h *SearchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		http.Error(w, "query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	filter := ParseEntryFilter(r)
	page := ParsePageParams(r)

	entries, err := h.store.Search(r.Context(), q, filter, page)
	if err != nil {
		h.logger.Error("search failed", "query", q, "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(entries); err != nil {
		h.logger.Error("failed to encode search results", "error", err)
	}
}
