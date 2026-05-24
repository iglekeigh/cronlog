package store

import (
	"testing"
	"time"
)

func insertPurgeFixture(t *testing.T, s *Store, job string, exitCode int) {
	t.Helper()
	err := s.Insert(Entry{
		JobName:   job,
		StartedAt: time.Now(),
		ExitCode:  exitCode,
		Output:    "output",
	})
	if err != nil {
		t.Fatalf("insert fixture: %v", err)
	}
}

func TestPurgeByJob_RemovesMatchingEntries(t *testing.T) {
	s := newTestStore(t)

	insertPurgeFixture(t, s, "job-a", 0)
	insertPurgeFixture(t, s, "job-a", 1)
	insertPurgeFixture(t, s, "job-b", 0)

	n, err := s.PurgeByJob("job-a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 2 {
		t.Errorf("expected 2 deleted, got %d", n)
	}

	entries, err := s.ListAll()
	if err != nil {
		t.Fatalf("list all: %v", err)
	}
	if len(entries) != 1 || entries[0].JobName != "job-b" {
		t.Errorf("expected only job-b to remain, got %+v", entries)
	}
}

func TestPurgeByJob_EmptyName(t *testing.T) {
	s := newTestStore(t)
	_, err := s.PurgeByJob("")
	if err == nil {
		t.Error("expected error for empty job name, got nil")
	}
}

func TestPurgeByStatus_RemovesMatchingEntries(t *testing.T) {
	s := newTestStore(t)

	insertPurgeFixture(t, s, "job-a", 0)
	insertPurgeFixture(t, s, "job-b", 1)
	insertPurgeFixture(t, s, "job-c", 1)

	n, err := s.PurgeByStatus(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 2 {
		t.Errorf("expected 2 deleted, got %d", n)
	}

	entries, err := s.ListAll()
	if err != nil {
		t.Fatalf("list all: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 remaining entry, got %d", len(entries))
	}
}

func TestPurgeAll_RemovesEverything(t *testing.T) {
	s := newTestStore(t)

	insertPurgeFixture(t, s, "job-a", 0)
	insertPurgeFixture(t, s, "job-b", 1)

	n, err := s.PurgeAll()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 2 {
		t.Errorf("expected 2 deleted, got %d", n)
	}

	entries, err := s.ListAll()
	if err != nil {
		t.Fatalf("list all: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected no entries after purge, got %d", len(entries))
	}
}

func TestPurgeAll_EmptyStore(t *testing.T) {
	s := newTestStore(t)

	n, err := s.PurgeAll()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0 deleted from empty store, got %d", n)
	}
}
