# middleware.go

Request logging and observability middleware for the HTTP server.

- `recordingResponseWriter` wraps `http.ResponseWriter` to capture the response status
  code and duration while preserving `http.Flusher` for streaming endpoints.
- `observe` records Prometheus metrics and emits a structured access-log line via
  `log/slog` to stdout.
- Access-log fields: `method`, `path`, `status`, `duration` (seconds), `client_ip`.
- Logging is disabled when `config.Config.DisableRequestLogging` is `true`.

Related files: `http.go`, `metrics.go`, `ratelimit.go` (for `clientIP`), `metrics_test.go`.
