package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// mockLabelStore implements a minimal label store for testing.
type mockLabelStore struct {
	labels  map[int64]map[string]string
	byLabel []labelEntry
	setErr  error
	getErr  error
	listErr error
}

type labelEntry struct {
	ID     int64
	Job    string
	Labels map[string]string
}

func newMockLabelStore() *mockLabelStore {
	return &mockLabelStore{
		labels: make(map[int64]map[string]string),
	}
}

func (m *mockLabelStore) SetLabels(entryID int64, labels map[string]string) error {
	if m.setErr != nil {
		return m.setErr
	}
	m.labels[entryID] = labels
	return nil
}

func (m *mockLabelStore) GetLabels(entryID int64) (map[string]string, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if v, ok := m.labels[entryID]; ok {
		return v, nil
	}
	return map[string]string{}, nil
}

func (m *mockLabelStore) ListByLabel(key, value string) ([]labelEntry, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.byLabel, nil
}

func newLabelLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
}

func TestLabelHandler_MethodNotAllowed(t *testing.T) {
	h := NewLabelHandler(newMockLabelStore(), newLabelLogger())
	req := httptest.NewRequest(http.MethodDelete, "/labels/1", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestLabelHandler_SetLabels(t *testing.T) {
	store := newMockLabelStore()
	h := NewLabelHandler(store, newLabelLogger())

	body, _ := json.Marshal(map[string]string{"env": "prod", "team": "ops"})
	req := httptest.NewRequest(http.MethodPut, "/labels/42", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}
	if store.labels[42]["env"] != "prod" {
		t.Errorf("expected label env=prod to be stored")
	}
}

func TestLabelHandler_GetLabels(t *testing.T) {
	store := newMockLabelStore()
	store.labels[7] = map[string]string{"region": "us-east", "tier": "free"}
	h := NewLabelHandler(store, newLabelLogger())

	req := httptest.NewRequest(http.MethodGet, "/labels/7", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var result map[string]string
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if result["region"] != "us-east" {
		t.Errorf("expected region=us-east, got %s", result["region"])
	}
}

func TestLabelHandler_MissingID(t *testing.T) {
	h := NewLabelHandler(newMockLabelStore(), newLabelLogger())
	req := httptest.NewRequest(http.MethodGet, "/labels/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestLabelHandler_InvalidID(t *testing.T) {
	h := NewLabelHandler(newMockLabelStore(), newLabelLogger())
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/labels/%s", uuid.New().String()), nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestLabelHandler_SetLabels_InvalidBody(t *testing.T) {
	h := NewLabelHandler(newMockLabelStore(), newLabelLogger())
	req := httptest.NewRequest(http.MethodPut, "/labels/1", strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestLabelHandler_StoreError(t *testing.T) {
	store := newMockLabelStore()
	store.getErr = fmt.Errorf("db error")
	h := NewLabelHandler(store, newLabelLogger())

	req := httptest.NewRequest(http.MethodGet, "/labels/1", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}
