package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/user/cronlog/internal/store"
)

func TestListEntries_FilterByJob(t *testing.T) {
	s := newTestStore(t)
	now := time.Now()
	s.Insert(store.Entry{ID: "1", JobName: "backup", Status: "success", StartedAt: now})
	s.Insert(store.Entry{ID: "2", JobName: "cleanup", Status: "failure", StartedAt: now})

	h := NewHandler(s, testLogger(t))
	req := httptest.NewRequest(http.MethodGet, "/entries?job=backup", nil)
	rec := httptest.NewRecorder()
	h.ListEntries(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var entries []store.Entry
	if err := json.NewDecoder(rec.Body).Decode(&entries); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(entries) != 1 || entries[0].JobName != "backup" {
		t.Errorf("expected only backup entry, got %+v", entries)
	}
}

func TestListEntries_FilterByStatus(t *testing.T) {
	s := newTestStore(t)
	now := time.Now()
	s.Insert(store.Entry{ID: "1", JobName: "j", Status: "success", StartedAt: now})
	s.Insert(store.Entry{ID: "2", JobName: "j", Status: "failure", StartedAt: now})

	h := NewHandler(s, testLogger(t))
	req := httptest.NewRequest(http.MethodGet, "/entries?status=failure", nil)
	rec := httptest.NewRecorder()
	h.ListEntries(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var entries []store.Entry
	if err := json.NewDecoder(rec.Body).Decode(&entries); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(entries) != 1 || entries[0].Status != "failure" {
		t.Errorf("expected only failure entry, got %+v", entries)
	}
}
