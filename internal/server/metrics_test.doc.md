# metrics_test.go

Tests for the `/metrics` endpoint and the Prometheus-style metrics collector.

- `TestMetricsEndpoint` starts the full HTTP router, makes a request to `/health`,
  then scrapes `/metrics`.
- Verifies the response status, `text/plain` content type, and the presence of
  counter and histogram lines for the recorded request.

Related files: `metrics.go`, `middleware.go`, `http.go`.
