package handlers

import (
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const refreshCookieName = "nomdb_refresh"

// cookieSecure decides whether the refresh cookie is marked Secure.
// COOKIE_SECURE overrides; otherwise it follows the request scheme.
func cookieSecure(r *http.Request) bool {
	switch strings.ToLower(os.Getenv("COOKIE_SECURE")) {
	case "true":
		return true
	case "false":
		return false
	}
	return r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
}

// setRefreshCookie stores the refresh token in an httpOnly cookie so it is
// unreachable for JavaScript (XSS). Path-scoped to the auth endpoints and
// SameSite=Strict as the CSRF baseline (plus checkAuthOrigin).
func setRefreshCookie(w http.ResponseWriter, r *http.Request, token string, maxAge time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    token,
		Path:     "/api/auth",
		MaxAge:   int(maxAge.Seconds()),
		HttpOnly: true,
		Secure:   cookieSecure(r),
		SameSite: http.SameSiteStrictMode,
	})
}

// clearRefreshCookie removes the refresh cookie (logout / invalid session).
func clearRefreshCookie(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    "",
		Path:     "/api/auth",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   cookieSecure(r),
		SameSite: http.SameSiteStrictMode,
	})
}

// refreshTokenFromRequest prefers the httpOnly cookie and falls back to the
// request body value (transition support for pre-cookie clients).
func refreshTokenFromRequest(r *http.Request, bodyToken string) (token string, fromCookie bool) {
	if c, err := r.Cookie(refreshCookieName); err == nil && c.Value != "" {
		return c.Value, true
	}
	return bodyToken, false
}

// checkAuthOrigin rejects cookie-authenticated requests whose Origin header
// does not match this host or the configured ALLOWED_ORIGINS (CSRF defense
// in depth on top of SameSite=Strict).
func checkAuthOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true // non-browser clients (curl, tests) send no Origin
	}
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	if strings.EqualFold(u.Host, r.Host) {
		return true
	}
	for _, allowed := range strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",") {
		allowed = strings.TrimSpace(allowed)
		if allowed != "" && strings.EqualFold(allowed, origin) {
			return true
		}
	}
	return false
}
