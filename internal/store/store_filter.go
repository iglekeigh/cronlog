package store

import (
	"time"
)

// Filter holds criteria for querying log entries.
type Filter struct {
	JobName string
	Status  string
	Since   *time.Time
	Until   *time.Time
	Offset  int
	Limit   int
}

// ListFiltered returns entries matching the given filter criteria.
func (s *Store) ListFiltered(f Filter) ([]Entry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []Entry
	for _, e := range s.entries {
		if f.JobName != "" && e.JobName != f.JobName {
			continue
		}
		if f.Status != "" && e.Status != f.Status {
			continue
		}
		if f.Since != nil && e.StartedAt.Before(*f.Since) {
			continue
		}
		if f.Until != nil && e.StartedAt.After(*f.Until) {
			continue
		}
		results = append(results, e)
	}

	total := len(results)
	if f.Offset >= total {
		return []Entry{}, nil
	}
	results = results[f.Offset:]
	if f.Limit > 0 && len(results) > f.Limit {
		results = results[:f.Limit]
	}
	return results, nil
}
