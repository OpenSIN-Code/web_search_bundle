// Purpose: Unit tests for experiment loop logic.
// Docs: loop_test.doc.md

package experiment

import (
	"testing"
)

func TestNewLoop_InvalidRegex(t *testing.T) {
	cfg := Config{
		MetricRegex: "[invalid regex",
	}
	_, err := NewLoop(cfg)
	if err == nil {
		t.Fatal("expected error with invalid regex, got nil")
	}
}

func TestEvaluate_HigherIsBetter(t *testing.T) {
	cfg := Config{
		MetricName:     "test_metric",
		MetricRegex:    `test:\s*([0-9\.]+)`,
		HigherIsBetter: true,
	}

	loop, err := NewLoop(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// First run establishes the baseline
	result1 := &Result{
		MetricValue: 10.5,
	}
	kept, reason := loop.Evaluate(result1)
	if !kept {
		t.Errorf("expected baseline to be accepted, got false")
	}
	if reason != "accepted: new baseline established" {
		t.Errorf("expected baseline reason, got %q", reason)
	}

	// Second run with higher value
	result2 := &Result{
		MetricValue: 12.0,
	}
	kept, _ = loop.Evaluate(result2)
	if !kept {
		t.Errorf("expected higher value to be accepted, got false")
	}

	// Third run with lower value
	result3 := &Result{
		MetricValue: 11.0,
	}
	kept, _ = loop.Evaluate(result3)
	if kept {
		t.Errorf("expected lower value to be discarded, got true")
	}
}

func TestEvaluate_LowerIsBetter(t *testing.T) {
	cfg := Config{
		MetricName:     "test_metric",
		MetricRegex:    `test:\s*([0-9\.]+)`,
		HigherIsBetter: false,
	}

	loop, err := NewLoop(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Establish baseline
	result1 := &Result{
		MetricValue: 10.5,
	}
	kept, _ := loop.Evaluate(result1)
	if !kept {
		t.Fatal("expected baseline to be accepted")
	}

	// Second run with lower value (better)
	result2 := &Result{
		MetricValue: 9.0,
	}
	kept, _ = loop.Evaluate(result2)
	if !kept {
		t.Errorf("expected lower value to be accepted, got false")
	}

	// Third run with higher value (worse)
	result3 := &Result{
		MetricValue: 9.5,
	}
	kept, _ = loop.Evaluate(result3)
	if kept {
		t.Errorf("expected higher value to be discarded, got true")
	}
}
