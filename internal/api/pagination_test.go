package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParsePageParams_Defaults(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/entries", nil)
	p := ParsePageParams(req)

	if p.Limit != defaultLimit {
		t.Errorf("expected default limit %d, got %d", defaultLimit, p.Limit)
	}
	if p.Offset != 0 {
		t.Errorf("expected default offset 0, got %d", p.Offset)
	}
}

func TestParsePageParams_CustomValues(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/entries?limit=10&offset=20", nil)
	p := ParsePageParams(req)

	if p.Limit != 10 {
		t.Errorf("expected limit 10, got %d", p.Limit)
	}
	if p.Offset != 20 {
		t.Errorf("expected offset 20, got %d", p.Offset)
	}
}

func TestParsePageParams_ClampMaxLimit(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/entries?limit=9999", nil)
	p := ParsePageParams(req)

	if p.Limit != maxLimit {
		t.Errorf("expected limit clamped to %d, got %d", maxLimit, p.Limit)
	}
}

func TestParsePageParams_InvalidValues(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/entries?limit=abc&offset=-5", nil)
	p := ParsePageParams(req)

	if p.Limit != defaultLimit {
		t.Errorf("expected default limit %d on invalid input, got %d", defaultLimit, p.Limit)
	}
	if p.Offset != 0 {
		t.Errorf("expected default offset 0 on invalid input, got %d", p.Offset)
	}
}

func TestParsePageParams_ZeroLimit(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/entries?limit=0", nil)
	p := ParsePageParams(req)

	if p.Limit != defaultLimit {
		t.Errorf("expected default limit %d for zero limit, got %d", defaultLimit, p.Limit)
	}
}
