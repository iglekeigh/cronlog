package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewRouter_Routes(t *testing.T) {
	s := newTestStore(t)
	router := NewRouter(s, testLogger())

	cases := []struct {
		method string
		path   string
		want   int
	}{
		{http.MethodGet, "/entries", http.StatusOK},
		{http.MethodGet, "/stats", http.StatusOK},
		{http.MethodGet, "/healthz", http.StatusOK},
		{http.MethodGet, "/notfound", http.StatusNotFound},
	}

	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != tc.want {
				t.Errorf("%s %s: expected %d, got %d", tc.method, tc.path, tc.want, rec.Code)
			}
		})
	}
}
