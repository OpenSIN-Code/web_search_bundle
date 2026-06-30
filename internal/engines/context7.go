// SPDX-License-Identifier: MIT
// Copyright (c) 2024 OpenSIN-Code

package engines

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// context7Engine uses the @anthropics/context7-mcp via npx to resolve library IDs
// and query up-to-date documentation.
type context7Engine struct {
	name    string
	enabled bool
}

func NewContext7Engine() *context7Engine {
	return &context7Engine{
		name:    "context7",
		enabled: true,
	}
}

func (e *context7Engine) Name() string {
	return e.name
}

func (e *context7Engine) IsEnabled() bool {
	return e.enabled
}

func (e *context7Engine) SetEnabled(b bool) {
	e.enabled = b
}

func (e *context7Engine) Search(ctx context.Context, query string, limit int) ([]Result, error) {
	// context7 is a two-step process: first resolve library ID, then query docs.
	// We expose it as a single search with special query syntax:
	// "lib:<library name> <query>" or just "<library name> <query>"
	// If no library is specified, we try to infer from the query.

	libID, searchQuery := e.parseQuery(query)
	if libID == "" {
		// Try to auto-resolve from query
		var err error
		libID, err = e.resolveLibraryID(ctx, query)
		if err != nil || libID == "" {
			return nil, fmt.Errorf("could not resolve library for query: %s", query)
		}
	}

	docs, err := e.queryDocs(ctx, libID, searchQuery)
	if err != nil {
		return nil, err
	}

	results := make([]Result, 0, len(docs))
	for _, d := range docs {
		results = append(results, Result{
			Title:      fmt.Sprintf("%s - %s", libID, d.Title),
			URL:        d.URL,
			Snippet:    d.Snippet,
			Source:     "context7",
			Score:      d.Score,
			Timestamp:  time.Now(),
		})
	}
	return results[:min(limit, len(results))], nil
}

func (e *context7Engine) parseQuery(query string) (libID, searchQuery string) {
	// Support "lib:react useState" or "react useState" format
	parts := strings.Fields(query)
	if len(parts) >= 2 && parts[0] == "lib:" {
		return parts[1], strings.Join(parts[2:], " ")
	}
	// Heuristic: first word might be library name if it's a known one
	knownLibs := map[string]bool{
		"react": true, "nextjs": true, "next.js": true, "vue": true,
		"svelte": true, "prisma": true, "supabase": true, "tailwind": true,
		"django": true, "fastapi": true, "express": true, "nestjs": true,
		"lodash": true, "axios": true, "typescript": true, "go": true,
		"rust": true, "python": true, "node": true, "deno": true, "bun": true,
	}
	if len(parts) > 1 && knownLibs[strings.ToLower(parts[0])] {
		return parts[0], strings.Join(parts[1:], " ")
	}
	return "", query
}

func (e *context7Engine) resolveLibraryID(ctx context.Context, query string) (string, error) {
	// Call context7-mcp resolve-library-id tool via npx
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "npx", "-y", "@anthropics/context7-mcp", "resolve-library-id", query)
	cmd.Env = append(os.Environ(), "NODE_OPTIONS=--max-old-space-size=512")

	out, err := cmd.CombinedOutput()
	if err != nil {
		stderr := string(out)
		if strings.Contains(stderr, "library") && strings.Contains(stderr, "id") {
			for _, line := range strings.Split(stderr, "\n") {
				line = strings.TrimSpace(line)
				if line != "" && !strings.Contains(strings.ToLower(line), "error") {
					return line, nil
				}
			}
		}
		return "", fmt.Errorf("resolve-library-id failed: %w (output: %s)", err, string(out))
	}

	libID := strings.TrimSpace(string(out))
	if libID == "" {
		return "", fmt.Errorf("empty library ID returned")
	}
	return libID, nil
}

func (e *context7Engine) queryDocs(ctx context.Context, libID, topic string) ([]docResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "npx", "-y", "@anthropics/context7-mcp", "query-docs", libID, topic)
	cmd.Env = append(os.Environ(), "NODE_OPTIONS=--max-old-space-size=512")

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("query-docs failed: %w (output: %s)", err, string(out))
	}

	var result struct {
		Docs []docResult `json:"docs"`
	}
	if err := json.Unmarshal(out, &result); err != nil {
		return e.parsePlainTextOutput(string(out)), nil
	}
	return result.Docs, nil
}

type docResult struct {
	Title   string  `json:"title"`
	URL     string  `json:"url"`
	Snippet string  `json:"snippet"`
	Score   float64 `json:"score"`
}

func (e *context7Engine) parsePlainTextOutput(out string) []docResult {
	var results []docResult
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "---") || strings.HasPrefix(line, "===") {
			continue
		}
		results = append(results, docResult{
			Title:   "Documentation",
			URL:     "",
			Snippet: line,
			Score:   0.5,
		})
	}
	return results
}