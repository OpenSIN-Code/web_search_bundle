// Purpose: Generate self-contained HTML briefings with embedded video frames.
// Docs: internal/briefing/video_html.doc.md
package briefing

import (
	"encoding/base64"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/OpenSIN-Code/web_search_bundle/internal/engines"
	"github.com/OpenSIN-Code/web_search_bundle/internal/prompts"
)

// VideoBriefOptions configures HTML generation.
type VideoBriefOptions struct {
	Analysis    *engines.VideoAnalysis
	Prompt      *prompts.BuiltPrompt
	Synthesis   string
	EmbedFrames bool
	JPEGQuality int
	MaxWidth    int
	OutputPath  string
	Title       string
}

// embeddedFrame is a frame embedded in the HTML briefing.
type embeddedFrame struct {
	Index     int
	Timestamp string
	DataURL   string
	Width     int
	Height    int
	SizeKB    int
}

// GenerateVideoBriefHTML creates a self-contained HTML file with embedded frames.
func GenerateVideoBriefHTML(opts VideoBriefOptions) (string, error) {
	if opts.Analysis == nil {
		return "", fmt.Errorf("analysis required")
	}
	if opts.JPEGQuality == 0 {
		opts.JPEGQuality = 75
	}
	if opts.MaxWidth == 0 {
		opts.MaxWidth = 1024
	}
	if opts.Title == "" {
		opts.Title = opts.Analysis.Title
	}
	if !opts.EmbedFrames {
		opts.EmbedFrames = true
	}

	var frames []embeddedFrame

	if opts.EmbedFrames {
		for i, f := range opts.Analysis.Frames {
			data, width, height, err := loadImageAsDataURL(f.Path, opts.MaxWidth, opts.JPEGQuality)
			if err != nil {
				fmt.Fprintf(os.Stderr, "⚠ frame %d skipped: %v\n", i, err)
				continue
			}
			frames = append(frames, embeddedFrame{
				Index:     i,
				Timestamp: formatDuration(f.Timestamp),
				DataURL:   data,
				Width:     width,
				Height:    height,
				SizeKB:    len(data) / 1024,
			})
		}
	}

	if opts.OutputPath == "" {
		slug := slugify(opts.Analysis.Title)
		if slug == "" {
			slug = "video"
		}
		home, _ := os.UserHomeDir()
		dir := filepath.Join(home, "Documents", "SIN-Briefings")
		if err := os.MkdirAll(dir, 0755); err != nil {
			return "", fmt.Errorf("briefings dir: %w", err)
		}
		opts.OutputPath = filepath.Join(dir, fmt.Sprintf("%s-%s.html", slug, time.Now().Format("2006-01-02-1504")))
	}

	if err := os.MkdirAll(filepath.Dir(opts.OutputPath), 0755); err != nil {
		return "", err
	}
	content := renderVideoHTML(opts, frames)
	if err := os.WriteFile(opts.OutputPath, []byte(content), 0644); err != nil {
		return "", err
	}

	return opts.OutputPath, nil
}

func loadImageAsDataURL(path string, maxWidth, jpegQuality int) (string, int, int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", 0, 0, err
	}
	mime := "image/jpeg"
	if len(data) >= 8 {
		switch {
		case data[0] == 0x89 && string(data[1:4]) == "PNG":
			mime = "image/png"
		case data[0] == 0xFF && data[1] == 0xD8:
			mime = "image/jpeg"
		case string(data[0:4]) == "RIFF" && string(data[8:12]) == "WEBP":
			mime = "image/webp"
		case data[0] == 'G' && data[1] == 'I' && data[2] == 'F':
			mime = "image/gif"
		}
	}
	width, height := parseJPEGDimensions(data)
	if width == 0 {
		width = maxWidth
		height = int(float64(maxWidth) * 0.5625)
	}
	_ = jpegQuality
	encoded := base64.StdEncoding.EncodeToString(data)
	return fmt.Sprintf("data:%s;base64,%s", mime, encoded), width, height, nil
}

func parseJPEGDimensions(data []byte) (int, int) {
	if len(data) < 2 || data[0] != 0xFF || data[1] != 0xD8 {
		return 0, 0
	}
	for i := 2; i < len(data)-1; {
		if data[i] != 0xFF {
			break
		}
		marker := data[i+1]
		if marker >= 0xC0 && marker <= 0xCF && marker != 0xC4 && marker != 0xC8 && marker != 0xCC {
			if i+9 < len(data) {
				height := int(data[i+5])<<8 | int(data[i+6])
				width := int(data[i+7])<<8 | int(data[i+8])
				return width, height
			}
		}
		if i+3 >= len(data) {
			break
		}
		segLen := int(data[i+2])<<8 | int(data[i+3])
		i += 2 + segLen
	}
	return 0, 0
}

func renderVideoHTML(opts VideoBriefOptions, frames []embeddedFrame) string {
	a := opts.Analysis
	escapedTitle := html.EscapeString(opts.Title)

	var framesGrid strings.Builder
	for _, f := range frames {
		framesGrid.WriteString(fmt.Sprintf(`
            <div class="frame" id="frame-%d" onclick="openModal('%d')">
                <img src="%s" alt="Frame at %s" loading="lazy">
                <div class="frame-label">t=%s</div>
            </div>`, f.Index, f.Index, f.DataURL, f.Timestamp, f.Timestamp))
	}

	synthesisBlock := ""
	if opts.Synthesis != "" && opts.Prompt != nil {
		synthesisBlock = fmt.Sprintf(`
        <section class="synthesis">
            <h2>🤖 AI Analysis</h2>
            <div class="synthesis-content">%s</div>
            <div class="synthesis-meta">
                <span class="badge">Model: %s</span>
                <span class="badge">Preset: %s</span>
                <span class="badge">%s</span>
            </div>
        </section>`,
			html.EscapeString(opts.Synthesis), opts.Prompt.Model, opts.Prompt.Preset, opts.Prompt.TokenHint)
	}

	transcriptBlock := ""
	if a.Transcript != "" && a.TranscriptSource != "none" {
		transcriptBlock = fmt.Sprintf(`
        <section class="transcript">
            <h2>📝 Transcript</h2>
            <div class="transcript-meta">Source: %s</div>
            <div class="transcript-content">%s</div>
        </section>`,
			html.EscapeString(a.TranscriptSource), html.EscapeString(a.Transcript))
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>%s — Video Brief</title>
<style>
:root {
    --bg: #0f0f10; --bg-elev: #1a1a1c; --bg-card: #232326; --border: #2e2e32;
    --text: #e8e8ea; --text-dim: #9a9aa0; --accent: #6366f1; --accent-hover: #818cf8;
    --success: #10b981; --warn: #f59e0b;
}
* { box-sizing: border-box; }
html, body { margin: 0; padding: 0; background: var(--bg); color: var(--text);
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Inter, Roboto, sans-serif;
    font-size: 15px; line-height: 1.6; }
.container { max-width: 1200px; margin: 0 auto; padding: 2rem 1.5rem; }
header { border-bottom: 1px solid var(--border); padding-bottom: 1.5rem; margin-bottom: 2rem; }
h1 { margin: 0 0 0.5rem 0; font-size: 1.8rem; font-weight: 600; }
.meta { color: var(--text-dim); font-size: 0.9rem; display: flex; flex-wrap: wrap; gap: 1rem; }
.badge { display: inline-block; padding: 0.2rem 0.6rem; background: var(--bg-card); border: 1px solid var(--border); border-radius: 4px; font-size: 0.75rem; color: var(--text-dim); font-family: "SF Mono", Monaco, monospace; }
.badge.accent { background: var(--accent); color: white; border-color: var(--accent); }
section { margin-bottom: 2.5rem; }
h2 { font-size: 1.25rem; font-weight: 600; margin: 0 0 1rem 0; padding-bottom: 0.5rem; border-bottom: 1px solid var(--border); }
.frames-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(220px, 1fr)); gap: 1rem; }
.frame { position: relative; background: var(--bg-card); border: 1px solid var(--border); border-radius: 8px; overflow: hidden; cursor: pointer; transition: transform 0.15s, border-color 0.15s; aspect-ratio: 16/9; }
.frame:hover { transform: translateY(-2px); border-color: var(--accent); }
.frame img { width: 100%%; height: 100%%; object-fit: cover; display: block; }
.frame-label { position: absolute; bottom: 0; left: 0; right: 0; background: linear-gradient(to top, rgba(0,0,0,0.85), transparent); padding: 1.5rem 0.6rem 0.4rem 0.6rem; font-family: "SF Mono", Monaco, monospace; font-size: 0.85rem; color: white; font-weight: 600; }
.synthesis, .transcript { background: var(--bg-elev); border: 1px solid var(--border); border-radius: 8px; padding: 1.5rem; }
.synthesis-content, .transcript-content { white-space: pre-wrap; font-size: 0.95rem; }
.synthesis-meta { margin-top: 1rem; display: flex; flex-wrap: wrap; gap: 0.5rem; }
.modal { display: none; position: fixed; inset: 0; background: rgba(0,0,0,0.92); z-index: 1000; align-items: center; justify-content: center; padding: 2rem; }
.modal.active { display: flex; }
.modal img { max-width: 95%%; max-height: 90vh; border-radius: 4px; box-shadow: 0 20px 60px rgba(0,0,0,0.5); }
.modal-close { position: absolute; top: 1rem; right: 1.5rem; background: transparent; border: none; color: white; font-size: 2rem; cursor: pointer; padding: 0.5rem 1rem; }
.modal-close:hover { color: var(--accent); }
.modal-label { position: absolute; bottom: 1rem; left: 50%%; transform: translateX(-50%%); background: rgba(0,0,0,0.7); color: white; padding: 0.4rem 1rem; border-radius: 4px; font-family: "SF Mono", Monaco, monospace; }
footer { margin-top: 3rem; padding-top: 1.5rem; border-top: 1px solid var(--border); color: var(--text-dim); font-size: 0.8rem; text-align: center; }
footer code { background: var(--bg-card); padding: 0.1rem 0.4rem; border-radius: 3px; font-size: 0.75rem; }
@media print { body { background: white; color: black; } .frame { break-inside: avoid; border: 1px solid #ccc; } .frame-label { background: #f5f5f5; color: black; } .modal { display: none !important; } }
@media (max-width: 640px) { .frames-grid { grid-template-columns: repeat(auto-fill, minmax(140px, 1fr)); gap: 0.5rem; } h1 { font-size: 1.4rem; } .container { padding: 1rem; } }
</style>
</head>
<body>
<div class="container">
    <header>
        <h1>🎬 %s</h1>
        <div class="meta">
            <span class="meta-item"><strong>Source:</strong> %s</span>
            <span class="meta-item"><strong>Duration:</strong> %s</span>
            <span class="meta-item"><strong>Mode:</strong> %s</span>
            <span class="meta-item"><strong>Frames:</strong> %d</span>
            <span class="meta-item"><strong>Transcript:</strong> %s</span>
        </div>
        <div style="margin-top: 0.75rem;">
            <span class="badge accent">Generated %s</span>
            <span class="badge">sin-websearch vbrief</span>
        </div>
    </header>
    %s
    <section>
        <h2>🖼️ Extracted Frames (%d)</h2>
        <p style="color: var(--text-dim); font-size: 0.9rem;">Click any frame to view full-size. Each frame is timestamped for cross-referencing with the transcript.</p>
        <div class="frames-grid">
            %s
        </div>
    </section>
    %s
    <footer>
        <p>Generated by <code>sin-websearch vbrief</code> on %s</p>
        <p>Re-run: <code>sin-websearch vbrief "%s"</code></p>
        <p style="margin-top: 0.5rem;">Self-contained HTML • All frames embedded as base64 • Works offline</p>
    </footer>
</div>
<div class="modal" id="modal" onclick="closeModal()">
    <button class="modal-close" onclick="closeModal()">&times;</button>
    <img id="modal-img" src="" alt="">
    <div class="modal-label" id="modal-label"></div>
</div>
<script>
function openModal(idx) {
    const frame = document.getElementById('frame-' + idx);
    const img = frame.querySelector('img');
    const label = frame.querySelector('.frame-label');
    document.getElementById('modal-img').src = img.src;
    document.getElementById('modal-label').textContent = label.textContent;
    document.getElementById('modal').classList.add('active');
}
function closeModal() { document.getElementById('modal').classList.remove('active'); }
document.addEventListener('keydown', (e) => { if (e.key === 'Escape') closeModal(); });
</script>
</body>
</html>`,
		escapedTitle, escapedTitle, html.EscapeString(a.Source), a.Duration.Round(time.Second),
		a.Mode, len(a.Frames), html.EscapeString(a.TranscriptSource), time.Now().Format("2006-01-02 15:04"),
		synthesisBlock, len(frames), framesGrid.String(), transcriptBlock,
		time.Now().Format("2006-01-02 15:04:05"), html.EscapeString(a.URL),
	)
}

func slugify(s string) string {
	s = strings.ToLower(s)
	var out []rune
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			out = append(out, r)
		case r == ' ' || r == '-' || r == '_':
			out = append(out, '-')
		}
	}
	result := string(out)
	for strings.Contains(result, "--") {
		result = strings.ReplaceAll(result, "--", "-")
	}
	result = strings.Trim(result, "-")
	if len(result) > 60 {
		result = result[:60]
	}
	return result
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
