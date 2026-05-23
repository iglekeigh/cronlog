package store

import (
	"context"
	"fmt"
)

// ListOptions controls filtering and pagination for log entry queries.
type ListOptions struct {
	JobName string // optional filter by job name
	Limit   int
	Offset  int
}

// ListPaginated returns log entries with optional job filtering and pagination.
func (s *Store) ListPaginated(ctx context.Context, opts ListOptions) ([]LogEntry, error) {
	if opts.Limit <= 0 {
		opts.Limit = 50
	}

	var (
		rows interface{ Close() error }
		err  error
	)

	if opts.JobName != "" {
		rows, err = s.db.QueryContext(ctx,
			`SELECT id, job_name, started_at, duration_ms, exit_code, output
			   FROM log_entries
			  WHERE job_name = ?
			  ORDER BY started_at DESC
			  LIMIT ? OFFSET ?`,
			opts.JobName, opts.Limit, opts.Offset,
		)
	} else {
		rows, err = s.db.QueryContext(ctx,
			`SELECT id, job_name, started_at, duration_ms, exit_code, output
			   FROM log_entries
			  ORDER BY started_at DESC
			  LIMIT ? OFFSET ?`,
			opts.Limit, opts.Offset,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("store: list paginated: %w", err)
	}

	defer rows.(*sqlRows).Close()
	return scanEntries(rows.(*sqlRows))
}
