// SPDX-License-Identifier: MIT
// Purpose: Streaming MCP tool that returns search results as NDJSON progress events.
// Docs: internal/mcp/streaming.doc.md
package mcp

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// streamMetaLine is the first NDJSON line emitted, describing the query.
type streamMetaLine struct {
	Type        string `json:"type"`
	Query       string `json:"query"`
	SourcesCount int   `json:"sources_count"`
}

// streamResultLine is one NDJSON line per source as it completes.
type streamResultLine struct {
	Type   string        `json:"type"`
	Source string        `json:"source"`
	Items  []interface{} `json:"items"`
	Error  string        `json:"error"`
}

// streamDoneLine is the final NDJSON line with totals.
type streamDoneLine struct {
	Type         string `json:"type"`
	TotalSources int    `json:"total_sources"`
	TotalItems   int    `json:"total_items"`
}

// RegisterSearchStreamTool registers the websearch_search_stream tool on the server.
// Call this from setup() (or any caller) to wire the streaming tool without editing server.go.
func RegisterSearchStreamTool(s *Server) {
	tool := mcp.Tool{
		Name:        "websearch_search_stream",
		Description: "Multi-source web search returning results as NDJSON progress events (meta → result per source → done). Sources arrive in completion order.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"query": map[string]string{"type": "string", "description": "Search query"},
			},
			Required: []string{"query"},
		},
	}
	s.server.AddTool(tool, s.handleSearchStream)
}

// handleSearchStream calls SearchStream, collects channel results in arrival order,
// and returns them as a single NDJSON document.
func (s *Server) handleSearchStream(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query := argString(req, "query")
	if query == "" {
		return mcp.NewToolResultError("query required"), nil
	}

	ch, err := s.orchestrator.SearchStream(ctx, query)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var sb strings.Builder
	enc := json.NewEncoder(&sb)

	totalSources := 0
	totalItems := 0

	// Meta line — sources_count is unknown until we finish, so we emit it as 0
	// and the done line carries the real totals.
	meta := streamMetaLine{Type: "meta", Query: query, SourcesCount: 0}
	if err := enc.Encode(meta); err != nil {
		return mcp.NewToolResultError("meta encode: " + err.Error()), nil
	}

	// Collect results from the channel as they arrive (completion order).
	for sr := range ch {
		totalSources++
		// Convert []engines.Result to []interface{} for generic JSON marshalling.
		items := make([]interface{}, len(sr.Items))
		for i, item := range sr.Items {
			items[i] = item
		}
		totalItems += len(sr.Items)

		line := streamResultLine{
			Type:   "result",
			Source: sr.Source,
			Items:  items,
			Error:  sr.Error,
		}
		if err := enc.Encode(line); err != nil {
			return mcp.NewToolResultError("result encode: " + err.Error()), nil
		}
	}

	// Done line with final totals.
	done := streamDoneLine{Type: "done", TotalSources: totalSources, TotalItems: totalItems}
	if err := enc.Encode(done); err != nil {
		return mcp.NewToolResultError("done encode: " + err.Error()), nil
	}

	return mcp.NewToolResultText(sb.String()), nil
}
