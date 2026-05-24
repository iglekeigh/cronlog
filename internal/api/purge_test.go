package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

type mockPurgeStore struct {
	purgeByJobFn    func(name string) (int64, error)
	purgeByStatusFn func(status string) (int64, error)
	purgeAllFn      func() (int64, error)
}

func (m *mockPurgeStore) PurgeByJob(name string) (int64, error) {
	return m.purgeByJobFn(name)
}
func (m *mockPurgeStore) PurgeByStatus(status string) (int64, error) {
	return m.purgeByStatusFn(status)
}
func (m *mockPurgeStore) PurgeAll() (int64, error) { return m.purgeAllFn() }

func newPurgeLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func TestPurgeHandler_MethodNotAllowed(t *testing.T) {
	h := NewPurgeHandler(&mockPurgeStore{}, newPurgeLogger())
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/purge", nil))
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}

func TestPurgeHandler_ByJob(t *testing.T) {
	store := &mockPurgeStore{
		purgeByJobFn: func(name string) (int64, error) {
			if name != "backup" {
				t.Errorf("unexpected job name: %s", name)
			}
			return 3, nil
		},
	}
	h := NewPurgeHandler(store, newPurgeLogger())
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodDelete, "/purge?job=backup", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp purgeResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Deleted != 3 {
		t.Errorf("expected 3 deleted, got %d", resp.Deleted)
	}
}

func TestPurgeHandler_ByStatus(t *testing.T) {
	store := &mockPurgeStore{
		purgeByStatusFn: func(status string) (int64, error) { return 5, nil },
	}
	h := NewPurgeHandler(store, newPurgeLogger())
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodDelete, "/purge?status=failure", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestPurgeHandler_All(t *testing.T) {
	store := &mockPurgeStore{
		purgeAllFn: func() (int64, error) { return 10, nil },
	}
	h := NewPurgeHandler(store, newPurgeLogger())
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodDelete, "/purge", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestPurgeHandler_StoreError(t *testing.T) {
	store := &mockPurgeStore{
		purgeAllFn: func() (int64, error) { return 0, errors.New("db error") },
	}
	h := NewPurgeHandler(store, newPurgeLogger())
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodDelete, "/purge", nil))
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}
}
