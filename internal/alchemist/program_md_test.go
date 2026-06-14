// SPDX-License-Identifier: MIT
// Purpose: Smoke tests for the ProgramMD parser.
// Docs: program_md_test.doc.md

package alchemist

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadProgramMD(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "program.md")

	content := `# Test Program

## Hypothesis Queue

- [ ] Increase batch size from 32 to 64
- [ ] Add dropout 0.1
- [ ] Tune learning rate

## Learnings

- Baseline established at 0.42
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	prog, err := LoadProgramMD(path)
	if err != nil {
		t.Fatalf("LoadProgramMD failed: %v", err)
	}

	hypotheses := prog.Hypotheses()
	if len(hypotheses) != 3 {
		t.Fatalf("expected 3 hypotheses, got %d: %v", len(hypotheses), hypotheses)
	}
	if !strings.Contains(hypotheses[0], "batch size") {
		t.Errorf("unexpected first hypothesis: %q", hypotheses[0])
	}

	learnings := prog.Learnings()
	if len(learnings) != 1 {
		t.Fatalf("expected 1 learning, got %d", len(learnings))
	}
	if !strings.Contains(learnings[0], "Baseline") {
		t.Errorf("unexpected learning: %q", learnings[0])
	}
}

func TestProgramMDNextHypothesis(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "program.md")
	content := "## Hypothesis Queue\n\n- [ ] First idea\n- [ ] Second idea\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	prog, err := LoadProgramMD(path)
	if err != nil {
		t.Fatal(err)
	}

	first := prog.NextHypothesis()
	if first != "First idea" {
		t.Errorf("first hypothesis = %q, want %q", first, "First idea")
	}
	if len(prog.Hypotheses()) != 1 {
		t.Errorf("expected 1 remaining hypothesis, got %d", len(prog.Hypotheses()))
	}

	second := prog.NextHypothesis()
	if second != "Second idea" {
		t.Errorf("second hypothesis = %q, want %q", second, "Second idea")
	}

	third := prog.NextHypothesis()
	if third != "" {
		t.Errorf("expected empty hypothesis, got %q", third)
	}
}

func TestProgramMDAddHypothesis(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "program.md")
	content := "## Hypothesis Queue\n\n- [ ] Existing idea\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	prog, err := LoadProgramMD(path)
	if err != nil {
		t.Fatal(err)
	}

	prog.AddHypothesis("New idea")
	if len(prog.Hypotheses()) != 2 {
		t.Fatalf("expected 2 hypotheses after add, got %d", len(prog.Hypotheses()))
	}

	// Persisted file should contain the new hypothesis.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "New idea") {
		t.Errorf("persisted file does not contain new hypothesis")
	}
}

func TestProgramMDAppendLearning(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "program.md")
	content := "## Learnings\n\n- Old learning\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	prog, err := LoadProgramMD(path)
	if err != nil {
		t.Fatal(err)
	}

	prog.AppendLearning("Fresh learning")
	learnings := prog.Learnings()
	if len(learnings) != 2 {
		t.Fatalf("expected 2 learnings, got %d", len(learnings))
	}
	if !strings.Contains(learnings[1], "Fresh learning") {
		t.Errorf("unexpected learning: %q", learnings[1])
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "Fresh learning") {
		t.Errorf("persisted file does not contain fresh learning")
	}
}

func TestLoadProgramMDMissing(t *testing.T) {
	_, err := LoadProgramMD(filepath.Join(t.TempDir(), "missing.md"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
