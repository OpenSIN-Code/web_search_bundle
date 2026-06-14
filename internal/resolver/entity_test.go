// SPDX-License-Identifier: MIT
// Purpose: Hermetic unit tests for entity resolution.
// Docs: internal/resolver/entity_test.doc.md
package resolver

import (
	"context"
	"testing"
)

func TestNewEntityResolver(t *testing.T) {
	r := NewEntityResolver()
	if r == nil {
		t.Fatal("expected non-nil resolver")
	}
	if r.cache == nil {
		t.Error("expected cache map to be initialized")
	}
}

func TestResolveKnownTopic(t *testing.T) {
	r := NewEntityResolver()
	entity, err := r.Resolve(context.Background(), "openclaw")
	if err != nil {
		t.Fatalf("Resolve error: %v", err)
	}
	if entity.Query != "openclaw" {
		t.Errorf("query=%q, want openclaw", entity.Query)
	}
	if len(entity.XHandles) == 0 {
		t.Error("expected X handles for known topic")
	}
	if len(entity.GitHubUsers) == 0 {
		t.Error("expected GitHub users for known topic")
	}
	if len(entity.GitHubRepos) == 0 {
		t.Error("expected GitHub repos for known topic")
	}
	if len(entity.Subreddits) == 0 {
		t.Error("expected subreddits for known topic")
	}
}

func TestResolveCaseInsensitive(t *testing.T) {
	r := NewEntityResolver()
	entity, err := r.Resolve(context.Background(), "OpenClaw")
	if err != nil {
		t.Fatalf("Resolve error: %v", err)
	}
	if len(entity.XHandles) == 0 {
		t.Error("expected X handles for case-insensitive match")
	}
}

func TestResolveCachesResult(t *testing.T) {
	r := NewEntityResolver()
	ctx := context.Background()
	first, err := r.Resolve(ctx, "openclaw")
	if err != nil {
		t.Fatalf("Resolve error: %v", err)
	}
	second, err := r.Resolve(ctx, "openclaw")
	if err != nil {
		t.Fatalf("Resolve error: %v", err)
	}
	if first != second {
		t.Error("expected cached entity to be returned on second call")
	}
}

func TestResolveUnknownTopic(t *testing.T) {
	r := NewEntityResolver()
	entity, err := r.Resolve(context.Background(), "random unknown topic")
	if err != nil {
		t.Fatalf("Resolve error: %v", err)
	}
	if entity.Query != "random unknown topic" {
		t.Errorf("query=%q, want original query", entity.Query)
	}
	if len(entity.XHandles) != 0 {
		t.Error("expected no X handles for unknown topic")
	}
	if len(entity.GitHubRepos) != 0 {
		t.Error("expected no GitHub repos for unknown topic")
	}
}

func TestExpandQueries(t *testing.T) {
	e := &ResolvedEntity{
		Query:       "openclaw",
		XHandles:    []string{"@OpenClaw", "@steipete"},
		Subreddits:  []string{"openclaw", "ClaudeCode"},
		GitHubRepos: []string{"openclaw/openclaw"},
	}
	queries := e.ExpandQueries()
	want := []string{"openclaw", "from:OpenClaw", "from:steipete", "subreddit:openclaw", "subreddit:ClaudeCode", "repo:openclaw/openclaw"}
	if len(queries) != len(want) {
		t.Fatalf("expected %d queries, got %d", len(want), len(queries))
	}
	for i, q := range want {
		if queries[i] != q {
			t.Errorf("queries[%d]=%q, want %q", i, queries[i], q)
		}
	}
}

func TestExpandQueriesEmpty(t *testing.T) {
	e := &ResolvedEntity{Query: "empty"}
	queries := e.ExpandQueries()
	if len(queries) != 1 {
		t.Fatalf("expected 1 query, got %d", len(queries))
	}
	if queries[0] != "empty" {
		t.Errorf("query=%q, want empty", queries[0])
	}
}
