# health_test.go

Tests for the expanded `/health` endpoint.

- `TestHandleHealthDirect` calls `handleHealth` directly and verifies the JSON
  payload contains `status`, `uptime`, and `version`.
- `TestHandleHealthViaRouter` exercises the full HTTP router via `HTTPServer.Handler()`
  to ensure the endpoint is wired and returns JSON.

Related files: `handlers.go`, `http.go`, `config/config.go`.
