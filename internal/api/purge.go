package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type PurgeStore interface {
	PurgeByJob(name string) (int64, error)
	PurgeByStatus(status string) (int64, error)
	PurgeAll() (int64, error)
}

type PurgeHandler struct {
	store  PurgeStore
	logger *slog.Logger
}

func NewPurgeHandler(store PurgeStore, logger *slog.Logger) *PurgeHandler {
	return &PurgeHandler{store: store, logger: logger}
}

type purgeResponse struct {
	Deleted int64  `json:"deleted"`
	Message string `json:"message"`
}

func (h *PurgeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	job := r.URL.Query().Get("job")
	status := r.URL.Query().Get("status")

	var (
		deleted int64
		err     error
	)

	switch {
	case job != "":
		deleted, err = h.store.PurgeByJob(job)
	case status != "":
		deleted, err = h.store.PurgeByStatus(status)
	default:
		deleted, err = h.store.PurgeAll()
	}

	if err != nil {
		h.logger.Error("purge failed", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("purge completed", "deleted", deleted, "job", job, "status", status)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(purgeResponse{
		Deleted: deleted,
		Message: "purge successful",
	})
}
