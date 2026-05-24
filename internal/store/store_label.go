package store

import (
	"database/sql"
	"fmt"
)

// Label represents a key-value metadata pair attached to a log entry.
type Label struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// ensureLabelTable creates the labels table if it does not exist.
func ensureLabelTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS labels (
			entry_id INTEGER NOT NULL,
			key      TEXT NOT NULL,
			value    TEXT NOT NULL,
			PRIMARY KEY (entry_id, key),
			FOREIGN KEY (entry_id) REFERENCES entries(id) ON DELETE CASCADE
		)`)
	return err
}

// SetLabels replaces all labels for the given entry.
func (s *Store) SetLabels(entryID int64, labels map[string]string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	if _, err := tx.Exec(`DELETE FROM labels WHERE entry_id = ?`, entryID); err != nil {
		return fmt.Errorf("delete labels: %w", err)
	}

	for k, v := range labels {
		if _, err := tx.Exec(`INSERT INTO labels (entry_id, key, value) VALUES (?, ?, ?)`, entryID, k, v); err != nil {
			return fmt.Errorf("insert label %q: %w", k, err)
		}
	}

	return tx.Commit()
}

// GetLabels returns all labels for the given entry.
func (s *Store) GetLabels(entryID int64) (map[string]string, error) {
	rows, err := s.db.Query(`SELECT key, value FROM labels WHERE entry_id = ?`, entryID)
	if err != nil {
		return nil, fmt.Errorf("query labels: %w", err)
	}
	defer rows.Close()

	labels := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		labels[k] = v
	}
	return labels, rows.Err()
}

// ListByLabel returns entry IDs that have a label matching key (and optionally value).
func (s *Store) ListByLabel(key, value string) ([]int64, error) {
	var (
		rows *sql.Rows
		err  error
	)
	if value == "" {
		rows, err = s.db.Query(`SELECT entry_id FROM labels WHERE key = ?`, key)
	} else {
		rows, err = s.db.Query(`SELECT entry_id FROM labels WHERE key = ? AND value = ?`, key, value)
	}
	if err != nil {
		return nil, fmt.Errorf("list by label: %w", err)
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
