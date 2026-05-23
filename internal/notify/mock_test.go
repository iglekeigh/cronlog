package notify

import (
	"sync"
)

// MockNotifier records calls for use in integration tests across packages.
type MockNotifier struct {
	mu       sync.Mutex
	Failures []JobFailure
	Err      error
}

// Notify records the failure and returns the preset error.
func (m *MockNotifier) Notify(failure JobFailure) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Failures = append(m.Failures, failure)
	return m.Err
}

// Count returns the number of recorded failures.
func (m *MockNotifier) Count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.Failures)
}

// Reset clears recorded failures.
func (m *MockNotifier) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Failures = nil
}
