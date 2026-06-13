// Purpose: Score results for relevance, virality, and humor.
// Docs: internal/judge/humor.doc.md
package judge

import (
	"regexp"
	"strings"
)

// HumorJudge scores content for virality and humor signals.
type HumorJudge struct {
	viralPatterns []*regexp.Regexp
	humorSignals  []string
}

// Score is a combined content score.
type Score struct {
	Relevance  float64 `json:"relevance"`
	Virality   float64 `json:"virality"`
	Humor      float64 `json:"humor"`
	Engagement int     `json:"engagement"`
}

// NewHumorJudge creates a judge with default patterns.
func NewHumorJudge() *HumorJudge {
	return &HumorJudge{
		viralPatterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)\b(lol|lmao|rofl|dead|💀|😂|🤣)\b`),
			regexp.MustCompile(`(?i)(this|that) is (so|actually|literally) `),
			regexp.MustCompile(`(?i)(mic drop|chef's kiss|peak)`),
		},
		humorSignals: []string{
			"sarcastic", "ironic", "satire", "joke", "meme",
			"one-liner", "hot take", "unpopular opinion",
		},
	}
}

// ScoreResult evaluates a single text result.
func (j *HumorJudge) ScoreResult(text string, upvotes, views int) Score {
	score := Score{Engagement: upvotes + views/100}

	if upvotes > 1000 {
		score.Virality = 0.9
	} else if upvotes > 500 {
		score.Virality = 0.7
	} else if upvotes > 100 {
		score.Virality = 0.5
	} else {
		score.Virality = float64(upvotes) / 200.0
	}

	lower := strings.ToLower(text)
	humorMatches := 0
	for _, pattern := range j.viralPatterns {
		if pattern.MatchString(lower) {
			humorMatches++
		}
	}
	for _, signal := range j.humorSignals {
		if strings.Contains(lower, signal) {
			humorMatches++
		}
	}

	score.Humor = float64(humorMatches) / 5.0
	if score.Humor > 1.0 {
		score.Humor = 1.0
	}
	score.Relevance = 0.5 // Placeholder for LLM-based relevance.

	return score
}

// BestTakes returns the most viral/humorous items.
func (j *HumorJudge) BestTakes(items []struct {
	Text    string
	Upvotes int
}, limit int) []string {
	type scored struct {
		text  string
		score float64
	}

	var scoredItems []scored
	for _, item := range items {
		s := j.ScoreResult(item.Text, item.Upvotes, 0)
		scoredItems = append(scoredItems, scored{item.Text, s.Virality*0.6 + s.Humor*0.4})
	}

	for i := 0; i < len(scoredItems); i++ {
		for k := i + 1; k < len(scoredItems); k++ {
			if scoredItems[k].score > scoredItems[i].score {
				scoredItems[i], scoredItems[k] = scoredItems[k], scoredItems[i]
			}
		}
	}

	var best []string
	for i := 0; i < limit && i < len(scoredItems); i++ {
		best = append(best, scoredItems[i].text)
	}
	return best
}
