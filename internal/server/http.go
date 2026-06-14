// SPDX-License-Identifier: MIT
// Purpose: HTTP REST API server wiring and lifecycle.
// Docs: internal/server/http.doc.md
package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/OpenSIN-Code/web_search_bundle/internal/config"
	"github.com/OpenSIN-Code/web_search_bundle/internal/orchestrator"
	"github.com/OpenSIN-Code/web_search_bundle/internal/resolver"
)

// HTTPServer exposes sin-websearch via REST API.
type HTTPServer struct {
	cfg          *config.Config
	orchestrator *orchestrator.Orchestrator
	resolver     *resolver.EntityResolver
	server       *http.Server
	metrics      *metricsCollector
	logger       *slog.Logger
	startedAt    time.Time
}

// NewHTTPServer creates a new HTTP server.
func NewHTTPServer(cfg *config.Config, orch *orchestrator.Orchestrator) *HTTPServer {
	return &HTTPServer{
		cfg:          cfg,
		orchestrator: orch,
		resolver:     resolver.NewEntityResolver(),
		metrics:      newMetricsCollector(),
		logger:       slog.New(slog.NewTextHandler(os.Stdout, nil)),
		startedAt:    time.Now(),
	}
}

// Handler returns the fully wired HTTP handler for the server. It is used by
// Start() and by tests that want to exercise routing/middleware without opening
// a real TCP listener.
func (s *HTTPServer) Handler() http.Handler {
	mux := http.NewServeMux()
	obs := s.observe

	mux.HandleFunc("/health", obs(s.handleHealth))
	mux.HandleFunc("/metrics", obs(s.handleMetrics))

	// Rate limit all mutating/expensive API endpoints per client IP.
	rl := s.newRateLimiter()
	mux.HandleFunc("/api/v1/search", obs(rl(s.authMiddleware(s.handleSearch))))
	mux.HandleFunc("/api/v1/search/stream", obs(rl(s.authMiddleware(s.handleSearchStream))))
	mux.HandleFunc("/api/v1/pulse", obs(rl(s.authMiddleware(s.handlePulse))))
	mux.HandleFunc("/api/v1/resolve", obs(rl(s.authMiddleware(s.handleResolve))))
	mux.HandleFunc("/api/v1/alchemist", obs(rl(s.authMiddleware(s.handleAlchemist))))
	mux.HandleFunc("/api/v1/alchemist/swarm", obs(rl(s.authMiddleware(s.handleAlchemistSwarm))))
	mux.HandleFunc("/api/v1/watch", obs(rl(s.authMiddleware(s.handleWatch))))
	mux.HandleFunc("/api/v1/vbrief", obs(rl(s.authMiddleware(s.handleVideoBrief))))
	mux.HandleFunc("/api/v1/vprompt", obs(rl(s.authMiddleware(s.handleVideoPrompt))))

	return mux
}

// Start runs the HTTP server.
func (s *HTTPServer) Start() error {
	port := 8787
	if s.cfg != nil && s.cfg.HTTPPort != 0 {
		port = s.cfg.HTTPPort
	}

	s.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      s.Handler(),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	return s.server.ListenAndServe()
}

// newRateLimiter creates the per-IP rate limiter based on config defaults.
func (s *HTTPServer) newRateLimiter() func(http.HandlerFunc) http.HandlerFunc {
	rps := 10.0
	burst := 20
	if s.cfg != nil {
		if s.cfg.RateLimitRPS > 0 {
			rps = s.cfg.RateLimitRPS
		}
		if s.cfg.RateLimitBurst > 0 {
			burst = s.cfg.RateLimitBurst
		}
	}
	store := newLimiterStore(rps, burst)
	return store.rateLimitMiddleware
}

// Shutdown gracefully shuts down the server.
func (s *HTTPServer) Shutdown(ctx context.Context) error {
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}
