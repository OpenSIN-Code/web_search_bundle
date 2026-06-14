// SPDX-License-Identifier: MIT
// Purpose: HTTP handlers for alchemist and swarm endpoints.
// Docs: alchemist_handler.doc.md
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
)

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
		"report":          report,
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
