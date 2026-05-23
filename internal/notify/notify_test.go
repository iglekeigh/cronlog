package notify

import (
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestNotify_Disabled(t *testing.T) {
	n := New(Config{Enabled: false}, testLogger())
	err := n.Notify(JobFailure{
		JobName:    "backup",
		ExitCode:   1,
		Output:     "disk full",
		OccurredAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("expected no error when disabled, got: %v", err)
	}
}

func TestNotify_SMTPError(t *testing.T) {
	n := New(Config{
		Enabled:  true,
		SMTPHost: "127.0.0.1",
		SMTPPort: 1, // no server listening
		From:     "cronlog@example.com",
		To:       []string{"admin@example.com"},
	}, testLogger())

	err := n.Notify(JobFailure{
		JobName:    "cleanup",
		ExitCode:   2,
		Output:     "permission denied",
		OccurredAt: time.Now(),
	})
	if err == nil {
		t.Fatal("expected error when SMTP unreachable, got nil")
	}
	if !strings.Contains(err.Error(), "notify: send mail") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestBuildBody(t *testing.T) {
	n := New(Config{}, testLogger())
	f := JobFailure{
		JobName:    "report",
		ExitCode:   127,
		Output:     "command not found",
		OccurredAt: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
	}
	body := n.buildBody(f)

	for _, want := range []string{"report", "127", "command not found", "2024-01-15"} {
		if !strings.Contains(body, want) {
			t.Errorf("body missing %q; body: %s", want, body)
		}
	}
}

func TestBuildMessage(t *testing.T) {
	n := New(Config{
		From: "from@example.com",
		To:   []string{"to@example.com"},
	}, testLogger())

	msg := string(n.buildMessage("Test Subject", "Test body"))

	for _, want := range []string{"From: from@example.com", "To: to@example.com", "Subject: Test Subject", "Test body"} {
		if !strings.Contains(msg, want) {
			t.Errorf("message missing %q", want)
		}
	}
}
