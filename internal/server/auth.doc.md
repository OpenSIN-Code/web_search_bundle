Bearer-token authentication middleware for the HTTP API.

- `auth.go` provides `authMiddleware`, which wraps handlers and enforces `Authorization: Bearer <token>` when `config.Config.Token` is configured.
- If no token is configured (nil config or empty token), all requests are allowed. This keeps local development simple while allowing production hardening.
- Used by `http.go` to protect all `/api/v1/*` endpoints.

Related files: `http.go`, `config/config.go`, `auth_test.go`.
