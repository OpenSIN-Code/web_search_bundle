// Purpose: Unit tests for the clustering package.
// Docs: internal/clustering/cluster_test.doc.md
package clustering

import (
	"testing"
)

func TestCluster(t *testing.T) {
	items := []ClusterItem{
		{Source: "reddit", Title: "OpenClaw raises 10M", URL: "https://reddit.com/r/1", Score: 0.5},
		{Source: "hn", Title: "OpenClaw raises 10M funding", URL: "https://news.ycombinator.com/2", Score: 0.6},
		{Source: "reddit", Title: "Kanye drops new album", URL: "https://reddit.com/r/3", Score: 0.1},
	}
	c := NewClusterer()
	clusters := c.Cluster(items)
	if len(clusters) != 2 {
		t.Fatalf("expected 2 clusters, got %d", len(clusters))
	}
	if len(clusters[0].Items) != 2 {
		t.Fatalf("expected first cluster to have 2 items, got %d", len(clusters[0].Items))
	}
}
