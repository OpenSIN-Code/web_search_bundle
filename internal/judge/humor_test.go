// SPDX-License-Identifier: MIT
// Purpose: Hermetic unit tests for the humor/virality judge.
// Docs: internal/judge/humor_test.doc.md
package judge

import (
	"testing"
)

func TestNewHumorJudge(t *testing.T) {
	j := NewHumorJudge()
	if j == nil {
		t.Fatal("expected non-nil judge")
	}
	if len(j.viralPatterns) == 0 {
		t.Error("expected viral patterns")
	}
	if len(j.humorSignals) == 0 {
		t.Error("expected humor signals")
	}
}

func TestScoreResultViralityThresholds(t *testing.T) {
	j := NewHumorJudge()
	text := "regular release"
	tests := []struct {
		upvotes int
		want    float64
	}{
		{0, 0.0},
		{50, 0.25},
		{100, 0.5},
		{500, 0.5},
		{501, 0.7},
		{1000, 0.7},
		{1001, 0.9},
		{2000, 0.9},
	}
	for _, tc := range tests {
		score := j.ScoreResult(text, tc.upvotes, 0)
		if score.Virality != tc.want {
			t.Errorf("upvotes=%d virality=%v, want %v", tc.upvotes, score.Virality, tc.want)
		}
	}
}

func TestScoreResultEngagement(t *testing.T) {
	j := NewHumorJudge()
	score := j.ScoreResult("text", 10, 100)
	if score.Engagement != 11 {
		t.Errorf("engagement=%d, want 11", score.Engagement)
	}
}

func TestScoreResultHumor(t *testing.T) {
	j := NewHumorJudge()
	score := j.ScoreResult("lol 😂 this is literally peak", 10, 0)
	if score.Humor <= 0.0 {
		t.Errorf("expected humor > 0, got %v", score.Humor)
	}
	if score.Humor > 1.0 {
		t.Errorf("expected humor capped at 1.0, got %v", score.Humor)
	}
}

func TestScoreResultNoHumor(t *testing.T) {
	j := NewHumorJudge()
	score := j.ScoreResult("quarterly earnings report", 10, 0)
	if score.Humor != 0.0 {
		t.Errorf("expected humor 0, got %v", score.Humor)
	}
	if score.Relevance != 0.5 {
		t.Errorf("expected placeholder relevance 0.5, got %v", score.Relevance)
	}
}

func TestBestTakesSortsAndLimits(t *testing.T) {
	j := NewHumorJudge()
	items := []struct {
		Text    string
		Upvotes int
	}{
		{"just a regular update", 10},
		{"lol this is literally peak", 1000},
		{"another regular post", 50},
	}
	best := j.BestTakes(items, 2)
	if len(best) != 2 {
		t.Fatalf("expected 2 best takes, got %d", len(best))
	}
	if best[0] != "lol this is literally peak" {
		t.Errorf("expected top take to be most viral/humor, got %q", best[0])
	}
}

func TestBestTakesLimitLargerThanItems(t *testing.T) {
	j := NewHumorJudge()
	items := []struct {
		Text    string
		Upvotes int
	}{
		{"mic drop: Go is now faster", 100},
	}
	best := j.BestTakes(items, 5)
	if len(best) != 1 {
		t.Errorf("expected 1 best take, got %d", len(best))
	}
}

func TestBestTakesEmpty(t *testing.T) {
	j := NewHumorJudge()
	best := j.BestTakes(nil, 3)
	if len(best) != 0 {
		t.Errorf("expected 0 best takes, got %d", len(best))
	}
}
