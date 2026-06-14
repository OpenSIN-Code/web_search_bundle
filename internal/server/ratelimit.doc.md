What it does: Implements per-IP token-bucket rate limiting for the HTTP API using `golang.org/x/time/rate`.

Dependencies: Used by `internal/server/http.go` as middleware around all `/api/v1/*` handlers. Config values come from `internal/config.Config` (`rate_limit_rps`, `rate_limit_burst`).

Important config values & limits:
- Default `rate_limit_rps`: 10.0 requests per second.
- Default `rate_limit_burst`: 20 requests.
- Returns HTTP 429 with `Retry-After: 1` when exceeded.
- Supports `X-Forwarded-For` and `X-Real-Ip` for proxied deployments.

Caveats:
- In-memory only; limiters are not shared across multiple server instances. For multi-instance deployments, use a Redis/shared store.
- Per-IP map grows unbounded; a sweeper to remove stale entries is not implemented.
- The `X-Forwarded-For` header can be spoofed if not sanitized by a trusted proxy.
