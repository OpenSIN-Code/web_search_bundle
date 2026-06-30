// SPDX-License-Identifier: MIT
// Purpose: Hermetic unit tests for the semantic cache layer.

package cache

import (
	"errors"
	"path/filepath"
	"testing"
	"time"
)

func newSemanticTestCache(t *testing.T, embedder Embedder) *SemanticCache {
	t.Helper()
	path := filepath.Join(t.TempDir(), "sem_cache.db")
	inner, err := New(path)
	if err != nil {
		t.Fatalf("New error: %v", err)
	}
	t.Cleanup(func() { _ = inner.Close() })
	return NewSemanticCache(inner, embedder)
}

// ---------------------------------------------------------------------------
// cosineSimilarity
// ---------------------------------------------------------------------------

func TestCosineSimilarityIdentical(t *testing.T) {
	v := []float32{1, 2, 3}
	sim := cosineSimilarity(v, v)
	if sim != 1.0 {
		t.Errorf("identical vectors: got %f, want 1.0", sim)
	}
}

func TestCosineSimilarityOrthogonal(t *testing.T) {
	a := []float32{1, 0}
	b := []float32{0, 1}
	sim := cosineSimilarity(a, b)
	if sim != 0 {
		t.Errorf("orthogonal vectors: got %f, want 0", sim)
	}
}

func TestCosineSimilaritySimilar(t *testing.T) {
	a := []float32{1, 2, 3}
	b := []float32{2, 4, 6} // same direction, scaled
	sim := cosineSimilarity(a, b)
	if sim < 0.999 {
		t.Errorf("parallel vectors: got %f, want ~1.0", sim)
	}
}

func TestCosineSimilarityZeroVector(t *testing.T) {
	sim := cosineSimilarity([]float32{0, 0}, []float32{1, 1})
	if sim != 0 {
		t.Errorf("zero vector: got %f, want 0", sim)
	}
}

func TestCosineSimilarityEmpty(t *testing.T) {
	sim := cosineSimilarity(nil, []float32{1})
	if sim != 0 {
		t.Errorf("empty vector: got %f, want 0", sim)
	}
}

// ---------------------------------------------------------------------------
// TFIDFEmbedder
// ---------------------------------------------------------------------------

func TestTFIDFEmbedder(t *testing.T) {
	emb := TFIDFEmbedder{}
	vec, err := emb.Embed("hello world hello")
	if err != nil {
		t.Fatalf("Embed error: %v", err)
	}
	if len(vec) == 0 {
		t.Fatal("expected non-empty vector")
	}
}

func TestTFIDFEmbedderEmpty(t *testing.T) {
	emb := TFIDFEmbedder{}
	vec, err := emb.Embed("")
	if err != nil {
		t.Fatalf("Embed error: %v", err)
	}
	if vec == nil {
		t.Fatal("expected non-nil vector for empty query")
	}
}

// ---------------------------------------------------------------------------
// SemanticCache with TFIDF
// ---------------------------------------------------------------------------

func TestSemanticCacheExactHit(t *testing.T) {
	sc := newSemanticTestCache(t, TFIDFEmbedder{})
	if err := sc.Set("golang concurrency", []string{"reddit"}, "result", time.Hour); err != nil {
		t.Fatalf("Set: %v", err)
	}
	data, found, err := sc.Get("golang concurrency", []string{"test"})
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !found {
		t.Fatal("expected exact hit")
	}
	if string(data) != `"result"` {
		t.Errorf("unexpected payload: %s", data)
	}
}

func TestSemanticCacheSemanticHit(t *testing.T) {
	sc := newSemanticTestCache(t, TFIDFEmbedder{})
	// Store under one phrasing.
	if err := sc.Set("how to test go code", []string{"reddit"}, "answer", time.Hour); err != nil {
		t.Fatalf("Set: %v", err)
	}
	// Query with overlapping words — TF-IDF bag-of-words should match.
	data, found, err := sc.Get("test go code how to", []string{"test"})
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !found {
		t.Fatal("expected semantic hit for reordered words")
	}
	if string(data) != `"answer"` {
		t.Errorf("unexpected payload: %s", data)
	}
}

func TestSemanticCacheMiss(t *testing.T) {
	sc := newSemanticTestCache(t, TFIDFEmbedder{})
	if err := sc.Set("golang concurrency", []string{"reddit"}, "r1", time.Hour); err != nil {
		t.Fatalf("Set: %v", err)
	}
	_, found, err := sc.Get("completely different topic xyz", []string{"test"})
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if found {
		t.Error("expected miss for unrelated query")
	}
}

func TestSemanticCacheNilEmbedder(t *testing.T) {
	sc := newSemanticTestCache(t, NilEmbedder{})
	if err := sc.Set("query", []string{"reddit"}, "result", time.Hour); err != nil {
		t.Fatalf("Set: %v", err)
	}
	// Exact match still works because keys map is populated.
	data, found, err := sc.Get("query", []string{"test"})
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !found {
		t.Fatal("expected exact hit with nil embedder")
	}
	if string(data) != `"result"` {
		t.Errorf("unexpected payload: %s", data)
	}
	// No semantic match possible.
	_, found, err = sc.Get("totally different query", []string{"test"})
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if found {
		t.Error("expected miss with nil embedder for new query")
	}
}

// ---------------------------------------------------------------------------
// Graceful degradation
// ---------------------------------------------------------------------------

type errorEmbedder struct{}

func (errorEmbedder) Embed(string) ([]float32, error) {
	return nil, errors.New("boom")
}

func TestSemanticCacheEmbedderError(t *testing.T) {
	sc := newSemanticTestCache(t, errorEmbedder{})
	if err := sc.Set("query a", []string{"reddit"}, "result", time.Hour); err != nil {
		t.Fatalf("Set: %v", err)
	}
	// Exact match works (key stored even though embedding failed).
	data, found, err := sc.Get("query a", []string{"test"})
	if err != nil {
		t.Fatalf("Get exact: %v", err)
	}
	if !found {
		t.Fatal("expected exact hit despite embedder error")
	}
	if string(data) != `"result"` {
		t.Errorf("unexpected payload: %s", data)
	}
	// Semantic lookup degrades gracefully — no hit, no error.
	_, found, err = sc.Get("query a variant", []string{"test"})
	if err != nil {
		t.Fatalf("Get semantic: %v", err)
	}
	if found {
		t.Error("expected miss when embedder errors")
	}
}

func TestNewEmbedderDefaultsToTFIDF(t *testing.T) {
	// SIN_NIM_API_KEY is not set in test env → should get TFIDFEmbedder.
	e := NewEmbedder()
	if _, ok := e.(TFIDFEmbedder); !ok {
		t.Errorf("expected TFIDFEmbedder when no API key, got %T", e)
	}
}
