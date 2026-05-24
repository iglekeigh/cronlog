package store

import (
	"fmt"
	"time"
)

// PurgeResult holds the outcome of a purge operation.
type PurgeResult struct {
	Deleted int64
	Duration time.Duration
}

// PurgeByJob removes all log entries for a specific job name.
// Returns the number of rows deleted.
func (s *Store) PurgeByJob(job string) (int64, error) {
	if job == "" {
		return 0, fmt.Errorf("job name must not be empty")
	}

	res, err := s.db.Exec(`DELETE FROM log_entries WHERE job_name = ?`, job)
	if err != nil {
		return 0, fmt.Errorf("purge by job: %w", err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("purge by job rows affected: %w", err)
	}

	return n, nil
}

// PurgeByStatus removes all log entries with the given exit status.
// Returns the number of rows deleted.
func (s *Store) PurgeByStatus(status int) (int64, error) {
	res, err := s.db.Exec(`DELETE FROM log_entries WHERE exit_code = ?`, status)
	if err != nil {
		return 0, fmt.Errorf("purge by status: %w", err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("purge by status rows affected: %w", err)
	}

	return n, nil
}

// PurgeAll removes every log entry from the store.
// Returns the number of rows deleted.
func (s *Store) PurgeAll() (int64, error) {
	res, err := s.db.Exec(`DELETE FROM log_entries`)
	if err != nil {
		return 0, fmt.Errorf("purge all: %w", err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("purge all rows affected: %w", err)
	}

	return n, nil
}
