package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/user/cronlog/internal/store"
)

// LabelStore defines the store operations used by LabelHandler.
type LabelStore interface {
	SetLabels(entryID int64, labels map[string]string) error
	GetLabels(entryID int64) (map[string]string, error)
	ListByLabel(key, value string) ([]int64, error)
}

// LabelHandler handles label CRUD for log entries.
type LabelHandler struct {
	store  LabelStore
	logger *slog.Logger
}

// NewLabelHandler creates a new LabelHandler.
func NewLabelHandler(s LabelStore, l *slog.Logger) *LabelHandler {
	return &LabelHandler{store: s, logger: l}
}

// ServeHTTP routes label requests.
// GET  /labels?key=env&value=prod  — list entry IDs by label
// GET  /labels/{id}               — get labels for entry
// PUT  /labels/{id}               — set labels for entry (body: JSON object)
func (h *LabelHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if idStr := r.PathValue("id"); idStr != "" {
			h.getLabels(w, r, idStr)
		} else {
			h.listByLabel(w, r)
		}
	case http.MethodPut:
		if idStr := r.PathValue("id"); idStr != "" {
			h.setLabels(w, r, idStr)
		} else {
			http.Error(w, "entry id required", http.StatusBadRequest)
		}
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *LabelHandler) getLabels(w http.ResponseWriter, _ *http.Request, idStr string) {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	labels, err := h.store.GetLabels(id)
	if err != nil {
		h.logger.Error("get labels", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(labels) //nolint:errcheck
}

func (h *LabelHandler) setLabels(w http.ResponseWriter, r *http.Request, idStr string) {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	var labels map[string]string
	if err := json.NewDecoder(r.Body).Decode(&labels); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if err := h.store.SetLabels(id, labels); err != nil {
		h.logger.Error("set labels", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *LabelHandler) listByLabel(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "key is required", http.StatusBadRequest)
		return
	}
	value := r.URL.Query().Get("value")
	ids, err := h.store.ListByLabel(key, value)
	if err != nil {
		h.logger.Error("list by label", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if ids == nil {
		ids = []int64{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"entry_ids": ids, "count": len(ids)}) //nolint:errcheck
}

// compile-time check
var _ LabelStore = (*store.Store)(nil)
