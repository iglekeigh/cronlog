package api_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/user/cronlog/internal/api"
	"github.com/user/cronlog/internal/runner"
	"github.com/user/cronlog/internal/store"
)

func newReplayLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func newReplayStore(t *testing.T) *store.Store {
	t.Helper()
	s, err := store.New(":memory:")
	if err != nil {
		t.Fatalf("store.New: %v", err)
	}
	return s
}

func TestReplayHandler_MethodNotAllowed(t *testing.T) {
	s := newReplayStore(t)
	h := api.NewReplayHandler(s, nil, newReplayLogger())

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/replay?id=1", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestReplayHandler_MissingID(t *testing.T) {
	s := newReplayStore(t)
	h := api.NewReplayHandler(s, nil, newReplayLogger())

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/replay", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestReplayHandler_InvalidID(t *testing.T) {
	s := newReplayStore(t)
	h := api.NewReplayHandler(s, nil, newReplayLogger())

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/replay?id=abc", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestReplayHandler_EntryNotFound(t *testing.T) {
	s := newReplayStore(t)
	h := api.NewReplayHandler(s, nil, newReplayLogger())

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/replay?id=9999", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestReplayHandler_Success(t *testing.T) {
	s := newReplayStore(t)
	log := newReplayLogger()

	// Insert a seed entry so GetByID succeeds.
	entry, err := s.Insert(context.Background(), store.Entry{
		JobName: "backup",
		Command: "echo hello",
		Status:  "success",
		Output:  "hello",
	})
	if err != nil {
		t.Fatalf("Insert: %v", err)
	}

	r := runner.New(s, log, nil)
	h := api.NewReplayHandler(s, r, log)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/replay?id="+strconv.FormatInt(entry.ID, 10), nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var got store.Entry
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.JobName != "backup" {
		t.Errorf("expected job name backup, got %q", got.JobName)
	}
}
