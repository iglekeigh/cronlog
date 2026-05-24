package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/user/cronlog/internal/store"
)

// --- mock tag store ---

type mockTagStore struct {
	tags  map[int64][]store.Tag
	byTag map[string][]int64
	err   error
}

func newMockTagStore() *mockTagStore {
	return &mockTagStore{
		tags:  make(map[int64][]store.Tag),
		byTag: make(map[string][]int64),
	}
}

func (m *mockTagStore) SetTags(id int64, tags []store.Tag) error {
	if m.err != nil {
		return m.err
	}
	m.tags[id] = tags
	return nil
}

func (m *mockTagStore) GetTags(id int64) ([]store.Tag, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.tags[id], nil
}

func (m *mockTagStore) ListByTag(key, _ string) ([]int64, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.byTag[key], nil
}

// --- helpers ---

func newTagLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, nil))
}

// --- tests ---

func TestTagHandler_MethodNotAllowed(t *testing.T) {
	h := NewTagHandler(newMockTagStore(), newTagLogger())
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/tags", nil))
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rr.Code)
	}
}

func TestTagHandler_ListByTag_MissingKey(t *testing.T) {
	h := NewTagHandler(newMockTagStore(), newTagLogger())
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/tags", nil))
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestTagHandler_ListByTag_Success(t *testing.T) {
	ms := newMockTagStore()
	ms.byTag["env"] = []int64{1, 2}
	h := NewTagHandler(ms, newTagLogger())

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/tags?key=env", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&resp) //nolint:errcheck
	ids := resp["entry_ids"].([]interface{})
	if len(ids) != 2 {
		t.Errorf("expected 2 ids, got %d", len(ids))
	}
}

func TestTagHandler_GetTags_Success(t *testing.T) {
	ms := newMockTagStore()
	ms.tags[7] = []store.Tag{{Key: "env", Value: "prod"}}
	h := NewTagHandler(ms, newTagLogger())

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/tags/7", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestTagHandler_SetTags_Success(t *testing.T) {
	ms := newMockTagStore()
	h := NewTagHandler(ms, newTagLogger())

	body := `{"tags":[{"key":"env","value":"prod"}]}`
	req := httptest.NewRequest(http.MethodPut, "/tags/3", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rr.Code)
	}
	if len(ms.tags[3]) != 1 {
		t.Errorf("expected 1 stored tag, got %d", len(ms.tags[3]))
	}
}

func TestTagHandler_SetTags_StoreError(t *testing.T) {
	ms := newMockTagStore()
	ms.err = fmt.Errorf("db error")
	h := NewTagHandler(ms, newTagLogger())

	body := `{"tags":[]}`
	req := httptest.NewRequest(http.MethodPut, "/tags/1", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}
