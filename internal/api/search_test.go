package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cronlog/internal/api"
	"github.com/cronlog/internal/store"
)

type searchStore struct {
	*testStoreBase
	results []store.Entry
	err     error
}

func (s *searchStore) Search(_ context.Context, q string, _ store.EntryFilter, _ store.PageParams) ([]store.Entry, error) {
	return s.results, s.err
}

func newSearchLogger() *testLogger {
	return &testLogger{}
}

func TestSearch_MissingQuery(t *testing.T) {
	h := api.NewSearchHandler(newTestStore(), newSearchLogger())
	req := httptest.NewRequest(http.MethodGet, "/search", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)
	if rw.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rw.Code)
	}
}

func TestSearch_MethodNotAllowed(t *testing.T) {
	h := api.NewSearchHandler(newTestStore(), newSearchLogger())
	req := httptest.NewRequest(http.MethodPost, "/search?q=foo", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)
	if rw.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rw.Code)
	}
}

func TestSearch_ReturnsResults(t *testing.T) {
	expected := []store.Entry{
		{ID: 1, Job: "backup", Output: "backup completed", StartedAt: time.Now()},
	}
	s := newTestStore()
	s.entries = expected
	h := api.NewSearchHandler(s, newSearchLogger())
	req := httptest.NewRequest(http.MethodGet, "/search?q=backup", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)
	if rw.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rw.Code)
	}
	var got []store.Entry
	if err := json.NewDecoder(rw.Body).Decode(&got); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(got) != 1 || got[0].Job != "backup" {
		t.Fatalf("unexpected results: %+v", got)
	}
}
