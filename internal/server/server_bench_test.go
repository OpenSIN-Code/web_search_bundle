// SPDX-License-Identifier: MIT
// Purpose: Benchmark HTTP server hot paths without external network calls.
// Docs: handlers.doc.md
package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/OpenSIN-Code/web_search_bundle/internal/config"
	"github.com/OpenSIN-Code/web_search_bundle/internal/engines"
	"github.com/OpenSIN-Code/web_search_bundle/internal/orchestrator"
)

// mockEngine returns synthetic results without network calls.
type mockEngine struct{}

func (mockEngine) Name() string { return "mock" }

func (mockEngine) Search(ctx context.Context, query string, limit int) ([]engines.Result, error) {
	return []engines.Result{
		{Title: "Result 1", URL: "https://example.com/1", Source: "mock", Engagement: 100, Score: 0.9},
		{Title: "Result 2", URL: "https://example.com/2", Source: "mock", Engagement: 80, Score: 0.8},
	}, nil
}

func BenchmarkAuthMiddlewareNoToken(b *testing.B) {
	s := NewHTTPServer(nil, nil)
	next := func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }
	handler := s.authMiddleware(next)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/search", nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		handler(rr, req)
	}
}

func BenchmarkAuthMiddlewareValidToken(b *testing.B) {
	s := NewHTTPServer(&config.Config{Token: "secret"}, nil)
	next := func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }
	handler := s.authMiddleware(next)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/search", nil)
	req.Header.Set("Authorization", "Bearer secret")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		handler(rr, req)
	}
}

func BenchmarkAuthMiddlewareInvalidToken(b *testing.B) {
	s := NewHTTPServer(&config.Config{Token: "secret"}, nil)
	next := func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }
	handler := s.authMiddleware(next)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/search", nil)
	req.Header.Set("Authorization", "Bearer wrong")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		handler(rr, req)
	}
}

func BenchmarkHandleHealth(b *testing.B) {
	s := NewHTTPServer(nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		s.handleHealth(rr, req)
	}
}

func BenchmarkHandleSearch(b *testing.B) {
	orch := orchestrator.New([]engines.Engine{mockEngine{}})
	s := NewHTTPServer(nil, orch)
	body := map[string]string{"query": "openclaw"}
	data, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/search", bytes.NewReader(data))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		s.handleSearch(rr, req)
	}
}

func BenchmarkHandlePulse(b *testing.B) {
	orch := orchestrator.New([]engines.Engine{mockEngine{}})
	s := NewHTTPServer(nil, orch)
	body := map[string]string{"topic": "openclaw"}
	data, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/pulse", bytes.NewReader(data))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		s.handlePulse(rr, req)
	}
}

func BenchmarkHandleResolve(b *testing.B) {
	s := NewHTTPServer(nil, nil)
	body := map[string]string{"name": "openclaw"}
	data, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/resolve", bytes.NewReader(data))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		s.handleResolve(rr, req)
	}
}

func BenchmarkBuildAlchemistConfig(b *testing.B) {
	req := AlchemistRequest{
		RepoPath:       "/tmp/repo",
		ProgramFile:    "program.md",
		TargetFile:     "train.py",
		MetricName:     "metric",
		MetricRegex:    `metric:\s*([0-9\.]+)`,
		RunCmd:         "python train.py",
		MaxExperiments: 10,
		Budget:         "30s",
		Runtime:        "5m",
		Safety:         "headless",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := buildAlchemistConfig(req)
		if err != nil {
			b.Fatalf("build config: %v", err)
		}
	}
}

func BenchmarkWriteJSON(b *testing.B) {
	payload := map[string]interface{}{
		"status": "ok",
		"time":   time.Now().UTC(),
		"items":  []int{1, 2, 3, 4, 5},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		writeJSON(rr, http.StatusOK, payload)
	}
}

func BenchmarkWriteSSE(b *testing.B) {
	payload := map[string]string{"query": "openclaw"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		writeSSE(rr, "result", payload)
	}
}
