package store

import (
	"strings"
	"time"
)

// SearchEntry represents a log entry matching a search query.
type SearchEntry struct {
	ID        int64     `json:"id"`
	JobName   string    `json:"job_name"`
	Status    string    `json:"status"`
	Output    string    `json:"output"`
	StartedAt time.Time `json:"started_at"`
	Duration  float64   `json:"duration_seconds"`
}

// Search returns log entries whose output or job name contains the given query string.
// Results are limited to at most limit rows, offset by offset.
func (s *Store) Search(query string, limit, offset int) ([]SearchEntry, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	q := `
		SELECT id, job_name, status, output, started_at, duration_seconds
		FROM log_entries
		WHERE job_name LIKE ? OR output LIKE ?
		ORDER BY started_at DESC
		LIMIT ? OFFSET ?
	`
	pattern := "%" + strings.ReplaceAll(query, "%", "\\%") + "%"

	rows, err := s.db.Query(q, pattern, pattern, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []SearchEntry
	for rows.Next() {
		var e SearchEntry
		if err := rows.Scan(&e.ID, &e.JobName, &e.Status, &e.Output, &e.StartedAt, &e.Duration); err != nil {
			return nil, err
		}
		results = append(results, e)
	}
	if results == nil {
		results = []SearchEntry{}
	}
	return results, rows.Err()
}
