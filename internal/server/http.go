// Purpose: HTTP REST API server for sin-websearch.
// Docs: internal/server/http.doc.md
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/OpenSIN-Code/web_search_bundle/internal/alchemist"
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
	mux.HandleFunc("/api/v1/search", s.handleSearch)
	mux.HandleFunc("/api/v1/search/stream", s.handleSearchStream)
	mux.HandleFunc("/api/v1/pulse", s.handlePulse)
	mux.HandleFunc("/api/v1/resolve", s.handleResolve)
	mux.HandleFunc("/api/v1/alchemist", s.handleAlchemist)
	mux.HandleFunc("/api/v1/alchemist/swarm", s.handleAlchemistSwarm)

	s.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.cfg.HTTPPort),
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

func (s *HTTPServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().UTC(),
	})
}

func (s *HTTPServer) handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Query string `json:"query"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if req.Query == "" {
		http.Error(w, "query required", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	res, err := s.orchestrator.Search(ctx, req.Query)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (s *HTTPServer) handlePulse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Topic string `json:"topic"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	res, err := s.orchestrator.Pulse(ctx, req.Topic)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (s *HTTPServer) handleResolve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	entity, err := s.resolver.Resolve(ctx, req.Name)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, entity)
}

func (s *HTTPServer) handleSearchStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Query string `json:"query"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if req.Query == "" {
		http.Error(w, "query required", http.StatusBadRequest)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)

	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	stream, err := s.orchestrator.SearchStream(ctx, req.Query)
	if err != nil {
		writeSSE(w, "error", map[string]string{"error": err.Error()})
		flusher.Flush()
		return
	}

	writeSSE(w, "start", map[string]string{"query": req.Query})
	flusher.Flush()

	for result := range stream {
		data, _ := json.Marshal(result)
		writeSSE(w, "result", json.RawMessage(data))
		flusher.Flush()
	}

	writeSSE(w, "done", map[string]string{})
	flusher.Flush()
}

func writeSSE(w http.ResponseWriter, event string, data interface{}) {
	payload, _ := json.Marshal(data)
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, payload)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func redact(s string) string {
	if len(s) <= 8 {
		return "***"
	}
	return s[:4] + strings.Repeat("*", len(s)-8) + s[len(s)-4:]
}

// AlchemistRequest configures a single autonomous research loop.
type AlchemistRequest struct {
	RepoPath       string `json:"repo_path"`
	ProgramFile    string `json:"program_file"`
	TargetFile     string `json:"target_file"`
	MetricName     string `json:"metric_name"`
	MetricRegex    string `json:"metric_regex"`
	HigherIsBetter bool   `json:"higher_is_better"`
	RunCmd         string `json:"run_cmd"`
	MaxExperiments int    `json:"max_experiments"`
	Budget         string `json:"budget"`
	Runtime        string `json:"runtime"`
	Safety         string `json:"safety"`
}

// AlchemistSwarmRequest configures a multi-strategy parallel run.
type AlchemistSwarmRequest struct {
	AlchemistRequest
	Strategies []string `json:"strategies"`
	MaxWorkers int      `json:"max_workers"`
	FirstWin   bool     `json:"first_win"`
}

func (s *HTTPServer) handleAlchemist(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AlchemistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request: " + err.Error()})
		return
	}

	cfg, err := buildAlchemistConfig(req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), cfg.MaxRuntime)
	defer cancel()

	daemon, err := alchemist.NewDaemon(cfg)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "init daemon: " + err.Error()})
		return
	}
	defer daemon.Close()

	report, err := daemon.Run(ctx)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "run: " + err.Error()})
		return
	}

	md, _ := report.RenderMarkdown()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"report":       report,
		"report_markdown": md,
	})
}

func (s *HTTPServer) handleAlchemistSwarm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AlchemistSwarmRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request: " + err.Error()})
		return
	}

	baseCfg, err := buildAlchemistConfig(req.AlchemistRequest)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	// Safety: cap total runtime to avoid runaway HTTP requests.
	if baseCfg.MaxRuntime == 0 {
		var cancel context.CancelFunc
		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Hour)
		defer cancel()
		r = r.WithContext(ctx)
	}

	if len(req.Strategies) == 0 {
		req.Strategies = []string{"conservative", "aggressive", "creative", "minimal"}
	}

	swarmCfg := alchemist.SwarmConfig{
		BaseConfig: baseCfg,
		Strategies: req.Strategies,
		MaxWorkers: req.MaxWorkers,
		FirstWin:   req.FirstWin,
		SharedDB:   true,
	}

	swarm, err := alchemist.NewSwarm(swarmCfg)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "init swarm: " + err.Error()})
		return
	}
	defer swarm.Close()

	report, err := swarm.Run(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "run: " + err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"report":          report,
		"report_markdown": report.RenderMarkdown(),
	})
}

func buildAlchemistConfig(req AlchemistRequest) (alchemist.Config, error) {
	cfg := alchemist.Config{
		ProgramFile:    req.ProgramFile,
		TargetFile:     req.TargetFile,
		MetricName:     req.MetricName,
		MetricRegex:    req.MetricRegex,
		HigherIsBetter: req.HigherIsBetter,
		RunCmd:         []string{"sh", "-c", req.RunCmd},
		MaxExperiments: req.MaxExperiments,
		Safety:         alchemist.SafetyMode(req.Safety),
	}

	if cfg.ProgramFile == "" {
		cfg.ProgramFile = "program.md"
	}
	if cfg.TargetFile == "" {
		cfg.TargetFile = "train.py"
	}
	if cfg.MetricName == "" {
		cfg.MetricName = "metric"
	}
	if cfg.MetricRegex == "" {
		cfg.MetricRegex = `metric:\s*([0-9\.]+)`
	}
	if cfg.Safety == "" {
		cfg.Safety = alchemist.SafetyHeadless
	}
	if cfg.Safety != alchemist.SafetyHeadless && cfg.Safety != alchemist.SafetyAutoCommit && cfg.Safety != alchemist.SafetyInteractive {
		return alchemist.Config{}, fmt.Errorf("invalid safety mode: %s", cfg.Safety)
	}

	if req.RepoPath == "" {
		var err error
		req.RepoPath, err = os.Getwd()
		if err != nil {
			return alchemist.Config{}, err
		}
	}
	cfg.RepoPath = req.RepoPath

	budget := req.Budget
	if budget == "" {
		budget = "30s"
	}
	budgetDur, err := time.ParseDuration(budget)
	if err != nil {
		return alchemist.Config{}, fmt.Errorf("invalid budget: %w", err)
	}
	cfg.Budget = budgetDur

	runtime := req.Runtime
	if runtime == "" {
		runtime = "5m"
	}
	runtimeDur, err := time.ParseDuration(runtime)
	if err != nil {
		return alchemist.Config{}, fmt.Errorf("invalid runtime: %w", err)
	}
	cfg.MaxRuntime = runtimeDur

	if strings.TrimSpace(req.RunCmd) == "" {
		return alchemist.Config{}, fmt.Errorf("run_cmd required")
	}

	return cfg, nil
}
