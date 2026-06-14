HTTP handlers for read-only and streaming API endpoints.

- `handlers.go` contains `handleHealth`, `handleSearch`, `handlePulse`, `handleResolve`, and `handleSearchStream`.
- Shared response helpers `writeJSON` and `writeSSE` live here.
- All handlers are wrapped by `authMiddleware` in `http.go`.

Related files: `http.go`, `auth.go`, `http_test.go`.
