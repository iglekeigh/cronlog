package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/user/cronlog/internal/runner"
	"github.com/user/cronlog/internal/store"
)

// ReplayHandler re-runs the command associated with a stored log entry.
type ReplayHandler struct {
	store  *store.Store
	runner *runner.Runner
	log    *slog.Logger
}

func NewReplayHandler(s *store.Store, r *runner.Runner, log *slog.Logger) *ReplayHandler {
	return &ReplayHandler{store: s, runner: r, log: log}
}

func (h *ReplayHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "missing required parameter: id", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id: must be an integer", http.StatusBadRequest)
		return
	}

	entry, err := h.store.GetByID(r.Context(), id)
	if err != nil {
		h.log.Error("replay: failed to fetch entry", "id", id, "err", err)
		http.Error(w, "entry not found", http.StatusNotFound)
		return
	}

	h.log.Info("replaying job", "job", entry.JobName, "original_id", id)

	newEntry, err := h.runner.Run(r.Context(), entry.JobName, entry.Command)
	if err != nil {
		h.log.Error("replay: runner error", "job", entry.JobName, "err", err)
		http.Error(w, "failed to replay job", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(newEntry)
}
