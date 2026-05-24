package store

import (
	"testing"
	"time"
)

func insertSearchFixture(t *testing.T, s *Store, jobName, status, output string) {
	t.Helper()
	_, err := s.db.Exec(
		`INSERT INTO log_entries (job_name, status, output, started_at, duration_seconds) VALUES (?, ?, ?, ?, ?)`,
		jobName, status, output, time.Now(), 1.0,
	)
	if err != nil {
		t.Fatalf("insertSearchFixture: %v", err)
	}
}

func TestSearch_NoResults(t *testing.T) {
	s := newTestStore(t)
	insertSearchFixture(t, s, "backup", "success", "all files copied")

	results, err := s.Search("nonexistent", 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestSearch_MatchesJobName(t *testing.T) {
	s := newTestStore(t)
	insertSearchFixture(t, s, "db-backup", "success", "done")
	insertSearchFixture(t, s, "cleanup", "success", "done")

	results, err := s.Search("db-backup", 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].JobName != "db-backup" {
		t.Errorf("unexpected job name: %s", results[0].JobName)
	}
}

func TestSearch_MatchesOutput(t *testing.T) {
	s := newTestStore(t)
	insertSearchFixture(t, s, "nightly", "failure", "error: disk full")
	insertSearchFixture(t, s, "nightly", "success", "completed ok")

	results, err := s.Search("disk full", 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Status != "failure" {
		t.Errorf("expected failure status, got %s", results[0].Status)
	}
}

func TestSearch_Pagination(t *testing.T) {
	s := newTestStore(t)
	for i := 0; i < 5; i++ {
		insertSearchFixture(t, s, "job", "success", "output line")
	}

	results, err := s.Search("output", 2, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results with limit, got %d", len(results))
	}

	results2, err := s.Search("output", 2, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results2) != 2 {
		t.Errorf("expected 2 results with offset, got %d", len(results2))
	}
}
