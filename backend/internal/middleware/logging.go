package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nomdb/backend/internal/logger"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
	bytes      int64
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytes += int64(n)
	return n, err
}

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if request ID already exists in header
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			// Generate new UUID for request ID
			requestID = uuid.New().String()
		}

		// Add request ID to response headers
		w.Header().Set("X-Request-ID", requestID)

		// Add request ID to context
		ctx := context.WithValue(r.Context(), logger.RequestIDKey, requestID)

		// Call next handler with updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip logging and metrics for health check requests (Docker
		// healthchecks, uptime monitors, load balancers) regardless of client
		if r.URL.Path == "/api/health" {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()

		// Get request ID from context
		requestID := ""
		if id, ok := r.Context().Value(logger.RequestIDKey).(string); ok {
			requestID = id
		}

		// Wrap the response writer to capture status code and bytes
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
			bytes:          0,
		}

		// Get client IP (handle X-Forwarded-For for proxies)
		clientIP := r.RemoteAddr
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			clientIP = strings.Split(forwarded, ",")[0]
		}

		// Log incoming request
		logger.Debug("→ %s %s from %s", r.Method, r.URL.Path, clientIP)

		// Call the next handler
		next.ServeHTTP(rw, r)

		// Calculate duration
		duration := time.Since(start)

		// Record metrics
		GetMetrics().RecordRequest(r.Method, r.URL.Path, rw.statusCode, duration)

		// Log response with structured logging
		logger.LogRequest(
			r.Method,
			r.URL.Path,
			requestID,
			clientIP,
			duration,
			rw.statusCode,
			rw.bytes,
		)
	})
}

func getStatusEmoji(statusCode int) string {
	switch {
	case statusCode >= 500:
		return "💥"
	case statusCode >= 400:
		return "⚠️ "
	case statusCode >= 300:
		return "↪️ "
	case statusCode >= 200:
		return "✅"
	default:
		return "ℹ️ "
	}
}
