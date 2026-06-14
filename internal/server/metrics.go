// SPDX-License-Identifier: MIT
// Purpose: Prometheus-style metrics collection and exposition for the HTTP server.
// Docs: internal/server/metrics.doc.md
package server

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// durationBuckets are the histogram bucket boundaries for request latency. They
// mirror the classic Prometheus default buckets but stop at 10 seconds because
// the server already enforces 120 second write timeouts.
var durationBuckets = []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10}

// requestKey uniquely identifies a counter sample by method, path and status.
type requestKey struct {
	method, path, status string
}

// durationKey uniquely identifies a histogram sample by method and path.
type durationKey struct {
	method, path string
}

// durationStats holds the running count, sum and per-bucket counts for a histogram.
type durationStats struct {
	count   int64
	sum     float64
	buckets map[float64]int64
}

// metricsCollector stores request counters and duration histograms in memory.
// It is safe for concurrent use by multiple HTTP goroutines.
type metricsCollector struct {
	mu        sync.RWMutex
	requests  map[requestKey]int64
	durations map[durationKey]*durationStats
}

// newMetricsCollector creates an empty collector.
func newMetricsCollector() *metricsCollector {
	return &metricsCollector{
		requests:  make(map[requestKey]int64),
		durations: make(map[durationKey]*durationStats),
	}
}

// recordRequest increments the counter for the given method/path/status.
func (m *metricsCollector) recordRequest(method, path, status string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requests[requestKey{method, path, status}]++
}

// recordDuration observes a request duration into the histogram for method/path.
func (m *metricsCollector) recordDuration(method, path string, duration time.Duration) {
	secs := duration.Seconds()
	key := durationKey{method, path}
	m.mu.Lock()
	defer m.mu.Unlock()

	stats := m.durations[key]
	if stats == nil {
		stats = &durationStats{buckets: make(map[float64]int64)}
		m.durations[key] = stats
	}
	stats.count++
	stats.sum += secs
	for _, b := range durationBuckets {
		if secs <= b {
			stats.buckets[b]++
		}
	}
}

// handleMetrics exposes the current metrics in Prometheus text format.
func (s *HTTPServer) handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	body := s.metrics.exposition()
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(body)); err != nil {
		s.logger.Warn("failed to write metrics response", "error", err)
	}
}

// exposition renders the in-memory metrics as a Prometheus exposition string.
func (m *metricsCollector) exposition() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var b strings.Builder

	fmt.Fprintf(&b, "# HELP http_requests_total Total HTTP requests handled\n")
	fmt.Fprintf(&b, "# TYPE http_requests_total counter\n")

	type reqEntry struct {
		key   requestKey
		count int64
	}
	reqEntries := make([]reqEntry, 0, len(m.requests))
	for k, v := range m.requests {
		reqEntries = append(reqEntries, reqEntry{k, v})
	}
	sort.Slice(reqEntries, func(i, j int) bool {
		a, c := reqEntries[i].key, reqEntries[j].key
		if a.method != c.method {
			return a.method < c.method
		}
		if a.path != c.path {
			return a.path < c.path
		}
		return a.status < c.status
	})
	for _, e := range reqEntries {
		fmt.Fprintf(&b, "http_requests_total{method=%q,path=%q,status=%q} %d\n",
			e.key.method, e.key.path, e.key.status, e.count)
	}

	fmt.Fprintf(&b, "# HELP http_request_duration_seconds HTTP request duration distribution\n")
	fmt.Fprintf(&b, "# TYPE http_request_duration_seconds histogram\n")

	type durEntry struct {
		key   durationKey
		stats *durationStats
	}
	durEntries := make([]durEntry, 0, len(m.durations))
	for k, v := range m.durations {
		durEntries = append(durEntries, durEntry{k, v})
	}
	sort.Slice(durEntries, func(i, j int) bool {
		a, c := durEntries[i].key, durEntries[j].key
		if a.method != c.method {
			return a.method < c.method
		}
		return a.path < c.path
	})
	for _, e := range durEntries {
		for _, bucket := range durationBuckets {
			fmt.Fprintf(&b, "http_request_duration_seconds_bucket{method=%q,path=%q,le=%q} %d\n",
				e.key.method, e.key.path, formatFloat(bucket), e.stats.buckets[bucket])
		}
		fmt.Fprintf(&b, "http_request_duration_seconds_bucket{method=%q,path=%q,le=%q} %d\n",
			e.key.method, e.key.path, "+Inf", e.stats.count)
		fmt.Fprintf(&b, "http_request_duration_seconds_sum{method=%q,path=%q} %g\n",
			e.key.method, e.key.path, e.stats.sum)
		fmt.Fprintf(&b, "http_request_duration_seconds_count{method=%q,path=%q} %d\n",
			e.key.method, e.key.path, e.stats.count)
	}

	return b.String()
}

// formatFloat renders a float in the shortest decimal form Prometheus expects.
func formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}
