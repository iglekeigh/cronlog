package store

import (
	"testing"
	"time"
)

func TestListFiltered_ByJob(t *testing.T) {
	s := newTestStore(t)
	now := time.Now()
	s.Insert(Entry{ID: "1", JobName: "alpha", Status: "success", StartedAt: now})
	s.Insert(Entry{ID: "2", JobName: "beta", Status: "failure", StartedAt: now})
	s.Insert(Entry{ID: "3", JobName: "alpha", Status: "failure", StartedAt: now})

	results, err := s.ListFiltered(Filter{JobName: "alpha"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestListFiltered_ByStatus(t *testing.T) {
	s := newTestStore(t)
	now := time.Now()
	s.Insert(Entry{ID: "1", JobName: "alpha", Status: "success", StartedAt: now})
	s.Insert(Entry{ID: "2", JobName: "beta", Status: "failure", StartedAt: now})

	results, err := s.ListFiltered(Filter{Status: "failure"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 || results[0].ID != "2" {
		t.Errorf("expected entry 2, got %+v", results)
	}
}

func TestListFiltered_BySince(t *testing.T) {
	s := newTestStore(t)
	old := time.Now().Add(-48 * time.Hour)
	recent := time.Now()
	s.Insert(Entry{ID: "1", JobName: "j", Status: "success", StartedAt: old})
	s.Insert(Entry{ID: "2", JobName: "j", Status: "success", StartedAt: recent})

	cutoff := time.Now().Add(-24 * time.Hour)
	results, err := s.ListFiltered(Filter{Since: &cutoff})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 || results[0].ID != "2" {
		t.Errorf("expected only recent entry, got %+v", results)
	}
}

func TestListFiltered_Pagination(t *testing.T) {
	s := newTestStore(t)
	now := time.Now()
	for i := 0; i < 5; i++ {
		s.Insert(Entry{ID: string(rune('a' + i)), JobName: "j", Status: "success", StartedAt: now})
	}

	results, err := s.ListFiltered(Filter{Offset: 2, Limit: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 paginated results, got %d", len(results))
	}
}
