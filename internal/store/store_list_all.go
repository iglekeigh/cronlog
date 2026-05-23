package store

import (
	"context"
	"database/sql"
	"fmt"
)

// ListAll returns every log entry in the store without pagination or filtering.
// It is intended for use by aggregation endpoints (e.g. stats).
func (s *Store) ListAll(ctx context.Context) ([]Entry, error) {
	const q = `
		SELECT id, job, exit_code, status, output, started_at, finished_at
		FROM log_entries
		ORDER BY started_at DESC`

	rows, err := s.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("store.ListAll query: %w", err)
	}
	defer rows.Close()

	var entries []Entry
	for rows.Next() {
		var e Entry
		var output sql.NullString
		if err := rows.Scan(
			&e.ID, &e.Job, &e.ExitCode, &e.Status,
			&output, &e.StartedAt, &e.FinishedAt,
		); err != nil {
			return nil, fmt.Errorf("store.ListAll scan: %w", err)
		}
		if output.Valid {
			e.Output = output.String
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store.ListAll rows: %w", err)
	}
	return entries, nil
}
