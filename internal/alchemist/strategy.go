// SPDX-License-Identifier: MIT
// Purpose: Define strategies for Swarm workers.
// Docs: strategy.doc.md

package alchemist

import (
	"fmt"
)

// Strategy defines how a swarm worker approaches experiments
type Strategy struct {
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	MaxMutation   int     `json:"max_mutation"`   // max lines changed per experiment
	RiskAppetite  float64 `json:"risk_appetite"`  // 0.0 (conservative) → 1.0 (aggressive)
	Model         string  `json:"model"`          // LLM model to use for code generation
	Temperature   float64 `json:"temperature"`    // creativity knob
	PromptOverlay string  `json:"prompt_overlay"` // system prompt suffix
}

// BuiltinStrategies returns the 4 standard swarm strategies
func BuiltinStrategies() map[string]Strategy {
	return map[string]Strategy{
		"conservative": {
			Name:         "conservative",
			Description:  "Minimal changes, low risk. Single-function edits only.",
			MaxMutation:  20,
			RiskAppetite: 0.1,
			Model:        "anthropic/claude-haiku-4",
			Temperature:  0.2,
			PromptOverlay: `You are a conservative engineer. Make minimal, surgical changes.
- Modify at most 1 function per experiment.
- Prefer changing constants, thresholds, or small logic.
- Never restructure or refactor.
- If in doubt, make the smaller change.`,
		},
		"aggressive": {
			Name:         "aggressive",
			Description:  "Large refactors, higher risk. Structural changes welcome.",
			MaxMutation:  200,
			RiskAppetite: 0.8,
			Model:        "anthropic/claude-sonnet-4",
			Temperature:  0.6,
			PromptOverlay: `You are an aggressive optimizer. Make bold structural changes.
- Rewrite entire functions if needed.
- Try new algorithms, data structures, concurrency patterns.
- Don't fear breaking things — the verification gate protects you.
- Aim for 10%+ improvements, not 1%.`,
		},
		"creative": {
			Name:         "creative",
			Description:  "Unconventional approaches. Cross-pollinate from other domains.",
			MaxMutation:  150,
			RiskAppetite: 0.7,
			Model:        "anthropic/claude-opus-4",
			Temperature:  0.9,
			PromptOverlay: `You are a creative researcher. Borrow ideas from unrelated fields.
- Apply patterns from game dev, graphics, databases, biology.
- Try weird combinations: bloom filters + worker pools, SIMD + channels.
- Ask "what would Linus/John Carmock/Rob Pike do here?"
- Document your reasoning in commit messages.`,
		},
		"minimal": {
			Name:         "minimal",
			Description:  "Control group. Only 1-5 line changes to isolate variables.",
			MaxMutation:  5,
			RiskAppetite: 0.05,
			Model:        "anthropic/claude-haiku-4",
			Temperature:  0.1,
			PromptOverlay: `You make ONLY single-line or trivial changes.
- Change one constant, one flag, one operator.
- Never add or remove functions.
- Goal: isolate the effect of tiny variations.
- This is the scientific control group.`,
		},
		"literature-driven": {
			Name:         "literature-driven",
			Description:  "Hypotheses come from sin-websearch literature scan.",
			MaxMutation:  100,
			RiskAppetite: 0.5,
			Model:        "anthropic/claude-sonnet-4",
			Temperature:  0.4,
			PromptOverlay: `You implement techniques found in recent academic/industry literature.
- Each hypothesis is grounded in a specific paper, blog post, or talk.
- Cite the source in the commit message.
- Prefer techniques with proven results in similar domains.`,
		},
	}
}

// GetStrategy returns a strategy by name (falls back to conservative)
func GetStrategy(name string) Strategy {
	if s, ok := BuiltinStrategies()[name]; ok {
		return s
	}
	return BuiltinStrategies()["conservative"]
}

// StrategyNames returns all available strategy names
func StrategyNames() []string {
	names := make([]string, 0, len(BuiltinStrategies()))
	for name := range BuiltinStrategies() {
		names = append(names, name)
	}
	return names
}

// String implements Stringer
func (s Strategy) String() string {
	return fmt.Sprintf("%s (risk=%.2f, max_mutation=%d)", s.Name, s.RiskAppetite, s.MaxMutation)
}
