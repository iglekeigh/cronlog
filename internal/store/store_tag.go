package store

import (
	"database/sql"
	"fmt"
	"strings"
)

// Tag represents a key-value label attached to a log entry.
type Tag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// SetTags replaces all tags for the given entry ID.
func (s *Store) SetTags(entryID int64, tags []Tag) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	_, err = tx.Exec(`DELETE FROM entry_tags WHERE entry_id = ?`, entryID)
	if err != nil {
		return fmt.Errorf("delete old tags: %w", err)
	}

	for _, t := range tags {
		_, err = tx.Exec(
			`INSERT INTO entry_tags (entry_id, key, value) VALUES (?, ?, ?)`,
			entryID, t.Key, t.Value,
		)
		if err != nil {
			return fmt.Errorf("insert tag %q: %w", t.Key, err)
		}
	}

	return tx.Commit()
}

// GetTags returns all tags for the given entry ID.
func (s *Store) GetTags(entryID int64) ([]Tag, error) {
	rows, err := s.db.Query(
		`SELECT key, value FROM entry_tags WHERE entry_id = ? ORDER BY key`,
		entryID,
	)
	if err != nil {
		return nil, fmt.Errorf("query tags: %w", err)
	}
	defer rows.Close()

	var tags []Tag
	for rows.Next() {
		var t Tag
		if err := rows.Scan(&t.Key, &t.Value); err != nil {
			return nil, fmt.Errorf("scan tag: %w", err)
		}
		tags = append(tags, t)
	}
	return tags, rows.Err()
}

// ListByTag returns entry IDs that have a tag matching key=value.
func (s *Store) ListByTag(key, value string) ([]int64, error) {
	query := `SELECT entry_id FROM entry_tags WHERE key = ?`
	args := []interface{}{key}

	if value != "" {
		query += ` AND value = ?`
		args = append(args, value)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("list by tag: %w", err)
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

// ensureTagTable creates the entry_tags table if it does not exist.
// Called during store initialisation.
func ensureTagTable(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS entry_tags (
		id       INTEGER PRIMARY KEY AUTOINCREMENT,
		entry_id INTEGER NOT NULL,
		key      TEXT    NOT NULL,
		value    TEXT    NOT NULL DEFAULT '',
		CONSTRAINT fk_entry FOREIGN KEY (entry_id) REFERENCES entries(id) ON DELETE CASCADE,
		UNIQUE (entry_id, key)
	)`)
	if err != nil {
		return fmt.Errorf("create entry_tags table: %w", err)
	}
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_tags_key_value ON entry_tags(key, value)`)
	_ = strings.TrimSpace("") // keep import
	return err
}
