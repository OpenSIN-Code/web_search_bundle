Tests for the HTTP authentication middleware.

- Covers the no-token, missing-token, invalid-token, and valid-token cases.
- Uses `httptest` to exercise `authMiddleware` directly.

Related files: `auth.go`, `http.go`, `config/config.go`.
