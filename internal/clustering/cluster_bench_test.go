// SPDX-License-Identifier: MIT
// Purpose: Benchmark cluster grouping and Jaccard similarity.
// Docs: cluster.doc.md
package clustering

import (
	"fmt"
	"testing"
)

func makeItems(n int) []ClusterItem {
	items := make([]ClusterItem, n)
	for i := 0; i < n; i++ {
		items[i] = ClusterItem{
			Source: fmt.Sprintf("source-%d", i%5),
			Title:  fmt.Sprintf("Breaking news about Go version %d released today", i%10),
			URL:    fmt.Sprintf("https://example.com/%d", i),
			Score:  float64(i),
		}
	}
	return items
}

func BenchmarkClustererCluster100(b *testing.B) {
	items := makeItems(100)
	c := NewClusterer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = c.Cluster(items)
	}
}

func BenchmarkClustererCluster1000(b *testing.B) {
	items := makeItems(1000)
	c := NewClusterer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = c.Cluster(items)
	}
}

func BenchmarkSimilarity(b *testing.B) {
	c := NewClusterer()
	a := "Go 1.25 release brings new features and performance improvements"
	bb := "Go 1.25 release includes performance improvements and new features"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = c.similarity(a, bb)
	}
}
