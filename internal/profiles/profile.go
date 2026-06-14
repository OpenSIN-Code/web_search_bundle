// SPDX-License-Identifier: MIT
// Purpose: Research profile definitions for multi-agent missions.
// Docs: internal/profiles/profile.doc.md
package profiles

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Profile is a research mission blueprint.
type Profile struct {
	Name         string             `yaml:"name" json:"name"`
	Description  string             `yaml:"description" json:"description"`
	Version      string             `yaml:"version" json:"version"`
	Agents       AgentConfig        `yaml:"agents" json:"agents"`
	Sources      SourceConfig       `yaml:"sources" json:"sources"`
	Output       OutputConfig       `yaml:"output" json:"output"`
	Verification VerificationConfig `yaml:"verification" json:"verification"`
	Tags         []string           `yaml:"tags,omitempty" json:"tags,omitempty"`
}

// AgentConfig configures explore and librarian agents.
type AgentConfig struct {
	Explore   ExploreAgentConfig   `yaml:"explore" json:"explore"`
	Librarian LibrarianAgentConfig `yaml:"librarian" json:"librarian"`
}

// ExploreAgentConfig configures parallel research agents.
type ExploreAgentConfig struct {
	Count              int                `yaml:"count" json:"count"`
	Parallel           bool               `yaml:"parallel" json:"parallel"`
	FocusDistribution  map[string]float64 `yaml:"focus_distribution,omitempty" json:"focus_distribution,omitempty"`
	Timeout            string             `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	MaxResultsPerAgent int                `yaml:"max_results_per_agent,omitempty" json:"max_results_per_agent,omitempty"`
}

// LibrarianAgentConfig configures synthesis agents.
type LibrarianAgentConfig struct {
	Count int      `yaml:"count" json:"count"`
	Tasks []string `yaml:"tasks" json:"tasks"`
}

// SourceConfig defines allowed sources.
type SourceConfig struct {
	Required []string `yaml:"required" json:"required"`
	Optional []string `yaml:"optional,omitempty" json:"optional,omitempty"`
	Excluded []string `yaml:"excluded,omitempty" json:"excluded,omitempty"`
}

// OutputConfig defines mission output.
type OutputConfig struct {
	Primary   string   `yaml:"primary" json:"primary"`
	Secondary []string `yaml:"secondary,omitempty" json:"secondary,omitempty"`
	MaxClaims int      `yaml:"max_claims,omitempty" json:"max_claims,omitempty"`
}

// VerificationConfig defines citation rules.
type VerificationConfig struct {
	MinSourcesPerClaim  int     `yaml:"min_sources_per_claim" json:"min_sources_per_claim"`
	ConfidenceThreshold float64 `yaml:"confidence_threshold" json:"confidence_threshold"`
	FlagContested       bool    `yaml:"flag_contested" json:"flag_contested"`
}

// Registry holds loaded profiles.
type Registry struct {
	profiles map[string]*Profile
	dir      string
}

// NewRegistry creates a profile registry from a directory.
func NewRegistry(dir string) (*Registry, error) {
	r := &Registry{profiles: make(map[string]*Profile), dir: dir}
	if err := r.loadBuiltin(); err != nil {
		return nil, err
	}
	if dir != "" {
		if err := r.loadFromDir(dir); err != nil {
			fmt.Fprintf(os.Stderr, "⚠ custom profiles: %v\n", err)
		}
	}
	return r, nil
}

// Get returns a profile by name.
func (r *Registry) Get(name string) (*Profile, error) {
	p, ok := r.profiles[name]
	if !ok {
		return nil, fmt.Errorf("profile not found: %s", name)
	}
	return p, nil
}

// List returns all available profile names.
func (r *Registry) List() []string {
	var names []string
	for n := range r.profiles {
		names = append(names, n)
	}
	return names
}

// Add registers a profile.
func (r *Registry) Add(p *Profile) error {
	if p.Name == "" {
		return fmt.Errorf("profile name required")
	}
	r.profiles[p.Name] = p
	return nil
}

func (r *Registry) loadBuiltin() error {
	for _, p := range builtinProfiles() {
		if err := r.Add(p); err != nil {
			return err
		}
	}
	return nil
}

func (r *Registry) loadFromDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := filepath.Ext(e.Name())
		if ext != ".yaml" && ext != ".yml" {
			continue
		}
		p, err := LoadFromFile(filepath.Join(dir, e.Name()))
		if err != nil {
			fmt.Fprintf(os.Stderr, "⚠ skip %s: %v\n", e.Name(), err)
			continue
		}
		if err := r.Add(p); err != nil {
			fmt.Fprintf(os.Stderr, "⚠ skip %s: %v\n", e.Name(), err)
		}
	}
	return nil
}

// LoadFromFile loads a profile from YAML.
func LoadFromFile(path string) (*Profile, error) {
	data, err := os.ReadFile(path) // #nosec G304 — caller chooses profile file
	if err != nil {
		return nil, err
	}
	return LoadFromBytes(data)
}

// LoadFromBytes parses a profile from YAML.
func LoadFromBytes(data []byte) (*Profile, error) {
	var p Profile
	if err := yaml.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("parse profile: %w", err)
	}
	if p.Name == "" {
		return nil, fmt.Errorf("profile missing name")
	}
	applyDefaults(&p)
	return &p, nil
}

func applyDefaults(p *Profile) {
	if p.Agents.Explore.Count == 0 {
		p.Agents.Explore.Count = 5
	}
	if p.Agents.Librarian.Count == 0 {
		p.Agents.Librarian.Count = 2
	}
	if p.Agents.Explore.Timeout == "" {
		p.Agents.Explore.Timeout = "30s"
	}
	if p.Agents.Explore.MaxResultsPerAgent == 0 {
		p.Agents.Explore.MaxResultsPerAgent = 10
	}
	if p.Verification.MinSourcesPerClaim == 0 {
		p.Verification.MinSourcesPerClaim = 2
	}
	if p.Verification.ConfidenceThreshold == 0 {
		p.Verification.ConfidenceThreshold = 0.7
	}
	if p.Output.Primary == "" {
		p.Output.Primary = "html-brief"
	}
}

func builtinProfiles() []*Profile {
	return []*Profile{
		competitiveAnalysis(),
		personDossier(),
		marketLandscape(),
		crisisMonitoring(),
		productLaunch(),
		technicalDeepDive(),
	}
}

func competitiveAnalysis() *Profile {
	return &Profile{
		Name:        "competitive-analysis",
		Description: "Deep competitive analysis between products",
		Version:     "1.0",
		Agents: AgentConfig{
			Explore: ExploreAgentConfig{
				Count: 5, Parallel: true,
				FocusDistribution: map[string]float64{"technical": 0.25, "community": 0.25, "market": 0.20, "creator": 0.15, "users": 0.15},
				Timeout:           "45s", MaxResultsPerAgent: 15,
			},
			Librarian: LibrarianAgentConfig{Count: 2, Tasks: []string{"synthesize", "verify"}},
		},
		Sources:      SourceConfig{Required: []string{"github", "reddit", "hackernews", "polymarket"}, Optional: []string{"youtube", "x", "brave"}},
		Output:       OutputConfig{Primary: "html-brief", MaxClaims: 50},
		Verification: VerificationConfig{MinSourcesPerClaim: 2, ConfidenceThreshold: 0.7, FlagContested: true},
	}
}

func personDossier() *Profile {
	return &Profile{
		Name:        "person-dossier",
		Description: "Meeting prep research on a person",
		Version:     "1.0",
		Agents: AgentConfig{
			Explore:   ExploreAgentConfig{Count: 4, Parallel: true, FocusDistribution: map[string]float64{"x": 0.3, "github": 0.3, "reddit": 0.2, "youtube": 0.2}, Timeout: "30s", MaxResultsPerAgent: 10},
			Librarian: LibrarianAgentConfig{Count: 2, Tasks: []string{"synthesize", "verify"}},
		},
		Sources:      SourceConfig{Required: []string{"github", "reddit", "brave"}, Optional: []string{"youtube", "x"}},
		Output:       OutputConfig{Primary: "html-brief", MaxClaims: 30},
		Verification: VerificationConfig{MinSourcesPerClaim: 2, ConfidenceThreshold: 0.6, FlagContested: true},
	}
}

func marketLandscape() *Profile {
	return &Profile{
		Name:        "market-landscape",
		Description: "Industry overview",
		Version:     "1.0",
		Agents: AgentConfig{
			Explore:   ExploreAgentConfig{Count: 6, Parallel: true, FocusDistribution: map[string]float64{"market": 0.3, "technical": 0.2, "community": 0.2, "creator": 0.15, "users": 0.15}, Timeout: "60s", MaxResultsPerAgent: 15},
			Librarian: LibrarianAgentConfig{Count: 3, Tasks: []string{"synthesize", "verify", "compare"}},
		},
		Sources:      SourceConfig{Required: []string{"brave", "reddit", "hackernews", "polymarket"}, Optional: []string{"youtube", "github", "x"}},
		Output:       OutputConfig{Primary: "html-brief", MaxClaims: 100},
		Verification: VerificationConfig{MinSourcesPerClaim: 2, ConfidenceThreshold: 0.7, FlagContested: true},
	}
}

func crisisMonitoring() *Profile {
	return &Profile{
		Name:        "crisis-monitoring",
		Description: "PR crisis monitoring",
		Version:     "1.0",
		Agents: AgentConfig{
			Explore:   ExploreAgentConfig{Count: 3, Parallel: true, FocusDistribution: map[string]float64{"x": 0.4, "reddit": 0.4, "market": 0.2}, Timeout: "20s", MaxResultsPerAgent: 10},
			Librarian: LibrarianAgentConfig{Count: 1, Tasks: []string{"synthesize"}},
		},
		Sources:      SourceConfig{Required: []string{"reddit", "hackernews", "brave"}, Optional: []string{"x", "youtube"}},
		Output:       OutputConfig{Primary: "html-brief", MaxClaims: 20},
		Verification: VerificationConfig{MinSourcesPerClaim: 1, ConfidenceThreshold: 0.5, FlagContested: true},
	}
}

func productLaunch() *Profile {
	return &Profile{
		Name:        "product-launch",
		Description: "New release tracking",
		Version:     "1.0",
		Agents: AgentConfig{
			Explore:   ExploreAgentConfig{Count: 5, Parallel: true, FocusDistribution: map[string]float64{"github": 0.3, "hackernews": 0.3, "x": 0.2, "reddit": 0.2}, Timeout: "45s", MaxResultsPerAgent: 12},
			Librarian: LibrarianAgentConfig{Count: 2, Tasks: []string{"synthesize", "verify"}},
		},
		Sources:      SourceConfig{Required: []string{"github", "hackernews", "reddit", "brave"}, Optional: []string{"youtube", "x", "polymarket"}},
		Output:       OutputConfig{Primary: "html-brief", MaxClaims: 50},
		Verification: VerificationConfig{MinSourcesPerClaim: 2, ConfidenceThreshold: 0.7, FlagContested: true},
	}
}

func technicalDeepDive() *Profile {
	return &Profile{
		Name:        "technical-deep-dive",
		Description: "Architecture research",
		Version:     "1.0",
		Agents: AgentConfig{
			Explore:   ExploreAgentConfig{Count: 4, Parallel: true, FocusDistribution: map[string]float64{"github": 0.4, "hackernews": 0.3, "blogs": 0.3}, Timeout: "45s", MaxResultsPerAgent: 15},
			Librarian: LibrarianAgentConfig{Count: 2, Tasks: []string{"synthesize", "verify"}},
		},
		Sources:      SourceConfig{Required: []string{"github", "hackernews", "brave"}, Optional: []string{"youtube", "reddit"}},
		Output:       OutputConfig{Primary: "html-brief", MaxClaims: 50},
		Verification: VerificationConfig{MinSourcesPerClaim: 2, ConfidenceThreshold: 0.8, FlagContested: true},
	}
}
