// SPDX-License-Identifier: MIT
// Purpose: Verify MCP tool annotations and output schemas for all 8 tools.
package mcp

import (
	"context"
	"testing"

	"github.com/OpenSIN-Code/web_search_bundle/internal/orchestrator"
)

type stubOrchestrator struct{}

func (stubOrchestrator) Search(_ context.Context, _ string) (*orchestrator.SearchResult, error) {
	return &orchestrator.SearchResult{}, nil
}

func (stubOrchestrator) Pulse(_ context.Context, _ string) (*orchestrator.SearchResult, error) {
	return &orchestrator.SearchResult{}, nil
}

func (stubOrchestrator) SearchStream(_ context.Context, _ string) (<-chan orchestrator.StreamResult, error) {
	return nil, nil
}

func boolPtrVal(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

func TestToolAnnotations(t *testing.T) {
	srv := NewServer(stubOrchestrator{})
	tools := srv.server.ListTools()
	if len(tools) != 8 {
		t.Fatalf("expected 8 tools, got %d", len(tools))
	}

	readOnlyTools := []string{
		"websearch_search",
		"websearch_pulse",
		"websearch_resolve",
		"websearch_watch",
		"websearch_video_brief",
		"websearch_video_prompt",
		"websearch_status",
	}

	for _, name := range readOnlyTools {
		t.Run(name+"_annotations", func(t *testing.T) {
			st, ok := tools[name]
			if !ok {
				t.Fatalf("tool %s not found", name)
			}
			if got := boolPtrVal(st.Tool.Annotations.ReadOnlyHint); got != true {
				t.Errorf("ReadOnlyHint = %v, want true", got)
			}
			if got := boolPtrVal(st.Tool.Annotations.DestructiveHint); got != false {
				t.Errorf("DestructiveHint = %v, want false", got)
			}
			if got := boolPtrVal(st.Tool.Annotations.IdempotentHint); got != true {
				t.Errorf("IdempotentHint = %v, want true", got)
			}
			if got := boolPtrVal(st.Tool.Annotations.OpenWorldHint); got != true {
				t.Errorf("OpenWorldHint = %v, want true", got)
			}
		})
	}

	t.Run("websearch_alchemist_annotations", func(t *testing.T) {
		st, ok := tools["websearch_alchemist"]
		if !ok {
			t.Fatal("websearch_alchemist not found")
		}
		if got := boolPtrVal(st.Tool.Annotations.ReadOnlyHint); got != false {
			t.Errorf("ReadOnlyHint = %v, want false", got)
		}
		if got := boolPtrVal(st.Tool.Annotations.DestructiveHint); got != true {
			t.Errorf("DestructiveHint = %v, want true", got)
		}
		if got := boolPtrVal(st.Tool.Annotations.IdempotentHint); got != false {
			t.Errorf("IdempotentHint = %v, want false", got)
		}
		if got := boolPtrVal(st.Tool.Annotations.OpenWorldHint); got != true {
			t.Errorf("OpenWorldHint = %v, want true", got)
		}
	})
}

func TestOutputSchemas(t *testing.T) {
	srv := NewServer(stubOrchestrator{})
	tools := srv.server.ListTools()

	cases := []struct {
		name       string
		properties map[string]string
	}{
		{"websearch_search", map[string]string{"results": "array", "clusters": "array", "errors": "object"}},
		{"websearch_pulse", map[string]string{"results": "array", "clusters": "array", "errors": "object"}},
		{"websearch_resolve", map[string]string{"handles": "object", "confidence": "number"}},
		{"websearch_watch", map[string]string{"frames": "array", "transcript": "string", "duration": "string"}},
		{"websearch_video_brief", map[string]string{"html": "string", "preset": "string"}},
		{"websearch_video_prompt", map[string]string{"prompt": "string", "model": "string", "tokens": "integer"}},
		{"websearch_alchemist", map[string]string{"experiments": "integer", "best_metric": "string", "report": "string"}},
	}

	for _, tc := range cases {
		t.Run(tc.name+"_output_schema", func(t *testing.T) {
			st, ok := tools[tc.name]
			if !ok {
				t.Fatalf("tool %s not found", tc.name)
			}
			if st.Tool.OutputSchema.Type != "object" {
				t.Errorf("OutputSchema.Type = %q, want %q", st.Tool.OutputSchema.Type, "object")
			}
			props := st.Tool.OutputSchema.Properties
			if props == nil {
				t.Fatal("OutputSchema.Properties is nil")
			}
			if len(props) != len(tc.properties) {
				t.Errorf("OutputSchema.Properties has %d keys, want %d", len(props), len(tc.properties))
			}
			for prop, expectedType := range tc.properties {
				p, ok := props[prop]
				if !ok {
					t.Errorf("missing property %q", prop)
					continue
				}
				pm, ok := p.(map[string]any)
				if !ok {
					t.Errorf("property %q is %T, not map[string]any", prop, p)
					continue
				}
				pt, _ := pm["type"].(string)
				if pt != expectedType {
					t.Errorf("property %q type = %q, want %q", prop, pt, expectedType)
				}
			}
		})
	}
}

func TestToolNamesAndInputSchemasUnchanged(t *testing.T) {
	srv := NewServer(stubOrchestrator{})
	tools := srv.server.ListTools()

	expected := map[string][]string{
		"websearch_search":       {"query"},
		"websearch_pulse":        {"topic"},
		"websearch_resolve":      {"name"},
		"websearch_watch":        {"url"},
		"websearch_video_brief":  {"url"},
		"websearch_video_prompt": {"url"},
		"websearch_alchemist":    {"run_cmd"},
	}

	for name, required := range expected {
		t.Run(name+"_input", func(t *testing.T) {
			st, ok := tools[name]
			if !ok {
				t.Fatalf("tool %s not found", name)
			}
			if st.Tool.InputSchema.Type != "object" {
				t.Errorf("InputSchema.Type = %q, want object", st.Tool.InputSchema.Type)
			}
			for _, r := range required {
				found := false
				for _, rr := range st.Tool.InputSchema.Required {
					if rr == r {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("required field %q missing from InputSchema.Required", r)
				}
			}
		})
	}
}
