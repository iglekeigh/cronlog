package retention_test

import (
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/example/cronlog/internal/retention"
)

// fakeDeleter records the cutoff passed to DeleteOlderThan.
type fakeDeleter struct {
	cutoff  time.Time
	deleted int64
	err     error
}

func (f *fakeDeleter) DeleteOlderThan(cutoff time.Time) (int64, error) {
	f.cutoff = cutoff
	return f.deleted, f.err
}

func logger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestApply_NoMaxAge(t *testing.T) {
	d := &fakeDeleter{deleted: 5}
	r := retention.NewRunner(retention.Policy{MaxAge: 0}, d, logger())

	n, err := r.Apply()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0 deletions when MaxAge=0, got %d", n)
	}
	if !d.cutoff.IsZero() {
		t.Error("DeleteOlderThan should not have been called")
	}
}

func TestApply_WithMaxAge(t *testing.T) {
	d := &fakeDeleter{deleted: 3}
	maxAge := 24 * time.Hour
	r := retention.NewRunner(retention.Policy{MaxAge: maxAge}, d, logger())

	before := time.Now().Add(-maxAge)
	n, err := r.Apply()
	after := time.Now().Add(-maxAge)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 3 {
		t.Errorf("expected 3 deletions, got %d", n)
	}
	if d.cutoff.Before(before) || d.cutoff.After(after) {
		t.Errorf("cutoff %v out of expected range [%v, %v]", d.cutoff, before, after)
	}
}

func TestApply_StoreError(t *testing.T) {
	sentinel := errors.New("db gone")
	d := &fakeDeleter{err: sentinel}
	r := retention.NewRunner(retention.Policy{MaxAge: time.Hour}, d, logger())

	_, err := r.Apply()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

func TestRunEvery_StopsOnClose(t *testing.T) {
	d := &fakeDeleter{deleted: 0}
	r := retention.NewRunner(retention.Policy{MaxAge: time.Hour}, d, logger())

	done := make(chan struct{})
	r.RunEvery(10*time.Millisecond, done)
	time.Sleep(35 * time.Millisecond)
	close(done)
	// Just ensure no panic / deadlock — goroutine exits cleanly.
}
