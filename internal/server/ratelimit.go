// SPDX-License-Identifier: MIT
// Purpose: token-bucket rate limiting for the HTTP API.
// Docs: internal/server/ratelimit.doc.md
package server

import (
	"net"
	"net/http"
	"sync"

	"golang.org/x/time/rate"
)

// limiterStore keeps a map of client IP -> limiter. It is protected by a mutex.
type limiterStore struct {
	mu       sync.RWMutex
	limiters map[string]*rate.Limiter
	rps      rate.Limit
	burst    int
}

// newLimiterStore creates a store with the given rate and burst.
func newLimiterStore(rps float64, burst int) *limiterStore {
	if rps <= 0 {
		rps = 10.0
	}
	if burst <= 0 {
		burst = int(rps * 2)
		if burst < 1 {
			burst = 1
		}
	}
	return &limiterStore{
		limiters: make(map[string]*rate.Limiter),
		rps:      rate.Limit(rps),
		burst:    burst,
	}
}

// getLimiter returns the limiter for the given IP, creating one if needed.
func (s *limiterStore) getLimiter(ip string) *rate.Limiter {
	s.mu.RLock()
	lim, ok := s.limiters[ip]
	s.mu.RUnlock()
	if ok {
		return lim
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	lim, ok = s.limiters[ip]
	if ok {
		return lim
	}
	lim = rate.NewLimiter(s.rps, s.burst)
	s.limiters[ip] = lim
	return lim
}

// rateLimitMiddleware returns an HTTP middleware that applies per-IP token-bucket
// limiting. If the limit is exceeded it returns 429 Too Many Requests with a
// Retry-After header.
func (s *limiterStore) rateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		lim := s.getLimiter(ip)
		if !lim.Allow() {
			w.Header().Set("Retry-After", "1")
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next(w, r)
	}
}

// clientIP extracts the client IP from the request, preferring X-Forwarded-For
// when the server is behind a proxy. It falls back to RemoteAddr.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the comma-separated list.
		for i := 0; i < len(xff); i++ {
			if xff[i] == ',' {
				xff = xff[:i]
				break
			}
		}
		if host, _, err := net.SplitHostPort(xff); err == nil {
			return host
		}
		return xff
	}
	if xri := r.Header.Get("X-Real-Ip"); xri != "" {
		if host, _, err := net.SplitHostPort(xri); err == nil {
			return host
		}
		return xri
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
