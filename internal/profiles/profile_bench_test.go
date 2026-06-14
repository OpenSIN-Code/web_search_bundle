// SPDX-License-Identifier: MIT
// Purpose: Benchmark profile YAML parsing, registry creation, and lookups.
// Docs: profile.doc.md
package profiles

import (
	"fmt"
	"testing"
)

var sampleProfile = []byte(`
name: benchmark-profile
description: A profile used for benchmark testing
version: "1.0"
agents:
  explore:
    count: 5
    parallel: true
    focus_distribution:
      github: 0.4
      hackernews: 0.3
      blogs: 0.3
    timeout: 45s
    max_results_per_agent: 15
  librarian:
    count: 2
    tasks:
      - synthesize
      - verify
sources:
  required:
    - github
    - hackernews
    - brave
  optional:
    - youtube
    - reddit
output:
  primary: html-brief
  secondary:
    - json
    - markdown
  max_claims: 50
verification:
  min_sources_per_claim: 2
  confidence_threshold: 0.8
  flag_contested: true
tags:
  - benchmark
  - test
`)

func BenchmarkLoadFromBytes(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := LoadFromBytes(sampleProfile)
		if err != nil {
			b.Fatalf("load from bytes: %v", err)
		}
	}
}

func BenchmarkNewRegistryBuiltin(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := NewRegistry("")
		if err != nil {
			b.Fatalf("new registry: %v", err)
		}
	}
}

func BenchmarkRegistryGet(b *testing.B) {
	r, err := NewRegistry("")
	if err != nil {
		b.Fatalf("setup registry: %v", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := r.Get("competitive-analysis")
		if err != nil {
			b.Fatalf("get profile: %v", err)
		}
	}
}

func BenchmarkRegistryList(b *testing.B) {
	r, err := NewRegistry("")
	if err != nil {
		b.Fatalf("setup registry: %v", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.List()
	}
}

func BenchmarkRegistryAdd(b *testing.B) {
	r, err := NewRegistry("")
	if err != nil {
		b.Fatalf("setup registry: %v", err)
	}
	profiles := make([]*Profile, 10000)
	for i := range profiles {
		profiles[i] = &Profile{Name: fmt.Sprintf("added-%d", i)}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := r.Add(profiles[i%len(profiles)]); err != nil {
			b.Fatalf("add profile: %v", err)
		}
	}
}
