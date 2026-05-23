package api

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func testMiddlewareLogger(buf *bytes.Buffer) *slog.Logger {
	return slog.New(slog.NewTextHandler(buf, nil))
}

func TestRequestLogger_LogsRequest(t *testing.T) {
	var buf bytes.Buffer
	logger := testMiddlewareLogger(&buf)

	handler := RequestLogger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	logOutput := buf.String()
	if !strings.Contains(logOutput, "http request") {
		t.Errorf("expected log to contain 'http request', got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "/healthz") {
		t.Errorf("expected log to contain path '/healthz', got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "status=200") {
		t.Errorf("expected log to contain 'status=200', got: %s", logOutput)
	}
}

func TestRequestLogger_CapturesStatusCode(t *testing.T) {
	var buf bytes.Buffer
	logger := testMiddlewareLogger(&buf)

	handler := RequestLogger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))

	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !strings.Contains(buf.String(), "status=404") {
		t.Errorf("expected status=404 in log output, got: %s", buf.String())
	}
}

func TestRecoverer_HandlesPanic(t *testing.T) {
	var buf bytes.Buffer
	logger := testMiddlewareLogger(&buf)

	handler := Recoverer(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("something went wrong")
	}))

	req := httptest.NewRequest(http.MethodGet, "/boom", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
	if !strings.Contains(buf.String(), "panic recovered") {
		t.Errorf("expected 'panic recovered' in log, got: %s", buf.String())
	}
}
