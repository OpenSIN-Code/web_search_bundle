// SPDX-License-Identifier: MIT
// Purpose: tests for the HTTP rate-limiting middleware.
package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimitMiddlewareAllowsUnderLimit(t *testing.T) {
	store := newLimiterStore(100, 100)
	handler := store.rateLimitMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/search", nil)
	rr := httptest.NewRecorder()
	handler(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestRateLimitMiddlewareBlocksOverLimit(t *testing.T) {
	store := newLimiterStore(1, 1)
	handler := store.rateLimitMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// First request consumes the single burst token.
	req1 := httptest.NewRequest(http.MethodGet, "/api/v1/search", nil)
	rr1 := httptest.NewRecorder()
	handler(rr1, req1)
	if rr1.Code != http.StatusOK {
		t.Fatalf("expected first request 200, got %d", rr1.Code)
	}

	// Second request from same IP should be rate limited.
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/search", nil)
	rr2 := httptest.NewRecorder()
	handler(rr2, req2)
	if rr2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rr2.Code)
	}
	if rr2.Header().Get("Retry-After") != "1" {
		t.Fatalf("expected Retry-After header, got %q", rr2.Header().Get("Retry-After"))
	}
}

func TestRateLimitMiddlewarePerIP(t *testing.T) {
	store := newLimiterStore(1, 1)
	handler := store.rateLimitMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// First IP consumes the token.
	req1 := httptest.NewRequest(http.MethodGet, "/api/v1/search", nil)
	req1.RemoteAddr = "1.2.3.4:1234"
	rr1 := httptest.NewRecorder()
	handler(rr1, req1)
	if rr1.Code != http.StatusOK {
		t.Fatalf("expected 200 for first IP, got %d", rr1.Code)
	}

	// Second IP should still be allowed.
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/search", nil)
	req2.RemoteAddr = "5.6.7.8:5678"
	rr2 := httptest.NewRecorder()
	handler(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("expected 200 for second IP, got %d", rr2.Code)
	}
}

func TestRateLimitMiddlewareRefillsOverTime(t *testing.T) {
	store := newLimiterStore(100, 1)
	handler := store.rateLimitMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/search", nil)
	// First request consumes the burst token.
	rr1 := httptest.NewRecorder()
	handler(rr1, req)
	if rr1.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr1.Code)
	}

	// Second request should be blocked immediately.
	rr2 := httptest.NewRecorder()
	handler(rr2, req)
	if rr2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 immediately, got %d", rr2.Code)
	}

	// Wait for the bucket to refill (rps=100 means 1 token per 10ms).
	time.Sleep(15 * time.Millisecond)

	rr3 := httptest.NewRecorder()
	handler(rr3, req)
	if rr3.Code != http.StatusOK {
		t.Fatalf("expected 200 after refill, got %d", rr3.Code)
	}
}

func TestClientIPPrefersXForwardedFor(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/search", nil)
	req.RemoteAddr = "1.2.3.4:1234"
	req.Header.Set("X-Forwarded-For", "9.8.7.6, 1.2.3.4")
	if got := clientIP(req); got != "9.8.7.6" {
		t.Fatalf("expected 9.8.7.6, got %q", got)
	}
}

func TestClientIPDefaultsToRemoteAddr(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/search", nil)
	req.RemoteAddr = "1.2.3.4:1234"
	if got := clientIP(req); got != "1.2.3.4" {
		t.Fatalf("expected 1.2.3.4, got %q", got)
	}
}

func TestNewLimiterStoreDefaults(t *testing.T) {
	store := newLimiterStore(0, 0)
	if store.rps != 10.0 {
		t.Fatalf("expected default rps 10.0, got %v", store.rps)
	}
	if store.burst != 20 {
		t.Fatalf("expected default burst 20, got %d", store.burst)
	}
}
