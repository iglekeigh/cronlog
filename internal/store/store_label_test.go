package store

import (
	"testing"
	"time"
)

func TestSetAndGetLabels(t *testing.T) {
	s := newTestStore(t)

	id, err := s.Insert(Entry{
		Job:       "label-job",
		Status:    "success",
		Output:    "ok",
		StartedAt: time.Now(),
		Duration:  1,
	})
	if err != nil {
		t.Fatalf("insert: %v", err)
	}

	labels := map[string]string{"env": "prod", "team": "ops"}
	if err := s.SetLabels(id, labels); err != nil {
		t.Fatalf("SetLabels: %v", err)
	}

	got, err := s.GetLabels(id)
	if err != nil {
		t.Fatalf("GetLabels: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 labels, got %d", len(got))
	}
	if got["env"] != "prod" || got["team"] != "ops" {
		t.Errorf("unexpected labels: %v", got)
	}
}

func TestSetLabels_Replaces(t *testing.T) {
	s := newTestStore(t)

	id, err := s.Insert(Entry{Job: "replace-job", Status: "success", StartedAt: time.Now()})
	if err != nil {
		t.Fatalf("insert: %v", err)
	}

	_ = s.SetLabels(id, map[string]string{"a": "1", "b": "2"})
	_ = s.SetLabels(id, map[string]string{"c": "3"})

	got, err := s.GetLabels(id)
	if err != nil {
		t.Fatalf("GetLabels: %v", err)
	}
	if len(got) != 1 || got["c"] != "3" {
		t.Errorf("expected only label c=3, got %v", got)
	}
}

func TestGetLabels_Empty(t *testing.T) {
	s := newTestStore(t)

	id, _ := s.Insert(Entry{Job: "empty-labels", Status: "success", StartedAt: time.Now()})

	got, err := s.GetLabels(id)
	if err != nil {
		t.Fatalf("GetLabels: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty labels, got %v", got)
	}
}

func TestListByLabel_KeyOnly(t *testing.T) {
	s := newTestStore(t)

	id1, _ := s.Insert(Entry{Job: "j1", Status: "success", StartedAt: time.Now()})
	id2, _ := s.Insert(Entry{Job: "j2", Status: "failure", StartedAt: time.Now()})
	_, _ = s.Insert(Entry{Job: "j3", Status: "success", StartedAt: time.Now()})

	_ = s.SetLabels(id1, map[string]string{"region": "us-east"})
	_ = s.SetLabels(id2, map[string]string{"region": "eu-west"})

	ids, err := s.ListByLabel("region", "")
	if err != nil {
		t.Fatalf("ListByLabel: %v", err)
	}
	if len(ids) != 2 {
		t.Errorf("expected 2 entries with label 'region', got %d", len(ids))
	}
}

func TestListByLabel_KeyValue(t *testing.T) {
	s := newTestStore(t)

	id1, _ := s.Insert(Entry{Job: "j1", Status: "success", StartedAt: time.Now()})
	id2, _ := s.Insert(Entry{Job: "j2", Status: "failure", StartedAt: time.Now()})

	_ = s.SetLabels(id1, map[string]string{"env": "prod"})
	_ = s.SetLabels(id2, map[string]string{"env": "staging"})

	ids, err := s.ListByLabel("env", "prod")
	if err != nil {
		t.Fatalf("ListByLabel: %v", err)
	}
	if len(ids) != 1 || ids[0] != id1 {
		t.Errorf("expected entry %d, got %v", id1, ids)
	}
}
