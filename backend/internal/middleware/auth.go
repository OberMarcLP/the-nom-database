package middleware

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/nomdb/backend/internal/auth"
	"github.com/nomdb/backend/internal/database"
	"github.com/nomdb/backend/internal/logger"
	"github.com/nomdb/backend/internal/models"
)

// AuthMode defines the authentication mode
type AuthMode string

const (
	AuthModeNone  AuthMode = "none"  // No authentication (testing only)
	AuthModeLocal AuthMode = "local" // Local JWT authentication
	AuthModeOAuth AuthMode = "oauth" // OAuth only
	AuthModeBoth  AuthMode = "both"  // Both local and OAuth
)

var (
	currentAuthMode AuthMode
	jwtService      *auth.JWTService
)

// InitAuthMiddleware initializes the authentication middleware
func InitAuthMiddleware(jwtSvc *auth.JWTService) {
	if jwtSvc != nil {
		jwtService = jwtSvc
	}

	// Get auth mode from environment
	mode := os.Getenv("AUTH_MODE")
	switch mode {
	case "none":
		currentAuthMode = AuthModeNone
		logger.Warn("⚠️  Authentication DISABLED - only use for testing!")
	case "local":
		currentAuthMode = AuthModeLocal
		logger.Info("🔐 Authentication mode: Local (JWT)")
	case "oauth":
		currentAuthMode = AuthModeOAuth
		logger.Info("🔐 Authentication mode: OAuth only")
	case "both", "":
		currentAuthMode = AuthModeBoth
		logger.Info("🔐 Authentication mode: Both (Local + OAuth)")
	default:
		logger.Warn("Unknown AUTH_MODE '%s', defaulting to 'both'", mode)
		currentAuthMode = AuthModeBoth
	}
}

// SetJWTService sets the JWT service (called after initialization)
func SetJWTService(jwtSvc *auth.JWTService) {
	jwtService = jwtSvc
}

// GetAuthMode returns the current authentication mode
func GetAuthMode() AuthMode {
	return currentAuthMode
}

// AuthMiddleware validates JWT tokens and adds user to context
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth if mode is 'none'
		if currentAuthMode == AuthModeNone {
			// Create a dummy user for testing
			dummyUser := &models.User{
				ID:       1,
				Email:    "test@example.com",
				Username: "testuser",
				IsAdmin:  true,
				IsActive: true,
			}
			ctx := context.WithValue(r.Context(), models.UserContextKey, dummyUser)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing authorization header", http.StatusUnauthorized)
			return
		}

		// Check for Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// Validate token
		claims, err := jwtService.ValidateAccessToken(tokenString)
		if err != nil {
			if err == auth.ErrExpiredToken {
				http.Error(w, "Token has expired", http.StatusUnauthorized)
			} else {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
			}
			return
		}

		// Fetch user from database to ensure they still exist and are active
		ctx := r.Context()
		var user models.User
		err = database.GetPool().QueryRow(ctx,
			`SELECT id, email, username, password_hash, provider, provider_id, full_name, avatar_url,
			is_active, is_admin, email_verified, password_must_change, last_login_at, created_at, updated_at
			FROM users WHERE id = $1`, claims.UserID).Scan(
			&user.ID, &user.Email, &user.Username, &user.PasswordHash, &user.Provider, &user.ProviderID,
			&user.FullName, &user.AvatarURL, &user.IsActive, &user.IsAdmin,
			&user.EmailVerified, &user.PasswordMustChange, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt)

		if err != nil {
			logger.Warn("User not found for token: %d", claims.UserID)
			http.Error(w, "User not found", http.StatusUnauthorized)
			return
		}

		// Check if user is active
		if !user.IsActive {
			http.Error(w, "Account is disabled", http.StatusUnauthorized)
			return
		}

		// Add user to context
		ctx = context.WithValue(ctx, models.UserContextKey, &user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuthMiddleware validates JWT tokens if present but allows anonymous access
func OptionalAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth if mode is 'none'
		if currentAuthMode == AuthModeNone {
			dummyUser := &models.User{
				ID:       1,
				Email:    "test@example.com",
				Username: "testuser",
				IsAdmin:  true,
				IsActive: true,
			}
			ctx := context.WithValue(r.Context(), models.UserContextKey, dummyUser)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// Try to extract token
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// No token, continue as anonymous
			next.ServeHTTP(w, r)
			return
		}

		// Parse token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			// Invalid format, continue as anonymous
			next.ServeHTTP(w, r)
			return
		}

		tokenString := parts[1]
		claims, err := jwtService.ValidateAccessToken(tokenString)
		if err != nil {
			// Invalid token, continue as anonymous
			next.ServeHTTP(w, r)
			return
		}

		// Fetch user
		ctx := r.Context()
		var user models.User
		err = database.GetPool().QueryRow(ctx,
			`SELECT id, email, username, password_hash, provider, provider_id, full_name, avatar_url,
			is_active, is_admin, email_verified, password_must_change, last_login_at, created_at, updated_at
			FROM users WHERE id = $1`, claims.UserID).Scan(
			&user.ID, &user.Email, &user.Username, &user.PasswordHash, &user.Provider, &user.ProviderID,
			&user.FullName, &user.AvatarURL, &user.IsActive, &user.IsAdmin,
			&user.EmailVerified, &user.PasswordMustChange, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt)

		if err != nil || !user.IsActive {
			// User not found or inactive, continue as anonymous
			next.ServeHTTP(w, r)
			return
		}

		// Add user to context
		ctx = context.WithValue(ctx, models.UserContextKey, &user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AdminOnlyMiddleware ensures the user is an admin
func AdminOnlyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(models.UserContextKey).(*models.User)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if !user.IsAdmin {
			http.Error(w, "Forbidden - admin access required", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// GetUserFromRequest extracts the user from the request context
func GetUserFromRequest(r *http.Request) (*models.User, bool) {
	user, ok := r.Context().Value(models.UserContextKey).(*models.User)
	return user, ok
}
