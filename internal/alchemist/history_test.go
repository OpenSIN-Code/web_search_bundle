// SPDX-License-Identifier: MIT
// Purpose: Smoke tests for the SQLite history store.
// Docs: history_test.doc.md

package alchemist

import (
	"testing"
	"time"
)

func TestHistoryInsertAndAll(t *testing.T) {
	hist, err := NewHistory(t.TempDir())
	if err != nil {
		t.Fatalf("NewHistory failed: %v", err)
	}
	defer hist.Close()

	r := ExperimentRecord{
		Timestamp:    time.Now(),
		Hypothesis:   "Test hypothesis",
		MetricBefore: 0.1,
		MetricAfter:  0.2,
		Delta:        0.1,
		Duration:     time.Second,
		Decision:     "committed",
		CommitSHA:    "abc123def456",
	}
	if err := hist.Insert(r); err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	all, err := hist.All()
	if err != nil {
		t.Fatalf("All failed: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("expected 1 record, got %d", len(all))
	}
	if all[0].Hypothesis != "Test hypothesis" {
		t.Errorf("hypothesis = %q, want %q", all[0].Hypothesis, "Test hypothesis")
	}
	if all[0].CommitSHA != "abc123def456" {
		t.Errorf("commit sha = %q, want %q", all[0].CommitSHA, "abc123def456")
	}
}

func TestHistorySummary(t *testing.T) {
	hist, err := NewHistory(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer hist.Close()

	records := []ExperimentRecord{
		{Timestamp: time.Now(), Hypothesis: "h1", MetricAfter: 0.5, Decision: "committed", Delta: 0.1},
		{Timestamp: time.Now(), Hypothesis: "h2", MetricAfter: 0.4, Decision: "discarded", Delta: -0.1},
		{Timestamp: time.Now(), Hypothesis: "h3", MetricAfter: 0.0, Decision: "error", Delta: 0},
	}
	for _, r := range records {
		if err := hist.Insert(r); err != nil {
			t.Fatal(err)
		}
	}

	summary, err := hist.Summary()
	if err != nil {
		t.Fatal(err)
	}
	if summary["total_experiments"] != 3 {
		t.Errorf("total_experiments = %v, want 3", summary["total_experiments"])
	}
	if summary["committed"] != 1 {
		t.Errorf("committed = %v, want 1", summary["committed"])
	}
	if summary["discarded"] != 1 {
		t.Errorf("discarded = %v, want 1", summary["discarded"])
	}
	if summary["errors"] != 1 {
		t.Errorf("errors = %v, want 1", summary["errors"])
	}
	if summary["best_delta"] != 0.1 {
		t.Errorf("best_delta = %v, want 0.1", summary["best_delta"])
	}
}

func TestHistoryRecent(t *testing.T) {
	hist, err := NewHistory(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer hist.Close()

	for i := 0; i < 5; i++ {
		if err := hist.Insert(ExperimentRecord{
			Timestamp:  time.Now(),
			Hypothesis: "h",
			Decision:   "committed",
		}); err != nil {
			t.Fatal(err)
		}
	}

	recent, err := hist.Recent(2)
	if err != nil {
		t.Fatal(err)
	}
	if len(recent) != 2 {
		t.Errorf("expected 2 recent records, got %d", len(recent))
	}
}

func TestHistorySafeRate(t *testing.T) {
	if safeRate(1, 0) != 0 {
		t.Error("safeRate(1,0) should be 0")
	}
	if safeRate(1, 2) != 0.5 {
		t.Error("safeRate(1,2) should be 0.5")
	}
}
