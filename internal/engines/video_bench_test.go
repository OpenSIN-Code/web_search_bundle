// SPDX-License-Identifier: MIT
// Purpose: Benchmark video subtitle parsing and HTML stripping.
// Docs: video.doc.md
package engines

import "testing"

func sampleVTT() string {
	return `WEBVTT

00:00:00.000 --> 00:00:05.000
Hello world

00:00:05.000 --> 00:00:10.000
This is a test subtitle

00:00:10.000 --> 00:00:15.000
Hello world

00:00:15.000 --> 00:00:20.000
<b>Bold text</b> and <i>italic text</i>
`
}

func BenchmarkParseVTT(b *testing.B) {
	vtt := sampleVTT()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parseVTT(vtt)
	}
}

func BenchmarkDedupeLines(b *testing.B) {
	lines := []string{
		"hello world", "this is a test", "hello world", "another line",
		"this is a test", "final line", "hello world",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = dedupeLines(lines)
	}
}

func BenchmarkStripHTML(b *testing.B) {
	s := "<p>This is <b>bold</b> and <a href=\"https://example.com\">a link</a>.</p>"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = stripHTML(s)
	}
}

func BenchmarkDetectVideoSource(b *testing.B) {
	urls := []string{
		"https://www.youtube.com/watch?v=dQw4w9WgXcQ",
		"https://youtu.be/dQw4w9WgXcQ",
		"https://vimeo.com/123456789",
		"https://example.com/video.mp4",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = detectVideoSource(urls[i%len(urls)])
	}
}
