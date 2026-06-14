// SPDX-License-Identifier: MIT
// Purpose: Benchmark claim extraction and verification pipeline.
// Docs: claim.doc.md
package verify

import (
	"testing"

	"github.com/OpenSIN-Code/web_search_bundle/internal/engines"
)

func sampleResults() []engines.Result {
	return []engines.Result{
		{Title: "Go 1.25 released", Snippet: "Go 1.25 was released on June 1. It has 5 new features.", Source: "blog", URL: "https://example.com/1"},
		{Title: "Go 1.25 overview", Snippet: "Go 1.25 was released on June 1. It has 5 new features.", Source: "news", URL: "https://example.com/2"},
		{Title: "Rust 1.85 released", Snippet: "Rust 1.85 was released on May 20. It has 3 new features.", Source: "blog", URL: "https://example.com/3"},
		{Title: "Another Go feature", Snippet: "Go 1.25 includes improved generics.", Source: "docs", URL: "https://example.com/4"},
	}
}

func BenchmarkClaimExtractorExtract(b *testing.B) {
	e := NewClaimExtractor(DefaultDiscipline())
	results := sampleResults()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = e.Extract(results)
	}
}

func BenchmarkEngineVerify(b *testing.B) {
	eng := NewEngine(DefaultDiscipline())
	results := sampleResults()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = eng.Verify("Go 1.25", results)
	}
}

func BenchmarkSplitSentences(b *testing.B) {
	text := "Go 1.25 was released. It has 5 new features. It improves performance!"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = splitSentences(text)
	}
}

func BenchmarkIsFactual(b *testing.B) {
	s := "Go 1.25 has 5 new features."
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = isFactual(s)
	}
}

func BenchmarkCategorize(b *testing.B) {
	s := "Go 1.25 has 5 new features."
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = categorize(s)
	}
}

func BenchmarkMergeClaims(b *testing.B) {
	claims := []Claim{
		{Text: "Go 1.25 was released", Sources: []Citation{{Source: "a"}}},
		{Text: "Go 1.25 has 5 features", Sources: []Citation{{Source: "b"}}},
		{Text: "Go 1.25 was released", Sources: []Citation{{Source: "c"}}},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mergeClaims(claims)
	}
}
