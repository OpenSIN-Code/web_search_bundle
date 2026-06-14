// SPDX-License-Identifier: MIT
// Purpose: Unit tests for experiment loop Run and Evaluate edge cases.
// Docs: loop_run_test.doc.md

package experiment

import (
	"context"
	"runtime"
	"testing"
	"time"
)

func TestNewLoop_ValidRegex(t *testing.T) {
	cfg := Config{MetricRegex: `value:\s*([0-9\.]+)`}
	loop, err := NewLoop(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if loop.re == nil {
		t.Error("expected compiled regex")
	}
}

func TestRun_SuccessExtractsMetric(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping shell test on windows")
	}
	cfg := Config{
		MetricRegex:    `metric:\s*([0-9\.]+)`,
		MetricName:     "metric",
		HigherIsBetter: true,
		Budget:         5 * time.Second,
		RunCmd:         []string{"sh", "-c", "echo 'metric: 3.14'"},
	}
	loop, err := NewLoop(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	res, err := loop.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Error != nil {
		t.Fatalf("unexpected run error: %v", res.Error)
	}
	if res.TimedOut {
		t.Error("expected no timeout")
	}
	if res.MetricValue != 3.14 {
		t.Errorf("expected metric 3.14, got %v", res.MetricValue)
	}
}

func TestRun_CommandError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping shell test on windows")
	}
	cfg := Config{
		MetricRegex: `metric:\s*([0-9\.]+)`,
		MetricName:  "metric",
		Budget:      5 * time.Second,
		RunCmd:      []string{"sh", "-c", "exit 1"},
	}
	loop, err := NewLoop(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	res, err := loop.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Error == nil {
		t.Fatal("expected run error")
	}
}

func TestRun_Timeout(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping shell test on windows")
	}
	cfg := Config{
		MetricRegex: `metric:\s*([0-9\.]+)`,
		MetricName:  "metric",
		Budget:      100 * time.Millisecond,
		RunCmd:      []string{"sh", "-c", "sleep 2"},
	}
	loop, err := NewLoop(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	res, err := loop.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.TimedOut {
		t.Error("expected timeout")
	}
}

func TestRun_MetricParseError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping shell test on windows")
	}
	cfg := Config{
		MetricRegex: `metric:\s*([0-9\.]+)`,
		MetricName:  "metric",
		Budget:      5 * time.Second,
		RunCmd:      []string{"sh", "-c", "echo 'metric: abc'"},
	}
	loop, err := NewLoop(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	res, err := loop.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Error != nil {
		t.Fatalf("unexpected run error: %v", res.Error)
	}
	if res.MetricValue != 0 {
		t.Errorf("expected metric value 0 for parse error, got %v", res.MetricValue)
	}
}

func TestEvaluate_Error(t *testing.T) {
	cfg := Config{MetricRegex: `.*`, MetricName: "x"}
	loop, _ := NewLoop(cfg)
	kept, reason := loop.Evaluate(&Result{Error: context.Canceled})
	if kept {
		t.Error("expected discarded")
	}
	if reason != "discarded: error or timeout" {
		t.Errorf("unexpected reason: %s", reason)
	}
}

func TestEvaluate_Timeout(t *testing.T) {
	cfg := Config{MetricRegex: `.*`, MetricName: "x"}
	loop, _ := NewLoop(cfg)
	kept, reason := loop.Evaluate(&Result{TimedOut: true})
	if kept {
		t.Error("expected discarded")
	}
	if reason != "discarded: error or timeout" {
		t.Errorf("unexpected reason: %s", reason)
	}
}

func TestEvaluate_NoMetric(t *testing.T) {
	cfg := Config{MetricRegex: `.*`, MetricName: "x"}
	loop, _ := NewLoop(cfg)
	kept, reason := loop.Evaluate(&Result{})
	if kept {
		t.Error("expected discarded")
	}
	if reason != "discarded: no metric extracted" {
		t.Errorf("unexpected reason: %s", reason)
	}
}

func TestEvaluate_ImprovedMessage(t *testing.T) {
	cfg := Config{MetricRegex: `.*`, MetricName: "x", HigherIsBetter: true}
	loop, _ := NewLoop(cfg)
	loop.Evaluate(&Result{MetricValue: 1.0})
	kept, reason := loop.Evaluate(&Result{MetricValue: 2.0})
	if !kept {
		t.Error("expected kept")
	}
	if reason != "accepted: new best x = 2.0000" {
		t.Errorf("unexpected reason: %s", reason)
	}
}

func TestEvaluate_DiscardedMessage(t *testing.T) {
	cfg := Config{MetricRegex: `.*`, MetricName: "x", HigherIsBetter: true}
	loop, _ := NewLoop(cfg)
	loop.Evaluate(&Result{MetricValue: 5.0})
	kept, reason := loop.Evaluate(&Result{MetricValue: 3.0})
	if kept {
		t.Error("expected discarded")
	}
	if reason != "discarded: metric 3.0000 did not beat baseline 5.0000" {
		t.Errorf("unexpected reason: %s", reason)
	}
}
