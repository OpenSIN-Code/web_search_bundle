// Purpose: Tests for the MCP server tool handlers.
// Docs: server_test.doc.md

package mcp

import (
	"context"
	"testing"

	"github.com/OpenSIN-Code/web_search_bundle/internal/engines"
	"github.com/OpenSIN-Code/web_search_bundle/internal/orchestrator"
	"github.com/mark3labs/mcp-go/mcp"
)

// mockOrchestrator is a test double that returns canned results.
type mockOrchestrator struct {
	searchResult *orchestrator.SearchResult
	pulseResult  *orchestrator.SearchResult
	streamResult []orchestrator.StreamResult
}

func (m *mockOrchestrator) Search(ctx context.Context, topic string) (*orchestrator.SearchResult, error) {
	return m.searchResult, nil
}

func (m *mockOrchestrator) Pulse(ctx context.Context, topic string) (*orchestrator.SearchResult, error) {
	return m.pulseResult, nil
}

func (m *mockOrchestrator) SearchStream(ctx context.Context, topic string) (<-chan orchestrator.StreamResult, error) {
	ch := make(chan orchestrator.StreamResult, len(m.streamResult))
	for _, r := range m.streamResult {
		ch <- r
	}
	close(ch)
	return ch, nil
}

func newMockServer() *Server {
	mock := &mockOrchestrator{
		searchResult: &orchestrator.SearchResult{
			Results: []engines.Result{
				{Source: "reddit", Title: "Test", URL: "https://example.com", Snippet: "hello"},
			},
		},
		pulseResult: &orchestrator.SearchResult{
			Results: []engines.Result{
				{Source: "hackernews", Title: "Pulse", URL: "https://hn.example.com", Snippet: "world"},
			},
		},
		streamResult: []orchestrator.StreamResult{
			{Source: "reddit", Items: []engines.Result{{Title: "Stream"}}},
		},
	}
	return NewServer(mock)
}

func TestHandleSearch(t *testing.T) {
	s := newMockServer()
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{"query": "OpenClaw"}

	res, err := s.handleSearch(context.Background(), req)
	if err != nil {
		t.Fatalf("handleSearch error: %v", err)
	}
	if len(res.Content) == 0 {
		t.Fatal("expected content")
	}
	if res.IsError {
		t.Fatal("expected no error")
	}
	// Result should be JSON containing the mock result.
	text := res.Content[0].(mcp.TextContent).Text
	if !contains(text, "reddit") || !contains(text, "Test") {
		t.Errorf("unexpected result: %s", text)
	}
}

func TestHandleSearchMissingQuery(t *testing.T) {
	s := newMockServer()
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	res, err := s.handleSearch(context.Background(), req)
	if err != nil {
		t.Fatalf("handleSearch error: %v", err)
	}
	if !res.IsError {
		t.Fatal("expected error for missing query")
	}
}

func TestHandlePulse(t *testing.T) {
	s := newMockServer()
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{"topic": "AI"}

	res, err := s.handlePulse(context.Background(), req)
	if err != nil {
		t.Fatalf("handlePulse error: %v", err)
	}
	if res.IsError {
		t.Fatal("expected no error")
	}
	text := res.Content[0].(mcp.TextContent).Text
	if !contains(text, "hackernews") {
		t.Errorf("unexpected result: %s", text)
	}
}

func TestHandleResolve(t *testing.T) {
	s := newMockServer()
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{"name": "Peter Steinberger"}

	res, err := s.handleResolve(context.Background(), req)
	if err != nil {
		t.Fatalf("handleResolve error: %v", err)
	}
	if res.IsError {
		t.Fatal("expected no error")
	}
	text := res.Content[0].(mcp.TextContent).Text
	if !contains(text, "Peter Steinberger") {
		t.Errorf("expected resolved entity in result: %s", text)
	}
}

func TestHandleResolveMissingName(t *testing.T) {
	s := newMockServer()
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	res, err := s.handleResolve(context.Background(), req)
	if err != nil {
		t.Fatalf("handleResolve error: %v", err)
	}
	if !res.IsError {
		t.Fatal("expected error for missing name")
	}
}

func TestHandleWatchMissingURL(t *testing.T) {
	s := newMockServer()
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	res, err := s.handleWatch(context.Background(), req)
	if err != nil {
		t.Fatalf("handleWatch error: %v", err)
	}
	if !res.IsError {
		t.Fatal("expected error for missing url")
	}
}

func TestHandleVideoBriefMissingURL(t *testing.T) {
	s := newMockServer()
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	res, err := s.handleVideoBrief(context.Background(), req)
	if err != nil {
		t.Fatalf("handleVideoBrief error: %v", err)
	}
	if !res.IsError {
		t.Fatal("expected error for missing url")
	}
}

func TestHandleVideoPromptMissingURL(t *testing.T) {
	s := newMockServer()
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	res, err := s.handleVideoPrompt(context.Background(), req)
	if err != nil {
		t.Fatalf("handleVideoPrompt error: %v", err)
	}
	if !res.IsError {
		t.Fatal("expected error for missing url")
	}
}

func TestHandleAlchemistMissingRunCmd(t *testing.T) {
	s := newMockServer()
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{"repo_path": t.TempDir()}

	res, err := s.handleAlchemist(context.Background(), req)
	if err != nil {
		t.Fatalf("handleAlchemist error: %v", err)
	}
	if !res.IsError {
		t.Fatal("expected error for missing run_cmd")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
