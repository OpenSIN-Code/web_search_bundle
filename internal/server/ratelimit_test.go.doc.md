What it does: Tests the per-IP token-bucket rate limiter in `internal/server/ratelimit.go`.

Dependencies: Tests `rateLimitMiddleware`, `clientIP`, and `newLimiterStore`.

Important scenarios covered:
- Requests under the limit return 200.
- Requests over the limit return 429 with `Retry-After`.
- Limiting is per-IP (different IPs have separate buckets).
- Token bucket refills over time.
- `X-Forwarded-For` is preferred over `RemoteAddr` for client IP extraction.
- Default values are applied when config is zero/invalid.
