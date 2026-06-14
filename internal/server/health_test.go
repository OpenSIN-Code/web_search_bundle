// SPDX-License-Identifier: MIT
// Purpose: Tests for the expanded /health endpoint.
// Docs: internal/server/health_test.doc.md
package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/OpenSIN-Code/web_search_bundle/internal/config"
)

func TestHandleHealthDirect(t *testing.T) {
	s := NewHTTPServer(nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	s.handleHealth(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	var body map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode health body: %v\n%s", err, rr.Body.String())
	}
	if body["status"] != "ok" {
		t.Errorf("expected status ok, got %v", body["status"])
	}
	if body["uptime"] == "" {
		t.Error("expected non-empty uptime")
	}
	if body["version"] != config.Version {
		t.Errorf("expected version %q, got %v", config.Version, body["version"])
	}
}

func TestHandleHealthViaRouter(t *testing.T) {
	s := NewHTTPServer(&config.Config{DisableRequestLogging: true}, nil)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/health")
	if err != nil {
		t.Fatalf("get health: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Errorf("expected JSON content type, got %q", ct)
	}
	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode health body: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("expected status ok, got %v", body["status"])
	}
	if body["version"] != config.Version {
		t.Errorf("expected version %q, got %v", config.Version, body["version"])
	}
	if _, ok := body["uptime"].(string); !ok {
		t.Errorf("expected uptime string, got %T", body["uptime"])
	}
}
