// SPDX-License-Identifier: MIT
// Purpose: Hermetic unit tests for citation discipline and claim types.
// Docs: internal/verify/claim_test.doc.md
package verify

import "testing"

func TestDefaultDiscipline(t *testing.T) {
	d := DefaultDiscipline()
	if d == nil {
		t.Fatal("DefaultDiscipline returned nil")
	}
	if d.MinSourcesPerClaim != 2 {
		t.Errorf("MinSourcesPerClaim = %d, want 2", d.MinSourcesPerClaim)
	}
	if d.ConfidenceThreshold != 0.7 {
		t.Errorf("ConfidenceThreshold = %v, want 0.7", d.ConfidenceThreshold)
	}
	if !d.FlagContested {
		t.Error("FlagContested = false, want true")
	}
}

func TestStatusConstants(t *testing.T) {
	want := map[string]string{
		"verified":     StatusVerified,
		"weak":         StatusWeak,
		"contested":    StatusContested,
		"contradicted": StatusContradicted,
		"unverified":   StatusUnverified,
	}
	for name, got := range want {
		if got == "" {
			t.Errorf("status constant %s is empty", name)
		}
	}
}
