// Purpose: Benchmark humor/virality scoring and best-takes selection.
// Docs: internal/judge/humor.doc.md
package judge

import "testing"

func makeText(n int) []string {
	texts := make([]string, n)
	base := []string{
		"Go 1.25 release is literally the best lol",
		"This is so peak performance",
		"Another release announcement with no jokes",
		"Mic drop: Go is now faster",
		"Just a regular update with nothing special",
	}
	for i := 0; i < n; i++ {
		texts[i] = base[i%len(base)]
	}
	return texts
}

func BenchmarkHumorJudgeScoreResult(b *testing.B) {
	j := NewHumorJudge()
	text := "Go 1.25 is literally the best release lol 😂"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = j.ScoreResult(text, 512, 10000)
	}
}

func BenchmarkHumorJudgeBestTakes10(b *testing.B) {
	j := NewHumorJudge()
	texts := makeText(10)
	items := make([]struct {
		Text    string
		Upvotes int
	}, len(texts))
	for i, t := range texts {
		items[i] = struct {
			Text    string
			Upvotes int
		}{Text: t, Upvotes: (i + 1) * 100}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = j.BestTakes(items, 3)
	}
}

func BenchmarkHumorJudgeBestTakes100(b *testing.B) {
	j := NewHumorJudge()
	texts := makeText(100)
	items := make([]struct {
		Text    string
		Upvotes int
	}, len(texts))
	for i, t := range texts {
		items[i] = struct {
			Text    string
			Upvotes int
		}{Text: t, Upvotes: (i + 1) * 10}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = j.BestTakes(items, 5)
	}
}
