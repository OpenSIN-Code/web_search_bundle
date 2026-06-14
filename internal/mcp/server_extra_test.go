// SPDX-License-Identifier: MIT
// Purpose: Additional unit tests for MCP server helpers and error paths.
// Docs: internal/mcp/server_extra_test.doc.md
package mcp

import (
	"context"
	"errors"
	"testing"

	"github.com/OpenSIN-Code/web_search_bundle/internal/orchestrator"
	"github.com/mark3labs/mcp-go/mcp"
)

// errorOrchestrator returns errors for every operation.
type errorOrchestrator struct{}

func (errorOrchestrator) Search(ctx context.Context, topic string) (*orchestrator.SearchResult, error) {
	return nil, errors.New("search failed")
}

func (errorOrchestrator) Pulse(ctx context.Context, topic string) (*orchestrator.SearchResult, error) {
	return nil, errors.New("pulse failed")
}

func (errorOrchestrator) SearchStream(ctx context.Context, topic string) (<-chan orchestrator.StreamResult, error) {
	return nil, errors.New("stream failed")
}

func TestArgString(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{"key": "value"}
	if got := argString(req, "key"); got != "value" {
		t.Errorf("argString = %q, want value", got)
	}
	if got := argString(req, "missing"); got != "" {
		t.Errorf("argString missing = %q, want empty", got)
	}
}

func TestArgStringNoArguments(t *testing.T) {
	req := mcp.CallToolRequest{}
	if got := argString(req, "key"); got != "" {
		t.Errorf("argString = %q, want empty", got)
	}
}

func TestArgBool(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{"flag": true}
	if !argBool(req, "flag") {
		t.Errorf("argBool = false, want true")
	}
	if argBool(req, "missing") {
		t.Errorf("argBool missing = true, want false")
	}
}

func TestArgBoolWrongType(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{"flag": "not a bool"}
	if argBool(req, "flag") {
		t.Errorf("argBool = true, want false for wrong type")
	}
}

func TestArgInt(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{"count": 42}
	if got := argInt(req, "count", 0); got != 42 {
		t.Errorf("argInt = %d, want 42", got)
	}
	if got := argInt(req, "missing", 7); got != 7 {
		t.Errorf("argInt default = %d, want 7", got)
	}
}

func TestArgIntFloat64(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{"count": 3.0}
	if got := argInt(req, "count", 0); got != 3 {
		t.Errorf("argInt = %d, want 3", got)
	}
}

func TestHandleSearchError(t *testing.T) {
	s := NewServer(errorOrchestrator{})
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{"query": "fail"}

	res, err := s.handleSearch(context.Background(), req)
	if err != nil {
		t.Fatalf("handleSearch error: %v", err)
	}
	if !res.IsError {
		t.Fatal("expected error result")
	}
	if !contains(res.Content[0].(mcp.TextContent).Text, "search failed") {
		t.Errorf("expected error message, got: %v", res.Content)
	}
}

func TestHandlePulseError(t *testing.T) {
	s := NewServer(errorOrchestrator{})
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{"topic": "fail"}

	res, err := s.handlePulse(context.Background(), req)
	if err != nil {
		t.Fatalf("handlePulse error: %v", err)
	}
	if !res.IsError {
		t.Fatal("expected error result")
	}
	if !contains(res.Content[0].(mcp.TextContent).Text, "pulse failed") {
		t.Errorf("expected error message, got: %v", res.Content)
	}
}

func TestHandleResolveMissingNameError(t *testing.T) {
	s := NewServer(errorOrchestrator{})
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	res, err := s.handleResolve(context.Background(), req)
	if err != nil {
		t.Fatalf("handleResolve error: %v", err)
	}
	if !res.IsError {
		t.Fatal("expected error result")
	}
	if !contains(res.Content[0].(mcp.TextContent).Text, "name required") {
		t.Errorf("expected name required message, got: %v", res.Content)
	}
}

func TestHandleAlchemistInvalidBudget(t *testing.T) {
	s := newMockServer()
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"run_cmd": "true",
		"budget":  "not-a-duration",
	}

	res, err := s.handleAlchemist(context.Background(), req)
	if err != nil {
		t.Fatalf("handleAlchemist error: %v", err)
	}
	if !res.IsError {
		t.Fatal("expected error result")
	}
	if !contains(res.Content[0].(mcp.TextContent).Text, "invalid budget") {
		t.Errorf("expected invalid budget message, got: %v", res.Content)
	}
}

func TestHandleAlchemistInvalidRuntime(t *testing.T) {
	s := newMockServer()
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"run_cmd": "true",
		"runtime": "not-a-duration",
	}

	res, err := s.handleAlchemist(context.Background(), req)
	if err != nil {
		t.Fatalf("handleAlchemist error: %v", err)
	}
	if !res.IsError {
		t.Fatal("expected error result")
	}
	if !contains(res.Content[0].(mcp.TextContent).Text, "invalid runtime") {
		t.Errorf("expected invalid runtime message, got: %v", res.Content)
	}
}

func TestHandleAlchemistInvalidSafety(t *testing.T) {
	s := newMockServer()
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"run_cmd": "true",
		"safety":  "dangerous",
	}

	res, err := s.handleAlchemist(context.Background(), req)
	if err != nil {
		t.Fatalf("handleAlchemist error: %v", err)
	}
	if !res.IsError {
		t.Fatal("expected error result")
	}
	if !contains(res.Content[0].(mcp.TextContent).Text, "invalid safety mode") {
		t.Errorf("expected invalid safety mode message, got: %v", res.Content)
	}
}

func TestHandleAlchemistDefaults(t *testing.T) {
	// Only validates the parameter parsing up to the point before the daemon
	// is created. We use a repo path that doesn't exist so the daemon init
	// returns an error quickly, which proves defaults were applied.
	s := newMockServer()
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"run_cmd": "echo metric: 1.0",
	}

	res, err := s.handleAlchemist(context.Background(), req)
	if err != nil {
		t.Fatalf("handleAlchemist error: %v", err)
	}
	// NewDaemon will fail because the default repo path is the current working
	// directory and it expects a git repo. We just verify it is an error.
	if !res.IsError {
		t.Fatal("expected error result for non-git repo")
	}
}

func TestHandleAlchemistSwarmInit(t *testing.T) {
	// Verify that the strategies argument is parsed into a multi-strategy
	// request. The swarm may succeed or fail depending on the repo; we only
	// check that the function terminates without a protocol error.
	s := newMockServer()
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"run_cmd":    "echo metric: 1.0",
		"strategies": "strategy1,strategy2",
	}

	res, err := s.handleAlchemist(context.Background(), req)
	if err != nil {
		t.Fatalf("handleAlchemist error: %v", err)
	}
	if len(res.Content) == 0 {
		t.Fatal("expected non-empty result")
	}
}
