package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/user/cronlog/internal/store"
)

func TestHandleStats_Empty(t *testing.T) {
	s := newTestStore(t)
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	h := NewHandler(s, log)

	req := httptest.NewRequest(http.MethodGet, "/stats", nil)
	rw := httptest.NewRecorder()
	h.handleStats(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rw.Code)
	}

	var resp StatsResponse
	if err := json.NewDecoder(rw.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(resp.Jobs) != 0 {
		t.Errorf("expected empty jobs, got %d", len(resp.Jobs))
	}
}

func TestHandleStats_Aggregation(t *testing.T) {
	s := newTestStore(t)
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	h := NewHandler(s, log)

	ctx := context.Background()
	now := time.Now()

	entries := []store.Entry{
		{Job: "backup", ExitCode: 0, Status: "success", StartedAt: now, FinishedAt: now},
		{Job: "backup", ExitCode: 1, Status: "failure", StartedAt: now, FinishedAt: now},
		{Job: "cleanup", ExitCode: 0, Status: "success", StartedAt: now, FinishedAt: now},
	}
	for _, e := range entries {
		if err := s.Insert(ctx, e); err != nil {
			t.Fatalf("insert: %v", err)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/stats", nil)
	rw := httptest.NewRecorder()
	h.handleStats(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rw.Code)
	}

	var resp StatsResponse
	if err := json.NewDecoder(rw.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(resp.Jobs) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(resp.Jobs))
	}

	byJob := make(map[string]JobStats)
	for _, j := range resp.Jobs {
		byJob[j.Job] = j
	}

	backup := byJob["backup"]
	if backup.Total != 2 || backup.Successes != 1 || backup.Failures != 1 {
		t.Errorf("unexpected backup stats: %+v", backup)
	}

	cleanup := byJob["cleanup"]
	if cleanup.Total != 1 || cleanup.Successes != 1 || cleanup.Failures != 0 {
		t.Errorf("unexpected cleanup stats: %+v", cleanup)
	}
}
