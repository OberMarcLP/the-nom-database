package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nomdb/backend/internal/models"
)

// The Register and Login tests below only exercise input validation, which
// runs before password hashing and before any database access.

func TestRegister_ValidationErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		body    string
		wantMsg string
	}{
		{"malformed json", `{"email":`, "Invalid request body"},
		{"empty body", ``, "Invalid request body"},
		{"missing all fields", `{}`, "Email, username, and password are required"},
		{"missing email", `{"username":"nomuser","password":"longenough1"}`, "Email, username, and password are required"},
		{"missing username", `{"email":"nom@example.com","password":"longenough1"}`, "Email, username, and password are required"},
		{"missing password", `{"email":"nom@example.com","username":"nomuser"}`, "Email, username, and password are required"},
		{"email without at sign", `{"email":"nomexample.com","username":"nomuser","password":"longenough1"}`, "Invalid email format"},
		{"email without dot", `{"email":"nom@examplecom","username":"nomuser","password":"longenough1"}`, "Invalid email format"},
		{"password too short", `{"email":"nom@example.com","username":"nomuser","password":"1234567"}`, "Password must be at least 8 characters"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rr := httptest.NewRecorder()
			r := newJSONRequestWithVars(http.MethodPost, "/api/auth/register", tt.body, nil)

			Register(rr, r)

			if rr.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
			}
			if !strings.Contains(rr.Body.String(), tt.wantMsg) {
				t.Errorf("body = %q, want it to contain %q", rr.Body.String(), tt.wantMsg)
			}
		})
	}
}

func TestLogin_ValidationErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		body    string
		wantMsg string
	}{
		{"malformed json", `{`, "Invalid request body"},
		{"empty body", ``, "Invalid request body"},
		{"missing both fields", `{}`, "Email and password are required"},
		{"missing password", `{"email":"nom@example.com"}`, "Email and password are required"},
		{"missing email", `{"password":"secret"}`, "Email and password are required"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rr := httptest.NewRecorder()
			r := newJSONRequestWithVars(http.MethodPost, "/api/auth/login", tt.body, nil)

			Login(rr, r)

			if rr.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
			}
			if !strings.Contains(rr.Body.String(), tt.wantMsg) {
				t.Errorf("body = %q, want it to contain %q", rr.Body.String(), tt.wantMsg)
			}
		})
	}
}

func TestIsValidEmail(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		email string
		want  bool
	}{
		{"typical address", "user@example.com", true},
		{"short but complete", "a@b.c", true},
		{"empty string", "", false},
		{"no at sign", "userexample.com", false},
		{"no dot", "user@examplecom", false},
		{"below minimum length", "a@b", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := isValidEmail(tt.email); got != tt.want {
				t.Errorf("isValidEmail(%q) = %v, want %v", tt.email, got, tt.want)
			}
		})
	}
}

func TestIsDuplicateKeyError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"postgres duplicate key message", errors.New(`ERROR: duplicate key value violates unique constraint "users_email_key"`), true},
		{"unique constraint message", errors.New("violates unique constraint"), true},
		{"unrelated error", errors.New("connection refused"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := isDuplicateKeyError(tt.err); got != tt.want {
				t.Errorf("isDuplicateKeyError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestGetUserFromContext(t *testing.T) {
	t.Parallel()

	t.Run("no user in context", func(t *testing.T) {
		t.Parallel()

		r := httptest.NewRequest(http.MethodGet, "/api/whoami", nil)
		user, ok := GetUserFromContext(r)
		if ok {
			t.Error("ok = true, want false when no user is set")
		}
		if user != nil {
			t.Errorf("user = %+v, want nil", user)
		}
	})

	t.Run("user in context", func(t *testing.T) {
		t.Parallel()

		want := &models.User{ID: 7, Username: "nomuser"}
		r := withUser(httptest.NewRequest(http.MethodGet, "/api/whoami", nil), want)

		user, ok := GetUserFromContext(r)
		if !ok {
			t.Fatal("ok = false, want true when a user is set")
		}
		if user != want {
			t.Errorf("user = %+v, want the exact instance stored in the context", user)
		}
	})
}

func TestGetUserIDFromPath(t *testing.T) {
	t.Parallel()

	t.Run("missing id", func(t *testing.T) {
		t.Parallel()

		r := httptest.NewRequest(http.MethodGet, "/api/users", nil)
		if _, err := GetUserIDFromPath(r); err == nil {
			t.Error("error = nil, want an error for a missing id")
		}
	})

	t.Run("non-numeric id", func(t *testing.T) {
		t.Parallel()

		r := newRequestWithVars(http.MethodGet, "/api/users/abc", nil, map[string]string{"id": "abc"})
		if _, err := GetUserIDFromPath(r); err == nil {
			t.Error("error = nil, want an error for a non-numeric id")
		}
	})

	t.Run("valid id", func(t *testing.T) {
		t.Parallel()

		r := newRequestWithVars(http.MethodGet, "/api/users/42", nil, map[string]string{"id": "42"})
		id, err := GetUserIDFromPath(r)
		if err != nil {
			t.Fatalf("error = %v, want nil", err)
		}
		if id != 42 {
			t.Errorf("id = %d, want 42", id)
		}
	})
}
