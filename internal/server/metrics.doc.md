# metrics.go

Prometheus-style metrics collection and exposition for the HTTP server.

- Uses only the Go standard library (`sync`, `sort`, `fmt`, `strings`).
- Stores `http_requests_total` counters labelled by `method`, `path`, and `status`.
- Stores `http_request_duration_seconds` histograms labelled by `method` and `path`.
- Exposes the `/metrics` route in Prometheus text format (`text/plain; version=0.0.4`).
- Histogram buckets: 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10 seconds.
- Populated by the observability middleware in `middleware.go` and consumed by `http.go`.

Related files: `http.go`, `middleware.go`, `metrics_test.go`.
