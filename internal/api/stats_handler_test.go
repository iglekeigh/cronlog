package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cronlog/internal/store"
)

type mockStatsStore struct {
	result *store.StatsResult
	err    error
}

func (m *mockStatsStore) GetStats() (*store.StatsResult, error) {
	return m.result, m.err
}

func TestStatsHandler_Success(t *testing.T) {
	ms := &mockStatsStore{
		result: &store.StatsResult{
			TotalRuns: 5,
			Successes: 4,
			Failures:  1,
			Jobs: []store.JobStats{
				{JobName: "backup", Total: 5, Successes: 4, Failures: 1, LastStatus: "success"},
			},
		},
	}
	h := NewStatsHandler(ms, testLogger())

	req := httptest.NewRequest(http.MethodGet, "/stats", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var result store.StatsResult
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if result.TotalRuns != 5 {
		t.Errorf("expected TotalRuns=5, got %d", result.TotalRuns)
	}
	if len(result.Jobs) != 1 {
		t.Errorf("expected 1 job, got %d", len(result.Jobs))
	}
}

func TestStatsHandler_StoreError(t *testing.T) {
	ms := &mockStatsStore{err: fmt.Errorf("db unavailable")}
	h := NewStatsHandler(ms, testLogger())

	req := httptest.NewRequest(http.MethodGet, "/stats", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
}

func TestStatsHandler_MethodNotAllowed(t *testing.T) {
	ms := &mockStatsStore{result: &store.StatsResult{}}
	h := NewStatsHandler(ms, testLogger())

	req := httptest.NewRequest(http.MethodPost, "/stats", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}
