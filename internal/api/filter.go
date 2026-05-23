package api

import (
	"net/http"
	"time"
)

// EntryFilter holds query parameters for filtering log entries.
type EntryFilter struct {
	JobName  string
	Status   string
	Since    *time.Time
	Until    *time.Time
}

// ParseEntryFilter extracts filter parameters from an HTTP request.
func ParseEntryFilter(r *http.Request) EntryFilter {
	q := r.URL.Query()
	filter := EntryFilter{
		JobName: q.Get("job"),
		Status:  q.Get("status"),
	}

	if s := q.Get("since"); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			filter.Since = &t
		}
	}

	if u := q.Get("until"); u != "" {
		if t, err := time.Parse(time.RFC3339, u); err == nil {
			filter.Until = &t
		}
	}

	return filter
}
