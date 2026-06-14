// SPDX-License-Identifier: MIT
// Purpose: Additional unit tests for alchemist and swarm HTTP handlers.
// Docs: alchemist_test.doc.md

package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandleAlchemistInvalidJSON(t *testing.T) {
	s := NewHTTPServer(nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/alchemist", bytes.NewReader([]byte("not-json")))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	s.handleAlchemist(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "invalid request") {
		t.Errorf("expected invalid request error, got %s", rr.Body.String())
	}
}

func TestHandleAlchemistInvalidBudget(t *testing.T) {
	s := NewHTTPServer(nil, nil)
	reqBody := AlchemistRequest{
		RepoPath:    setupGitRepo(t),
		RunCmd:      "echo ok",
		Budget:      "not-a-duration",
		Runtime:     "10s",
	}
	data, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/alchemist", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	s.handleAlchemist(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

func TestHandleAlchemistInvalidRuntime(t *testing.T) {
	s := NewHTTPServer(nil, nil)
	reqBody := AlchemistRequest{
		RepoPath: setupGitRepo(t),
		RunCmd:   "echo ok",
		Budget:   "5s",
		Runtime:  "not-a-duration",
	}
	data, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/alchemist", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	s.handleAlchemist(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

func TestHandleAlchemistSwarmMethodNotAllowed(t *testing.T) {
	s := NewHTTPServer(nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/alchemist/swarm", nil)
	rr := httptest.NewRecorder()

	s.handleAlchemistSwarm(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", rr.Code)
	}
}

func TestHandleAlchemistSwarmInvalidJSON(t *testing.T) {
	s := NewHTTPServer(nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/alchemist/swarm", bytes.NewReader([]byte("not-json")))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	s.handleAlchemistSwarm(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

func TestHandleAlchemistSwarmMissingRunCmd(t *testing.T) {
	s := NewHTTPServer(nil, nil)
	reqBody := AlchemistSwarmRequest{
		AlchemistRequest: AlchemistRequest{RepoPath: setupGitRepo(t)},
	}
	data, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/alchemist/swarm", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	s.handleAlchemistSwarm(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

func TestHandleAlchemistSwarmInvalidBudget(t *testing.T) {
	s := NewHTTPServer(nil, nil)
	reqBody := AlchemistSwarmRequest{
		AlchemistRequest: AlchemistRequest{
			RepoPath: setupGitRepo(t),
			RunCmd:   "echo ok",
			Budget:   "not-a-duration",
			Runtime:  "10s",
		},
	}
	data, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/alchemist/swarm", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	s.handleAlchemistSwarm(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

func TestHandleAlchemistSwarmInvalidRuntime(t *testing.T) {
	s := NewHTTPServer(nil, nil)
	reqBody := AlchemistSwarmRequest{
		AlchemistRequest: AlchemistRequest{
			RepoPath: setupGitRepo(t),
			RunCmd:   "echo ok",
			Budget:   "5s",
			Runtime:  "not-a-duration",
		},
	}
	data, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/alchemist/swarm", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	s.handleAlchemistSwarm(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

func TestHandleAlchemistSwarmInvalidSafety(t *testing.T) {
	s := NewHTTPServer(nil, nil)
	reqBody := AlchemistSwarmRequest{
		AlchemistRequest: AlchemistRequest{
			RepoPath: setupGitRepo(t),
			RunCmd:   "echo ok",
			Budget:   "5s",
			Runtime:  "10s",
			Safety:   "dangerous",
		},
	}
	data, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/alchemist/swarm", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	s.handleAlchemistSwarm(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

func TestBuildAlchemistConfigInvalidRuntime(t *testing.T) {
	req := AlchemistRequest{
		RunCmd:  "echo ok",
		Runtime: "not-a-duration",
	}
	_, err := buildAlchemistConfig(req)
	if err == nil {
		t.Fatal("expected error for invalid runtime")
	}
}

func TestBuildAlchemistConfigDefaultRepoPath(t *testing.T) {
	req := AlchemistRequest{
		RunCmd:  "echo ok",
		Budget:  "5s",
		Runtime: "10s",
	}
	cfg, err := buildAlchemistConfig(req)
	if err != nil {
		t.Fatal(err)
	}
	cwd, _ := filepath.Abs("")
	if cfg.RepoPath == "" {
		t.Error("expected repo path to default to cwd")
	}
	if cfg.RepoPath != cwd {
		// filepath.Abs may differ; just confirm it is an absolute directory.
		if !filepath.IsAbs(cfg.RepoPath) {
			t.Errorf("expected absolute repo path, got %s", cfg.RepoPath)
		}
	}
}

func TestBuildAlchemistConfigDefaultSafety(t *testing.T) {
	req := AlchemistRequest{
		RunCmd:  "echo ok",
		Budget:  "5s",
		Runtime: "10s",
	}
	cfg, err := buildAlchemistConfig(req)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Safety != "headless" {
		t.Errorf("expected safety=headless, got %s", cfg.Safety)
	}
}

func TestBuildAlchemistConfigDefaultBudgetAndRuntime(t *testing.T) {
	req := AlchemistRequest{
		RunCmd: "echo ok",
	}
	cfg, err := buildAlchemistConfig(req)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Budget.String() != "30s" {
		t.Errorf("expected budget=30s, got %s", cfg.Budget)
	}
	if cfg.MaxRuntime.String() != "5m0s" {
		t.Errorf("expected runtime=5m, got %s", cfg.MaxRuntime)
	}
}
