// Purpose: Run verification pipeline on search results.
// Docs: internal/verify/engine.doc.md
package verify

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/OpenSIN-Code/web_search_bundle/internal/engines"
)

// Engine runs the verification pipeline.
type Engine struct {
	discipline *CitationDiscipline
	extractor  *ClaimExtractor
}

// NewEngine creates a verification engine.
func NewEngine(d *CitationDiscipline) *Engine {
	if d == nil {
		d = DefaultDiscipline()
	}
	return &Engine{discipline: d, extractor: NewClaimExtractor(d)}
}

// Verify runs verification on search results.
func (e *Engine) Verify(topic string, results []engines.Result) *VerificationReport {
	claims := e.extractor.Extract(results)
	if e.discipline.FlagContested {
		claims = e.detectContradictions(claims)
	}
	return e.buildReport(topic, claims)
}

func (e *Engine) detectContradictions(claims []Claim) []Claim {
	for i := range claims {
		for j := i + 1; j < len(claims); j++ {
			if claims[i].Category == "statistic" && claims[j].Category == "statistic" {
				if e.opposes(claims[i].Text, claims[j].Text) {
					if claims[i].Confidence < claims[j].Confidence {
						claims[i].Status = StatusContradicted
						claims[i].Contradictions = append(claims[i].Contradictions, claims[j].Sources...)
					} else {
						claims[j].Status = StatusContradicted
						claims[j].Contradictions = append(claims[j].Contradictions, claims[i].Sources...)
					}
				}
			}
		}
	}
	return claims
}

func (e *Engine) opposes(a, b string) bool {
	return false
}

func (e *Engine) buildReport(topic string, claims []Claim) *VerificationReport {
	report := &VerificationReport{
		Topic:       topic,
		TotalClaims: len(claims),
		Claims:      claims,
		GeneratedAt: time.Now(),
	}

	var totalConf float64
	for i := range claims {
		c := &claims[i]
		// Compute confidence based on source count.
		c.Confidence = float64(len(c.Sources)) / float64(e.discipline.MinSourcesPerClaim)
		if c.Confidence > 1.0 {
			c.Confidence = 1.0
		}
		totalConf += c.Confidence

		if c.Status == "" {
			if c.Confidence >= e.discipline.ConfidenceThreshold {
				c.Status = StatusVerified
			} else if len(c.Contradictions) > 0 {
				c.Status = StatusContradicted
			} else if len(c.Sources) > 0 {
				c.Status = StatusWeak
			} else {
				c.Status = StatusUnverified
			}
		}

		switch c.Status {
		case StatusVerified:
			report.Verified++
			report.StrongClaims = append(report.StrongClaims, *c)
		case StatusWeak:
			report.Weak++
			report.WeakClaims = append(report.WeakClaims, *c)
		case StatusContested:
			report.Contested++
			report.ContestedClaims = append(report.ContestedClaims, *c)
		case StatusContradicted:
			report.Contradicted++
		case StatusUnverified:
			report.Unverified++
		}
	}

	if len(claims) > 0 {
		report.AvgConfidence = totalConf / float64(len(claims))
	}
	sort.Slice(report.StrongClaims, func(i, j int) bool { return report.StrongClaims[i].Confidence > report.StrongClaims[j].Confidence })
	sort.Slice(report.ContestedClaims, func(i, j int) bool {
		return report.ContestedClaims[i].Confidence > report.ContestedClaims[j].Confidence
	})
	return report
}

// FormatText renders a human-readable report.
func (r *VerificationReport) FormatText() string {
	var sb strings.Builder
	sb.WriteString("🔍 VERIFICATION REPORT\n======================\n\n")
	sb.WriteString("Topic: " + r.Topic + "\n")
	sb.WriteString("Generated: " + r.GeneratedAt.Format("2006-01-02 15:04:05") + "\n\n")
	sb.WriteString(fmt.Sprintf("📊 Summary:\n  Total: %d\n  ✅ Verified: %d\n  ⚠️ Weak: %d\n  ⚖️ Contested: %d\n  ❌ Contradicted: %d\n  Avg confidence: %.2f\n\n",
		r.TotalClaims, r.Verified, r.Weak, r.Contested, r.Contradicted, r.AvgConfidence))

	if len(r.StrongClaims) > 0 {
		sb.WriteString("✅ STRONG CLAIMS:\n")
		for i, c := range r.StrongClaims {
			if i >= 10 {
				break
			}
			sb.WriteString(fmt.Sprintf("  [%.2f] %s\n", c.Confidence, truncate(c.Text, 80)))
		}
		sb.WriteString("\n")
	}
	if len(r.ContestedClaims) > 0 {
		sb.WriteString("⚖️ CONTESTED CLAIMS:\n")
		for i, c := range r.ContestedClaims {
			if i >= 5 {
				break
			}
			sb.WriteString(fmt.Sprintf("  [%.2f] %s\n", c.Confidence, truncate(c.Text, 80)))
		}
	}
	return sb.String()
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
