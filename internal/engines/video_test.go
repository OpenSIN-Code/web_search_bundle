// SPDX-License-Identifier: MIT
// Purpose: Hermetic tests for video analysis helpers and orchestration.
// Docs: video_test.doc.md
package engines

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/OpenSIN-Code/web_search_bundle/internal/imaging"
	"github.com/OpenSIN-Code/web_search_bundle/internal/sidecar"
)

// newTestSidecarManagerWithBinaries creates a sidecar manager with fake yt-dlp and ffmpeg.
func newTestSidecarManagerWithBinaries(t *testing.T, ytdlpScript, ffmpegScript string) (*sidecar.Manager, string) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	binDir := filepath.Join(home, ".sin-websearch", "bin")
	if err := os.MkdirAll(binDir, 0750); err != nil {
		t.Fatal(err)
	}

	fakeYtdlp := filepath.Join(binDir, "yt-dlp")
	if err := os.WriteFile(fakeYtdlp, []byte(ytdlpScript), 0755); err != nil {
		t.Fatal(err)
	}

	fakeFfmpeg := filepath.Join(binDir, "ffmpeg")
	if err := os.WriteFile(fakeFfmpeg, []byte(ffmpegScript), 0755); err != nil {
		t.Fatal(err)
	}

	sc, err := sidecar.NewManager()
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	return sc, home
}

const fakeYtdlpMetadata = `#!/bin/sh
if [ "$2" = "--dump-json" ]; then
  echo '{"title":"Test Video","duration":120}'
  exit 0
fi
output=""
prev=""
for i in "$@"; do
  if [ "$prev" = "-o" ]; then
    output="$i"
  fi
  prev="$i"
done
printf '%s\n' "WEBVTT" "" "00:00:00.000 --> 00:00:02.000" "Hello world" "" "00:00:02.000 --> 00:00:04.000" "This is a longer caption line to exceed the one hundred character minimum length threshold" > "${output}.en.vtt"
exit 0
`

const fakeFfmpegFrames = `#!/bin/sh
output=""
for i in "$@"; do
  output="$i"
done
dir=$(dirname "$output")
if echo "$output" | grep -q "frame_"; then
  touch "${dir}/frame_0001.jpg"
  touch "${dir}/frame_0002.jpg"
fi
exit 0
`

func TestVideoCleanup(t *testing.T) {
	dir := t.TempDir()
	a := &VideoAnalysis{WorkDir: dir}
	e := &VideoEngine{}
	if err := e.Cleanup(a); err != nil {
		t.Fatalf("Cleanup error: %v", err)
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Errorf("expected workdir to be removed")
	}
}

func TestVideoEngineGetMetadata(t *testing.T) {
	sc, _ := newTestSidecarManagerWithBinaries(t, fakeYtdlpMetadata, fakeFfmpegFrames)
	e := NewVideoEngine(sc)
	meta, err := e.getMetadata(context.Background(), "https://youtube.com/watch?v=1")
	if err != nil {
		t.Fatalf("getMetadata error: %v", err)
	}
	if meta.Title != "Test Video" {
		t.Errorf("title = %q, want Test Video", meta.Title)
	}
	if meta.Duration != 120 {
		t.Errorf("duration = %d, want 120", meta.Duration)
	}
}

func TestVideoEngineGetMetadataFallback(t *testing.T) {
	sc, _ := newTestSidecarManagerWithBinaries(t, "#!/bin/sh\nexit 1\n", fakeFfmpegFrames)
	e := NewVideoEngine(sc)
	meta, err := e.getMetadata(context.Background(), "https://example.com/video.mp4")
	if err != nil {
		t.Fatalf("getMetadata error: %v", err)
	}
	if meta.Title != "https://example.com/video.mp4" {
		t.Errorf("title = %q, want URL fallback", meta.Title)
	}
	if meta.Duration != 0 {
		t.Errorf("duration = %d, want 0", meta.Duration)
	}
}

func TestVideoEngineExtractFrames(t *testing.T) {
	sc, _ := newTestSidecarManagerWithBinaries(t, fakeYtdlpMetadata, fakeFfmpegFrames)
	e := NewVideoEngine(sc)
	sessionDir := t.TempDir()
	opts := WatchOptions{URL: "https://youtube.com/watch?v=1", MaxFrames: 10, Resolution: 768}
	frames, err := e.extractFrames(context.Background(), opts, sessionDir, 120)
	if err != nil {
		t.Fatalf("extractFrames error: %v", err)
	}
	if len(frames) != 2 {
		t.Fatalf("expected 2 frames, got %d", len(frames))
	}
	if frames[0].Index != 0 {
		t.Errorf("first frame index = %d, want 0", frames[0].Index)
	}
}

func TestVideoEngineGetTranscriptNative(t *testing.T) {
	sc, _ := newTestSidecarManagerWithBinaries(t, fakeYtdlpMetadata, fakeFfmpegFrames)
	e := NewVideoEngine(sc)
	sessionDir := t.TempDir()
	opts := WatchOptions{URL: "https://youtube.com/watch?v=1", Whisper: "none"}
	meta := &videoMetadata{Title: "Test", Duration: 120}
	text, source, audioPath, err := e.getTranscript(context.Background(), opts, sessionDir, meta)
	if err != nil {
		t.Fatalf("getTranscript error: %v", err)
	}
	if source != "native" {
		t.Errorf("source = %q, want native", source)
	}
	if text == "" {
		t.Errorf("expected non-empty transcript")
	}
	if audioPath != "" {
		t.Errorf("audioPath = %q, want empty", audioPath)
	}
}

func TestVideoEngineWatch(t *testing.T) {
	sc, _ := newTestSidecarManagerWithBinaries(t, fakeYtdlpMetadata, fakeFfmpegFrames)
	e := NewVideoEngine(sc)
	opts := WatchOptions{URL: "https://youtube.com/watch?v=1", MaxFrames: 10, Resolution: 768}
	analysis, err := e.Watch(context.Background(), opts)
	if err != nil {
		t.Fatalf("Watch error: %v", err)
	}
	if analysis.Title != "Test Video" {
		t.Errorf("title = %q, want Test Video", analysis.Title)
	}
	if analysis.Source != "youtube" {
		t.Errorf("source = %q, want youtube", analysis.Source)
	}
	if len(analysis.Frames) != 2 {
		t.Errorf("frames = %d, want 2", len(analysis.Frames))
	}
	if analysis.TranscriptSource != "native" {
		t.Errorf("transcript source = %q, want native", analysis.TranscriptSource)
	}
}

func TestVideoEngineWatchFocusedMode(t *testing.T) {
	sc, _ := newTestSidecarManagerWithBinaries(t, fakeYtdlpMetadata, fakeFfmpegFrames)
	e := NewVideoEngine(sc)
	opts := WatchOptions{URL: "https://youtube.com/watch?v=1", Start: "10", End: "20", MaxFrames: 5, Resolution: 768}
	analysis, err := e.Watch(context.Background(), opts)
	if err != nil {
		t.Fatalf("Watch error: %v", err)
	}
	if analysis.Mode != "focused" {
		t.Errorf("mode = %q, want focused", analysis.Mode)
	}
}

func TestVideoEngineWatchWithResizeEmptyFrames(t *testing.T) {
	// Fake ffmpeg exits 0 but writes no frames so the resize short-circuit is exercised.
	sc, _ := newTestSidecarManagerWithBinaries(t, "#!/bin/sh\nexit 0\n", "#!/bin/sh\nexit 0\n")
	e := NewVideoEngine(sc)
	outDir := t.TempDir()
	opts := WatchOptions{URL: "https://unknown.example.com/video", MaxFrames: 10, Resolution: 768, OutDir: outDir}
	analysis, err := e.WatchWithResize(context.Background(), opts, imaging.ResizeOptions{MaxWidth: 100, JPEGQuality: 75})
	if err != nil {
		t.Fatalf("WatchWithResize error: %v", err)
	}
	if len(analysis.Frames) != 0 {
		t.Errorf("expected 0 frames, got %d", len(analysis.Frames))
	}
}

func TestVideoEngineWatchWithOutDir(t *testing.T) {
	sc, _ := newTestSidecarManagerWithBinaries(t, fakeYtdlpMetadata, fakeFfmpegFrames)
	e := NewVideoEngine(sc)
	outDir := t.TempDir()
	opts := WatchOptions{URL: "https://youtube.com/watch?v=1", OutDir: outDir, MaxFrames: 10, Resolution: 768}
	analysis, err := e.Watch(context.Background(), opts)
	if err != nil {
		t.Fatalf("Watch error: %v", err)
	}
	if analysis.WorkDir != outDir {
		t.Errorf("workDir = %q, want %q", analysis.WorkDir, outDir)
	}
}

func TestVideoEngineExtractFramesMissingFfmpeg(t *testing.T) {
	// Re-create sidecar manager with only yt-dlp so ffmpeg is missing.
	home := t.TempDir()
	t.Setenv("HOME", home)
	binDir := filepath.Join(home, ".sin-websearch", "bin")
	if err := os.MkdirAll(binDir, 0750); err != nil {
		t.Fatal(err)
	}
	fakeYtdlp := filepath.Join(binDir, "yt-dlp")
	if err := os.WriteFile(fakeYtdlp, []byte(fakeYtdlpMetadata), 0755); err != nil {
		t.Fatal(err)
	}
	sc, err := sidecar.NewManager()
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	e := NewVideoEngine(sc)
	_, err = e.extractFrames(context.Background(), WatchOptions{URL: "https://example.com"}, t.TempDir(), 120)
	if err == nil {
		t.Fatal("expected error for missing ffmpeg")
	}
}

func TestVideoEngineGetTranscriptDisabled(t *testing.T) {
	sc, _ := newTestSidecarManagerWithBinaries(t, "#!/bin/sh\nexit 1\n", fakeFfmpegFrames)
	e := NewVideoEngine(sc)
	sessionDir := t.TempDir()
	opts := WatchOptions{URL: "https://example.com/video", Whisper: "none"}
	meta := &videoMetadata{Title: "Test", Duration: 120}
	_, source, _, err := e.getTranscript(context.Background(), opts, sessionDir, meta)
	if err == nil {
		t.Fatal("expected error")
	}
	if source != "disabled" {
		t.Errorf("source = %q, want disabled", source)
	}
}
