package api

import (
	"log/slog"
	"net/http"

	"github.com/user/cronlog/internal/store"
)

func NewRouter(s *store.Store, log *slog.Logger) http.Handler {
	h := NewHandler(s, log)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", healthz)
	mux.HandleFunc("GET /api/v1/entries", h.ListEntries)
	mux.HandleFunc("GET /api/v1/stats", h.HandleStats)
	mux.HandleFunc("GET /api/v1/export", h.HandleExport)

	return RequestLogger(log, Recoverer(mux))
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}
