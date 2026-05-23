package store

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// LogEntry represents a single cron job log record.
type LogEntry struct {
	ID        int64
	JobName   string
	StartedAt time.Time
	FinishedAt time.Time
	ExitCode  int
	Output    string
	Success   bool
}

// Store manages persistence of cron log entries.
type Store struct {
	db *sql.DB
}

// New opens (or creates) the SQLite database at the given path and
// runs the schema migrations.
func New(path string) (*Store, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("store: open db: %w", err)
	}
	if err := migrate(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("store: migrate: %w", err)
	}
	return &Store{db: db}, nil
}

// Close releases the underlying database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

// Insert persists a new log entry and returns its assigned ID.
func (s *Store) Insert(e *LogEntry) (int64, error) {
	res, err := s.db.Exec(
		`INSERT INTO log_entries (job_name, started_at, finished_at, exit_code, output, success)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		e.JobName, e.StartedAt.UTC(), e.FinishedAt.UTC(), e.ExitCode, e.Output, e.Success,
	)
	if err != nil {
		return 0, fmt.Errorf("store: insert: %w", err)
	}
	return res.LastInsertId()
}

// ListByJob returns all log entries for the given job name, newest first.
func (s *Store) ListByJob(jobName string) ([]LogEntry, error) {
	rows, err := s.db.Query(
		`SELECT id, job_name, started_at, finished_at, exit_code, output, success
		 FROM log_entries WHERE job_name = ? ORDER BY started_at DESC`, jobName,
	)
	if err != nil {
		return nil, fmt.Errorf("store: list: %w", err)
	}
	defer rows.Close()
	return scanRows(rows)
}

// DeleteOlderThan removes entries whose started_at is before the cutoff.
func (s *Store) DeleteOlderThan(cutoff time.Time) (int64, error) {
	res, err := s.db.Exec(
		`DELETE FROM log_entries WHERE started_at < ?`, cutoff.UTC(),
	)
	if err != nil {
		return 0, fmt.Errorf("store: delete old: %w", err)
	}
	return res.RowsAffected()
}

func migrate(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS log_entries (
		id          INTEGER PRIMARY KEY AUTOINCREMENT,
		job_name    TEXT    NOT NULL,
		started_at  DATETIME NOT NULL,
		finished_at DATETIME NOT NULL,
		exit_code   INTEGER NOT NULL DEFAULT 0,
		output      TEXT    NOT NULL DEFAULT '',
		success     BOOLEAN NOT NULL DEFAULT 1
	)`)
	return err
}

func scanRows(rows *sql.Rows) ([]LogEntry, error) {
	var entries []LogEntry
	for rows.Next() {
		var e LogEntry
		if err := rows.Scan(&e.ID, &e.JobName, &e.StartedAt, &e.FinishedAt,
			&e.ExitCode, &e.Output, &e.Success); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}
