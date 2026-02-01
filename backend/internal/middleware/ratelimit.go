package middleware

import (
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

const staleLimiterAge = 10 * time.Minute // Remove limiters not used in this long (per-key lazy expiration)

// IPRateLimiter manages rate limiters for different IP addresses
type IPRateLimiter struct {
	ips      map[string]*rate.Limiter
	lastUsed map[string]time.Time
	mu       *sync.RWMutex
	r        rate.Limit
	b        int
	stopCh   chan struct{} // Channel to stop cleanup goroutine
}

// NewIPRateLimiter creates a new IP-based rate limiter
// r: requests per second
// b: burst size (max requests in short burst)
func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	return &IPRateLimiter{
		ips:      make(map[string]*rate.Limiter),
		lastUsed: make(map[string]time.Time),
		mu:       &sync.RWMutex{},
		r:        r,
		b:        b,
		stopCh:   make(chan struct{}),
	}
}

// GetLimiter returns the rate limiter for the specified IP address.
// Per-key lazy expiration: entries not used in staleLimiterAge are removed when next seen.
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	now := time.Now()
	if last, ok := i.lastUsed[ip]; ok && now.Sub(last) > staleLimiterAge {
		delete(i.ips, ip)
		delete(i.lastUsed, ip)
	}

	limiter, exists := i.ips[ip]
	if !exists {
		limiter = rate.NewLimiter(i.r, i.b)
		i.ips[ip] = limiter
	}
	i.lastUsed[ip] = now

	return limiter
}

// CleanupStaleEntries removes inactive rate limiters (run periodically)
func (i *IPRateLimiter) CleanupStaleEntries() {
	i.mu.Lock()
	defer i.mu.Unlock()

	now := time.Now()
	for ip, last := range i.lastUsed {
		if now.Sub(last) > staleLimiterAge {
			delete(i.ips, ip)
			delete(i.lastUsed, ip)
		}
	}
}

// RateLimitMiddleware creates a rate limiting middleware
// Example: 100 requests per minute = rate.Every(time.Minute/100), burst: 20
func RateLimitMiddleware(limiter *IPRateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get IP address (handle X-Forwarded-For and X-Real-IP headers)
			ip := getIPAddress(r)

			// Get rate limiter for this IP
			ipLimiter := limiter.GetLimiter(ip)

			// Check if request is allowed
			if !ipLimiter.Allow() {
				http.Error(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// trustProxyHeaders indicates whether to trust X-Forwarded-For headers
// Set TRUST_PROXY=true only when running behind a trusted reverse proxy
var trustProxyHeaders = os.Getenv("TRUST_PROXY") == "true"

// getIPAddress extracts the real IP address from the request
// Only trusts proxy headers if TRUST_PROXY=true
func getIPAddress(r *http.Request) string {
	// Only trust proxy headers if explicitly enabled
	if trustProxyHeaders {
		// Check X-Forwarded-For header (format: "client, proxy1, proxy2, ...")
		// Take the leftmost IP as the client IP
		forwarded := r.Header.Get("X-Forwarded-For")
		if forwarded != "" {
			// Split by comma and get the first (leftmost) IP
			ips := strings.Split(forwarded, ",")
			if len(ips) > 0 {
				clientIP := strings.TrimSpace(ips[0])
				// Validate it's a proper IP
				if isValidIP(clientIP) {
					return clientIP
				}
			}
		}

		// Check X-Real-IP header
		realIP := r.Header.Get("X-Real-IP")
		if realIP != "" && isValidIP(realIP) {
			return realIP
		}
	}

	// Fall back to RemoteAddr (strip port if present)
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// RemoteAddr doesn't have a port
		return r.RemoteAddr
	}
	return ip
}

// isValidIP checks if the string is a valid IP address
func isValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

// StartCleanupTask starts a background goroutine to clean up stale rate limiters
func (i *IPRateLimiter) StartCleanupTask(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				i.CleanupStaleEntries()
			case <-i.stopCh:
				return
			}
		}
	}()
}

// Stop stops the cleanup goroutine gracefully
func (i *IPRateLimiter) Stop() {
	close(i.stopCh)
}
