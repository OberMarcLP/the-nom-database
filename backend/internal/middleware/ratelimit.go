package middleware

import (
	"net/http"
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

// getIPAddress extracts the real IP address from the request
func getIPAddress(r *http.Request) string {
	// Check X-Forwarded-For header (used by proxies/load balancers)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// Take the first IP if multiple are present
		return forwarded
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// StartCleanupTask starts a background goroutine to clean up stale rate limiters
func (i *IPRateLimiter) StartCleanupTask(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			i.CleanupStaleEntries()
		}
	}()
}
