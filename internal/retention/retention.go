// Package retention implements log entry cleanup based on configured policies.
package retention

import (
	"fmt"
	"log/slog"
	"time"
)

// Deleter is the subset of store.Store used by the retention runner.
type Deleter interface {
	DeleteOlderThan(cutoff time.Time) (int64, error)
}

// Policy holds retention configuration.
type Policy struct {
	// MaxAge is how long to keep log entries. Zero means keep forever.
	MaxAge time.Duration
}

// Runner applies retention policies against the store.
type Runner struct {
	policy Policy
	store  Deleter
	logger *slog.Logger
}

// NewRunner creates a Runner with the given policy and store.
func NewRunner(p Policy, d Deleter, logger *slog.Logger) *Runner {
	if logger == nil {
		logger = slog.Default()
	}
	return &Runner{policy: p, store: d, logger: logger}
}

// Apply runs the retention policy once, deleting stale entries.
// It returns the number of rows removed and any error encountered.
func (r *Runner) Apply() (int64, error) {
	if r.policy.MaxAge <= 0 {
		r.logger.Info("retention: no max-age configured, skipping")
		return 0, nil
	}

	cutoff := time.Now().Add(-r.policy.MaxAge)
	r.logger.Info("retention: pruning entries",
		"max_age", r.policy.MaxAge,
		"cutoff", cutoff.Format(time.RFC3339),
	)

	n, err := r.store.DeleteOlderThan(cutoff)
	if err != nil {
		return 0, fmt.Errorf("retention: apply: %w", err)
	}

	r.logger.Info("retention: pruning complete", "deleted", n)
	return n, nil
}

// RunEvery starts a background goroutine that applies the policy on the
// given interval. It stops when the done channel is closed.
func (r *Runner) RunEvery(interval time.Duration, done <-chan struct{}) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if _, err := r.Apply(); err != nil {
					r.logger.Error("retention: periodic apply failed", "err", err)
				}
			case <-done:
				return
			}
		}
	}()
}
