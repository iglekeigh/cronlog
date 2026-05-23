package api

import (
	"net/http"
	"strconv"
)

const (
	defaultLimit = 50
	maxLimit      = 500
)

// PageParams holds parsed pagination query parameters.
type PageParams struct {
	Limit  int
	Offset int
}

// ParsePageParams extracts and validates pagination parameters from a request.
func ParsePageParams(r *http.Request) PageParams {
	q := r.URL.Query()

	limit := defaultLimit
	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	offset := 0
	if v := q.Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}

	return PageParams{Limit: limit, Offset: offset}
}
