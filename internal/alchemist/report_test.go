// SPDX-License-Identifier: MIT
// Purpose: Tests for alchemist report rendering.
// Docs: report_test.doc.md
package alchemist

import (
	"strings"
	"testing"
	"time"
)

func TestMorningReportRenderMarkdown(t *testing.T) {
	r := &MorningReport{
		StartTime:   time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
		EndTime:     time.Date(2026, 1, 2, 4, 5, 6, 0, time.UTC),
		Duration:    time.Hour,
		WorkBranch:  "test-branch",
		ProgramFile: "program.md",
		Summary: map[string]any{
			"total_experiments": 5,
			"committed":         2,
			"discarded":         2,
			"errors":            1,
			"success_rate":      0.4,
			"best_delta":        0.1234,
			"total_runtime":     "1h",
		},
		Experiments: []ExperimentRecord{
			{Hypothesis: "h1", Decision: "committed", MetricAfter: 1.0, Delta: 0.1, CommitSHA: "abc1234567890"},
			{Hypothesis: "h2", Decision: "discarded", MetricAfter: 2.0, Delta: -0.2},
		},
		Learnings:   []string{"learning one", "learning two"},
		TopCommits:  []ExperimentRecord{{Hypothesis: "h1", Decision: "committed", MetricAfter: 1.0, Delta: 0.1, CommitSHA: "abc1234567890"}},
		DiffPreview: "+ added line",
	}

	md, err := r.RenderMarkdown()
	if err != nil {
		t.Fatalf("RenderMarkdown: %v", err)
	}
	for _, want := range []string{"Alchemist Morning Report", "test-branch", "program.md", "learning one", "abc1234", "Total experiments"} {
		if !strings.Contains(md, want) {
			t.Errorf("report missing %q", want)
		}
	}
}

func TestMorningReportRenderMarkdownEmpty(t *testing.T) {
	r := &MorningReport{
		StartTime:   time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
		EndTime:     time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
		Duration:    0,
		WorkBranch:  "main",
		ProgramFile: "program.md",
		Summary:     map[string]any{},
	}
	md, err := r.RenderMarkdown()
	if err != nil {
		t.Fatalf("RenderMarkdown: %v", err)
	}
	if !strings.Contains(md, "No successful commits") {
		t.Error("expected empty top-commits message")
	}
}

func TestRecentN(t *testing.T) {
	r := &MorningReport{Experiments: []ExperimentRecord{
		{Hypothesis: "a"},
		{Hypothesis: "b"},
		{Hypothesis: "c"},
	}}
	if len(r.RecentN(2)) != 2 {
		t.Errorf("RecentN(2) = %d, want 2", len(r.RecentN(2)))
	}
	if len(r.RecentN(10)) != 3 {
		t.Errorf("RecentN(10) = %d, want 3", len(r.RecentN(10)))
	}
}

func TestToFloat(t *testing.T) {
	cases := []struct {
		in   any
		want float64
	}{
		{float64(1.5), 1.5},
		{float32(2.5), 2.5},
		{int(3), 3},
		{int64(4), 4},
		{"nope", 0},
	}
	for _, c := range cases {
		if got := toFloat(c.in); got != c.want {
			t.Errorf("toFloat(%v) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestShortSHA(t *testing.T) {
	if got := shortSHA("abcdef1234567890"); got != "abcdef1" {
		t.Errorf("shortSHA = %q, want abcdef1", got)
	}
	if got := shortSHA("abc"); got != "abc" {
		t.Errorf("shortSHA short = %q, want abc", got)
	}
}

func TestRenderSimple(t *testing.T) {
	r := &MorningReport{
		StartTime:   time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
		EndTime:     time.Date(2026, 1, 2, 4, 5, 6, 0, time.UTC),
		Duration:    time.Hour,
		WorkBranch:  "branch",
		ProgramFile: "program.md",
		Summary: map[string]any{
			"total_experiments": 1,
			"committed":         1,
			"discarded":         0,
			"errors":            0,
			"success_rate":      1.0,
			"best_delta":        0.1,
			"total_runtime":     "10m",
		},
		TopCommits:  []ExperimentRecord{{Hypothesis: "h", Decision: "committed", MetricAfter: 1.0, Delta: 0.1, CommitSHA: "sha"}},
		Experiments: []ExperimentRecord{{Hypothesis: "h", Decision: "committed", MetricAfter: 1.0, Delta: 0.1}},
		DiffPreview: "+line",
	}
	s := r.renderSimple()
	for _, want := range []string{"Alchemist Morning Report", "branch", "program.md", "h", "+line"} {
		if !strings.Contains(s, want) {
			t.Errorf("renderSimple missing %q", want)
		}
	}
}
