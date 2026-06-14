// SPDX-License-Identifier: MIT
// Purpose: Benchmark entity resolution and query expansion.
// Docs: entity.doc.md
package resolver

import (
	"context"
	"testing"
)

func BenchmarkResolveKnown(b *testing.B) {
	r := NewEntityResolver()
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = r.Resolve(ctx, "openclaw")
	}
}

func BenchmarkResolveUnknown(b *testing.B) {
	r := NewEntityResolver()
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = r.Resolve(ctx, "random unknown topic")
	}
}

func BenchmarkExpandQueries(b *testing.B) {
	e := &ResolvedEntity{
		Query:           "openclaw",
		XHandles:        []string{"@OpenClaw", "@steipete"},
		Subreddits:      []string{"openclaw", "ClaudeCode"},
		GitHubRepos:     []string{"openclaw/openclaw"},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = e.ExpandQueries()
	}
}
