// Purpose: Smoke tests for alchemist strategies.
// Docs: strategy_test.doc.md

package alchemist

import (
	"testing"
)

func TestBuiltinStrategies(t *testing.T) {
	strategies := BuiltinStrategies()
	if len(strategies) == 0 {
		t.Fatal("expected at least one strategy")
	}

	expected := []string{"conservative", "aggressive", "creative", "minimal", "literature-driven"}
	for _, name := range expected {
		s, ok := strategies[name]
		if !ok {
			t.Errorf("missing strategy %q", name)
			continue
		}
		if s.Name != name {
			t.Errorf("strategy name mismatch: got %q, want %q", s.Name, name)
		}
		if s.RiskAppetite < 0 || s.RiskAppetite > 1 {
			t.Errorf("risk appetite out of range for %s: %f", name, s.RiskAppetite)
		}
		if s.MaxMutation <= 0 {
			t.Errorf("max mutation must be positive for %s: %d", name, s.MaxMutation)
		}
	}
}

func TestGetStrategy(t *testing.T) {
	s := GetStrategy("creative")
	if s.Name != "creative" {
		t.Errorf("expected creative strategy, got %s", s.Name)
	}

	// Unknown names fall back to conservative.
	s = GetStrategy("does-not-exist")
	if s.Name != "conservative" {
		t.Errorf("expected fallback conservative, got %s", s.Name)
	}
}

func TestStrategyNames(t *testing.T) {
	names := StrategyNames()
	if len(names) == 0 {
		t.Fatal("expected strategy names")
	}
	seen := make(map[string]bool)
	for _, n := range names {
		if seen[n] {
			t.Errorf("duplicate strategy name %q", n)
		}
		seen[n] = true
	}
}

func TestStrategyString(t *testing.T) {
	s := Strategy{Name: "test", RiskAppetite: 0.5, MaxMutation: 42}
	got := s.String()
	want := "test (risk=0.50, max_mutation=42)"
	if got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}
