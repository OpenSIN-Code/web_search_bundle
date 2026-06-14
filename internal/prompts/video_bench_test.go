// SPDX-License-Identifier: MIT
// Purpose: Benchmark video prompt construction for Vision-LLM requests.
// Docs: video.doc.md
package prompts

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/OpenSIN-Code/web_search_bundle/internal/engines"
)

func makeFrames(n int) []engines.VideoFrame {
	frames := make([]engines.VideoFrame, n)
	for i := 0; i < n; i++ {
		frames[i] = engines.VideoFrame{
			Path:      fmt.Sprintf("/tmp/frame-%04d.jpg", i),
			Timestamp: time.Duration(i*5) * time.Second,
			Index:     i,
		}
	}
	return frames
}

func makeTranscript(words int) string {
	parts := make([]string, words)
	for i := 0; i < words; i++ {
		parts[i] = fmt.Sprintf("word%d", i)
	}
	return strings.Join(parts, " ")
}

func makeAnalysis(frames, words int) *engines.VideoAnalysis {
	return &engines.VideoAnalysis{
		URL:              "https://example.com/video",
		Title:            "Benchmark Video",
		Duration:         10 * time.Minute,
		Frames:           makeFrames(frames),
		Transcript:       makeTranscript(words),
		TranscriptSource: "whisper-local",
		FrameCount:       frames,
		Mode:             "full",
		Source:           "youtube",
	}
}

func BenchmarkBuildVideoPromptSmall(b *testing.B) {
	a := makeAnalysis(8, 200)
	req := VideoPromptRequest{Model: ModelClaude, Preset: PresetGeneral, Analysis: a}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = BuildVideoPrompt(req)
	}
}

func BenchmarkBuildVideoPromptMedium(b *testing.B) {
	a := makeAnalysis(16, 2000)
	req := VideoPromptRequest{Model: ModelGPT4o, Preset: PresetTutorial, Analysis: a}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = BuildVideoPrompt(req)
	}
}

func BenchmarkBuildVideoPromptLarge(b *testing.B) {
	a := makeAnalysis(32, 10000)
	req := VideoPromptRequest{Model: ModelGemini, Preset: PresetBugReport, Analysis: a}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = BuildVideoPrompt(req)
	}
}

func BenchmarkBuildVideoPromptAllPresets(b *testing.B) {
	a := makeAnalysis(8, 500)
	presets := []Preset{PresetGeneral, PresetBugReport, PresetTutorial, PresetHook, PresetTranscript, PresetComparison, PresetSummary}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, preset := range presets {
			_ = BuildVideoPrompt(VideoPromptRequest{Model: ModelGeneric, Preset: preset, Analysis: a})
		}
	}
}

func BenchmarkPresetList(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = PresetList()
	}
}

func BenchmarkModelList(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ModelList()
	}
}
