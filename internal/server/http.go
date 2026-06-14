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
	mux.HandleFunc("/api/v1/search", s.authMiddleware(s.handleSearch))
	mux.HandleFunc("/api/v1/search/stream", s.authMiddleware(s.handleSearchStream))
	mux.HandleFunc("/api/v1/pulse", s.authMiddleware(s.handlePulse))
	mux.HandleFunc("/api/v1/resolve", s.authMiddleware(s.handleResolve))
	mux.HandleFunc("/api/v1/alchemist", s.authMiddleware(s.handleAlchemist))
	mux.HandleFunc("/api/v1/alchemist/swarm", s.authMiddleware(s.handleAlchemistSwarm))
	mux.HandleFunc("/api/v1/watch", s.authMiddleware(s.handleWatch))
	mux.HandleFunc("/api/v1/vbrief", s.authMiddleware(s.handleVideoBrief))
	mux.HandleFunc("/api/v1/vprompt", s.authMiddleware(s.handleVideoPrompt))

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

// Shutdown gracefully shuts down the server.
func (s *HTTPServer) Shutdown(ctx context.Context) error {
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}
