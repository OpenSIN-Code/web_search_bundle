// SPDX-License-Identifier: MIT
// Purpose: Vision-LLM prompt templates for video analysis.
// Docs: internal/prompts/video.doc.md
package prompts

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/OpenSIN-Code/web_search_bundle/internal/engines"
)

// Model identifies the target Vision LLM.
type Model string

const (
	ModelClaude  Model = "claude"
	ModelGPT4o   Model = "gpt4o"
	ModelGemini  Model = "gemini"
	ModelGeneric Model = "generic"
)

// Preset is a pre-built prompt strategy.
type Preset string

const (
	PresetGeneral    Preset = "general"
	PresetBugReport  Preset = "bug"
	PresetTutorial   Preset = "tutorial"
	PresetHook       Preset = "hook"
	PresetTranscript Preset = "transcript"
	PresetComparison Preset = "compare"
	PresetSummary    Preset = "summary"
)

// VideoPromptRequest bundles everything needed to build a prompt.
type VideoPromptRequest struct {
	Model        Model
	Preset       Preset
	UserQuestion string
	Analysis     *engines.VideoAnalysis
}

// BuiltPrompt is the final prompt + metadata.
type BuiltPrompt struct {
	System     string   `json:"system"`
	User       string   `json:"user"`
	ImagePaths []string `json:"image_paths"`
	ImageCount int      `json:"image_count"`
	Transcript string   `json:"transcript"`
	TokenHint  string   `json:"token_hint"`
	Model      Model    `json:"model"`
	Preset     Preset   `json:"preset"`
}

// BuildVideoPrompt creates a model-aware prompt from a video analysis.
func BuildVideoPrompt(req VideoPromptRequest) *BuiltPrompt {
	if req.Model == "" {
		req.Model = ModelGeneric
	}
	if req.Preset == "" {
		req.Preset = PresetGeneral
	}
	if req.UserQuestion == "" {
		req.UserQuestion = "What happens in this video? Provide a grounded summary with timestamps."
	}

	prompt := &BuiltPrompt{
		Model:      req.Model,
		Preset:     req.Preset,
		Transcript: req.Analysis.Transcript,
	}
	for _, f := range req.Analysis.Frames {
		prompt.ImagePaths = append(prompt.ImagePaths, f.Path)
	}
	prompt.ImageCount = len(prompt.ImagePaths)
	prompt.System = buildSystemPrompt(req)
	prompt.User = buildUserPrompt(req)
	prompt.TokenHint = estimateTokens(prompt.ImageCount, len(strings.Fields(req.Analysis.Transcript)))
	return prompt
}

func buildSystemPrompt(req VideoPromptRequest) string {
	base := `You are a video analysis expert. You have been given:
1. A timestamped transcript of the audio track (source: ` + req.Analysis.TranscriptSource + `)
2. A series of extracted frames with t=MM:SS markers

MANDATORY RULES:
- Ground EVERY claim in a specific frame [t=MM:SS] or transcript timestamp
- Quote on-screen text verbatim when visible (code, slides, UI labels)
- Distinguish "seen on screen" vs "heard in audio" vs "inferred"
- If a question cannot be answered from the provided frames+transcript, say so explicitly
- Use the frame's t= marker (e.g. [t=1:23]) when citing visual evidence
`

	switch req.Preset {
	case PresetBugReport:
		base += `
BUG-REPORT MODE:
- Identify the EXACT frame where the issue first appears
- Describe what SHOULD be on screen vs what IS on screen
- Note any error messages, stack traces, or UI states verbatim
- Suggest likely root causes based on visual evidence
`
	case PresetTutorial:
		base += `
TUTORIAL MODE:
- Extract a numbered step-by-step guide
- For each step, cite the frame where it's demonstrated
- List exact commands/keystrokes/code shown on screen
- Note prerequisites and warnings mentioned in audio
`
	case PresetHook:
		base += `
HOOK-ANALYSIS MODE (content creator):
- Analyze the first 3-10 seconds in detail: visual, audio, text overlay
- Identify the hook technique (curiosity gap, pattern interrupt, bold claim, etc.)
- Note the on-screen text/captions style
- Rate the hook's effectiveness (1-10) with justification
`
	case PresetTranscript:
		base += `
TRANSCRIPT-ONLY MODE:
- Ignore frames entirely
- Summarize the spoken content with speaker attribution where possible
- Extract key quotes with timestamps
`
	case PresetComparison:
		base += `
COMPARISON MODE:
- Identify before/after states or A/B scenarios
- Build a side-by-side comparison table
- Cite frames for each state
`
	case PresetSummary:
		base += `
EXECUTIVE-SUMMARY MODE:
- 3-5 bullet points max
- Lead with the most important insight
- Include key timestamps for follow-up viewing
`
	}

	return base
}

func buildUserPrompt(req VideoPromptRequest) string {
	a := req.Analysis
	var sb strings.Builder
	sb.WriteString("# Video Analysis Request\n\n")
	sb.WriteString(fmt.Sprintf("**Title:** %s\n", a.Title))
	sb.WriteString(fmt.Sprintf("**Source:** %s | **Duration:** %s | **Mode:** %s\n",
		a.Source, a.Duration.Round(time.Second), a.Mode))
	sb.WriteString(fmt.Sprintf("**Frames provided:** %d | **Transcript source:** %s\n\n",
		len(a.Frames), a.TranscriptSource))

	if len(a.Frames) > 0 {
		sb.WriteString("## Extracted Frames (Read each image for visual context)\n\n")
		for _, f := range a.Frames {
			sb.WriteString(fmt.Sprintf("- [t=%s] `%s`\n",
				formatDuration(f.Timestamp), filepath.Base(f.Path)))
		}
		sb.WriteString("\n")
	}

	if a.Transcript != "" && a.TranscriptSource != "none" {
		sb.WriteString("## Transcript\n\n")
		transcript := a.Transcript
		if len(transcript) > 15000 {
			transcript = transcript[:15000] + "\n\n[... transcript truncated for context window ...]"
		}
		sb.WriteString(transcript)
		sb.WriteString("\n\n")
	}

	sb.WriteString("## Your Task\n\n")
	switch req.Model {
	case ModelClaude:
		sb.WriteString("I have attached the frames as images. Please Read each one and answer:\n\n")
	case ModelGPT4o:
		sb.WriteString("The frames are provided as image_url attachments in order. Analyze them and answer:\n\n")
	case ModelGemini:
		sb.WriteString("The frames are attached inline. Analyze them alongside the transcript and answer:\n\n")
	default:
		sb.WriteString("Analyze the frames and transcript, then answer:\n\n")
	}

	sb.WriteString(fmt.Sprintf("> %s\n\n", req.UserQuestion))
	sb.WriteString("Cite specific frames with [t=MM:SS] and quote on-screen text verbatim.")
	return sb.String()
}

func estimateTokens(frameCount, transcriptWords int) string {
	frameTokens := frameCount * 1500
	transcriptTokens := int(float64(transcriptWords) * 0.75)
	total := frameTokens + transcriptTokens
	switch {
	case total < 5000:
		return fmt.Sprintf("~%dk tokens (low)", total/1000)
	case total < 20000:
		return fmt.Sprintf("~%dk tokens (moderate)", total/1000)
	case total < 50000:
		return fmt.Sprintf("~%dk tokens (high)", total/1000)
	default:
		return fmt.Sprintf("~%dk tokens (very high — consider --start/--end)", total/1000)
	}
}

func formatDuration(d time.Duration) string {
	total := int(d.Seconds())
	h := total / 3600
	m := (total % 3600) / 60
	s := total % 60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%d:%02d", m, s)
}

// PresetList returns all available presets with descriptions.
func PresetList() map[Preset]string {
	return map[Preset]string{
		PresetGeneral:    "General analysis (default): what happened?",
		PresetBugReport:  "Bug diagnosis: find UI/UX issues",
		PresetTutorial:   "Extract step-by-step instructions",
		PresetHook:       "Content creator: analyze opening hooks",
		PresetTranscript: "Transcript-only (ignore frames)",
		PresetComparison: "Before/after, A/B comparison",
		PresetSummary:    "Executive 3-5 bullet summary",
	}
}

// ModelList returns all supported Vision models.
func ModelList() map[Model]string {
	return map[Model]string{
		ModelClaude:  "Anthropic Claude 3.5/3.7 Sonnet/Opus/Haiku",
		ModelGPT4o:   "OpenAI GPT-4o / GPT-4o-mini",
		ModelGemini:  "Google Gemini 1.5/2.0 Pro/Flash",
		ModelGeneric: "Generic (text + image list)",
	}
}
