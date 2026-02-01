package middleware

import (
	"compress/gzip"
	"net/http"
	"strings"
)

// compressWrapper wraps the response to capture Content-Type on WriteHeader
// (or on first Write) and only then enable gzip for non-image, non-compressed types.
type compressWrapper struct {
	http.ResponseWriter
	gz         *gzip.Writer
	skip       bool
	wroteHeader bool
}

func (c *compressWrapper) ensureWriteHeader(code int) {
	if c.wroteHeader {
		return
	}
	c.wroteHeader = true
	if c.skip {
		c.ResponseWriter.WriteHeader(code)
		return
	}
	contentType := c.Header().Get("Content-Type")
	if strings.Contains(contentType, "image/") ||
		strings.Contains(contentType, "video/") ||
		strings.Contains(contentType, "application/zip") ||
		strings.Contains(contentType, "application/gzip") {
		c.skip = true
		c.ResponseWriter.WriteHeader(code)
		return
	}
	c.Header().Set("Content-Encoding", "gzip")
	c.Header().Del("Content-Length")
	c.ResponseWriter.WriteHeader(code)
	c.gz = gzip.NewWriter(c.ResponseWriter)
}

func (c *compressWrapper) WriteHeader(code int) {
	c.ensureWriteHeader(code)
}

func (c *compressWrapper) Write(b []byte) (int, error) {
	if !c.wroteHeader {
		c.ensureWriteHeader(http.StatusOK)
	}
	if c.skip {
		return c.ResponseWriter.Write(b)
	}
	return c.gz.Write(b)
}

func (c *compressWrapper) Flush() {
	if c.gz != nil {
		c.gz.Flush()
	}
	if f, ok := c.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// CompressionMiddleware adds gzip compression to responses when client supports it.
// Content-Type is checked after the handler calls WriteHeader, so image/video/zip
// responses are not compressed.
func CompressionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		cw := &compressWrapper{ResponseWriter: w}
		next.ServeHTTP(cw, r)
		if cw.gz != nil {
			_ = cw.gz.Close()
		}
	})
}
