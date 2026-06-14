// SPDX-License-Identifier: MIT
// Purpose: Hermetic unit tests for claim extraction from search results.
// Docs: internal/verify/extractor_test.doc.md
package verify

import (
	"testing"

	"github.com/OpenSIN-Code/web_search_bundle/internal/engines"
)

func TestNewClaimExtractor(t *testing.T) {
	d := DefaultDiscipline()
	e := NewClaimExtractor(d)
	if e == nil {
		t.Fatal("NewClaimExtractor returned nil")
	}
	if e.discipline != d {
		t.Error("extractor did not store discipline")
	}
}

func TestExtractDedupesSentences(t *testing.T) {
	e := NewClaimExtractor(DefaultDiscipline())
	results := []engines.Result{
		{Title: "Go 125 released", Snippet: "Go 125 was released on June 1.", Source: "blog", URL: "https://example.com/1"},
		{Title: "Go 125 overview", Snippet: "Go 125 was released on June 1.", Source: "news", URL: "https://example.com/2"},
	}
	claims := e.Extract(results)
	found := false
	for _, c := range claims {
		if c.Text == "Go 125 was released on June 1" && len(c.Sources) == 1 {
			found = true
		}
	}
	if !found {
		t.Errorf("expected deduplicated claim with one source, got: %+v", claims)
	}
}

func TestExtractFiltersShortSentences(t *testing.T) {
	e := NewClaimExtractor(DefaultDiscipline())
	results := []engines.Result{
		{Title: "Hi", Snippet: "Hi.", Source: "blog", URL: "https://example.com/1"},
	}
	claims := e.Extract(results)
	if len(claims) != 0 {
		t.Errorf("expected no claims, got %d", len(claims))
	}
}

func TestExtractSkipsNonFactual(t *testing.T) {
	e := NewClaimExtractor(DefaultDiscipline())
	results := []engines.Result{
		{Title: "Story", Snippet: "The quick brown fox jumps over the lazy dog.", Source: "blog", URL: "https://example.com/1"},
	}
	claims := e.Extract(results)
	if len(claims) != 0 {
		t.Errorf("expected no claims, got %d", len(claims))
	}
}

func TestSplitSentences(t *testing.T) {
	got := splitSentences("Go 125 was released today. It has 5 new features! Great news for everyone?")
	if len(got) != 3 {
		t.Fatalf("sentences = %d, want 3: %v", len(got), got)
	}
	want := []string{"Go 125 was released today", "It has 5 new features", "Great news for everyone"}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("sentence[%d] = %q, want %q", i, got[i], w)
		}
	}
}

func TestSplitSentencesIgnoresShort(t *testing.T) {
	got := splitSentences("Hi. This is a much longer sentence that passes the length filter.")
	if len(got) != 1 {
		t.Fatalf("sentences = %d, want 1: %v", len(got), got)
	}
}

func TestIsFactual(t *testing.T) {
	if !isFactual("Go 125 has 5 new features.") {
		t.Error("expected numeric sentence to be factual")
	}
	if !isFactual("Go 125 was released.") {
		t.Error("expected 'was' marker to be factual")
	}
	if !isFactual("There are many features.") {
		t.Error("expected 'are' marker to be factual")
	}
	if isFactual("The quick brown fox jumps over the lazy dog.") {
		t.Error("expected non-factual sentence")
	}
}

func TestCategorize(t *testing.T) {
	if got := categorize("Go 125 has 5 features"); got != "statistic" {
		t.Errorf("categorize = %q, want statistic", got)
	}
	if got := categorize("Go was released today"); got != "statement" {
		t.Errorf("categorize = %q, want statement", got)
	}
}

func TestMergeClaims(t *testing.T) {
	claims := []Claim{
		{Text: "Go 125 was released", Sources: []Citation{{Source: "a"}}},
		{Text: "Go 125 has 5 features", Sources: []Citation{{Source: "b"}}},
		{Text: "Go 125 was released", Sources: []Citation{{Source: "c"}}},
	}
	merged := mergeClaims(claims)
	if len(merged) != 2 {
		t.Fatalf("merged = %d, want 2", len(merged))
	}
	for _, c := range merged {
		if c.Text == "Go 125 was released" && len(c.Sources) != 2 {
			t.Errorf("sources = %d, want 2", len(c.Sources))
		}
	}
}
