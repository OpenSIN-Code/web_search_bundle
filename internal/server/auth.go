// Purpose: API-key/token authentication middleware for the HTTP API.
// Docs: auth.doc.md
package server

import (
	"net/http"
	"strings"
)

// authMiddleware protects HTTP endpoints with a bearer token.
// If the server has no configured token, all requests are allowed.
func (s *HTTPServer) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.cfg == nil || s.cfg.Token == "" {
			next(w, r)
			return
		}

		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing or invalid authorization"})
			return
		}
		if strings.TrimPrefix(auth, "Bearer ") != s.cfg.Token {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token"})
			return
		}
		next(w, r)
	}
}
