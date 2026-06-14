// Purpose: Tests for HTTP API authentication middleware.
// Docs: auth_test.doc.md
package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/OpenSIN-Code/web_search_bundle/internal/config"
)

func TestAuthMiddlewareNoToken(t *testing.T) {
	s := NewHTTPServer(nil, nil)
	called := false
	handler := s.authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/search", nil)
	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if !called {
		t.Error("handler was not called")
	}
}

func TestAuthMiddlewareMissingToken(t *testing.T) {
	s := NewHTTPServer(&config.Config{Token: "secret"}, nil)
	called := false
	handler := s.authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/search", nil)
	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
	if called {
		t.Error("handler should not have been called")
	}
}

func TestAuthMiddlewareInvalidToken(t *testing.T) {
	s := NewHTTPServer(&config.Config{Token: "secret"}, nil)
	called := false
	handler := s.authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/search", nil)
	req.Header.Set("Authorization", "Bearer wrong")
	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
	if called {
		t.Error("handler should not have been called")
	}
}

func TestAuthMiddlewareValidToken(t *testing.T) {
	s := NewHTTPServer(&config.Config{Token: "secret"}, nil)
	called := false
	handler := s.authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/search", nil)
	req.Header.Set("Authorization", "Bearer secret")
	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if !called {
		t.Error("handler was not called")
	}
}
