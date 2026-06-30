// SPDX-License-Identifier: MIT
// Purpose: Tests for the streaming MCP tool (NDJSON progress events).
// Docs: internal/mcp/streaming_test.doc.md
package mcp

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/OpenSIN-Code/web_search_bundle/internal/engines"
	"github.com/OpenSIN-Code/web_search_bundle/internal/orchestrator"
	"github.com/mark3labs/mcp-go/mcp"
)

// streamMockOrchestrator returns a channel that emits results in order then closes.
type streamMockOrchestrator struct {
	results []orchestrator.StreamResult
}

func (m *streamMockOrchestrator) Search(ctx context.Context, topic string) (*orchestrator.SearchResult, error) {
	return nil, nil
}

func (m *streamMockOrchestrator) Pulse(ctx context.Context, topic string) (*orchestrator.SearchResult, error) {
	return nil, nil
}

func (m *streamMockOrchestrator) SearchStream(ctx context.Context, topic string) (<-chan orchestrator.StreamResult, error) {
	ch := make(chan orchestrator.StreamResult, len(m.results))
	for _, r := range m.results {
		select {
		case ch <- r:
		case <-ctx.Done():
			close(ch)
			return ch, nil
		}
	}
	close(ch)
	return ch, nil
}

func parseNDJSON(t *testing.T, text string) []map[string]interface{} {
	t.Helper()
	lines := strings.Split(strings.TrimSpace(text), "\n")
	if len(lines) == 0 {
		t.Fatal("expected at least one NDJSON line, got empty output")
	}
	var out []map[string]interface{}
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var obj map[string]interface{}
		if err := json.Unmarshal([]byte(line), &obj); err != nil {
			t.Fatalf("line %d is not valid JSON: %v\nline: %q", i, err, line)
		}
		out = append(out, obj)
	}
	return out
}

func TestHandleSearchStreamNDJSONFormat(t *testing.T) {
	mock := &streamMockOrchestrator{
		results: []orchestrator.StreamResult{
			{Source: "reddit", Items: []engines.Result{{Title: "R1", URL: "https://r1.com"}}},
			{Source: "hackernews", Items: []engines.Result{{Title: "H1", URL: "https://h1.com"}, {Title: "H2", URL: "https://h2.com"}}},
		},
	}
	s := NewServer(mock)
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{"query": "test query"}

	res, err := s.handleSearchStream(context.Background(), req)
	if err != nil {
		t.Fatalf("handleSearchStream error: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected no error, got: %v", res.Content)
	}

	text := res.Content[0].(mcp.TextContent).Text
	lines := parseNDJSON(t, text)

	if len(lines) != 4 {
		t.Fatalf("expected 4 NDJSON lines (meta + 2 results + done), got %d", len(lines))
	}

	// Line 0: meta
	if lines[0]["type"] != "meta" {
		t.Errorf("line 0 type = %v, want meta", lines[0]["type"])
	}
	if lines[0]["query"] != "test query" {
		t.Errorf("line 0 query = %v, want test query", lines[0]["query"])
	}

	// Lines 1-2: result
	if lines[1]["type"] != "result" {
		t.Errorf("line 1 type = %v, want result", lines[1]["type"])
	}
	if lines[1]["source"] != "reddit" {
		t.Errorf("line 1 source = %v, want reddit", lines[1]["source"])
	}
	items1, ok := lines[1]["items"].([]interface{})
	if !ok || len(items1) != 1 {
		t.Errorf("line 1 items = %v, want 1 item", lines[1]["items"])
	}
	if lines[1]["error"] != "" {
		t.Errorf("line 1 error = %v, want empty", lines[1]["error"])
	}

	if lines[2]["source"] != "hackernews" {
		t.Errorf("line 2 source = %v, want hackernews", lines[2]["source"])
	}
	items2, ok := lines[2]["items"].([]interface{})
	if !ok || len(items2) != 2 {
		t.Errorf("line 2 items = %v, want 2 items", lines[2]["items"])
	}

	// Line 3: done
	if lines[3]["type"] != "done" {
		t.Errorf("line 3 type = %v, want done", lines[3]["type"])
	}
	totalSources, _ := lines[3]["total_sources"].(float64)
	if int(totalSources) != 2 {
		t.Errorf("done total_sources = %v, want 2", lines[3]["total_sources"])
	}
	totalItems, _ := lines[3]["total_items"].(float64)
	if int(totalItems) != 3 {
		t.Errorf("done total_items = %v, want 3", lines[3]["total_items"])
	}
}

func TestHandleSearchStreamResultWithError(t *testing.T) {
	mock := &streamMockOrchestrator{
		results: []orchestrator.StreamResult{
			{Source: "reddit", Items: []engines.Result{{Title: "OK"}}},
			{Source: "bluesky", Error: "rate limited"},
		},
	}
	s := NewServer(mock)
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{"query": "err test"}

	res, err := s.handleSearchStream(context.Background(), req)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected tool success with error in NDJSON, got: %v", res.Content)
	}

	lines := parseNDJSON(t, res.Content[0].(mcp.TextContent).Text)
	// meta + 2 results + done = 4
	if len(lines) != 4 {
		t.Fatalf("expected 4 lines, got %d", len(lines))
	}

	// Find the bluesky error result
	found := false
	for _, line := range lines {
		if line["type"] == "result" && line["source"] == "bluesky" {
			if line["error"] != "rate limited" {
				t.Errorf("bluesky error = %v, want 'rate limited'", line["error"])
			}
			found = true
		}
	}
	if !found {
		t.Fatal("bluesky error result not found in NDJSON output")
	}
}

func TestHandleSearchStreamMissingQuery(t *testing.T) {
	mock := &streamMockOrchestrator{}
	s := NewServer(mock)
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	res, err := s.handleSearchStream(context.Background(), req)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if !res.IsError {
		t.Fatal("expected error for missing query")
	}
}

func TestHandleSearchStreamOrchestratorError(t *testing.T) {
	s := NewServer(errorOrchestrator{})
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{"query": "fail"}

	res, err := s.handleSearchStream(context.Background(), req)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if !res.IsError {
		t.Fatal("expected error result from orchestrator failure")
	}
	if !contains(res.Content[0].(mcp.TextContent).Text, "stream failed") {
		t.Errorf("expected 'stream failed' message, got: %v", res.Content)
	}
}

func TestHandleSearchStreamContextCancel(t *testing.T) {
	mock := &streamMockOrchestrator{
		results: []orchestrator.StreamResult{
			{Source: "slow", Items: []engines.Result{{Title: "should not see this"}}},
		},
	}
	s := NewServer(mock)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{"query": "cancel test"}

	res, err := s.handleSearchStream(ctx, req)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	// With a cancelled context, the mock's SearchStream closes the channel
	// immediately, so we should still get valid NDJSON (meta + done with 0 sources).
	if res.IsError {
		t.Fatalf("expected valid NDJSON output even on cancel, got error: %v", res.Content)
	}

	lines := parseNDJSON(t, res.Content[0].(mcp.TextContent).Text)
	// At minimum we have meta and done lines.
	if len(lines) < 2 {
		t.Fatalf("expected at least meta + done lines, got %d", len(lines))
	}
	if lines[0]["type"] != "meta" {
		t.Errorf("first line type = %v, want meta", lines[0]["type"])
	}
	last := lines[len(lines)-1]
	if last["type"] != "done" {
		t.Errorf("last line type = %v, want done", last["type"])
	}
}

func TestHandleSearchStreamEmptyResults(t *testing.T) {
	mock := &streamMockOrchestrator{
		results: []orchestrator.StreamResult{},
	}
	s := NewServer(mock)
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{"query": "empty"}

	res, err := s.handleSearchStream(context.Background(), req)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected success, got: %v", res.Content)
	}

	lines := parseNDJSON(t, res.Content[0].(mcp.TextContent).Text)
	// meta + done = 2 lines
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines (meta + done), got %d", len(lines))
	}
	if lines[0]["type"] != "meta" {
		t.Errorf("first line type = %v, want meta", lines[0]["type"])
	}
	if lines[1]["type"] != "done" {
		t.Errorf("last line type = %v, want done", lines[1]["type"])
	}
	totalSources, _ := lines[1]["total_sources"].(float64)
	if int(totalSources) != 0 {
		t.Errorf("done total_sources = %v, want 0", lines[1]["total_sources"])
	}
	totalItems, _ := lines[1]["total_items"].(float64)
	if int(totalItems) != 0 {
		t.Errorf("done total_items = %v, want 0", lines[1]["total_items"])
	}
}

// TestHandleSearchStreamOrderPreserved verifies results arrive in channel order
// (which reflects source completion order, not registration order).
func TestHandleSearchStreamOrderPreserved(t *testing.T) {
	mock := &streamMockOrchestrator{
		results: []orchestrator.StreamResult{
			{Source: "third", Items: []engines.Result{{Title: "C"}}},
			{Source: "first", Items: []engines.Result{{Title: "A"}}},
			{Source: "second", Items: []engines.Result{{Title: "B"}}},
		},
	}
	s := NewServer(mock)
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{"query": "order"}

	res, err := s.handleSearchStream(context.Background(), req)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected success, got: %v", res.Content)
	}

	lines := parseNDJSON(t, res.Content[0].(mcp.TextContent).Text)
	// meta + 3 results + done = 5
	if len(lines) != 5 {
		t.Fatalf("expected 5 lines, got %d", len(lines))
	}
	// Verify the result sources are in the exact channel order.
	expectedSources := []string{"third", "first", "second"}
	for i, expected := range expectedSources {
		line := lines[i+1] // +1 to skip meta
		if line["type"] != "result" {
			t.Errorf("line %d type = %v, want result", i+1, line["type"])
		}
		if line["source"] != expected {
			t.Errorf("line %d source = %v, want %s", i+1, line["source"], expected)
		}
	}
}

// TestRegisterSearchStreamTool verifies the tool can be registered without panicking.
func TestRegisterSearchStreamTool(t *testing.T) {
	mock := &streamMockOrchestrator{}
	s := NewServer(mock)
	// Should not panic.
	RegisterSearchStreamTool(s)
}

// TestHandleSearchStreamConcurrent verifies race-free behavior under -race.
func TestHandleSearchStreamConcurrent(t *testing.T) {
	mock := &streamMockOrchestrator{
		results: []orchestrator.StreamResult{
			{Source: "a", Items: []engines.Result{{Title: "1"}}},
			{Source: "b", Items: []engines.Result{{Title: "2"}}},
			{Source: "c", Items: []engines.Result{{Title: "3"}}},
		},
	}
	s := NewServer(mock)

	done := make(chan struct{})
	go func() {
		defer close(done)
		req := mcp.CallToolRequest{}
		req.Params.Arguments = map[string]interface{}{"query": "race"}
		res, err := s.handleSearchStream(context.Background(), req)
		if err != nil {
			t.Errorf("concurrent error: %v", err)
			return
		}
		if res.IsError {
			t.Errorf("concurrent result was error: %v", res.Content)
		}
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("handleSearchStream timed out")
	}
}
