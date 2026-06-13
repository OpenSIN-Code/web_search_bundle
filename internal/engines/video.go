// Purpose: Deep multimodal video analysis (frames + transcript).
// Docs: internal/engines/video.doc.md
package engines

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/OpenSIN-Code/web_search_bundle/internal/imaging"
	"github.com/OpenSIN-Code/web_search_bundle/internal/sidecar"
)

// VideoAnalysis is the result of analyzing a video.
type VideoAnalysis struct {
	URL              string        `json:"url"`
	Title            string        `json:"title"`
	Duration         time.Duration `json:"duration"`
	Frames           []VideoFrame  `json:"frames"`
	Transcript       string        `json:"transcript"`
	TranscriptSource string        `json:"transcript_source"`
	AudioPath        string        `json:"audio_path,omitempty"`
	WorkDir          string        `json:"work_dir"`
	FrameCount       int           `json:"frame_count"`
	Mode             string        `json:"mode"`
	StartOffset      time.Duration `json:"start_offset,omitempty"`
	EndOffset        time.Duration `json:"end_offset,omitempty"`
	Source           string        `json:"source"`
}

// VideoFrame represents an extracted frame.
type VideoFrame struct {
	Path      string        `json:"path"`
	Timestamp time.Duration `json:"timestamp"`
	Index     int           `json:"index"`
}

// WatchOptions configures video analysis.
type WatchOptions struct {
	URL        string
	Start      string
	End        string
	MaxFrames  int
	Resolution int
	FPS        float64
	Whisper    string
	OutDir     string
}

// VideoEngine analyzes videos multimodally.
type VideoEngine struct {
	sidecar     *sidecar.Manager
	workDir     string
	maxFrames   int
	maxFPS      float64
	whisperPref string
}

// NewVideoEngine creates a video engine.
func NewVideoEngine(sc *sidecar.Manager) *VideoEngine {
	workDir := filepath.Join(os.TempDir(), "sin-websearch-video")
	if err := os.MkdirAll(workDir, 0755); err != nil {
		// Best-effort temp directory creation; log and continue.
		fmt.Fprintf(os.Stderr, "video workdir: %v\n", err)
	}
	return &VideoEngine{
		sidecar:     sc,
		workDir:     workDir,
		maxFrames:   100,
		maxFPS:      2.0,
		whisperPref: detectWhisperPref(),
	}
}

// Watch analyzes a video.
func (e *VideoEngine) Watch(ctx context.Context, opts WatchOptions) (*VideoAnalysis, error) {
	if opts.MaxFrames == 0 {
		opts.MaxFrames = 80
	}
	if opts.Resolution == 0 {
		opts.Resolution = 768
	}
	if opts.Whisper == "" {
		opts.Whisper = e.whisperPref
	}

	hash := sha256.Sum256([]byte(opts.URL + opts.Start + opts.End))
	sessionDir := filepath.Join(e.workDir, hex.EncodeToString(hash[:])[:12])
	if opts.OutDir != "" {
		sessionDir = opts.OutDir
	}
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return nil, fmt.Errorf("session dir: %w", err)
	}

	analysis := &VideoAnalysis{
		URL:     opts.URL,
		WorkDir: sessionDir,
		Mode:    "full",
		Source:  detectVideoSource(opts.URL),
	}
	if opts.Start != "" || opts.End != "" {
		analysis.Mode = "focused"
	}

	meta, err := e.getMetadata(ctx, opts.URL)
	if err != nil {
		return nil, fmt.Errorf("metadata: %w", err)
	}
	analysis.Title = meta.Title
	analysis.Duration = time.Duration(meta.Duration) * time.Second

	frames, err := e.extractFrames(ctx, opts, sessionDir, meta.Duration)
	if err != nil {
		return nil, fmt.Errorf("frames: %w", err)
	}
	analysis.Frames = frames
	analysis.FrameCount = len(frames)

	transcript, source, audioPath, err := e.getTranscript(ctx, opts, sessionDir, meta)
	if err != nil {
		transcript = fmt.Sprintf("[transcript unavailable: %v]", err)
		source = "none"
	}
	analysis.Transcript = transcript
	analysis.TranscriptSource = source
	analysis.AudioPath = audioPath

	return analysis, nil
}

type videoMetadata struct {
	Title    string `json:"title"`
	Duration int    `json:"duration"`
}

func (e *VideoEngine) getMetadata(ctx context.Context, url string) (*videoMetadata, error) {
	ytdlp, err := e.sidecar.EnsureBinary("yt-dlp")
	if err != nil {
		return &videoMetadata{Title: url, Duration: 0}, nil
	}
	cmd := exec.CommandContext(ctx, ytdlp, url, "--dump-json", "--no-download", "--no-warnings", "--quiet")
	out, err := cmd.Output()
	if err != nil {
		return &videoMetadata{Title: url, Duration: 0}, nil
	}
	var meta videoMetadata
	if err := json.Unmarshal(out, &meta); err != nil {
		return &videoMetadata{Title: url, Duration: 0}, nil
	}
	return &meta, nil
}

func (e *VideoEngine) extractFrames(ctx context.Context, opts WatchOptions, sessionDir string, durationSec int) ([]VideoFrame, error) {
	ffmpeg, err := e.sidecar.EnsureBinary("ffmpeg")
	if err != nil {
		return nil, fmt.Errorf("ffmpeg not available: %w", err)
	}

	duration := float64(durationSec)
	startSec := parseVideoTime(opts.Start)
	endSec := parseVideoTime(opts.End)
	if endSec == 0 {
		endSec = duration
	}
	if endSec == 0 {
		endSec = 300 // fallback: 5 minutes
	}
	windowDuration := endSec - startSec

	fps := float64(opts.MaxFrames) / windowDuration
	if fps > e.maxFPS {
		fps = e.maxFPS
	}
	if opts.FPS > 0 {
		fps = opts.FPS
		if fps > e.maxFPS {
			fps = e.maxFPS
		}
	}

	args := []string{"-y"}
	if startSec > 0 {
		args = append(args, "-ss", fmt.Sprintf("%.2f", startSec))
	}
	args = append(args, "-i", opts.URL)
	if endSec > 0 && endSec < duration {
		args = append(args, "-to", fmt.Sprintf("%.2f", endSec))
	}
	args = append(args,
		"-vf", fmt.Sprintf("fps=%v,scale=%v:-1", fps, opts.Resolution),
		"-q:v", "2",
		filepath.Join(sessionDir, "frame_%04d.jpg"),
	)

	cmd := exec.CommandContext(ctx, ffmpeg, args...)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg: %w", err)
	}

	var frames []VideoFrame
	entries, err := os.ReadDir(sessionDir)
	if err != nil {
		return nil, err
	}
	idx := 0
	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), "frame_") {
			continue
		}
		ts := float64(idx) / fps
		frames = append(frames, VideoFrame{
			Path:      filepath.Join(sessionDir, entry.Name()),
			Timestamp: time.Duration(ts * float64(time.Second)),
			Index:     idx,
		})
		idx++
	}

	return frames, nil
}

func (e *VideoEngine) getTranscript(ctx context.Context, opts WatchOptions, sessionDir string, meta *videoMetadata) (string, string, string, error) {
	ytdlp, err := e.sidecar.EnsureBinary("yt-dlp")
	if err != nil {
		return "", "disabled", "", fmt.Errorf("yt-dlp unavailable: %w", err)
	}

	cmd := exec.CommandContext(ctx, ytdlp,
		opts.URL,
		"--write-auto-sub", "--write-sub",
		"--sub-lang", "en,de",
		"--skip-download",
		"--sub-format", "vtt",
		"-o", filepath.Join(sessionDir, "subs"),
		"--no-warnings", "--quiet",
	)
	if err := cmd.Run(); err == nil {
		for _, lang := range []string{"en", "de"} {
			path := filepath.Join(sessionDir, "subs."+lang+".vtt")
			if data, err := os.ReadFile(path); err == nil {
				text := parseVTT(string(data))
				if len(text) > 100 {
					return text, "native", "", nil
				}
			}
		}
	}

	if opts.Whisper == "none" {
		return "", "disabled", "", fmt.Errorf("whisper disabled, no native captions")
	}

	ffmpeg, err := e.sidecar.EnsureBinary("ffmpeg")
	if err != nil {
		return "", "none", "", fmt.Errorf("ffmpeg not available for audio extraction: %w", err)
	}

	audioPath := filepath.Join(sessionDir, "audio.wav")
	cmd = exec.CommandContext(ctx, ffmpeg,
		"-y", "-i", opts.URL,
		"-vn", "-acodec", "pcm_s16le", "-ar", "16000", "-ac", "1",
		audioPath,
	)
	if err := cmd.Run(); err != nil {
		return "", "none", audioPath, fmt.Errorf("audio extraction failed: %w", err)
	}

	text, err := transcribe(ctx, audioPath, opts.Whisper)
	if err != nil {
		return "", "none", audioPath, err
	}
	return text, "whisper-" + opts.Whisper, audioPath, nil
}

// WatchWithResize analyzes a video and resizes all frames in place.
func (e *VideoEngine) WatchWithResize(ctx context.Context, opts WatchOptions, resizeOpts imaging.ResizeOptions) (*VideoAnalysis, error) {
	analysis, err := e.Watch(ctx, opts)
	if err != nil {
		return nil, err
	}
	if len(analysis.Frames) == 0 {
		return analysis, nil
	}

	resizedDir := filepath.Join(analysis.WorkDir, "resized")
	inputPaths := make([]string, len(analysis.Frames))
	for i, f := range analysis.Frames {
		inputPaths[i] = f.Path
	}
	resizedPaths, err := imaging.ResizeBatch(inputPaths, resizedDir, resizeOpts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "⚠ frame resize failed: %v\n", err)
		return analysis, nil
	}
	for i := range analysis.Frames {
		if i < len(resizedPaths) {
			analysis.Frames[i].Path = resizedPaths[i]
		}
	}
	return analysis, nil
}

// Cleanup removes the working directory.
func (e *VideoEngine) Cleanup(analysis *VideoAnalysis) error {
	return os.RemoveAll(analysis.WorkDir)
}

func parseVideoTime(s string) float64 {
	if s == "" {
		return 0
	}
	if strings.Contains(s, ":") {
		parts := strings.Split(s, ":")
		if len(parts) == 2 {
			m, _ := strconv.Atoi(parts[0])
			sec, _ := strconv.Atoi(parts[1])
			return float64(m*60 + sec)
		}
		if len(parts) == 3 {
			h, _ := strconv.Atoi(parts[0])
			m, _ := strconv.Atoi(parts[1])
			sec, _ := strconv.Atoi(parts[2])
			return float64(h*3600 + m*60 + sec)
		}
	}
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func parseVTT(vtt string) string {
	var lines []string
	for _, line := range strings.Split(vtt, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || line == "WEBVTT" || strings.Contains(line, "-->") {
			continue
		}
		if _, err := strconv.Atoi(line); err == nil {
			continue
		}
		clean := stripHTML(line)
		if clean != "" {
			lines = append(lines, clean)
		}
	}
	return strings.Join(dedupeLines(lines), " ")
}

func dedupeLines(lines []string) []string {
	if len(lines) == 0 {
		return nil
	}
	var out []string
	prev := ""
	for _, l := range lines {
		if l != prev {
			out = append(out, l)
			prev = l
		}
	}
	return out
}

func stripHTML(s string) string {
	var out []byte
	inTag := false
	for i := 0; i < len(s); i++ {
		if s[i] == '<' {
			inTag = true
			continue
		}
		if s[i] == '>' {
			inTag = false
			continue
		}
		if !inTag {
			out = append(out, s[i])
		}
	}
	return string(out)
}

func detectWhisperPref() string {
	if os.Getenv("GROQ_API_KEY") != "" {
		return "groq"
	}
	if os.Getenv("OPENAI_API_KEY") != "" {
		return "openai"
	}
	return "none"
}

func detectVideoSource(url string) string {
	switch {
	case strings.Contains(url, "youtube.com") || strings.Contains(url, "youtu.be"):
		return "youtube"
	case strings.Contains(url, "tiktok.com"):
		return "tiktok"
	case strings.Contains(url, "instagram.com"):
		return "instagram"
	case strings.Contains(url, "x.com") || strings.Contains(url, "twitter.com"):
		return "x"
	case strings.Contains(url, "vimeo.com"):
		return "vimeo"
	case strings.Contains(url, "loom.com"):
		return "loom"
	default:
		if _, err := os.Stat(url); err == nil {
			return "local"
		}
		return "unknown"
	}
}
