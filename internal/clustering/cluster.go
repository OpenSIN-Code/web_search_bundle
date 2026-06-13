// Purpose: Merge duplicate or overlapping stories across sources.
// Docs: internal/clustering/cluster.doc.md
package clustering

import (
	"strings"
	"unicode"
)

// Cluster groups related items.
type Cluster struct {
	ID         string        `json:"id"`
	Title      string        `json:"title"`
	Sources    []string      `json:"sources"`
	Items      []ClusterItem `json:"items"`
	Engagement int           `json:"engagement"`
}

// ClusterItem is a single result within a cluster.
type ClusterItem struct {
	Source string  `json:"source"`
	Title  string  `json:"title"`
	URL    string  `json:"url"`
	Score  float64 `json:"score"`
}

// Clusterer groups items by title similarity.
type Clusterer struct {
	threshold float64
}

// NewClusterer creates a clusterer with default Jaccard threshold.
func NewClusterer() *Clusterer {
	return &Clusterer{threshold: 0.6}
}

// Cluster groups the given items.
func (c *Clusterer) Cluster(items []ClusterItem) []Cluster {
	var clusters []Cluster
	used := make(map[int]bool)

	for i, item := range items {
		if used[i] {
			continue
		}

		cluster := Cluster{
			ID:      normalize(item.Title),
			Title:   item.Title,
			Items:   []ClusterItem{item},
			Sources: []string{item.Source},
		}
		used[i] = true

		for j := i + 1; j < len(items); j++ {
			if used[j] {
				continue
			}
			if c.similarity(item.Title, items[j].Title) >= c.threshold {
				cluster.Items = append(cluster.Items, items[j])
				if !contains(cluster.Sources, items[j].Source) {
					cluster.Sources = append(cluster.Sources, items[j].Source)
				}
				used[j] = true
			}
		}

		clusters = append(clusters, cluster)
	}

	return clusters
}

func (c *Clusterer) similarity(a, b string) float64 {
	wordsA := wordSet(normalize(a))
	wordsB := wordSet(normalize(b))

	if len(wordsA) == 0 || len(wordsB) == 0 {
		return 0
	}

	intersection := 0
	for w := range wordsA {
		if wordsB[w] {
			intersection++
		}
	}

	union := len(wordsA) + len(wordsB) - intersection
	return float64(intersection) / float64(union)
}

func normalize(s string) string {
	s = strings.ToLower(s)
	var result []rune
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || r == ' ' {
			result = append(result, r)
		}
	}
	return string(result)
}

func wordSet(s string) map[string]bool {
	words := make(map[string]bool)
	for _, w := range strings.Fields(s) {
		if len(w) > 2 {
			words[w] = true
		}
	}
	return words
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
