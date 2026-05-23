package runner_test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/example/cronlog/internal/runner"
	"github.com/example/cronlog/internal/store"
)

func newTestStore(t *testing.T) *store.Store {
	t.Helper()
	s, err := store.New(":memory:")
	if err != nil {
		t.Fatalf("store.New: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestRun_Success(t *testing.T) {
	s := newTestStore(t)
	r := runner.New(s, testLogger())

	res, err := r.Run(context.Background(), "echo-job", "echo", "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Success {
		t.Errorf("expected success, got exit_code=%d", res.ExitCode)
	}
	if res.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", res.ExitCode)
	}
}

func TestRun_Failure(t *testing.T) {
	s := newTestStore(t)
	r := runner.New(s, testLogger())

	res, err := r.Run(context.Background(), "fail-job", "false")
	if err != nil {
		// err is expected when command exits non-zero
	}
	if res == nil {
		t.Fatal("expected result, got nil")
	}
	if res.Success {
		t.Error("expected failure result")
	}
	if res.ExitCode == 0 {
		t.Error("expected non-zero exit code")
	}
}

func TestRun_PersistsEntry(t *testing.T) {
	s := newTestStore(t)
	r := runner.New(s, testLogger())

	_, err := r.Run(context.Background(), "persist-job", "echo", "stored")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entries, listErr := s.ListByJob(context.Background(), "persist-job")
	if listErr != nil {
		t.Fatalf("ListByJob: %v", listErr)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries))
	}
}

func TestRun_CancelledContext(t *testing.T) {
	s := newTestStore(t)
	r := runner.New(s, testLogger())

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	res, _ := r.Run(ctx, "cancelled-job", "sleep", "10")
	if res != nil && res.Success {
		t.Error("expected non-success for cancelled context")
	}
}
