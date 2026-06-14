// Purpose: Unit tests for video prompt builders.
// Docs: internal/prompts/video_test.doc.md
package prompts

import (
	"strings"
	"testing"
	"time"

	"github.com/OpenSIN-Code/web_search_bundle/internal/engines"
)

func newTestAnalysis() *engines.VideoAnalysis {
	return &engines.VideoAnalysis{
		Title:            "Test Video",
		Source:           "youtube",
		Duration:         5*time.Minute + 30*time.Second,
		Mode:             "full",
		Transcript:       "hello world from the video",
		TranscriptSource: "whisper",
		Frames: []engines.VideoFrame{
			{Path: "/tmp/frame1.jpg", Timestamp: 1 * time.Second},
			{Path: "/tmp/frame2.jpg", Timestamp: 2 * time.Second},
		},
	}
}

func TestBuildVideoPromptDefaults(t *testing.T) {
	req := VideoPromptRequest{Analysis: newTestAnalysis()}
	p := BuildVideoPrompt(req)

	if p.Model != ModelGeneric {
		t.Errorf("Model = %s, want generic", p.Model)
	}
	if p.Preset != PresetGeneral {
		t.Errorf("Preset = %s, want general", p.Preset)
	}
	if p.ImageCount != 2 {
		t.Errorf("ImageCount = %d, want 2", p.ImageCount)
	}
	if !strings.Contains(p.User, "What happens in this video?") {
		t.Errorf("expected default question in user prompt, got: %s", p.User)
	}
	if !strings.Contains(p.System, "Ground EVERY claim") {
		t.Errorf("expected system prompt, got: %s", p.System)
	}
}

func TestBuildVideoPromptModelSpecific(t *testing.T) {
	for _, model := range []Model{ModelClaude, ModelGPT4o, ModelGemini, ModelGeneric} {
		req := VideoPromptRequest{
			Model:        model,
			Preset:       PresetSummary,
			UserQuestion: "Summarize this.",
			Analysis:     newTestAnalysis(),
		}
		p := BuildVideoPrompt(req)
		if p.Model != model {
			t.Errorf("Model = %s, want %s", p.Model, model)
		}
		if p.Preset != PresetSummary {
			t.Errorf("Preset = %s, want summary", p.Preset)
		}
		if !strings.Contains(p.User, modelSpecificPrefix(model)) {
			t.Errorf("expected %s model prefix, got: %s", model, p.User)
		}
	}
}

func modelSpecificPrefix(model Model) string {
	switch model {
	case ModelClaude:
		return "I have attached the frames as images"
	case ModelGPT4o:
		return "The frames are provided as image_url"
	case ModelGemini:
		return "The frames are attached inline"
	default:
		return "Analyze the frames and transcript"
	}
}

func TestBuildVideoPromptPresets(t *testing.T) {
	presets := []Preset{PresetBugReport, PresetTutorial, PresetHook, PresetTranscript, PresetComparison, PresetSummary}
	for _, preset := range presets {
		req := VideoPromptRequest{
			Preset:   preset,
			Analysis: newTestAnalysis(),
		}
		p := BuildVideoPrompt(req)
		if p.Preset != preset {
			t.Errorf("Preset = %s, want %s", p.Preset, preset)
		}
		if p.System == "" {
			t.Errorf("expected non-empty system prompt for preset %s", preset)
		}
	}
}

func TestBuildVideoPromptNoFrames(t *testing.T) {
	req := VideoPromptRequest{
		Analysis: &engines.VideoAnalysis{
			Title:            "No Frames",
			Duration:         1 * time.Minute,
			Mode:             "full",
			Transcript:       "audio only",
			TranscriptSource: "whisper",
		},
	}
	p := BuildVideoPrompt(req)
	if p.ImageCount != 0 {
		t.Errorf("ImageCount = %d, want 0", p.ImageCount)
	}
	if !strings.Contains(p.User, "audio only") {
		t.Errorf("expected transcript in user prompt, got: %s", p.User)
	}
}

func TestBuildVideoPromptTranscriptTruncation(t *testing.T) {
	longTranscript := strings.Repeat("word ", 5000)
	req := VideoPromptRequest{
		Analysis: &engines.VideoAnalysis{
			Title:            "Long",
			Duration:         1 * time.Minute,
			Mode:             "full",
			Transcript:       longTranscript,
			TranscriptSource: "whisper",
		},
	}
	p := BuildVideoPrompt(req)
	if !strings.Contains(p.User, "[... transcript truncated for context window ...]") {
		t.Errorf("expected truncation marker in user prompt")
	}
}

func TestEstimateTokens(t *testing.T) {
	cases := []struct {
		frames, words int
		want          string
	}{
		{0, 0, "~0k tokens (low)"},
		{1, 0, "~1k tokens (low)"},
		{10, 1000, "~15k tokens (moderate)"},
		{40, 1000, "~60k tokens (very high — consider --start/--end)"},
	}
	for _, tc := range cases {
		got := estimateTokens(tc.frames, tc.words)
		if !strings.HasPrefix(got, tc.want) {
			t.Errorf("estimateTokens(%d, %d) = %s, want prefix %s", tc.frames, tc.words, got, tc.want)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	cases := []struct {
		d    time.Duration
		want string
	}{
		{30 * time.Second, "0:30"},
		{90 * time.Second, "1:30"},
		{1*time.Hour + 30*time.Minute + 5*time.Second, "1:30:05"},
	}
	for _, tc := range cases {
		got := formatDuration(tc.d)
		if got != tc.want {
			t.Errorf("formatDuration(%v) = %s, want %s", tc.d, got, tc.want)
		}
	}
}

func TestPresetList(t *testing.T) {
	list := PresetList()
	if len(list) != 7 {
		t.Errorf("len(PresetList) = %d, want 7", len(list))
	}
	if _, ok := list[PresetGeneral]; !ok {
		t.Errorf("expected general preset in list")
	}
}

func TestModelList(t *testing.T) {
	list := ModelList()
	if len(list) != 4 {
		t.Errorf("len(ModelList) = %d, want 4", len(list))
	}
	if _, ok := list[ModelClaude]; !ok {
		t.Errorf("expected claude model in list")
	}
}
