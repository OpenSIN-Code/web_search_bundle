What: Production deployment guide for the sin-websearch HTTP API.

Dependency map: README.md links to this file. HTTP API behavior (port 8787, bearer token, rate limiting) is implemented in internal/server/http.go. Configuration keys are defined in internal/config/ and documented in README.md.

Important config values & limits:
- Default HTTP port: 8787 (override via http_port in sin-websearch.yaml).
- Default rate limit: 10 rps, burst 20 (override via rate_limit_rps and rate_limit_burst).
- Bearer token is optional but required for all endpoints when configured (via token key).

Why these decisions:
- Caddy is shown first because it handles automatic HTTPS, reducing operational overhead.
- nginx is shown as an alternative for environments with existing certificate management.
- systemd hardening flags (NoNewPrivileges, PrivateTmp, ProtectSystem) follow common service-hardening practice without blocking the API's normal operation.
- Docker example uses a published image placeholder (ghcr.io/opensin-code/sin-websearch:latest) so operators can replace it with their own build.
- Reverse proxy examples pass the Host and X-Forwarded-* headers so downstream behavior matches the public hostname.

Usage examples:
- See Caddyfile or nginx snippets for reverse-proxy TLS termination.
- See systemd unit for running the binary as a service on Linux.
- See docker-compose.yml for a containerized deployment behind Caddy.

Known caveats:
- Rate limiting is per-IP by default; behind a reverse proxy the limiter sees the proxy IP unless the proxy is configured to forward the real client IP.
- The guide assumes the binary is already built or a published image is available; it does not replace the local development setup in README.md.
- Security hardening in the systemd unit may need adjustment if the binary requires additional paths at runtime (e.g., for cache or log directories).
