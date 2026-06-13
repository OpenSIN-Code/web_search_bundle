# http_test.go

Tests for the HTTP REST API server.

- Imports: `server` (internal), `net/http`, `net/http/httptest`, `os`, `os/exec`, `path/filepath`, `strings`, `testing`
- Creates temporary git repos for alchemist endpoint tests.
- Tests: `/api/v1/alchemist` (single daemon), `/api/v1/alchemist/swarm`, missing `run_cmd` validation, method-not-allowed, and `buildAlchemistConfig` defaults/validation.
