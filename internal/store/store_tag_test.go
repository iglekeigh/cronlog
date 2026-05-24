package store

import (
	"testing"
)

func TestSetAndGetTags(t *testing.T) {
	s := newTestStore(t)

	entryID := insertTagFixture(t, s, "backup-job", "success")

	tags := []Tag{
		{Key: "env", Value: "prod"},
		{Key: "team", Value: "ops"},
	}
	if err := s.SetTags(entryID, tags); err != nil {
		t.Fatalf("SetTags: %v", err)
	}

	got, err := s.GetTags(entryID)
	if err != nil {
		t.Fatalf("GetTags: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(got))
	}
	if got[0].Key != "env" || got[0].Value != "prod" {
		t.Errorf("unexpected tag[0]: %+v", got[0])
	}
}

func TestSetTags_Replaces(t *testing.T) {
	s := newTestStore(t)
	entryID := insertTagFixture(t, s, "deploy", "success")

	_ = s.SetTags(entryID, []Tag{{Key: "env", Value: "staging"}})
	_ = s.SetTags(entryID, []Tag{{Key: "env", Value: "prod"}})

	got, err := s.GetTags(entryID)
	if err != nil {
		t.Fatalf("GetTags: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 tag after replace, got %d", len(got))
	}
	if got[0].Value != "prod" {
		t.Errorf("expected value prod, got %s", got[0].Value)
	}
}

func TestGetTags_Empty(t *testing.T) {
	s := newTestStore(t)
	entryID := insertTagFixture(t, s, "noop", "success")

	got, err := s.GetTags(entryID)
	if err != nil {
		t.Fatalf("GetTags: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected no tags, got %d", len(got))
	}
}

func TestListByTag(t *testing.T) {
	s := newTestStore(t)
	id1 := insertTagFixture(t, s, "job-a", "success")
	id2 := insertTagFixture(t, s, "job-b", "success")

	_ = s.SetTags(id1, []Tag{{Key: "env", Value: "prod"}})
	_ = s.SetTags(id2, []Tag{{Key: "env", Value: "staging"}})

	ids, err := s.ListByTag("env", "prod")
	if err != nil {
		t.Fatalf("ListByTag: %v", err)
	}
	if len(ids) != 1 || ids[0] != id1 {
		t.Errorf("expected [%d], got %v", id1, ids)
	}
}

func TestListByTag_KeyOnly(t *testing.T) {
	s := newTestStore(t)
	id1 := insertTagFixture(t, s, "job-a", "success")
	id2 := insertTagFixture(t, s, "job-b", "failure")

	_ = s.SetTags(id1, []Tag{{Key: "team", Value: "ops"}})
	_ = s.SetTags(id2, []Tag{{Key: "team", Value: "dev"}})

	ids, err := s.ListByTag("team", "")
	if err != nil {
		t.Fatalf("ListByTag: %v", err)
	}
	if len(ids) != 2 {
		t.Errorf("expected 2 entries with tag 'team', got %d", len(ids))
	}
}

func insertTagFixture(t *testing.T, s *Store, job, status string) int64 {
	t.Helper()
	id, err := s.Insert(Entry{JobName: job, Status: status, Output: "ok"})
	if err != nil {
		t.Fatalf("insertTagFixture: %v", err)
	}
	return id
}
