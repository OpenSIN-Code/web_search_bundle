// SPDX-License-Identifier: MIT
// Purpose: MCP server exposing sin-websearch tools to sin-code and other agents.
// Docs: internal/mcp/server.doc.md
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/OpenSIN-Code/web_search_bundle/internal/alchemist"
	"github.com/OpenSIN-Code/web_search_bundle/internal/briefing"
	"github.com/OpenSIN-Code/web_search_bundle/internal/engines"
	"github.com/OpenSIN-Code/web_search_bundle/internal/orchestrator"
	"github.com/OpenSIN-Code/web_search_bundle/internal/prompts"
	"github.com/OpenSIN-Code/web_search_bundle/internal/resolver"
	"github.com/OpenSIN-Code/web_search_bundle/internal/sidecar"
	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

// Orchestrator describes the search/pulse operations the MCP server needs.
type Orchestrator interface {
	Search(ctx context.Context, topic string) (*orchestrator.SearchResult, error)
	Pulse(ctx context.Context, topic string) (*orchestrator.SearchResult, error)
	SearchStream(ctx context.Context, topic string) (<-chan orchestrator.StreamResult, error)
}

// Server wraps the MCP server and application services.
type Server struct {
	orchestrator Orchestrator
	resolver     *resolver.EntityResolver
	server       *mcpserver.MCPServer
}

// NewServer creates and configures the MCP server.
func NewServer(orch Orchestrator) *Server {
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
		Annotations: mcp.ToolAnnotation{
			ReadOnlyHint:    mcp.ToBoolPtr(true),
			DestructiveHint: mcp.ToBoolPtr(false),
			IdempotentHint:  mcp.ToBoolPtr(true),
			OpenWorldHint:   mcp.ToBoolPtr(true),
		},
		OutputSchema: mcp.ToolOutputSchema{
			Type: "object",
			Properties: map[string]any{
				"results":  map[string]any{"type": "array"},
				"clusters": map[string]any{"type": "array"},
				"errors":   map[string]any{"type": "object"},
			},
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
		Annotations: mcp.ToolAnnotation{
			ReadOnlyHint:    mcp.ToBoolPtr(true),
			DestructiveHint: mcp.ToBoolPtr(false),
			IdempotentHint:  mcp.ToBoolPtr(true),
			OpenWorldHint:   mcp.ToBoolPtr(true),
		},
		OutputSchema: mcp.ToolOutputSchema{
			Type: "object",
			Properties: map[string]any{
				"results":  map[string]any{"type": "array"},
				"clusters": map[string]any{"type": "array"},
				"errors":   map[string]any{"type": "object"},
			},
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
		Annotations: mcp.ToolAnnotation{
			ReadOnlyHint:    mcp.ToBoolPtr(true),
			DestructiveHint: mcp.ToBoolPtr(false),
			IdempotentHint:  mcp.ToBoolPtr(true),
			OpenWorldHint:   mcp.ToBoolPtr(true),
		},
		OutputSchema: mcp.ToolOutputSchema{
			Type: "object",
			Properties: map[string]any{
				"handles":   map[string]any{"type": "object"},
				"confidence": map[string]any{"type": "number"},
			},
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
		Annotations: mcp.ToolAnnotation{
			ReadOnlyHint:    mcp.ToBoolPtr(true),
			DestructiveHint: mcp.ToBoolPtr(false),
			IdempotentHint:  mcp.ToBoolPtr(true),
			OpenWorldHint:   mcp.ToBoolPtr(true),
		},
		OutputSchema: mcp.ToolOutputSchema{
			Type: "object",
			Properties: map[string]any{
				"frames":     map[string]any{"type": "array"},
				"transcript": map[string]any{"type": "string"},
				"duration":   map[string]any{"type": "string"},
			},
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
		Annotations: mcp.ToolAnnotation{
			ReadOnlyHint:    mcp.ToBoolPtr(true),
			DestructiveHint: mcp.ToBoolPtr(false),
			IdempotentHint:  mcp.ToBoolPtr(true),
			OpenWorldHint:   mcp.ToBoolPtr(true),
		},
		OutputSchema: mcp.ToolOutputSchema{
			Type: "object",
			Properties: map[string]any{
				"html":   map[string]any{"type": "string"},
				"preset": map[string]any{"type": "string"},
			},
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
		Annotations: mcp.ToolAnnotation{
			ReadOnlyHint:    mcp.ToBoolPtr(true),
			DestructiveHint: mcp.ToBoolPtr(false),
			IdempotentHint:  mcp.ToBoolPtr(true),
			OpenWorldHint:   mcp.ToBoolPtr(true),
		},
		OutputSchema: mcp.ToolOutputSchema{
			Type: "object",
			Properties: map[string]any{
				"prompt": map[string]any{"type": "string"},
				"model":  map[string]any{"type": "string"},
				"tokens": map[string]any{"type": "integer"},
			},
		},
	}
	s.server.AddTool(videoPromptTool, s.handleVideoPrompt)

	alchemistTool := mcp.Tool{
		Name:        "websearch_alchemist",
		Description: "Run a Karpathy-style autonomous alchemist research loop or a multi-strategy swarm in a local git repo. Defaults to headless (no git mutations) for safety.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"repo_path":        map[string]string{"type": "string", "description": "Path to git repo (default: current working directory)"},
				"run_cmd":          map[string]string{"type": "string", "description": "Shell command that prints the metric to stdout (M3 verification gate)"},
				"target":           map[string]string{"type": "string", "description": "File the agent mutates (default: train.py)"},
				"metric":           map[string]string{"type": "string", "description": "Metric name (default: metric)"},
				"regex":            map[string]string{"type": "string", "description": "Regex with one capture group to extract metric (default: metric:\\s*([0-9\\.]+))"},
				"higher_is_better": map[string]string{"type": "boolean", "description": "Optimization direction (default: false)"},
				"max_experiments":  map[string]string{"type": "integer", "description": "Maximum experiments (default: 3)"},
				"budget":           map[string]string{"type": "string", "description": "Per-experiment time budget (default: 30s)"},
				"runtime":          map[string]string{"type": "string", "description": "Total loop runtime cap (default: 5m)"},
				"program":          map[string]string{"type": "string", "description": "Path to program.md (default: program.md)"},
				"safety":           map[string]string{"type": "string", "description": "headless|auto-commit|interactive (default: headless)"},
				"strategies":       map[string]string{"type": "string", "description": "Comma-separated strategy names for swarm mode. If empty, runs a single daemon."},
			},
			Required: []string{"run_cmd"},
		},
		Annotations: mcp.ToolAnnotation{
			ReadOnlyHint:    mcp.ToBoolPtr(false),
			DestructiveHint: mcp.ToBoolPtr(true),
			IdempotentHint:  mcp.ToBoolPtr(false),
			OpenWorldHint:   mcp.ToBoolPtr(true),
		},
		OutputSchema: mcp.ToolOutputSchema{
			Type: "object",
			Properties: map[string]any{
				"experiments":  map[string]any{"type": "integer"},
				"best_metric":  map[string]any{"type": "string"},
				"report":       map[string]any{"type": "string"},
			},
		},
	}
	s.server.AddTool(alchemistTool, s.handleAlchemist)
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

func argBool(req mcp.CallToolRequest, key string) bool {
	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok {
		return false
	}
	val, _ := args[key].(bool)
	return val
}

func argInt(req mcp.CallToolRequest, key string, defaultVal int) int {
	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok {
		return defaultVal
	}
	switch v := args[key].(type) {
	case int:
		return v
	case float64:
		return int(v)
	case int64:
		return int(v)
	default:
		return defaultVal
	}
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

func (s *Server) handleAlchemist(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	runCmd := argString(req, "run_cmd")
	if runCmd == "" {
		return mcp.NewToolResultError("run_cmd required (M3 verification gate)"), nil
	}

	repoPath := argString(req, "repo_path")
	if repoPath == "" {
		var err error
		repoPath, err = os.Getwd()
		if err != nil {
			return mcp.NewToolResultError("repo path: " + err.Error()), nil
		}
	}

	metric := argString(req, "metric")
	if metric == "" {
		metric = "metric"
	}
	regex := argString(req, "regex")
	if regex == "" {
		regex = `metric:\s*([0-9\.]+)`
	}
	program := argString(req, "program")
	if program == "" {
		program = "program.md"
	}
	target := argString(req, "target")
	if target == "" {
		target = "train.py"
	}

	budget := argString(req, "budget")
	if budget == "" {
		budget = "30s"
	}
	budgetDur, err := time.ParseDuration(budget)
	if err != nil {
		return mcp.NewToolResultError("invalid budget: " + err.Error()), nil
	}

	runtime := argString(req, "runtime")
	if runtime == "" {
		runtime = "5m"
	}
	runtimeDur, err := time.ParseDuration(runtime)
	if err != nil {
		return mcp.NewToolResultError("invalid runtime: " + err.Error()), nil
	}

	safety := alchemist.SafetyMode(argString(req, "safety"))
	if safety == "" {
		safety = alchemist.SafetyHeadless
	}
	if safety != alchemist.SafetyHeadless && safety != alchemist.SafetyAutoCommit && safety != alchemist.SafetyInteractive {
		return mcp.NewToolResultError("invalid safety mode: " + string(safety)), nil
	}

	maxExperiments := argInt(req, "max_experiments", 3)

	cfg := alchemist.Config{
		RepoPath:       repoPath,
		ProgramFile:    program,
		TargetFile:     target,
		MetricName:     metric,
		MetricRegex:    regex,
		HigherIsBetter: argBool(req, "higher_is_better"),
		Budget:         budgetDur,
		RunCmd:         []string{"sh", "-c", runCmd},
		MaxExperiments: maxExperiments,
		MaxRuntime:     runtimeDur,
		Safety:         safety,
	}

	strategies := argString(req, "strategies")
	if strategies != "" {
		// Swarm mode.
		var names []string
		for _, n := range strings.Split(strategies, ",") {
			n = strings.TrimSpace(n)
			if n != "" {
				names = append(names, n)
			}
		}
		swarmCfg := alchemist.SwarmConfig{
			BaseConfig: cfg,
			Strategies: names,
			MaxWorkers: 0,
			FirstWin:   false,
			SharedDB:   true,
		}
		swarm, err := alchemist.NewSwarm(swarmCfg)
		if err != nil {
			return mcp.NewToolResultError("init swarm: " + err.Error()), nil
		}
		defer swarm.Close()

		report, err := swarm.Run(ctx)
		if err != nil {
			return mcp.NewToolResultError("swarm run: " + err.Error()), nil
		}
		return mcp.NewToolResultText(report.RenderMarkdown()), nil
	}

	// Single daemon mode.
	daemon, err := alchemist.NewDaemon(cfg)
	if err != nil {
		return mcp.NewToolResultError("init daemon: " + err.Error()), nil
	}
	defer daemon.Close()

	report, err := daemon.Run(ctx)
	if err != nil {
		return mcp.NewToolResultError("daemon run: " + err.Error()), nil
	}
	md, err := report.RenderMarkdown()
	if err != nil {
		return mcp.NewToolResultError("render report: " + err.Error()), nil
	}
	return mcp.NewToolResultText(md), nil
}
