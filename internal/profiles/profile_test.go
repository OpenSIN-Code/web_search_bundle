// Purpose: Hermetic unit tests for research profile loading and registry.
// Docs: profile_test.doc.md
package profiles

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFromBytes(t *testing.T) {
	yaml := `
name: test-profile
description: A test profile
version: "1.0"
agents:
  explore:
    count: 3
    parallel: true
  librarian:
    count: 1
    tasks:
      - synthesize
sources:
  required:
    - github
output:
  primary: json
verification:
  min_sources_per_claim: 3
  confidence_threshold: 0.8
`
	p, err := LoadFromBytes([]byte(yaml))
	if err != nil {
		t.Fatalf("LoadFromBytes error: %v", err)
	}
	if p.Name != "test-profile" {
		t.Errorf("name = %q, want test-profile", p.Name)
	}
	if p.Agents.Explore.Count != 3 {
		t.Errorf("explore count = %d, want 3", p.Agents.Explore.Count)
	}
	if p.Agents.Librarian.Count != 1 {
		t.Errorf("librarian count = %d, want 1", p.Agents.Librarian.Count)
	}
	if p.Verification.MinSourcesPerClaim != 3 {
		t.Errorf("min sources = %d, want 3", p.Verification.MinSourcesPerClaim)
	}
	if p.Output.Primary != "json" {
		t.Errorf("output primary = %q, want json", p.Output.Primary)
	}
}

func TestLoadFromBytesMissingName(t *testing.T) {
	_, err := LoadFromBytes([]byte("description: no name"))
	if err == nil {
		t.Fatal("expected error for missing name")
	}
}

func TestLoadFromBytesDefaults(t *testing.T) {
	yaml := `name: defaults`
	p, err := LoadFromBytes([]byte(yaml))
	if err != nil {
		t.Fatalf("LoadFromBytes error: %v", err)
	}
	if p.Agents.Explore.Count != 5 {
		t.Errorf("default explore count = %d, want 5", p.Agents.Explore.Count)
	}
	if p.Agents.Librarian.Count != 2 {
		t.Errorf("default librarian count = %d, want 2", p.Agents.Librarian.Count)
	}
	if p.Agents.Explore.Timeout != "30s" {
		t.Errorf("default timeout = %q, want 30s", p.Agents.Explore.Timeout)
	}
	if p.Verification.ConfidenceThreshold != 0.7 {
		t.Errorf("default threshold = %v, want 0.7", p.Verification.ConfidenceThreshold)
	}
}

func TestRegistryBuiltin(t *testing.T) {
	r, err := NewRegistry("")
	if err != nil {
		t.Fatalf("NewRegistry error: %v", err)
	}
	for _, name := range []string{"competitive-analysis", "person-dossier", "market-landscape", "crisis-monitoring", "product-launch", "technical-deep-dive"} {
		p, err := r.Get(name)
		if err != nil {
			t.Errorf("builtin profile %s not found: %v", name, err)
			continue
		}
		if p.Name != name {
			t.Errorf("profile name = %q, want %q", p.Name, name)
		}
	}
	names := r.List()
	if len(names) != 6 {
		t.Errorf("expected 6 builtin profiles, got %d", len(names))
	}
}

func TestRegistryCustom(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "custom.yaml")
	if err := os.WriteFile(path, []byte("name: custom\nversion: \"1.0\"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	r, err := NewRegistry(dir)
	if err != nil {
		t.Fatalf("NewRegistry error: %v", err)
	}
	p, err := r.Get("custom")
	if err != nil {
		t.Fatalf("custom profile not found: %v", err)
	}
	if p.Name != "custom" {
		t.Errorf("name = %q, want custom", p.Name)
	}
}

func TestRegistryAdd(t *testing.T) {
	r, err := NewRegistry("")
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Add(&Profile{Name: "added"}); err != nil {
		t.Fatalf("Add error: %v", err)
	}
	if _, err := r.Get("added"); err != nil {
		t.Errorf("added profile not found: %v", err)
	}
	if err := r.Add(&Profile{Name: ""}); err == nil {
		t.Error("expected error for empty profile name")
	}
}

func TestRegistryGetMissing(t *testing.T) {
	r, err := NewRegistry("")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Get("missing"); err == nil {
		t.Error("expected error for missing profile")
	}
}

func TestLoadFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "profile.yaml")
	if err := os.WriteFile(path, []byte("name: from-file\n"), 0644); err != nil {
		t.Fatal(err)
	}
	p, err := LoadFromFile(path)
	if err != nil {
		t.Fatalf("LoadFromFile error: %v", err)
	}
	if p.Name != "from-file" {
		t.Errorf("name = %q, want from-file", p.Name)
	}
}
