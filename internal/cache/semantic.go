// SPDX-License-Identifier: MIT
// Purpose: Semantic (embedding-based) caching layer on top of the SQLite cache.
// If a new query is >threshold cosine-similar to a cached query, the cached
// result is returned without an API call.  Falls back to a pure-Go TF-IDF
// embedder when no NVIDIA NIM API key is available.
// Docs: internal/cache/semantic.doc.md
package cache

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// Embedder converts a query string into a dense float32 vector.
type Embedder interface {
	Embed(query string) ([]float32, error)
}

// SemanticCache wraps *Cache with an embedding-based similarity layer.
type SemanticCache struct {
	inner     *Cache
	embeddings map[string][]float32 // query → embedding vector (in-memory)
	keys      map[string]string     // query → HashKey for inner cache lookup
	threshold float64               // default 0.85
	mu        sync.RWMutex
	embedder  Embedder
}

// NewSemanticCache creates a SemanticCache around inner using the supplied embedder.
func NewSemanticCache(inner *Cache, embedder Embedder) *SemanticCache {
	return &SemanticCache{
		inner:      inner,
		embeddings: make(map[string][]float32),
		keys:       make(map[string]string),
		threshold:  0.85,
		embedder:   embedder,
	}
}

// Set stores a result in the cache together with its embedding.
func (sc *SemanticCache) Set(query string, sources []string, payload interface{}, ttl time.Duration) error {
	key := HashKey(query, sources)
	if err := sc.inner.Set(key, sources, payload, ttl); err != nil {
		return err
	}

	sc.mu.Lock()
	sc.keys[query] = key // always store for exact-match fast path
	if sc.embedder != nil {
		if emb, err := sc.embedder.Embed(query); err == nil && emb != nil {
			sc.embeddings[query] = emb
		}
	}
	sc.mu.Unlock()
	return nil
}

// Get first tries an exact key match, then iterates stored embeddings for a
// cosine similarity above the threshold.  Returns the cached payload on hit.
func (sc *SemanticCache) Get(query string) ([]byte, bool, error) {
	// Fast path: exact query match.
	sc.mu.RLock()
	key, exact := sc.keys[query]
	sc.mu.RUnlock()
	if exact {
		return sc.inner.Get(key)
	}

	// Semantic path: compute embedding and search.
	if sc.embedder == nil {
		return nil, false, nil
	}
	qEmb, err := sc.embedder.Embed(query)
	if err != nil || qEmb == nil {
		return nil, false, nil
	}

	sc.mu.RLock()
	defer sc.mu.RUnlock()

	var bestKey string
	var bestSim float64
	for storedQuery, emb := range sc.embeddings {
		sim := cosineSimilarity(qEmb, emb)
		if sim > bestSim {
			bestSim = sim
			bestKey = sc.keys[storedQuery]
		}
	}

	if bestSim >= sc.threshold && bestKey != "" {
		return sc.inner.Get(bestKey)
	}
	return nil, false, nil
}

// SetThreshold overrides the default 0.85 cosine-similarity threshold.
func (sc *SemanticCache) SetThreshold(t float64) {
	sc.mu.Lock()
	sc.threshold = t
	sc.mu.Unlock()
}

// ---------------------------------------------------------------------------
// Embedders
// ---------------------------------------------------------------------------

// NilEmbedder disables semantic matching by always returning nil.
type NilEmbedder struct{}

// Embed implements Embedder.
func (NilEmbedder) Embed(string) ([]float32, error) { return nil, nil }

// TFIDFEmbedder produces a fixed-dimensional term-frequency vector using the
// hashing trick.  Each token is hashed to a position in a 256-dim vector, so
// the same word always lands in the same slot — making cross-query cosine
// similarity meaningful without a shared vocabulary.  No external deps.
type TFIDFEmbedder struct{}

// tfidfDims is the fixed vector dimensionality for the hashing trick.
const tfidfDims = 256

// Embed implements Embedder.
func (TFIDFEmbedder) Embed(query string) ([]float32, error) {
	tokens := tokenize(query)
	vec := make([]float32, tfidfDims)
	for _, tok := range tokens {
		idx := hashToken(tok) % tfidfDims
		vec[idx]++
	}
	return vec, nil
}

// hashToken returns a non-crypto hash of a token (FNV-1a).
func hashToken(s string) int {
	var h uint32 = 2166136261
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= 16777619
	}
	return int(h)
}

// tokenize splits a string into lowercase word tokens.
func tokenize(s string) []string {
	s = strings.ToLower(s)
	fields := strings.Fields(s)
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		trimmed := strings.Trim(f, ".,!?;:\"'()[]{}")
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

// NIMEmbedder calls the NVIDIA NIM embeddings API.
type NIMEmbedder struct {
	APIKey string
	Model  string
	Client *http.Client
}

// NewNIMEmbedder creates a NIMEmbedder from an explicit API key.
func NewNIMEmbedder(apiKey string) *NIMEmbedder {
	return &NIMEmbedder{
		APIKey: apiKey,
		Model:  "nvidia/nv-embed-v1",
		Client: &http.Client{Timeout: 15 * time.Second},
	}
}

// Embed implements Embedder.
func (n *NIMEmbedder) Embed(query string) ([]float32, error) {
	if n.APIKey == "" {
		return nil, fmt.Errorf("NIM embedder: no API key")
	}
	body := map[string]string{
		"input": query,
		"model": n.Model,
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost,
		"https://integrate.api.nvidia.com/v1/embeddings", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+n.APIKey)

	resp, err := n.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("NIM embedder: HTTP %d", resp.StatusCode)
	}

	var result struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if len(result.Data) == 0 {
		return nil, fmt.Errorf("NIM embedder: empty response")
	}
	return result.Data[0].Embedding, nil
}

// NewEmbedder returns a NIMEmbedder when SIN_NIM_API_KEY is set, otherwise a
// TFIDFEmbedder (pure Go, no external dependencies).
func NewEmbedder() Embedder {
	if key := os.Getenv("SIN_NIM_API_KEY"); key != "" {
		return NewNIMEmbedder(key)
	}
	return TFIDFEmbedder{}
}

// ---------------------------------------------------------------------------
// Math
// ---------------------------------------------------------------------------

// cosineSimilarity computes the cosine of the angle between two float32
// vectors.  Vectors of differing length are compared over the shared prefix
// (sufficient for TF-IDF bag-of-words where dimension order is arbitrary but
// consistent per embedder).  Returns 0 for zero-length or zero-magnitude
// vectors.
func cosineSimilarity(a, b []float32) float64 {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	if n == 0 {
		return 0
	}

	var dot, normA, normB float64
	for i := 0; i < n; i++ {
		af := float64(a[i])
		bf := float64(b[i])
		dot += af * bf
		normA += af * af
		normB += bf * bf
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}
