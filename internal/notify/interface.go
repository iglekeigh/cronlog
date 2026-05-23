package notify

// Sender is the interface for sending failure notifications.
// It allows swapping in a mock during tests.
type Sender interface {
	Notify(failure JobFailure) error
}

// Ensure Notifier satisfies Sender at compile time.
var _ Sender = (*Notifier)(nil)
