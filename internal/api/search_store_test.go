package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andrebq/cronlog/internal/store"
)

// storeSearcher is a minimal interface for the search handler backed by the real store.
type storeSearcher interface {
	Search(query string, limit, offset int) ([]store.SearchEntry, error)
}

func TestSearch_StoreIntegration(t *testing.T) {
	log := newSearchLogger(t)

	// Use a mock that returns known data.
	mock := &mockSearchStore{
		entries: []store.SearchEntry{
			{ID: 1, JobName: "backup", Status: "success", Output: "all done"},
		},
	}

	h := NewSearchHandler(mock, log)

	req := httptest.NewRequest(http.MethodGet, "/search?q=backup", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var results []store.SearchEntry
	if err := json.NewDecoder(rec.Body).Decode(&results); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].JobName != "backup" {
		t.Errorf("unexpected job name: %s", results[0].JobName)
	}
	if mock.lastQuery != "backup" {
		t.Errorf("expected query 'backup', got '%s'", mock.lastQuery)
	}
}

func TestSearch_StoreError(t *testing.T) {
	log := newSearchLogger(t)
	mock := &mockSearchStore{err: errSearchFailed}

	h := NewSearchHandler(mock, log)

	req := httptest.NewRequest(http.MethodGet, "/search?q=anything", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
}
