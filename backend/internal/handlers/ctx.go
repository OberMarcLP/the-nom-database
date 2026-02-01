package handlers

import (
	"context"
	"net/http"
	"time"
)

// RequestTimeout is the default timeout for handler DB operations.
const RequestTimeout = 30 * time.Second

// RequestContext returns a context derived from the request with RequestTimeout.
// Callers must call the returned cancel function when done.
func RequestContext(r *http.Request) (context.Context, context.CancelFunc) {
	return context.WithTimeout(r.Context(), RequestTimeout)
}
