package notify

import (
	"testing"
)

func TestConfig_DefaultPort(t *testing.T) {
	cfg := Config{
		Enabled:  true,
		SMTPHost: "smtp.example.com",
		SMTPPort: 587,
		From:     "cronlog@example.com",
		To:       []string{"ops@example.com"},
	}

	if cfg.SMTPPort != 587 {
		t.Errorf("expected port 587, got %d", cfg.SMTPPort)
	}
	if len(cfg.To) != 1 {
		t.Errorf("expected 1 recipient, got %d", len(cfg.To))
	}
}

func TestConfig_MultipleRecipients(t *testing.T) {
	cfg := Config{
		Enabled: true,
		To:      []string{"a@example.com", "b@example.com", "c@example.com"},
	}

	if len(cfg.To) != 3 {
		t.Errorf("expected 3 recipients, got %d", len(cfg.To))
	}
}

func TestConfig_AuthOptional(t *testing.T) {
	// Notifier should not panic when no credentials are set
	n := New(Config{
		Enabled:  false,
		SMTPHost: "localhost",
		SMTPPort: 25,
		From:     "noreply@local",
		To:       []string{"admin@local"},
		// Username and Password intentionally empty
	}, testLogger())

	if n == nil {
		t.Fatal("expected non-nil Notifier")
	}
}
