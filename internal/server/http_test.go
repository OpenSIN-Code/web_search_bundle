// Purpose: Tests for the HTTP REST API server.
// Docs: http_test.doc.md

package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// setupGitRepo creates a temp git repo with a program.md and a train.py.
func setupGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	run := func(name string, args ...string) {
		t.Helper()
		cmd := exec.Command(name, args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%s %v failed: %v\n%s", name, args, err, out)
		}
	}

	run("git", "init")
	run("git", "config", "user.email", "test@example.com")
	run("git", "config", "user.name", "Test")

	if err := os.WriteFile(filepath.Join(dir, "program.md"), []byte("# Program\n\n## Hypothesis Queue\n\n- [ ] HTTP test hypothesis\n\n## Learnings\n\n- Initial\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "train.py"), []byte("print('metric: 0.5')\n"), 0644); err != nil {
		t.Fatal(err)
	}
	run("git", "add", "-A")
	run("git", "commit", "-m", "initial")
	return dir
}

func TestHandleAlchemist(t *testing.T) {
	repo := setupGitRepo(t)

	s := NewHTTPServer(nil, nil)

	reqBody := AlchemistRequest{
		RepoPath:       repo,
		TargetFile:     "train.py",
		MetricName:     "metric",
		MetricRegex:    `metric:\s*([0-9\.]+)`,
		RunCmd:         "echo 'metric: 0.8' > train.py && echo 'metric: 0.8'",
		MaxExperiments: 1,
		Budget:         "5s",
		Runtime:        "10s",
		Safety:         "headless",
	}
	data, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/alchemist", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	s.handleAlchemist(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp struct {
		Report struct {
			WorkBranch string `json:"WorkBranch"`
		} `json:"report"`
		ReportMarkdown string `json:"report_markdown"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v\n%s", err, rr.Body.String())
	}
	if resp.Report.WorkBranch == "" {
		t.Error("expected non-empty work branch in report")
	}
	if !strings.Contains(resp.ReportMarkdown, "Alchemist Morning Report") {
		t.Errorf("expected markdown report, got: %q", resp.ReportMarkdown)
	}
}

func TestHandleAlchemistMissingRunCmd(t *testing.T) {
	s := NewHTTPServer(nil, nil)

	reqBody := AlchemistRequest{RepoPath: setupGitRepo(t)}
	data, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/alchemist", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	s.handleAlchemist(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandleAlchemistSwarm(t *testing.T) {
	repo := setupGitRepo(t)

	s := NewHTTPServer(nil, nil)

	reqBody := AlchemistSwarmRequest{
		AlchemistRequest: AlchemistRequest{
			RepoPath:       repo,
			TargetFile:     "train.py",
			MetricName:     "metric",
			MetricRegex:    `metric:\s*([0-9\.]+)`,
			RunCmd:         "echo 'metric: 0.8' > train.py && echo 'metric: 0.8'",
			MaxExperiments: 1,
			Budget:         "5s",
			Runtime:        "20s",
			Safety:         "headless",
		},
		Strategies: []string{"minimal"},
		MaxWorkers: 1,
	}
	data, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/alchemist/swarm", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	s.handleAlchemistSwarm(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp struct {
		Report struct {
			Workers []struct {
				Strategy struct {
					Name string `json:"name"`
				} `json:"Strategy"`
			} `json:"Workers"`
		} `json:"report"`
		ReportMarkdown string `json:"report_markdown"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v\n%s", err, rr.Body.String())
	}
	if len(resp.Report.Workers) != 1 {
		t.Errorf("expected 1 worker, got %d", len(resp.Report.Workers))
	}
	if len(resp.Report.Workers) > 0 && resp.Report.Workers[0].Strategy.Name != "minimal" {
		t.Errorf("expected strategy minimal, got %s", resp.Report.Workers[0].Strategy.Name)
	}
	if !strings.Contains(resp.ReportMarkdown, "Alchemist Swarm Report") {
		t.Error("expected swarm markdown report")
	}
}

func TestHandleAlchemistMethodNotAllowed(t *testing.T) {
	s := NewHTTPServer(nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/alchemist", nil)
	rr := httptest.NewRecorder()
	s.handleAlchemist(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rr.Code)
	}
}

func TestBuildAlchemistConfigDefaults(t *testing.T) {
	req := AlchemistRequest{
		RunCmd:         "echo ok",
		MaxExperiments: 3,
		Budget:         "30s",
		Runtime:        "5m",
	}
	cfg, err := buildAlchemistConfig(req)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ProgramFile != "program.md" {
		t.Errorf("program file = %q, want program.md", cfg.ProgramFile)
	}
	if cfg.TargetFile != "train.py" {
		t.Errorf("target file = %q, want train.py", cfg.TargetFile)
	}
	if cfg.MetricName != "metric" {
		t.Errorf("metric name = %q, want metric", cfg.MetricName)
	}
	if cfg.Safety != "headless" {
		t.Errorf("safety = %q, want headless", cfg.Safety)
	}
}

func TestBuildAlchemistConfigInvalidSafety(t *testing.T) {
	req := AlchemistRequest{
		RunCmd: "echo ok",
		Safety: "dangerous",
	}
	_, err := buildAlchemistConfig(req)
	if err == nil {
		t.Fatal("expected error for invalid safety")
	}
}

func TestBuildAlchemistConfigInvalidBudget(t *testing.T) {
	req := AlchemistRequest{
		RunCmd: "echo ok",
		Budget: "not-a-duration",
	}
	_, err := buildAlchemistConfig(req)
	if err == nil {
		t.Fatal("expected error for invalid budget")
	}
}

func TestHandleWatchMissingURL(t *testing.T) {
	s := NewHTTPServer(nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/watch", bytes.NewReader([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	s.handleWatch(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestHandleWatchMethodNotAllowed(t *testing.T) {
	s := NewHTTPServer(nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/watch", nil)
	rr := httptest.NewRecorder()
	s.handleWatch(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rr.Code)
	}
}

func TestHandleVideoBriefMissingURL(t *testing.T) {
	s := NewHTTPServer(nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/vbrief", bytes.NewReader([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	s.handleVideoBrief(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestHandleVideoPromptMissingURL(t *testing.T) {
	s := NewHTTPServer(nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/vprompt", bytes.NewReader([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	s.handleVideoPrompt(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}
