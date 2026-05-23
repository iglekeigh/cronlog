package store_test

import (
	"os"
	"testing"
	"time"

	"github.com/example/cronlog/internal/store"
)

func newTestStore(t *testing.T) *store.Store {
	t.Helper()
	f, err := os.CreateTemp("", "cronlog-*.db")
	if err != nil {
		t.Fatalf("temp file: %v", err)
	}
	f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })

	s, err := store.New(f.Name())
	if err != nil {
		t.Fatalf("store.New: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestInsertAndList(t *testing.T) {
	s := newTestStore(t)

	now := time.Now().Truncate(time.Second)
	entry := &store.LogEntry{
		JobName:    "backup",
		StartedAt:  now,
		FinishedAt: now.Add(2 * time.Second),
		ExitCode:   0,
		Output:     "done",
		Success:    true,
	}

	id, err := s.Insert(entry)
	if err != nil {
		t.Fatalf("Insert: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive id, got %d", id)
	}

	entries, err := s.ListByJob("backup")
	if err != nil {
		t.Fatalf("ListByJob: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].JobName != "backup" {
		t.Errorf("job name mismatch: %s", entries[0].JobName)
	}
	if entries[0].ExitCode != 0 {
		t.Errorf("exit code mismatch: %d", entries[0].ExitCode)
	}
}

func TestDeleteOlderThan(t *testing.T) {
	s := newTestStore(t)

	old := time.Now().Add(-48 * time.Hour)
	recent := time.Now()

	for _, ts := range []time.Time{old, recent} {
		_, err := s.Insert(&store.LogEntry{
			JobName:    "cleanup",
			StartedAt:  ts,
			FinishedAt: ts.Add(time.Second),
			Success:    true,
		})
		if err != nil {
			t.Fatalf("Insert: %v", err)
		}
	}

	cutoff := time.Now().Add(-24 * time.Hour)
	deleted, err := s.DeleteOlderThan(cutoff)
	if err != nil {
		t.Fatalf("DeleteOlderThan: %v", err)
	}
	if deleted != 1 {
		t.Errorf("expected 1 deleted row, got %d", deleted)
	}

	entries, err := s.ListByJob("cleanup")
	if err != nil {
		t.Fatalf("ListByJob: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 remaining entry, got %d", len(entries))
	}
}

func TestListByJob_Empty(t *testing.T) {
	s := newTestStore(t)
	entries, err := s.ListByJob("nonexistent")
	if err != nil {
		t.Fatalf("ListByJob: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}
