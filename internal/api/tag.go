package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/user/cronlog/internal/store"
)

// tagStore is the subset of store.Store used by the tag handler.
type tagStore interface {
	SetTags(entryID int64, tags []store.Tag) error
	GetTags(entryID int64) ([]store.Tag, error)
	ListByTag(key, value string) ([]int64, error)
}

// TagHandler handles tag-related HTTP endpoints.
type TagHandler struct {
	store  tagStore
	logger *slog.Logger
}

// NewTagHandler creates a new TagHandler.
func NewTagHandler(s tagStore, l *slog.Logger) *TagHandler {
	return &TagHandler{store: s, logger: l}
}

// ServeHTTP routes /tags requests.
// GET  /tags?key=env&value=prod  — list entry IDs by tag
// GET  /tags/{id}                — get tags for an entry
// PUT  /tags/{id}                — replace tags for an entry
func (h *TagHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// strip leading /tags
	path := strings.TrimPrefix(r.URL.Path, "/tags")
	path = strings.TrimPrefix(path, "/")

	if path == "" {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		h.handleListByTag(w, r)
		return
	}

	id, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		http.Error(w, "invalid entry id", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.handleGetTags(w, r, id)
	case http.MethodPut:
		h.handleSetTags(w, r, id)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *TagHandler) handleListByTag(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "key is required", http.StatusBadRequest)
		return
	}
	value := r.URL.Query().Get("value")

	ids, err := h.store.ListByTag(key, value)
	if err != nil {
		h.logger.Error("list by tag", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if ids == nil {
		ids = []int64{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"entry_ids": ids}) //nolint:errcheck
}

func (h *TagHandler) handleGetTags(w http.ResponseWriter, _ *http.Request, id int64) {
	tags, err := h.store.GetTags(id)
	if err != nil {
		h.logger.Error("get tags", "entry_id", id, "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if tags == nil {
		tags = []store.Tag{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"tags": tags}) //nolint:errcheck
}

func (h *TagHandler) handleSetTags(w http.ResponseWriter, r *http.Request, id int64) {
	var body struct {
		Tags []store.Tag `json:"tags"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if err := h.store.SetTags(id, body.Tags); err != nil {
		h.logger.Error("set tags", "entry_id", id, "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
