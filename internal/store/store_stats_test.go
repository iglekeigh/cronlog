package store

import (
	"testing"
	"time"
)

func TestGetStats_Empty(t *testing.T) {
	s := newTestStore(t)

	result, err := s.GetStats()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalRuns != 0 {
		t.Errorf("expected 0 total runs, got %d", result.TotalRuns)
	}
	if len(result.Jobs) != 0 {
		t.Errorf("expected 0 jobs, got %d", len(result.Jobs))
	}
}

func TestGetStats_Aggregation(t *testing.T) {
	s := newTestStore(t)
	now := time.Now()

	entries := []LogEntry{
		{JobName: "backup", Status: "success", StartedAt: now, FinishedAt: now, Output: "", ExitCode: 0},
		{JobName: "backup", Status: "failure", StartedAt: now, FinishedAt: now, Output: "err", ExitCode: 1},
		{JobName: "cleanup", Status: "success", StartedAt: now, FinishedAt: now, Output: "", ExitCode: 0},
	}
	for _, e := range entries {
		if err := s.Insert(e); err != nil {
			t.Fatalf("insert: %v", err)
		}
	}

	result, err := s.GetStats()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.TotalRuns != 3 {
		t.Errorf("expected 3 total runs, got %d", result.TotalRuns)
	}
	if result.Successes != 2 {
		t.Errorf("expected 2 successes, got %d", result.Successes)
	}
	if result.Failures != 1 {
		t.Errorf("expected 1 failure, got %d", result.Failures)
	}
	if len(result.Jobs) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(result.Jobs))
	}

	// Jobs are ordered alphabetically: backup, cleanup
	backup := result.Jobs[0]
	if backup.JobName != "backup" {
		t.Errorf("expected job 'backup', got %q", backup.JobName)
	}
	if backup.Total != 2 {
		t.Errorf("expected backup total=2, got %d", backup.Total)
	}
	if backup.Failures != 1 {
		t.Errorf("expected backup failures=1, got %d", backup.Failures)
	}

	cleanup := result.Jobs[1]
	if cleanup.JobName != "cleanup" {
		t.Errorf("expected job 'cleanup', got %q", cleanup.JobName)
	}
	if cleanup.LastStatus != "success" {
		t.Errorf("expected cleanup last_status='success', got %q", cleanup.LastStatus)
	}
}
