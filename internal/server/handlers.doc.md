HTTP handlers for read-only and streaming API endpoints.

- `handlers.go` contains `handleHealth`, `handleSearch`, `handlePulse`, `handleResolve`, `handleSearchStream`, and the shared helpers `writeJSON` and `writeSSE`.
- `handleHealth` returns a JSON object with `status`, `uptime` (since server creation), and `version` (from `internal/config.Version`).
- Handlers are protected by `authMiddleware` in `http.go` and instrumented by the observability middleware in `middleware.go`.

Related files: `http.go`, `auth.go`, `middleware.go`, `metrics.go`, `http_test.go`.
