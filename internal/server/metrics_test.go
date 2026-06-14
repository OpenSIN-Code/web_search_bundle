// SPDX-License-Identifier: MIT
// Purpose: Tests for the /metrics endpoint and Prometheus-style metrics.
// Docs: internal/server/metrics_test.doc.md
package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/OpenSIN-Code/web_search_bundle/internal/config"
)

func TestMetricsEndpoint(t *testing.T) {
	s := NewHTTPServer(&config.Config{DisableRequestLogging: true}, nil)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	// Make a request that the middleware should record.
	resp, err := http.Get(ts.URL + "/health")
	if err != nil {
		t.Fatalf("get health: %v", err)
	}
	resp.Body.Close()

	// Scrape the metrics endpoint.
	resp, err = http.Get(ts.URL + "/metrics")
	if err != nil {
		t.Fatalf("get metrics: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "text/plain") {
		t.Errorf("expected text/plain content type, got %q", ct)
	}

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read metrics body: %v", err)
	}
	text := string(bytes)

	if !strings.Contains(text, "# HELP http_requests_total") {
		t.Error("expected http_requests_total help line")
	}
	if !strings.Contains(text, "# TYPE http_request_duration_seconds histogram") {
		t.Error("expected histogram type line")
	}
	want := `http_requests_total{method="GET",path="/health",status="200"} 1`
	if !strings.Contains(text, want) {
		t.Errorf("expected metrics to contain %q, got:\n%s", want, text)
	}
	if !strings.Contains(text, `http_request_duration_seconds_bucket{method="GET",path="/health",le="+Inf"} 1`) {
		t.Errorf("expected +Inf bucket for /health, got:\n%s", text)
	}
	if !strings.Contains(text, `http_request_duration_seconds_count{method="GET",path="/health"} 1`) {
		t.Errorf("expected count for /health, got:\n%s", text)
	}
}
