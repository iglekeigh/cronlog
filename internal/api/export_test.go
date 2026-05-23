package api_test

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/user/cronlog/internal/api"
	"github.com/user/cronlog/internal/store"
)

func TestHandleExport_JSON(t *testing.T) {
	s := newTestStore(t)
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	h := api.NewHandler(s, log)

	now := time.Now().UTC()
	_ = s.Insert(t.Context(), store.Entry{Job: "backup", Status: "success", ExitCode: 0, StartedAt: now})
	_ = s.Insert(t.Context(), store.Entry{Job: "backup", Status: "failure", ExitCode: 1, StartedAt: now})

	req := httptest.NewRequest(http.MethodGet, "/export?format=json", nil)
	rr := httptest.NewRecorder()
	h.HandleExport(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %s", ct)
	}

	var entries []store.Entry
	if err := json.NewDecoder(rr.Body).Decode(&entries); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}
}

func TestHandleExport_CSV(t *testing.T) {
	s := newTestStore(t)
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	h := api.NewHandler(s, log)

	now := time.Now().UTC()
	_ = s.Insert(t.Context(), store.Entry{Job: "sync", Status: "success", ExitCode: 0, StartedAt: now})

	req := httptest.NewRequest(http.MethodGet, "/export?format=csv", nil)
	rr := httptest.NewRecorder()
	h.HandleExport(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "text/csv" {
		t.Errorf("expected text/csv, got %s", ct)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "job") || !strings.Contains(body, "sync") {
		t.Errorf("unexpected CSV body: %s", body)
	}
}

func TestHandleExport_DefaultsToJSON(t *testing.T) {
	s := newTestStore(t)
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	h := api.NewHandler(s, log)

	req := httptest.NewRequest(http.MethodGet, "/export", nil)
	rr := httptest.NewRecorder()
	h.HandleExport(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %s", ct)
	}
}
