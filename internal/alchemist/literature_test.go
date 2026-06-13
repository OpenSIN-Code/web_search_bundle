// Purpose: Smoke tests for the LiteratureLoader hypothesis refresh.
// Docs: literature_test.doc.md

package alchemist

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// fakeSinWebsearch returns a script path that prints a valid mission JSON.
func fakeSinWebsearch(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	script := filepath.Join(dir, "sin-websearch")
	body := `#!/bin/sh
if [ "$1" = "mission" ]; then
  cat <<'JSON'
{
  "topic": "test optimization",
  "status": "complete",
  "all_results": [
    {"title": "Fast Training", "url": "https://example.com/1", "source": "reddit", "snippet": "use bigger batches", "upvotes": 120}
  ],
  "synthesis": "bigger batches help",
  "verification": {
    "total_claims": 2,
    "verified": 1,
    "contested": 1,
    "strong_claims": [
      {"text": "Increasing batch size improves throughput", "confidence": 0.85}
    ],
    "contested_claims": [
      {"text": "AdamW beats SGD everywhere", "confidence": 0.45}
    ]
  }
}
JSON
fi
`
	if err := os.WriteFile(script, []byte(body), 0755); err != nil {
		t.Fatal(err)
	}
	return script
}

func TestLiteratureLoaderShouldRefresh(t *testing.T) {
	l := NewLiteratureLoader(t.TempDir())
	l.SetRefreshEvery(3)

	if l.ShouldRefresh(0) {
		t.Error("ShouldRefresh(0) should be false")
	}
	if !l.ShouldRefresh(3) {
		t.Error("ShouldRefresh(3) should be true")
	}
	if l.ShouldRefresh(4) {
		t.Error("ShouldRefresh(4) should be false")
	}
	if !l.ShouldRefresh(6) {
		t.Error("ShouldRefresh(6) should be true")
	}

	l.SetRefreshEvery(0)
	if l.ShouldRefresh(3) {
		t.Error("ShouldRefresh should be disabled when refreshEvery=0")
	}
}

func TestLiteratureLoaderRefresh(t *testing.T) {
	l := NewLiteratureLoader(t.TempDir())
	l.sinWebsearchBin = fakeSinWebsearch(t)
	l.SetProfile("technical-deep-dive")

	res, err := l.Refresh(context.Background(), "test optimization")
	if err != nil {
		t.Fatalf("Refresh failed: %v", err)
	}
	if len(res.NewHypotheses) == 0 {
		t.Fatal("expected new hypotheses")
	}
	if len(res.VerifiedClaims) == 0 {
		t.Fatal("expected verified claims")
	}
	if len(res.ContestedClaims) == 0 {
		t.Fatal("expected contested claims")
	}
	if len(res.TopSources) == 0 {
		t.Fatal("expected top sources")
	}
	if res.TotalClaimsFound != 2 {
		t.Errorf("TotalClaimsFound = %d, want 2", res.TotalClaimsFound)
	}
}

func TestLiteratureLoaderRefreshBinaryNotFound(t *testing.T) {
	l := NewLiteratureLoader(t.TempDir())
	l.sinWebsearchBin = "" // simulate not on PATH

	res, err := l.Refresh(context.Background(), "topic")
	if err == nil {
		t.Fatal("expected error when binary not found")
	}
	if res == nil {
		t.Fatal("expected result even on error")
	}
	if res.Error == "" {
		t.Error("expected error message in result")
	}
}

func TestLiteratureLoaderInjectIntoProgramMD(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "program.md")
	if err := os.WriteFile(path, []byte("## Hypothesis Queue\n\n- [ ] Existing\n\n## Learnings\n\n- Old\n"), 0644); err != nil {
		t.Fatal(err)
	}

	prog, err := LoadProgramMD(path)
	if err != nil {
		t.Fatal(err)
	}

	res := &LiteratureResult{
		NewHypotheses:    []string{"H1", "H2"},
		VerifiedClaims:   []string{"Claim"},
		TotalClaimsFound: 1,
	}
	if err := NewLiteratureLoader(dir).InjectIntoProgramMD(prog, res); err != nil {
		t.Fatal(err)
	}

	if len(prog.Hypotheses()) != 3 {
		t.Errorf("expected 3 hypotheses, got %d", len(prog.Hypotheses()))
	}
	if len(prog.Learnings()) != 2 {
		t.Errorf("expected 2 learnings, got %d", len(prog.Learnings()))
	}
}

func TestLiteratureLoaderInjectEmptyResult(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "program.md")
	if err := os.WriteFile(path, []byte("## Hypothesis Queue\n\n- [ ] Existing\n"), 0644); err != nil {
		t.Fatal(err)
	}
	prog, err := LoadProgramMD(path)
	if err != nil {
		t.Fatal(err)
	}

	if err := NewLiteratureLoader(dir).InjectIntoProgramMD(prog, &LiteratureResult{}); err != nil {
		t.Fatal(err)
	}
	if len(prog.Hypotheses()) != 1 {
		t.Errorf("expected 1 hypothesis, got %d", len(prog.Hypotheses()))
	}
}
