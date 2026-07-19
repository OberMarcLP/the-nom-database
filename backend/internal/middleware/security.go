package middleware

import (
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/nomdb/backend/internal/errors"
	"github.com/nomdb/backend/internal/logger"
)

// SecurityHeadersMiddleware adds security headers to all responses
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prevent clickjacking attacks
		w.Header().Set("X-Frame-Options", "DENY")

		// Prevent MIME type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// Enable XSS protection (legacy but still useful)
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Referrer policy
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Content Security Policy (adjust based on your needs)
		// This is a basic policy - tighten for production
		csp := []string{
			"default-src 'self'",
			"img-src 'self' data: https:",
			"script-src 'self'",
			"style-src 'self' 'unsafe-inline'",
		}
		w.Header().Set("Content-Security-Policy", strings.Join(csp, "; "))

		// Strict Transport Security (HSTS) - only enable if using HTTPS
		// w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		// Permissions Policy (formerly Feature-Policy)
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		next.ServeHTTP(w, r)
	})
}

// MaxBytesMiddleware limits the size of request bodies to prevent memory exhaustion attacks
func MaxBytesMiddleware(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Limit request body size
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)

			next.ServeHTTP(w, r)
		})
	}
}

// ValidateContentTypeMiddleware ensures requests have valid Content-Type headers
func ValidateContentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip validation for GET, HEAD, OPTIONS
		if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}

		// Check for photo/avatar upload endpoints (multipart/form-data)
		if (strings.Contains(r.URL.Path, "/photos") || strings.Contains(r.URL.Path, "/avatar")) && r.Method == "POST" {
			contentType := r.Header.Get("Content-Type")
			if !strings.HasPrefix(contentType, "multipart/form-data") {
				http.Error(w, "Invalid Content-Type for file upload", http.StatusUnsupportedMediaType)
				return
			}
			next.ServeHTTP(w, r)
			return
		}

		// For other POST/PUT/PATCH requests, expect JSON
		if r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH" {
			contentType := r.Header.Get("Content-Type")
			if contentType != "" && !strings.Contains(contentType, "application/json") {
				http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

// SanitizeInputMiddleware provides basic input sanitization
func SanitizeInputMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Sanitize query parameters
		query := r.URL.Query()
		for key, values := range query {
			for i, value := range values {
				// Remove null bytes (can cause issues)
				query[key][i] = strings.ReplaceAll(value, "\x00", "")
			}
		}
		r.URL.RawQuery = query.Encode()

		next.ServeHTTP(w, r)
	})
}

// RecoveryMiddleware recovers from panics and returns a 500 error
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic with stack trace
				logger.Error("PANIC recovered: %v\nStack trace:\n%s", err, string(debug.Stack()))

				// Return structured error response
				errors.RespondWithError(w, errors.InternalError("An unexpected error occurred"))
			}
		}()

		next.ServeHTTP(w, r)
	})
}
