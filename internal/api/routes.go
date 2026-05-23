package api

import (
	"log/slog"
	"net/http"
)

// NewRouter builds and returns the HTTP mux with all routes registered.
// Middleware (logging, recovery) is applied to the entire mux.
func NewRouter(h *Handler, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", healthz)
	mux.HandleFunc("GET /api/v1/entries", h.ListEntries)
	mux.HandleFunc("GET /api/v1/entries/{job}", h.ListEntriesByJob)

	var handler http.Handler = mux
	handler = RequestLogger(logger)(handler)
	handler = Recoverer(logger)(handler)

	return handler
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}
