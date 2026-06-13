// Purpose: MCP server exposing sin-websearch tools to sin-code and other agents.
// Docs: internal/mcp/server.doc.md
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/OpenSIN-Code/web_search_bundle/internal/briefing"
	"github.com/OpenSIN-Code/web_search_bundle/internal/engines"
	"github.com/OpenSIN-Code/web_search_bundle/internal/orchestrator"
	"github.com/OpenSIN-Code/web_search_bundle/internal/prompts"
	"github.com/OpenSIN-Code/web_search_bundle/internal/resolver"
	"github.com/OpenSIN-Code/web_search_bundle/internal/sidecar"
	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

// Server wraps the MCP server and application services.
type Server struct {
	orchestrator *orchestrator.Orchestrator
	resolver     *resolver.EntityResolver
	server       *mcpserver.MCPServer
}

// NewServer creates and configures the MCP server.
func NewServer(orch *orchestrator.Orchestrator) *Server {
	s := &Server{
		orchestrator: orch,
		resolver:     resolver.NewEntityResolver(),
	}
	s.setup()
	return s
}

func (s *Server) setup() {
	s.server = mcpserver.NewMCPServer("sin-websearch", "1.0.0", mcpserver.WithToolCapabilities(true))

	searchTool := mcp.Tool{
		Name:        "websearch_search",
		Description: "Multi-source web search with caching and clustering",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"query": map[string]string{"type": "string", "description": "Search query"},
			},
			Required: []string{"query"},
		},
	}
	s.server.AddTool(searchTool, s.handleSearch)

	pulseTool := mcp.Tool{
		Name:        "websearch_pulse",
		Description: "Social pulse search focused on engagement",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]interface{}{"topic": map[string]string{"type": "string", "description": "Topic to analyze"}},
			Required:   []string{"topic"},
		},
	}
	s.server.AddTool(pulseTool, s.handlePulse)

	resolveTool := mcp.Tool{
		Name:        "websearch_resolve",
		Description: "Resolve a name or topic to platform handles",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]interface{}{"name": map[string]string{"type": "string", "description": "Name or topic to resolve"}},
			Required:   []string{"name"},
		},
	}
	s.server.AddTool(resolveTool, s.handleResolve)

	watchTool := mcp.Tool{
		Name:        "websearch_watch",
		Description: "Analyze a video: extract frames and transcript",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"url": map[string]string{"type": "string", "description": "Video URL or local path"},
			},
			Required: []string{"url"},
		},
	}
	s.server.AddTool(watchTool, s.handleWatch)

	videoBriefTool := mcp.Tool{
		Name:        "websearch_video_brief",
		Description: "Generate a self-contained HTML briefing for a video with embedded frames",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"url":    map[string]string{"type": "string", "description": "Video URL or local path"},
				"preset": map[string]string{"type": "string", "description": "general|bug|tutorial|hook|transcript|compare|summary"},
			},
			Required: []string{"url"},
		},
	}
	s.server.AddTool(videoBriefTool, s.handleVideoBrief)

	videoPromptTool := mcp.Tool{
		Name:        "websearch_video_prompt",
		Description: "Generate a Vision-LLM-ready prompt from a video",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"url":    map[string]string{"type": "string", "description": "Video URL or local path"},
				"model":  map[string]string{"type": "string", "description": "claude|gpt4o|gemini|generic"},
				"preset": map[string]string{"type": "string", "description": "general|bug|tutorial|hook|transcript|compare|summary"},
			},
			Required: []string{"url"},
		},
	}
	s.server.AddTool(videoPromptTool, s.handleVideoPrompt)
}

// Serve starts the stdio MCP server.
func (s *Server) Serve() error {
	return mcpserver.ServeStdio(s.server)
}

func argString(req mcp.CallToolRequest, key string) string {
	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok {
		return ""
	}
	val, _ := args[key].(string)
	return val
}

func (s *Server) handleSearch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query := argString(req, "query")
	if query == "" {
		return mcp.NewToolResultError("query required"), nil
	}
	res, err := s.orchestrator.Search(ctx, query)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	data, _ := json.MarshalIndent(res, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func (s *Server) handlePulse(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	topic := argString(req, "topic")
	if topic == "" {
		return mcp.NewToolResultError("topic required"), nil
	}
	res, err := s.orchestrator.Pulse(ctx, topic)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	data, _ := json.MarshalIndent(res, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func (s *Server) handleResolve(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name := argString(req, "name")
	if name == "" {
		return mcp.NewToolResultError("name required"), nil
	}
	entity, err := s.resolver.Resolve(ctx, name)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	data, _ := json.MarshalIndent(entity, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func (s *Server) handleWatch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	url := argString(req, "url")
	if url == "" {
		return mcp.NewToolResultError("url required"), nil
	}
	sc, err := sidecar.NewManager()
	if err != nil {
		return mcp.NewToolResultError("sidecar: " + err.Error()), nil
	}
	engine := engines.NewVideoEngine(sc)
	analysis, err := engine.Watch(ctx, engines.WatchOptions{URL: url})
	if err != nil {
		return mcp.NewToolResultError("watch: " + err.Error()), nil
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Video Analysis: %s\n\n", analysis.Title))
	sb.WriteString(fmt.Sprintf("Source: %s | Duration: %s | Mode: %s\n\n", analysis.Source, analysis.Duration, analysis.Mode))
	sb.WriteString("## Frames\n\n")
	for _, f := range analysis.Frames {
		sb.WriteString(fmt.Sprintf("- [t=%s] `%s`\n", formatDuration(f.Timestamp), f.Path))
	}
	sb.WriteString(fmt.Sprintf("\n## Transcript (%s)\n\n%s\n", analysis.TranscriptSource, analysis.Transcript))
	return mcp.NewToolResultText(sb.String()), nil
}

func (s *Server) handleVideoBrief(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	url := argString(req, "url")
	if url == "" {
		return mcp.NewToolResultError("url required"), nil
	}
	sc, err := sidecar.NewManager()
	if err != nil {
		return mcp.NewToolResultError("sidecar: " + err.Error()), nil
	}
	engine := engines.NewVideoEngine(sc)
	analysis, err := engine.Watch(ctx, engines.WatchOptions{URL: url})
	if err != nil {
		return mcp.NewToolResultError("watch: " + err.Error()), nil
	}
	preset := prompts.PresetGeneral
	if v := argString(req, "preset"); v != "" {
		preset = prompts.Preset(v)
	}
	built := prompts.BuildVideoPrompt(prompts.VideoPromptRequest{Analysis: analysis, Preset: preset})
	path, err := briefing.GenerateVideoBriefHTML(briefing.VideoBriefOptions{Analysis: analysis, Prompt: built, EmbedFrames: true})
	if err != nil {
		return mcp.NewToolResultError("html: " + err.Error()), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("Video briefing saved to: %s\nFrames: %d", path, analysis.FrameCount)), nil
}

func (s *Server) handleVideoPrompt(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	url := argString(req, "url")
	if url == "" {
		return mcp.NewToolResultError("url required"), nil
	}
	sc, err := sidecar.NewManager()
	if err != nil {
		return mcp.NewToolResultError("sidecar: " + err.Error()), nil
	}
	engine := engines.NewVideoEngine(sc)
	analysis, err := engine.Watch(ctx, engines.WatchOptions{URL: url})
	if err != nil {
		return mcp.NewToolResultError("watch: " + err.Error()), nil
	}
	model := prompts.ModelClaude
	if v := argString(req, "model"); v != "" {
		model = prompts.Model(v)
	}
	preset := prompts.PresetGeneral
	if v := argString(req, "preset"); v != "" {
		preset = prompts.Preset(v)
	}
	built := prompts.BuildVideoPrompt(prompts.VideoPromptRequest{Analysis: analysis, Model: model, Preset: preset})
	var sb strings.Builder
	sb.WriteString("## SYSTEM PROMPT\n\n" + built.System)
	sb.WriteString("\n\n## USER PROMPT\n\n" + built.User)
	sb.WriteString("\n\n## IMAGE ATTACHMENTS\n\n")
	for i, p := range built.ImagePaths {
		sb.WriteString(fmt.Sprintf("%d. `%s`\n", i+1, p))
	}
	return mcp.NewToolResultText(sb.String()), nil
}

func formatDuration(d interface{}) string {
	return fmt.Sprintf("%v", d)
}
