// SPDX-License-Identifier: MIT
// Purpose: Shared types and interfaces for all search engines.
// Docs: internal/engines/common.doc.md
package engines

import (
	"context"
	"time"
)

// Result is a normalized search result from any source.
type Result struct {
	Title      string    `json:"title"`
	URL        string    `json:"url"`
	Snippet    string    `json:"snippet"`
	Source     string    `json:"source"`
	Score      float64   `json:"score"`
	Engagement int       `json:"engagement"`
	Timestamp  time.Time `json:"timestamp"`
	Author     string    `json:"author,omitempty"`
	Subreddit  string    `json:"subreddit,omitempty"`
}

// Engine is the interface implemented by every search source.
type Engine interface {
	// Search runs the engine for the given query and returns normalized results.
	Search(ctx context.Context, query string, limit int) ([]Result, error)
	// Name returns the engine identifier.
	Name() string
}

// SourceContext is passed to engines during initialization.
type SourceContext struct {
	Config map[string]string
	Pool   interface{}
}
