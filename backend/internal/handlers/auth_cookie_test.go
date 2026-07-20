package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHashRefreshToken(t *testing.T) {
	t.Parallel()

	// SHA-256 of the empty string is a well-known constant.
	const emptySHA256 = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	if got := hashRefreshToken(""); got != emptySHA256 {
		t.Errorf("hashRefreshToken(\"\") = %q, want %q", got, emptySHA256)
	}

	if hashRefreshToken("token-a") != hashRefreshToken("token-a") {
		t.Error("hashRefreshToken is not deterministic for equal input")
	}
	if hashRefreshToken("token-a") == hashRefreshToken("token-b") {
		t.Error("hashRefreshToken produced identical digests for different tokens")
	}
	if got := hashRefreshToken("any"); len(got) != 64 {
		t.Errorf("digest length = %d, want 64 hex characters", len(got))
	}
}

// TestCookieSecure uses t.Setenv and therefore must not run in parallel.
func TestCookieSecure(t *testing.T) {
	tests := []struct {
		name           string
		env            string
		useTLS         bool
		forwardedProto string
		want           bool
	}{
		{"env true forces secure", "true", false, "", true},
		{"env TRUE is case-insensitive", "TRUE", false, "", true},
		{"env false overrides tls", "false", true, "https", false},
		{"unset plain http", "", false, "", false},
		{"unset with tls connection", "", true, "", true},
		{"unset with forwarded https", "", false, "https", true},
		{"unset with forwarded HTTPS uppercase", "", false, "HTTPS", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("COOKIE_SECURE", tt.env)

			target := "http://api.example.com/api/auth/refresh"
			if tt.useTLS {
				target = "https://api.example.com/api/auth/refresh"
			}
			r := httptest.NewRequest(http.MethodPost, target, nil)
			if tt.forwardedProto != "" {
				r.Header.Set("X-Forwarded-Proto", tt.forwardedProto)
			}

			if got := cookieSecure(r); got != tt.want {
				t.Errorf("cookieSecure() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestSetRefreshCookie uses t.Setenv and therefore must not run in parallel.
func TestSetRefreshCookie(t *testing.T) {
	t.Setenv("COOKIE_SECURE", "true")

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "http://api.example.com/api/auth/login", nil)

	setRefreshCookie(w, r, "refresh-token-value", 7*24*time.Hour)

	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("got %d cookies, want 1", len(cookies))
	}

	c := cookies[0]
	if c.Name != refreshCookieName {
		t.Errorf("Name = %q, want %q", c.Name, refreshCookieName)
	}
	if c.Value != "refresh-token-value" {
		t.Errorf("Value = %q, want the refresh token", c.Value)
	}
	if c.Path != "/api/auth" {
		t.Errorf("Path = %q, want %q", c.Path, "/api/auth")
	}
	if !c.HttpOnly {
		t.Error("HttpOnly = false, want true")
	}
	if !c.Secure {
		t.Error("Secure = false, want true with COOKIE_SECURE=true")
	}
	if c.SameSite != http.SameSiteStrictMode {
		t.Errorf("SameSite = %v, want %v", c.SameSite, http.SameSiteStrictMode)
	}
	if want := int((7 * 24 * time.Hour).Seconds()); c.MaxAge != want {
		t.Errorf("MaxAge = %d, want %d", c.MaxAge, want)
	}
}

// TestClearRefreshCookie uses t.Setenv and therefore must not run in parallel.
func TestClearRefreshCookie(t *testing.T) {
	t.Setenv("COOKIE_SECURE", "false")

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "http://api.example.com/api/auth/logout", nil)

	clearRefreshCookie(w, r)

	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("got %d cookies, want 1", len(cookies))
	}

	c := cookies[0]
	if c.Name != refreshCookieName {
		t.Errorf("Name = %q, want %q", c.Name, refreshCookieName)
	}
	if c.Value != "" {
		t.Errorf("Value = %q, want empty", c.Value)
	}
	if c.MaxAge >= 0 {
		t.Errorf("MaxAge = %d, want a negative value to delete the cookie", c.MaxAge)
	}
	if !c.HttpOnly {
		t.Error("HttpOnly = false, want true")
	}
}

func TestRefreshTokenFromRequest(t *testing.T) {
	t.Parallel()

	t.Run("cookie wins over body token", func(t *testing.T) {
		t.Parallel()

		r := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
		r.AddCookie(&http.Cookie{Name: refreshCookieName, Value: "cookie-token"})

		token, fromCookie := refreshTokenFromRequest(r, "body-token")
		if token != "cookie-token" {
			t.Errorf("token = %q, want %q", token, "cookie-token")
		}
		if !fromCookie {
			t.Error("fromCookie = false, want true")
		}
	})

	t.Run("falls back to body token without cookie", func(t *testing.T) {
		t.Parallel()

		r := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)

		token, fromCookie := refreshTokenFromRequest(r, "body-token")
		if token != "body-token" {
			t.Errorf("token = %q, want %q", token, "body-token")
		}
		if fromCookie {
			t.Error("fromCookie = true, want false")
		}
	})

	t.Run("unrelated cookie is ignored", func(t *testing.T) {
		t.Parallel()

		r := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
		r.AddCookie(&http.Cookie{Name: "other_cookie", Value: "irrelevant"})

		token, fromCookie := refreshTokenFromRequest(r, "body-token")
		if token != "body-token" || fromCookie {
			t.Errorf("got (%q, %v), want (%q, false)", token, fromCookie, "body-token")
		}
	})

	t.Run("empty cookie value falls back to body token", func(t *testing.T) {
		t.Parallel()

		r := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
		r.AddCookie(&http.Cookie{Name: refreshCookieName, Value: ""})

		token, fromCookie := refreshTokenFromRequest(r, "body-token")
		if token != "body-token" || fromCookie {
			t.Errorf("got (%q, %v), want (%q, false)", token, fromCookie, "body-token")
		}
	})
}

// TestCheckAuthOrigin uses t.Setenv and therefore must not run in parallel.
func TestCheckAuthOrigin(t *testing.T) {
	tests := []struct {
		name       string
		origin     string
		allowedEnv string
		want       bool
	}{
		{"no origin header allows non-browser clients", "", "", true},
		{"same host", "http://api.example.com", "", true},
		{"same host case-insensitive", "http://API.EXAMPLE.COM", "", true},
		{"foreign origin without allowlist", "http://evil.example", "", false},
		{"foreign origin in allowlist", "http://app.example.com", "http://app.example.com", true},
		{"allowlist with spaces and multiple entries", "http://b.example", "http://a.example, http://b.example", true},
		{"allowlist case-insensitive", "HTTP://APP.EXAMPLE.COM", "http://app.example.com", true},
		{"origin not in allowlist", "http://c.example", "http://a.example,http://b.example", false},
		{"malformed origin", "://missing-scheme", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("ALLOWED_ORIGINS", tt.allowedEnv)

			r := httptest.NewRequest(http.MethodPost, "http://api.example.com/api/auth/refresh", nil)
			if tt.origin != "" {
				r.Header.Set("Origin", tt.origin)
			}

			if got := checkAuthOrigin(r); got != tt.want {
				t.Errorf("checkAuthOrigin(origin=%q, allowed=%q) = %v, want %v", tt.origin, tt.allowedEnv, got, tt.want)
			}
		})
	}
}

func TestRequestContext(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/api/restaurants", nil)

	ctx, cancel := RequestContext(r)
	defer cancel()

	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("context has no deadline, want one derived from RequestTimeout")
	}
	remaining := time.Until(deadline)
	if remaining <= 0 || remaining > RequestTimeout {
		t.Errorf("deadline in %v, want within (0, %v]", remaining, RequestTimeout)
	}

	cancel()
	select {
	case <-ctx.Done():
		// expected: cancel closes the context
	default:
		t.Error("context is not done after cancel()")
	}
}
