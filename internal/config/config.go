package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the application configuration for cronlog.
type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Storage   StorageConfig   `yaml:"storage"`
	Retention RetentionConfig `yaml:"retention"`
	Notify    NotifyConfig    `yaml:"notify"`
}

// ServerConfig defines HTTP server settings.
type ServerConfig struct {
	Addr         string        `yaml:"addr"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

// StorageConfig defines where logs are persisted.
type StorageConfig struct {
	Driver string `yaml:"driver"` // sqlite, postgres
	DSN    string `yaml:"dsn"`
}

// RetentionConfig defines log retention policies.
type RetentionConfig struct {
	MaxAgeDays  int `yaml:"max_age_days"`
	MaxEntriesPerJob int `yaml:"max_entries_per_job"`
}

// NotifyConfig defines failure notification settings.
type NotifyConfig struct {
	Enabled    bool   `yaml:"enabled"`
	SMTPHost   string `yaml:"smtp_host"`
	SMTPPort   int    `yaml:"smtp_port"`
	From       string `yaml:"from"`
	To         []string `yaml:"to"`
}

// Load reads and parses a YAML config file from the given path.
func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("config: open %q: %w", path, err)
	}
	defer f.Close()

	cfg := defaults()
	if err := yaml.NewDecoder(f).Decode(cfg); err != nil {
		return nil, fmt.Errorf("config: decode: %w", err)
	}
	return cfg, nil
}

// defaults returns a Config populated with sensible defaults.
func defaults() *Config {
	return &Config{
		Server: ServerConfig{
			Addr:         ":8080",
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
		},
		Storage: StorageConfig{
			Driver: "sqlite",
			DSN:    "cronlog.db",
		},
		Retention: RetentionConfig{
			MaxAgeDays:       30,
			MaxEntriesPerJob: 500,
		},
		Notify: NotifyConfig{
			Enabled:  false,
			SMTPPort: 587,
		},
	}
}
