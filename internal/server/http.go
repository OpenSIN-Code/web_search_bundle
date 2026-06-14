// SPDX-License-Identifier: MIT
// Purpose: HTTP REST API server wiring and lifecycle.
// Docs: internal/server/http.doc.md
package server

import (
	"context"
	"fmt"
	"net/http"
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
}

// NewHTTPServer creates a new HTTP server.
func NewHTTPServer(cfg *config.Config, orch *orchestrator.Orchestrator) *HTTPServer {
	return &HTTPServer{
		cfg:          cfg,
		orchestrator: orch,
		resolver:     resolver.NewEntityResolver(),
	}
}

// Start runs the HTTP server.
func (s *HTTPServer) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)

	// Rate limit all mutating/expensive API endpoints per client IP.
	rl := s.newRateLimiter()
	mux.HandleFunc("/api/v1/search", rl(s.authMiddleware(s.handleSearch)))
	mux.HandleFunc("/api/v1/search/stream", rl(s.authMiddleware(s.handleSearchStream)))
	mux.HandleFunc("/api/v1/pulse", rl(s.authMiddleware(s.handlePulse)))
	mux.HandleFunc("/api/v1/resolve", rl(s.authMiddleware(s.handleResolve)))
	mux.HandleFunc("/api/v1/alchemist", rl(s.authMiddleware(s.handleAlchemist)))
	mux.HandleFunc("/api/v1/alchemist/swarm", rl(s.authMiddleware(s.handleAlchemistSwarm)))
	mux.HandleFunc("/api/v1/watch", rl(s.authMiddleware(s.handleWatch)))
	mux.HandleFunc("/api/v1/vbrief", rl(s.authMiddleware(s.handleVideoBrief)))
	mux.HandleFunc("/api/v1/vprompt", rl(s.authMiddleware(s.handleVideoPrompt)))

	port := 8787
	if s.cfg != nil && s.cfg.HTTPPort != 0 {
		port = s.cfg.HTTPPort
	}

	s.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      mux,
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
