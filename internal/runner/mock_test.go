package runner_test

import (
	"context"
	"fmt"

	"github.com/example/cronlog/internal/runner"
)

// mockExecutor satisfies runner.Executor for use in other package tests.
type mockExecutor struct {
	results map[string]*runner.Result
	err     error
}

func newMockExecutor() *mockExecutor {
	return &mockExecutor{results: make(map[string]*runner.Result)}
}

func (m *mockExecutor) Run(_ context.Context, jobName, _ string, _ ...string) (*runner.Result, error) {
	if m.err != nil {
		return nil, m.err
	}
	if r, ok := m.results[jobName]; ok {
		return r, nil
	}
	return nil, fmt.Errorf("mockExecutor: no result registered for job %q", jobName)
}

func (m *mockExecutor) register(jobName string, r *runner.Result) {
	m.results[jobName] = r
}
