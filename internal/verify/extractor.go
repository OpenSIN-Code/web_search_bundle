// Purpose: Extract claims from search results for verification.
// Docs: internal/verify/extractor.doc.md
package verify

import (
	"regexp"
	"strings"

	"github.com/OpenSIN-Code/web_search_bundle/internal/engines"
)

// ClaimExtractor extracts claims from search results.
type ClaimExtractor struct {
	discipline *CitationDiscipline
}

// NewClaimExtractor creates a claim extractor.
func NewClaimExtractor(d *CitationDiscipline) *ClaimExtractor {
	return &ClaimExtractor{discipline: d}
}

// Extract extracts simple claims from search results.
func (e *ClaimExtractor) Extract(results []engines.Result) []Claim {
	var claims []Claim
	seen := make(map[string]bool)

	for _, r := range results {
		text := r.Title + ". " + r.Snippet
		for _, sentence := range splitSentences(text) {
			if isFactual(sentence) && !seen[sentence] {
				seen[sentence] = true
				claim := Claim{
					Text:     sentence,
					Category: categorize(sentence),
					Sources: []Citation{{
						Source:     r.Source,
						URL:        r.URL,
						Snippet:    r.Snippet,
						Engagement: r.Engagement,
					}},
				}
				claims = append(claims, claim)
			}
		}
	}

	// Deduplicate by merging identical sentences.
	return mergeClaims(claims)
}

func splitSentences(text string) []string {
	re := regexp.MustCompile(`[.!?]+`)
	parts := re.Split(text, -1)
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if len(p) > 20 {
			out = append(out, p)
		}
	}
	return out
}

func isFactual(s string) bool {
	// Simple heuristic: contains numbers, dates, or known factual markers.
	if matched, _ := regexp.MatchString(`\d+`, s); matched {
		return true
	}
	markers := []string{"is", "are", "was", "were", "has", "have", "released", "announced", "launched"}
	lower := strings.ToLower(s)
	for _, m := range markers {
		if strings.Contains(lower, " "+m+" ") {
			return true
		}
	}
	return false
}

func categorize(s string) string {
	if matched, _ := regexp.MatchString(`\d+`, s); matched {
		return "statistic"
	}
	return "statement"
}

func mergeClaims(claims []Claim) []Claim {
	byText := make(map[string]*Claim)
	for i := range claims {
		c := &claims[i]
		if existing, ok := byText[c.Text]; ok {
			existing.Sources = append(existing.Sources, c.Sources...)
		} else {
			byText[c.Text] = c
		}
	}
	out := make([]Claim, 0, len(byText))
	for _, c := range byText {
		out = append(out, *c)
	}
	return out
}
