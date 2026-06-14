// SPDX-License-Identifier: MIT
// Purpose: Citation discipline configuration for verification.
// Docs: internal/verify/claim.doc.md
package verify

import "time"

// CitationDiscipline defines verification rules.
type CitationDiscipline struct {
	MinSourcesPerClaim  int
	ConfidenceThreshold float64
	FlagContested       bool
}

// DefaultDiscipline returns a default discipline.
func DefaultDiscipline() *CitationDiscipline {
	return &CitationDiscipline{
		MinSourcesPerClaim:  2,
		ConfidenceThreshold: 0.7,
		FlagContested:       true,
	}
}

// Citation links a claim to a source.
type Citation struct {
	Source     string    `json:"source"`
	URL        string    `json:"url"`
	Snippet    string    `json:"snippet"`
	Engagement int       `json:"engagement"`
	Retrieved  time.Time `json:"retrieved"`
}

// Claim is a verified statement.
type Claim struct {
	Text           string     `json:"text"`
	Category       string     `json:"category"`
	Confidence     float64    `json:"confidence"`
	Status         string     `json:"status"`
	Sources        []Citation `json:"sources"`
	Contradictions []Citation `json:"contradictions,omitempty"`
}

// Status values.
const (
	StatusVerified     = "verified"
	StatusWeak         = "weak"
	StatusContested    = "contested"
	StatusContradicted = "contradicted"
	StatusUnverified   = "unverified"
)

// VerificationReport aggregates claims.
type VerificationReport struct {
	Topic           string    `json:"topic"`
	TotalClaims     int       `json:"total_claims"`
	Verified        int       `json:"verified"`
	Weak            int       `json:"weak"`
	Contested       int       `json:"contested"`
	Contradicted    int       `json:"contradicted"`
	Unverified      int       `json:"unverified"`
	AvgConfidence   float64   `json:"avg_confidence"`
	Claims          []Claim   `json:"claims"`
	StrongClaims    []Claim   `json:"strong_claims,omitempty"`
	WeakClaims      []Claim   `json:"weak_claims,omitempty"`
	ContestedClaims []Claim   `json:"contested_claims,omitempty"`
	GeneratedAt     time.Time `json:"generated_at"`
}
