package api

import (
	"log/slog"
	"net/http"

	"github.com/user/cronlog/internal/store"
)

// NewRouter builds and returns an http.ServeMux with all API routes registered.
func NewRouter(s *store.Store, logger *slog.Logger) http.Handler {
	h := NewHandler(s, logger)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /entries", h.ListEntries)
	mux.HandleFunc("GET /stats", h.GetStats)
	mux.HandleFunc("GET /healthz", healthz)

	return mux
}

// healthz is a simple liveness probe.
func healthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
