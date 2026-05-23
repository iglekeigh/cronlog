package runner

import "context"

// Executor is the interface for running a named job command.
type Executor interface {
	Run(ctx context.Context, jobName, command string, args ...string) (*Result, error)
}
