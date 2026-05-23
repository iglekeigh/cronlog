package store

import (
	"database/sql"
	"fmt"
)

// JobStats holds aggregated statistics for a single job.
type JobStats struct {
	JobName    string `json:"job_name"`
	Total      int    `json:"total"`
	Successes  int    `json:"successes"`
	Failures   int    `json:"failures"`
	LastStatus string `json:"last_status"`
}

// StatsResult holds overall statistics across all jobs.
type StatsResult struct {
	TotalRuns  int        `json:"total_runs"`
	Successes  int        `json:"successes"`
	Failures   int        `json:"failures"`
	Jobs       []JobStats `json:"jobs"`
}

// GetStats returns aggregated run statistics from the store.
func (s *Store) GetStats() (*StatsResult, error) {
	rows, err := s.db.Query(`
		SELECT job_name,
			COUNT(*) AS total,
			SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END) AS successes,
			SUM(CASE WHEN status = 'failure' THEN 1 ELSE 0 END) AS failures,
			(SELECT status FROM log_entries le2
			 WHERE le2.job_name = le.job_name
			 ORDER BY started_at DESC LIMIT 1) AS last_status
		FROM log_entries le
		GROUP BY job_name
		ORDER BY job_name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("store: query stats: %w", err)
	}
	defer rows.Close()

	result := &StatsResult{}
	for rows.Next() {
		var js JobStats
		var lastStatus sql.NullString
		if err := rows.Scan(&js.JobName, &js.Total, &js.Successes, &js.Failures, &lastStatus); err != nil {
			return nil, fmt.Errorf("store: scan stats row: %w", err)
		}
		if lastStatus.Valid {
			js.LastStatus = lastStatus.String
		}
		result.TotalRuns += js.Total
		result.Successes += js.Successes
		result.Failures += js.Failures
		result.Jobs = append(result.Jobs, js)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: stats rows error: %w", err)
	}
	return result, nil
}
