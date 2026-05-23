package runner

import (
	"context"
	"log/slog"
	"os/exec"
	"time"

	"github.com/example/cronlog/internal/store"
)

// Result holds the outcome of a cron job execution.
type Result struct {
	JobName  string
	Command  string
	ExitCode int
	Output   string
	Started  time.Time
	Duration time.Duration
	Success  bool
}

// Runner executes cron job commands and persists results.
type Runner struct {
	store  *store.Store
	logger *slog.Logger
}

// New creates a Runner backed by the given store.
func New(s *store.Store, logger *slog.Logger) *Runner {
	return &Runner{store: s, logger: logger}
}

// Run executes the named command, records the result, and returns it.
func (r *Runner) Run(ctx context.Context, jobName, command string, args ...string) (*Result, error) {
	started := time.Now()

	cmd := exec.CommandContext(ctx, command, args...)
	out, err := cmd.CombinedOutput()

	duration := time.Since(started)
	exitCode := 0
	success := true

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
		success = false
	}

	result := &Result{
		JobName:  jobName,
		Command:  command,
		ExitCode: exitCode,
		Output:   string(out),
		Started:  started,
		Duration: duration,
		Success:  success,
	}

	r.logger.Info("job finished",
		"job", jobName,
		"exit_code", exitCode,
		"duration_ms", duration.Milliseconds(),
	)

	entry := store.Entry{
		JobName:   jobName,
		ExitCode:  exitCode,
		Output:    result.Output,
		StartedAt: started,
		Duration:  duration,
	}
	if insertErr := r.store.Insert(ctx, entry); insertErr != nil {
		r.logger.Error("failed to persist job result", "job", jobName, "err", insertErr)
		return result, insertErr
	}

	return result, nil
}
