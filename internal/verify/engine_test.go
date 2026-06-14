// SPDX-License-Identifier: MIT
// Purpose: Hermetic unit tests for the verification engine.
// Docs: internal/verify/engine_test.doc.md
package verify

import (
	"strings"
	"testing"

	"github.com/OpenSIN-Code/web_search_bundle/internal/engines"
)

func TestNewEngineDefaults(t *testing.T) {
	eng := NewEngine(nil)
	if eng == nil {
		t.Fatal("NewEngine(nil) returned nil")
	}
	if eng.discipline == nil {
		t.Fatal("expected default discipline")
	}
	if eng.discipline.MinSourcesPerClaim != 2 {
		t.Errorf("MinSourcesPerClaim = %d, want 2", eng.discipline.MinSourcesPerClaim)
	}
	if eng.extractor == nil {
		t.Fatal("expected extractor")
	}
}

func TestNewEngineCustomDiscipline(t *testing.T) {
	d := &CitationDiscipline{MinSourcesPerClaim: 1, ConfidenceThreshold: 0.5, FlagContested: false}
	eng := NewEngine(d)
	if eng.discipline != d {
		t.Error("NewEngine did not use provided discipline")
	}
}

func TestEngineVerify(t *testing.T) {
	eng := NewEngine(&CitationDiscipline{
		MinSourcesPerClaim:  1,
		ConfidenceThreshold: 0.5,
		FlagContested:       true,
	})
	results := []engines.Result{
		{Title: "Go 1.25 released", Snippet: "Go 1.25 was released on June 1. It has 5 new features.", Source: "blog", URL: "https://example.com/1"},
		{Title: "Go 1.25 overview", Snippet: "Go 1.25 was released on June 1. It has 5 new features.", Source: "news", URL: "https://example.com/2"},
	}
	report := eng.Verify("Go 1.25", results)
	if report.Topic != "Go 1.25" {
		t.Errorf("topic = %q, want Go 1.25", report.Topic)
	}
	if report.TotalClaims == 0 {
		t.Error("expected claims")
	}
	if report.GeneratedAt.IsZero() {
		t.Error("expected generated timestamp")
	}
	if report.Verified == 0 {
		t.Error("expected at least one verified claim")
	}
	if len(report.StrongClaims) == 0 {
		t.Error("expected strong claims")
	}
}

func TestEngineVerifyNoResults(t *testing.T) {
	eng := NewEngine(DefaultDiscipline())
	report := eng.Verify("empty", nil)
	if report.TotalClaims != 0 {
		t.Errorf("total claims = %d, want 0", report.TotalClaims)
	}
	if report.AvgConfidence != 0 {
		t.Errorf("avg confidence = %v, want 0", report.AvgConfidence)
	}
}

func TestEngineVerifyWeakClaim(t *testing.T) {
	eng := NewEngine(&CitationDiscipline{
		MinSourcesPerClaim:  2,
		ConfidenceThreshold: 0.9,
		FlagContested:       false,
	})
	results := []engines.Result{
		{Title: "Go 1.25 released", Snippet: "Go 1.25 was released on June 1.", Source: "blog", URL: "https://example.com/1"},
	}
	report := eng.Verify("Go 1.25", results)
	if report.Weak == 0 {
		t.Error("expected a weak claim when sources are below threshold")
	}
	if report.Verified != 0 {
		t.Error("expected no verified claim")
	}
}

func TestEngineVerifyUnverified(t *testing.T) {
	eng := NewEngine(&CitationDiscipline{
		MinSourcesPerClaim:  1,
		ConfidenceThreshold: 0.5,
		FlagContested:       false,
	})
	claims := []Claim{
		{Text: "The quick brown fox jumps over the lazy dog", Category: "statement"},
	}
	report := eng.buildReport("topic", claims)
	if report.Unverified != 1 {
		t.Errorf("unverified = %d, want 1", report.Unverified)
	}
	if report.Weak != 0 || report.Verified != 0 {
		t.Error("expected only unverified claims")
	}
}

func TestVerificationReportFormatText(t *testing.T) {
	report := &VerificationReport{
		Topic:       "Go 1.25",
		TotalClaims: 3,
		Verified:    2,
		Weak:        1,
		AvgConfidence: 0.75,
		StrongClaims: []Claim{
			{Text: "Go 1.25 was released", Confidence: 1.0},
		},
	}
	text := report.FormatText()
	if !strings.Contains(text, "VERIFICATION REPORT") {
		t.Error("expected report header")
	}
	if !strings.Contains(text, "Go 1.25") {
		t.Error("expected topic")
	}
	if !strings.Contains(text, "Verified: 2") {
		t.Error("expected verified count")
	}
	if !strings.Contains(text, "Go 1.25 was released") {
		t.Error("expected strong claim text")
	}
}

func TestVerificationReportFormatTextTruncates(t *testing.T) {
	long := strings.Repeat("a", 100)
	report := &VerificationReport{
		Topic:        "t",
		StrongClaims: []Claim{{Text: long, Confidence: 1.0}},
	}
	text := report.FormatText()
	if !strings.Contains(text, "...") {
		t.Error("expected truncated text")
	}
}
