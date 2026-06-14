// SPDX-License-Identifier: MIT
// Purpose: Request logging and observability middleware for the HTTP server.
// Docs: internal/server/middleware.doc.md
package server

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

// recordingResponseWriter wraps http.ResponseWriter so the middleware can
// capture the response status code and duration.
type recordingResponseWriter struct {
	http.ResponseWriter
	status int
	wrote  bool
}

// WriteHeader captures the status before delegating to the underlying writer.
func (rr *recordingResponseWriter) WriteHeader(status int) {
	rr.status = status
	rr.wrote = true
	rr.ResponseWriter.WriteHeader(status)
}

// Write captures an implicit 200 status if the handler never calls WriteHeader.
func (rr *recordingResponseWriter) Write(b []byte) (int, error) {
	if !rr.wrote {
		rr.status = http.StatusOK
		rr.wrote = true
	}
	return rr.ResponseWriter.Write(b)
}

// Flush preserves the http.Flusher interface for streaming handlers.
func (rr *recordingResponseWriter) Flush() {
	if f, ok := rr.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// observe wraps an HTTP handler with metrics recording and request logging.
// It records status codes, latencies and a structured access log line to stdout.
// Logging is skipped when config.DisableRequestLogging is true.
func (s *HTTPServer) observe(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &recordingResponseWriter{ResponseWriter: w}
		next(rec, r)

		status := rec.status
		if status == 0 {
			status = http.StatusOK
		}
		duration := time.Since(start)

		s.metrics.recordRequest(r.Method, r.URL.Path, strconv.Itoa(status))
		s.metrics.recordDuration(r.Method, r.URL.Path, duration)

		if s.cfg == nil || !s.cfg.DisableRequestLogging {
			ip := clientIP(r)
			s.logger.Info("request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", status),
				slog.Float64("duration", duration.Seconds()),
				slog.String("client_ip", ip),
			)
		}
	}
}
