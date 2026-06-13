// Purpose: Entity resolution: topic → handles, repos, subreddits, hashtags.
// Docs: internal/resolver/entity.doc.md
package resolver

import (
	"context"
	"strings"
	"sync"
)

// ResolvedEntity contains known platform handles for a topic.
type ResolvedEntity struct {
	Query           string
	XHandles        []string
	GitHubUsers     []string
	GitHubRepos     []string
	Subreddits      []string
	YouTubeChannels []string
	TikTokHashtags  []string
	Company         string
}

// EntityResolver maps topics to platform-specific entities.
type EntityResolver struct {
	cache map[string]*ResolvedEntity
	mu    sync.RWMutex
}

// NewEntityResolver creates a resolver with a small built-in knowledge map.
func NewEntityResolver() *EntityResolver {
	return &EntityResolver{cache: make(map[string]*ResolvedEntity)}
}

// Resolve returns the resolved entity for a topic.
func (r *EntityResolver) Resolve(ctx context.Context, topic string) (*ResolvedEntity, error) {
	key := strings.ToLower(topic)
	r.mu.RLock()
	if cached, ok := r.cache[key]; ok {
		r.mu.RUnlock()
		return cached, nil
	}
	r.mu.RUnlock()

	entity := &ResolvedEntity{Query: topic}
	var wg sync.WaitGroup
	wg.Add(5)

	go func() { defer wg.Done(); entity.XHandles = resolveXHandles(topic) }()
	go func() { defer wg.Done(); entity.GitHubUsers, entity.GitHubRepos = resolveGitHub(topic) }()
	go func() { defer wg.Done(); entity.Subreddits = resolveSubreddits(topic) }()
	go func() { defer wg.Done(); entity.YouTubeChannels = resolveYouTubeChannels(topic) }()
	go func() { defer wg.Done(); entity.TikTokHashtags = resolveTikTokHashtags(topic) }()

	wg.Wait()

	r.mu.Lock()
	r.cache[key] = entity
	r.mu.Unlock()

	return entity, nil
}

// ExpandQueries turns a resolved entity into concrete search queries.
func (e *ResolvedEntity) ExpandQueries() []string {
	var queries []string
	queries = append(queries, e.Query)
	for _, handle := range e.XHandles {
		queries = append(queries, "from:"+strings.TrimPrefix(handle, "@"))
	}
	for _, sub := range e.Subreddits {
		queries = append(queries, "subreddit:"+sub)
	}
	for _, repo := range e.GitHubRepos {
		queries = append(queries, "repo:"+repo)
	}
	return queries
}

func resolveXHandles(topic string) []string {
	known := map[string][]string{
		"peter steinberger": {"@steipete"},
		"openclaw":          {"@OpenClaw", "@steipete"},
		"kanye west":        {"@kanyewest"},
	}
	return known[strings.ToLower(topic)]
}

func resolveGitHub(topic string) ([]string, []string) {
	known := map[string]struct {
		users []string
		repos []string
	}{
		"openclaw":          {users: []string{"steipete"}, repos: []string{"openclaw/openclaw"}},
		"peter steinberger": {users: []string{"steipete"}},
	}
	if match, ok := known[strings.ToLower(topic)]; ok {
		return match.users, match.repos
	}
	return nil, nil
}

func resolveSubreddits(topic string) []string {
	known := map[string][]string{
		"openclaw":          {"openclaw", "ClaudeCode", "LocalLLaMA"},
		"peter steinberger": {"ClaudeCode"},
		"kanye west":        {"hiphopheads", "Kanye"},
	}
	return known[strings.ToLower(topic)]
}

func resolveYouTubeChannels(topic string) []string {
	return nil
}

func resolveTikTokHashtags(topic string) []string {
	return nil
}
