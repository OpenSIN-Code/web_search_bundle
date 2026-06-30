// SPDX-License-Identifier: MIT
// Purpose: Per-engine health metrics and stats tracking.
package orchestrator

import (
	"sync"
	"sync/atomic"
	"time"
)

// EngineStats tracks runtime metrics for a single engine.
type EngineStats struct {
	Name        string
	Requests    atomic.Int64
	Successes   atomic.Int64
	Failures    atomic.Int64
	CacheHits   atomic.Int64
	TotalLatency atomic.Int64 // nanoseconds
	LastUsed    atomic.Int64  // unix nano
}

// StatsRegistry holds per-engine stats, safe for concurrent access.
type StatsRegistry struct {
	mu    sync.RWMutex
	stats map[string]*EngineStats
}

// NewStatsRegistry creates an empty registry.
func NewStatsRegistry() *StatsRegistry {
	return &StatsRegistry{stats: make(map[string]*EngineStats)}
}

// Get returns the stats for an engine, creating if absent.
func (r *StatsRegistry) Get(name string) *EngineStats {
	r.mu.RLock()
	s, ok := r.stats[name]
	r.mu.RUnlock()
	if ok {
		return s
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok = r.stats[name]
	if !ok {
		s = &EngineStats{Name: name}
		r.stats[name] = s
	}
	return s
}

// RecordRequest records a request start.
func (r *StatsRegistry) RecordRequest(name string) {
	s := r.Get(name)
	s.Requests.Add(1)
	s.LastUsed.Store(time.Now().UnixNano())
}

// RecordSuccess records a successful response with latency.
func (r *StatsRegistry) RecordSuccess(name string, latency time.Duration) {
	s := r.Get(name)
	s.Successes.Add(1)
	s.TotalLatency.Add(int64(latency))
}

// RecordFailure records a failed response.
func (r *StatsRegistry) RecordFailure(name string) {
	s := r.Get(name)
	s.Failures.Add(1)
}

// RecordCacheHit records a cache hit (no engine call needed).
func (r *StatsRegistry) RecordCacheHit(name string) {
	s := r.Get(name)
	s.CacheHits.Add(1)
}

// Snapshot returns a copy of all engine stats for reporting.
func (r *StatsRegistry) Snapshot() []EngineStatsSnapshot {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]EngineStatsSnapshot, 0, len(r.stats))
	for _, s := range r.stats {
		req := s.Requests.Load()
		suc := s.Successes.Load()
		fail := s.Failures.Load()
		hits := s.CacheHits.Load()
		totalLat := s.TotalLatency.Load()
		lastUsed := s.LastUsed.Load()

		var avgLatency float64
		if suc > 0 {
			avgLatency = float64(totalLat) / float64(suc) / float64(time.Millisecond)
		}
		var successRate float64
		if req > 0 {
			successRate = float64(suc) / float64(req)
		}

		var lastUsedStr string
		if lastUsed > 0 {
			lastUsedStr = time.Unix(0, lastUsed).Format(time.RFC3339)
		}

		out = append(out, EngineStatsSnapshot{
			Name:        s.Name,
			Requests:    req,
			Successes:   suc,
			Failures:    fail,
			CacheHits:   hits,
			AvgLatency:  avgLatency,
			SuccessRate: successRate,
			LastUsed:    lastUsedStr,
		})
	}
	return out
}

// EngineStatsSnapshot is a point-in-time copy of engine metrics.
type EngineStatsSnapshot struct {
	Name        string  `json:"name"`
	Requests    int64   `json:"requests"`
	Successes   int64   `json:"successes"`
	Failures    int64   `json:"failures"`
	CacheHits   int64   `json:"cache_hits"`
	AvgLatency  float64 `json:"avg_latency_ms"`
	SuccessRate float64 `json:"success_rate"`
	LastUsed    string  `json:"last_used"`
}
