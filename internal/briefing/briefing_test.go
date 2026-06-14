// Purpose: Hermetic unit tests for video briefing HTML generation.
// Docs: briefing_test.doc.md

package briefing

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/OpenSIN-Code/web_search_bundle/internal/engines"
	"github.com/OpenSIN-Code/web_search_bundle/internal/prompts"
)

// minimalJPEG returns a valid JPEG header with known dimensions (32x16).
func minimalJPEG(t *testing.T) []byte {
	return []byte{
		0xFF, 0xD8, // SOI
		0xFF, 0xC0, // SOF0 marker
		0x00, 0x0B, // segment length
		0x08,       // precision
		0x00, 0x10, // height = 16
		0x00, 0x20, // width = 32
		0x01, 0x01, 0x11, 0x00, // components
	}
}

func TestSlugify(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"Hello World", "hello-world"},
		{"UPPER---case", "upper-case"},
		{"a-b_c 123", "a-b-c-123"},
		{"", ""},
		{"!@#", ""},
	}
	for _, c := range cases {
		got := slugify(c.in)
		if got != c.want {
			t.Errorf("slugify(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	cases := []struct {
		d    time.Duration
		want string
	}{
		{90 * time.Second, "1:30"},
		{60 * time.Second, "1:00"},
		{3661 * time.Second, "1:01:01"},
		{5 * time.Second, "0:05"},
	}
	for _, c := range cases {
		got := formatDuration(c.d)
		if got != c.want {
			t.Errorf("formatDuration(%v) = %q, want %q", c.d, got, c.want)
		}
	}
}

func TestParseJPEGDimensions(t *testing.T) {
	w, h := parseJPEGDimensions(minimalJPEG(t))
	if w != 32 || h != 16 {
		t.Errorf("parseJPEGDimensions = %dx%d, want 32x16", w, h)
	}
	w, h = parseJPEGDimensions([]byte{0xFF, 0xD9})
	if w != 0 || h != 0 {
		t.Errorf("expected 0x0 for invalid JPEG, got %dx%d", w, h)
	}
}

func TestLoadImageAsDataURL(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "frame.jpg")
	if err := os.WriteFile(path, minimalJPEG(t), 0644); err != nil {
		t.Fatal(err)
	}
	data, width, height, err := loadImageAsDataURL(path, 1024, 75)
	if err != nil {
		t.Fatalf("loadImageAsDataURL error: %v", err)
	}
	if !strings.HasPrefix(data, "data:image/jpeg;base64,") {
		t.Errorf("expected jpeg data URL, got %s", data)
	}
	if width != 32 || height != 16 {
		t.Errorf("loadImageAsDataURL dimensions = %dx%d, want 32x16", width, height)
	}
}

func TestLoadImageAsDataURLPNG(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "frame.png")
	// Minimal PNG header
	png := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	if err := os.WriteFile(path, png, 0644); err != nil {
		t.Fatal(err)
	}
	data, _, _, err := loadImageAsDataURL(path, 1024, 75)
	if err != nil {
		t.Fatalf("loadImageAsDataURL error: %v", err)
	}
	if !strings.HasPrefix(data, "data:image/png;base64,") {
		t.Errorf("expected png data URL, got %s", data)
	}
}

func TestLoadImageAsDataURLFileNotFound(t *testing.T) {
	_, _, _, err := loadImageAsDataURL("/nonexistent/file.jpg", 1024, 75)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGenerateVideoBriefHTML(t *testing.T) {
	dir := t.TempDir()
	framePath := filepath.Join(dir, "frame_0001.jpg")
	if err := os.WriteFile(framePath, minimalJPEG(t), 0644); err != nil {
		t.Fatal(err)
	}

	analysis := &engines.VideoAnalysis{
		URL:              "https://example.com/video",
		Title:            "Test Video",
		Duration:         120 * time.Second,
		Mode:             "full",
		Source:           "youtube",
		Transcript:       "hello world",
		TranscriptSource: "native",
		Frames: []engines.VideoFrame{
			{Path: framePath, Timestamp: 5 * time.Second, Index: 0},
		},
	}
	prompt := &prompts.BuiltPrompt{
		Model:     "claude",
		Preset:    "general",
		TokenHint: "1k",
	}
	outPath := filepath.Join(dir, "brief.html")
	path, err := GenerateVideoBriefHTML(VideoBriefOptions{
		Analysis:    analysis,
		Prompt:      prompt,
		Synthesis:   "summary",
		EmbedFrames: true,
		OutputPath:  outPath,
		Title:       "Test Video Brief",
	})
	if err != nil {
		t.Fatalf("GenerateVideoBriefHTML error: %v", err)
	}
	if path != outPath {
		t.Errorf("expected %s, got %s", outPath, path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	for _, want := range []string{"Test Video Brief", "youtube", "2m0s", "Extracted Frames", "AI Analysis", "Transcript", "hello world"} {
		if !strings.Contains(content, want) {
			t.Errorf("expected HTML to contain %q", want)
		}
	}
}

func TestGenerateVideoBriefHTMLNoAnalysis(t *testing.T) {
	_, err := GenerateVideoBriefHTML(VideoBriefOptions{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGenerateVideoBriefHTMLDefaults(t *testing.T) {
	dir := t.TempDir()
	framePath := filepath.Join(dir, "frame_0001.jpg")
	if err := os.WriteFile(framePath, minimalJPEG(t), 0644); err != nil {
		t.Fatal(err)
	}

	analysis := &engines.VideoAnalysis{
		Title: "Default Video",
		Frames: []engines.VideoFrame{
			{Path: framePath, Timestamp: 1 * time.Second, Index: 0},
		},
	}
	outPath := filepath.Join(dir, "brief.html")
	_, err := GenerateVideoBriefHTML(VideoBriefOptions{
		Analysis:   analysis,
		OutputPath: outPath,
	})
	if err != nil {
		t.Fatalf("GenerateVideoBriefHTML error: %v", err)
	}
	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "Default Video") {
		t.Error("expected title in HTML")
	}
}

func TestGenerateVideoBriefHTMLNoEmbed(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "brief.html")
	analysis := &engines.VideoAnalysis{
		Title: "No Embed",
		Frames: []engines.VideoFrame{
			{Path: filepath.Join(dir, "missing.jpg"), Timestamp: 1 * time.Second, Index: 0},
		},
	}
	_, err := GenerateVideoBriefHTML(VideoBriefOptions{
		Analysis:    analysis,
		OutputPath:  outPath,
		EmbedFrames: false,
	})
	if err != nil {
		t.Fatalf("GenerateVideoBriefHTML error: %v", err)
	}
	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "data:image") {
		t.Error("expected no embedded image data URL")
	}
}
