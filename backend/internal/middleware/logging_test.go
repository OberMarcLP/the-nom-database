package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nomdb/backend/internal/logger"
)

func TestRequestIDMiddleware(t *testing.T) {
	// Create a test handler that checks for request ID
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if request ID is in context
		if requestID, ok := r.Context().Value(logger.RequestIDKey).(string); ok {
			if requestID == "" {
				t.Error("Request ID in context is empty")
			}
			// Check if request ID is in response headers
			if headerID := w.Header().Get("X-Request-ID"); headerID == "" {
				t.Error("X-Request-ID header not set")
			} else if headerID != requestID {
				t.Errorf("X-Request-ID header (%s) doesn't match context value (%s)", headerID, requestID)
			}
			w.WriteHeader(http.StatusOK)
		} else {
			t.Error("Request ID not found in context")
			w.WriteHeader(http.StatusInternalServerError)
		}
	})

	// Wrap with RequestIDMiddleware
	handler := RequestIDMiddleware(testHandler)

	// Create test request
	req := httptest.NewRequest("GET", "/api/test", nil)
	rec := httptest.NewRecorder()

	// Execute request
	handler.ServeHTTP(rec, req)

	// Verify response
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	// Verify X-Request-ID header is present
	if requestID := rec.Header().Get("X-Request-ID"); requestID == "" {
		t.Error("X-Request-ID header not set in response")
	}
}

func TestRequestIDMiddleware_CustomRequestID(t *testing.T) {
	customID := "custom-request-id-12345"

	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if custom request ID is preserved
		if requestID, ok := r.Context().Value(logger.RequestIDKey).(string); ok {
			if requestID != customID {
				t.Errorf("Expected request ID to be %s, got %s", customID, requestID)
			}
		} else {
			t.Error("Request ID not found in context")
		}
		w.WriteHeader(http.StatusOK)
	})

	// Wrap with RequestIDMiddleware
	handler := RequestIDMiddleware(testHandler)

	// Create test request with custom X-Request-ID header
	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("X-Request-ID", customID)
	rec := httptest.NewRecorder()

	// Execute request
	handler.ServeHTTP(rec, req)

	// Verify custom ID is preserved in response
	if responseID := rec.Header().Get("X-Request-ID"); responseID != customID {
		t.Errorf("Expected X-Request-ID to be %s, got %s", customID, responseID)
	}
}

func TestLoggingMiddleware(t *testing.T) {
	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("test response"))
	})

	// Wrap with both RequestIDMiddleware and LoggingMiddleware
	handler := RequestIDMiddleware(LoggingMiddleware(testHandler))

	// Create test request
	req := httptest.NewRequest("GET", "/api/restaurants", nil)
	rec := httptest.NewRecorder()

	// Execute request
	handler.ServeHTTP(rec, req)

	// Verify response
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	// Verify response body
	if rec.Body.String() != "test response" {
		t.Errorf("Expected body 'test response', got '%s'", rec.Body.String())
	}
}

func TestLoggingMiddleware_HealthCheckFiltering(t *testing.T) {
	// Health checks are skipped for every client, not just Docker's wget
	tests := []struct {
		name      string
		userAgent string
	}{
		{"DockerWget", "Wget"},
		{"NoUserAgent", ""},
		{"UptimeMonitor", "Uptime-Kuma/1.23"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlerCalled := false
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true
				w.WriteHeader(http.StatusOK)
			})

			handler := LoggingMiddleware(testHandler)

			req := httptest.NewRequest("GET", "/api/health", nil)
			if tt.userAgent != "" {
				req.Header.Set("User-Agent", tt.userAgent)
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			// Verify handler was still called (filtering only affects logging, not execution)
			if !handlerCalled {
				t.Error("Expected handler to be called even for health checks")
			}

			if rec.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", rec.Code)
			}
		})
	}
}

func TestLoggingMiddleware_CapturesResponseData(t *testing.T) {
	responseBody := "test response body"
	expectedStatus := http.StatusCreated

	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(expectedStatus)
		_, _ = w.Write([]byte(responseBody))
	})

	// Wrap with LoggingMiddleware and RequestIDMiddleware
	handler := RequestIDMiddleware(LoggingMiddleware(testHandler))

	// Create test request
	req := httptest.NewRequest("POST", "/api/restaurants", nil)
	rec := httptest.NewRecorder()

	// Execute request
	handler.ServeHTTP(rec, req)

	// Verify status code
	if rec.Code != expectedStatus {
		t.Errorf("Expected status %d, got %d", expectedStatus, rec.Code)
	}

	// Verify response body
	if rec.Body.String() != responseBody {
		t.Errorf("Expected body '%s', got '%s'", responseBody, rec.Body.String())
	}

	// Verify Content-Length
	if rec.Body.Len() != len(responseBody) {
		t.Errorf("Expected body length %d, got %d", len(responseBody), rec.Body.Len())
	}
}

func TestLoggingMiddleware_XForwardedFor(t *testing.T) {
	forwardedIP := "192.168.1.100"

	// Create a test handler that doesn't need to do anything
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Wrap with LoggingMiddleware
	handler := RequestIDMiddleware(LoggingMiddleware(testHandler))

	// Create test request with X-Forwarded-For header
	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("X-Forwarded-For", forwardedIP+", 10.0.0.1")
	rec := httptest.NewRecorder()

	// Execute request
	handler.ServeHTTP(rec, req)

	// Verify response (logging should use forwarded IP, but we can't easily test that without mocking logger)
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
}

func TestResponseWriter_WriteHeader(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := &responseWriter{
		ResponseWriter: rec,
		statusCode:     http.StatusOK,
		bytes:          0,
	}

	// Write custom status code
	rw.WriteHeader(http.StatusCreated)

	// Verify status code is captured
	if rw.statusCode != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, rw.statusCode)
	}

	// Verify underlying writer received the status code
	if rec.Code != http.StatusCreated {
		t.Errorf("Expected recorder status %d, got %d", http.StatusCreated, rec.Code)
	}
}

func TestResponseWriter_Write(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := &responseWriter{
		ResponseWriter: rec,
		statusCode:     http.StatusOK,
		bytes:          0,
	}

	// Write data
	testData := []byte("test response data")
	n, err := rw.Write(testData)

	// Verify write succeeded
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify bytes written
	if n != len(testData) {
		t.Errorf("Expected to write %d bytes, wrote %d", len(testData), n)
	}

	// Verify byte counter
	if rw.bytes != int64(len(testData)) {
		t.Errorf("Expected byte counter to be %d, got %d", len(testData), rw.bytes)
	}

	// Verify underlying writer received the data
	if rec.Body.String() != string(testData) {
		t.Errorf("Expected body '%s', got '%s'", string(testData), rec.Body.String())
	}
}

func TestResponseWriter_MultipleWrites(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := &responseWriter{
		ResponseWriter: rec,
		statusCode:     http.StatusOK,
		bytes:          0,
	}

	// Write multiple chunks
	chunk1 := []byte("first ")
	chunk2 := []byte("second ")
	chunk3 := []byte("third")

	_, _ = rw.Write(chunk1)
	_, _ = rw.Write(chunk2)
	_, _ = rw.Write(chunk3)

	// Verify total bytes
	expectedBytes := int64(len(chunk1) + len(chunk2) + len(chunk3))
	if rw.bytes != expectedBytes {
		t.Errorf("Expected %d total bytes, got %d", expectedBytes, rw.bytes)
	}

	// Verify combined output
	expected := "first second third"
	if rec.Body.String() != expected {
		t.Errorf("Expected body '%s', got '%s'", expected, rec.Body.String())
	}
}

func TestGetStatusEmoji(t *testing.T) {
	tests := []struct {
		statusCode int
		emoji      string
	}{
		{200, "✅"},
		{201, "✅"},
		{299, "✅"},
		{300, "↪️ "},
		{301, "↪️ "},
		{302, "↪️ "},
		{399, "↪️ "},
		{400, "⚠️ "},
		{401, "⚠️ "},
		{404, "⚠️ "},
		{499, "⚠️ "},
		{500, "💥"},
		{502, "💥"},
		{503, "💥"},
		{100, "ℹ️ "},
		{199, "ℹ️ "},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("Status_%d", tt.statusCode), func(t *testing.T) {
			emoji := getStatusEmoji(tt.statusCode)
			if emoji != tt.emoji {
				t.Errorf("Expected emoji '%s' for status %d, got '%s'", tt.emoji, tt.statusCode, emoji)
			}
		})
	}
}

func BenchmarkRequestIDMiddleware(b *testing.B) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := RequestIDMiddleware(testHandler)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api/test", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}

func BenchmarkLoggingMiddleware(b *testing.B) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("test response"))
	})

	handler := RequestIDMiddleware(LoggingMiddleware(testHandler))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api/test", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}

func BenchmarkResponseWriter_Write(b *testing.B) {
	rec := httptest.NewRecorder()
	rw := &responseWriter{
		ResponseWriter: rec,
		statusCode:     http.StatusOK,
		bytes:          0,
	}

	testData := []byte("test data for benchmarking")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = rw.Write(testData)
	}
}
