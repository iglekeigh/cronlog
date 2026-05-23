package api

import (
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestParseEntryFilter_Empty(t *testing.T) {
	r := &http.Request{URL: &url.URL{RawQuery: ""}}
	f := ParseEntryFilter(r)
	if f.JobName != "" || f.Status != "" || f.Since != nil || f.Until != nil {
		t.Errorf("expected empty filter, got %+v", f)
	}
}

func TestParseEntryFilter_JobAndStatus(t *testing.T) {
	r := &http.Request{URL: &url.URL{RawQuery: "job=backup&status=failure"}}
	f := ParseEntryFilter(r)
	if f.JobName != "backup" {
		t.Errorf("expected job=backup, got %q", f.JobName)
	}
	if f.Status != "failure" {
		t.Errorf("expected status=failure, got %q", f.Status)
	}
}

func TestParseEntryFilter_ValidDates(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	sinceStr := now.Add(-24 * time.Hour).Format(time.RFC3339)
	untilStr := now.Format(time.RFC3339)

	r := &http.Request{URL: &url.URL{RawQuery: "since=" + sinceStr + "&until=" + untilStr}}
	f := ParseEntryFilter(r)

	if f.Since == nil {
		t.Fatal("expected Since to be set")
	}
	if f.Until == nil {
		t.Fatal("expected Until to be set")
	}
}

func TestParseEntryFilter_InvalidDates(t *testing.T) {
	r := &http.Request{URL: &url.URL{RawQuery: "since=not-a-date&until=also-bad"}}
	f := ParseEntryFilter(r)
	if f.Since != nil || f.Until != nil {
		t.Errorf("expected nil dates for invalid input, got since=%v until=%v", f.Since, f.Until)
	}
}
