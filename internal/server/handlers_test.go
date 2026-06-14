// SPDX-License-Identifier: MIT
// Purpose: Unit tests for search, pulse, resolve and streaming handlers.
// Docs: handlers_test.doc.md

package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/OpenSIN-Code/web_search_bundle/internal/engines"
	"github.com/OpenSIN-Code/web_search_bundle/internal/orchestrator"
)

// mockErrEngine fails every search so handlers can exercise the error path.
type mockErrEngine struct{}

func (mockErrEngine) Name() string { return "mock-err" }

func (mockErrEngine) Search(ctx context.Context, query string, limit int) ([]engines.Result, error) {
	return nil, context.DeadlineExceeded
}

func TestHandleHealth(t *testing.T) {
	s := NewHTTPServer(nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	s.handleHealth(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, `"status":"ok"`) {
		t.Errorf("expected ok status in body, got %s", body)
	}
}

func TestHandleSearchSuccess(t *testing.T) {
	orch := orchestrator.New([]engines.Engine{mockEngine{}})
	s := NewHTTPServer(nil, orch)

	body, _ := json.Marshal(map[string]string{"query": "openclaw"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/search", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	s.handleSearch(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["results"] == nil {
		t.Error("expected results in response")
	}
}

func TestHandleSearchMethodNotAllowed(t *testing.T) {
	s := NewHTTPServer(nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/search", nil)
	rr := httptest.NewRecorder()

	s.handleSearch(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", rr.Code)
	}
}

func TestHandleSearchInvalidJSON(t *testing.T) {
	s := NewHTTPServer(nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/search", bytes.NewReader([]byte("not-json")))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	s.handleSearch(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

func TestHandleSearchEmptyQuery(t *testing.T) {
	s := NewHTTPServer(nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/search", bytes.NewReader([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	s.handleSearch(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

func TestHandleSearchEngineError(t *testing.T) {
	orch := orchestrator.New([]engines.Engine{mockErrEngine{}})
	s := NewHTTPServer(nil, orch)

	body, _ := json.Marshal(map[string]string{"query": "openclaw"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/search", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	s.handleSearch(rr, req)

	// The orchestrator records per-engine errors but still returns a result.
	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["errors"] == nil {
		t.Errorf("expected errors in response body, got %s", rr.Body.String())
	}
}

func TestHandlePulseSuccess(t *testing.T) {
	orch := orchestrator.New([]engines.Engine{mockEngine{}})
	s := NewHTTPServer(nil, orch)

	body, _ := json.Marshal(map[string]string{"topic": "openclaw"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/pulse", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	s.handlePulse(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["results"] == nil {
		t.Error("expected results in response")
	}
}

func TestHandlePulseInvalidJSON(t *testing.T) {
	s := NewHTTPServer(nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/pulse", bytes.NewReader([]byte("not-json")))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	s.handlePulse(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

func TestHandleResolveSuccess(t *testing.T) {
	s := NewHTTPServer(nil, nil)
	body, _ := json.Marshal(map[string]string{"name": "openclaw"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/resolve", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	s.handleResolve(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["Query"] != "openclaw" {
		t.Errorf("expected Query=openclaw, got %v", resp["Query"])
	}
}

func TestHandleResolveInvalidJSON(t *testing.T) {
	s := NewHTTPServer(nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/resolve", bytes.NewReader([]byte("not-json")))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	s.handleResolve(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

func TestHandleResolveMethodNotAllowed(t *testing.T) {
	s := NewHTTPServer(nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/resolve", nil)
	rr := httptest.NewRecorder()

	s.handleResolve(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", rr.Code)
	}
}

func TestHandleSearchStreamSuccess(t *testing.T) {
	orch := orchestrator.New([]engines.Engine{mockEngine{}})
	s := NewHTTPServer(nil, orch)

	body, _ := json.Marshal(map[string]string{"query": "openclaw"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/search/stream", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	s.handleSearchStream(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if ct := rr.Header().Get("Content-Type"); ct != "text/event-stream" {
		t.Errorf("expected text/event-stream, got %s", ct)
	}
	bodyStr := rr.Body.String()
	if !strings.Contains(bodyStr, "event: start") {
		t.Errorf("expected start event, got %s", bodyStr)
	}
	if !strings.Contains(bodyStr, "event: result") {
		t.Errorf("expected result event, got %s", bodyStr)
	}
	if !strings.Contains(bodyStr, "event: done") {
		t.Errorf("expected done event, got %s", bodyStr)
	}
}

func TestHandleSearchStreamEngineError(t *testing.T) {
	orch := orchestrator.New([]engines.Engine{mockErrEngine{}})
	s := NewHTTPServer(nil, orch)

	body, _ := json.Marshal(map[string]string{"query": "openclaw"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/search/stream", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	s.handleSearchStream(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200 (stream starts), got %d", rr.Code)
	}
	bodyStr := rr.Body.String()
	if !strings.Contains(bodyStr, "event: result") {
		t.Errorf("expected result event, got %s", bodyStr)
	}
	if !strings.Contains(bodyStr, "context deadline exceeded") {
		t.Errorf("expected engine error in result event, got %s", bodyStr)
	}
}

func TestHandleSearchStreamMethodNotAllowed(t *testing.T) {
	s := NewHTTPServer(nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/search/stream", nil)
	rr := httptest.NewRecorder()

	s.handleSearchStream(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", rr.Code)
	}
}

func TestHandleSearchStreamInvalidJSON(t *testing.T) {
	s := NewHTTPServer(nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/search/stream", bytes.NewReader([]byte("not-json")))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	s.handleSearchStream(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

func TestHandleSearchStreamEmptyQuery(t *testing.T) {
	s := NewHTTPServer(nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/search/stream", bytes.NewReader([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	s.handleSearchStream(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}
