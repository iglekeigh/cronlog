package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/user/cronlog/internal/config"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "cronlog-*.yaml")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestLoad_Defaults(t *testing.T) {
	path := writeTempConfig(t, "{}\n")
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Server.Addr != ":8080" {
		t.Errorf("expected default addr :8080, got %q", cfg.Server.Addr)
	}
	if cfg.Retention.MaxAgeDays != 30 {
		t.Errorf("expected default max_age_days 30, got %d", cfg.Retention.MaxAgeDays)
	}
	if cfg.Notify.Enabled {
		t.Error("expected notifications disabled by default")
	}
}

func TestLoad_Override(t *testing.T) {
	yaml := `
server:
  addr: ":9090"
  read_timeout: 30s
storage:
  driver: postgres
  dsn: "postgres://localhost/cronlog"
retention:
  max_age_days: 7
  max_entries_per_job: 100
notify:
  enabled: true
  smtp_host: smtp.example.com
  smtp_port: 465
  from: cron@example.com
  to:
    - ops@example.com
`
	path := writeTempConfig(t, yaml)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Server.Addr != ":9090" {
		t.Errorf("expected :9090, got %q", cfg.Server.Addr)
	}
	if cfg.Server.ReadTimeout != 30*time.Second {
		t.Errorf("expected 30s read timeout, got %v", cfg.Server.ReadTimeout)
	}
	if cfg.Storage.Driver != "postgres" {
		t.Errorf("expected postgres driver, got %q", cfg.Storage.Driver)
	}
	if cfg.Retention.MaxAgeDays != 7 {
		t.Errorf("expected max_age_days 7, got %d", cfg.Retention.MaxAgeDays)
	}
	if !cfg.Notify.Enabled {
		t.Error("expected notifications enabled")
	}
	if len(cfg.Notify.To) != 1 || cfg.Notify.To[0] != "ops@example.com" {
		t.Errorf("unexpected notify.to: %v", cfg.Notify.To)
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := config.Load("/nonexistent/path/cronlog.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}
