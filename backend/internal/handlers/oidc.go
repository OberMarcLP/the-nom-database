package handlers

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/nomdb/backend/internal/database"
	"github.com/nomdb/backend/internal/logger"
	"github.com/nomdb/backend/internal/models"
	"golang.org/x/oauth2"
)

var (
	oidcProvider *oidc.Provider
	oidcConfig   *oauth2.Config
	oidcVerifier *oidc.IDTokenVerifier
	oidcStateStore = make(map[string]time.Time) // In production, use Redis
)

// InitOIDC initializes OIDC provider (Authentik or any OIDC-compliant provider)
func InitOIDC() error {
	issuerURL := os.Getenv("OIDC_ISSUER_URL")
	clientID := os.Getenv("OIDC_CLIENT_ID")
	clientSecret := os.Getenv("OIDC_CLIENT_SECRET")
	redirectURL := os.Getenv("OIDC_REDIRECT_URL")

	if issuerURL == "" || clientID == "" || clientSecret == "" {
		logger.Warn("OIDC not configured - OIDC login will not be available")
		logger.Debug("Required: OIDC_ISSUER_URL, OIDC_CLIENT_ID, OIDC_CLIENT_SECRET")
		return nil
	}

	if redirectURL == "" {
		redirectURL = "http://localhost:8080/api/auth/oidc/callback"
	}

	// Initialize OIDC provider
	ctx := context.Background()
	provider, err := oidc.NewProvider(ctx, issuerURL)
	if err != nil {
		return fmt.Errorf("failed to initialize OIDC provider: %w", err)
	}

	oidcProvider = provider
	oidcVerifier = provider.Verifier(&oidc.Config{ClientID: clientID})

	// Configure OAuth2
	oidcConfig = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	logger.Info("🔐 OIDC configured with issuer: %s", issuerURL)
	logger.Debug("OIDC redirect URL: %s", redirectURL)

	// Start cleanup task for state store
	go cleanupOIDCStates()

	return nil
}

// @Summary OIDC login
// @Description Initiate OIDC/Authentik login flow
// @Tags Auth
// @Produce json
// @Success 302 {string} string "Redirect to OIDC provider"
// @Router /auth/oidc/login [get]
func OIDCLogin(w http.ResponseWriter, r *http.Request) {
	if oidcConfig == nil {
		http.Error(w, "OIDC not configured", http.StatusServiceUnavailable)
		return
	}

	// Generate random state
	state, err := generateOIDCState()
	if err != nil {
		logger.Error("Failed to generate OIDC state: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Store state with expiry
	oidcStateStore[state] = time.Now().Add(10 * time.Minute)

	// Redirect to OIDC provider
	url := oidcConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// @Summary OIDC callback
// @Description Handle OIDC/Authentik callback
// @Tags Auth
// @Produce json
// @Param code query string true "Authorization code"
// @Param state query string true "OIDC state"
// @Success 200 {object} models.LoginResponse
// @Failure 400 {string} string "Invalid request"
// @Failure 401 {string} string "Invalid state or code"
// @Failure 500 {string} string "Internal server error"
// @Router /auth/oidc/callback [get]
func OIDCCallback(w http.ResponseWriter, r *http.Request) {
	if oidcConfig == nil || oidcVerifier == nil {
		http.Error(w, "OIDC not configured", http.StatusServiceUnavailable)
		return
	}

	// Verify state
	state := r.URL.Query().Get("state")
	expiry, exists := oidcStateStore[state]
	if !exists || time.Now().After(expiry) {
		http.Error(w, "Invalid or expired state", http.StatusBadRequest)
		return
	}
	delete(oidcStateStore, state)

	// Get authorization code
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		return
	}

	// Exchange code for token
	ctx := context.Background()
	oauth2Token, err := oidcConfig.Exchange(ctx, code)
	if err != nil {
		logger.Error("Failed to exchange code: %v", err)
		http.Error(w, "Failed to exchange authorization code", http.StatusInternalServerError)
		return
	}

	// Extract ID Token
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "No id_token in response", http.StatusInternalServerError)
		return
	}

	// Verify ID Token
	idToken, err := oidcVerifier.Verify(ctx, rawIDToken)
	if err != nil {
		logger.Error("Failed to verify ID token: %v", err)
		http.Error(w, "Failed to verify ID token", http.StatusUnauthorized)
		return
	}

	// Extract claims
	var claims struct {
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
		PreferredUsername string `json:"preferred_username"`
		Picture       string `json:"picture"`
		Sub           string `json:"sub"` // Subject - unique user ID
	}
	if err := idToken.Claims(&claims); err != nil {
		logger.Error("Failed to parse claims: %v", err)
		http.Error(w, "Failed to parse user information", http.StatusInternalServerError)
		return
	}

	// Validate required claims
	if claims.Email == "" || claims.Sub == "" {
		http.Error(w, "Missing required user information (email or sub)", http.StatusBadRequest)
		return
	}

	// Find or create user
	user, err := findOrCreateOIDCUser(ctx, &claims)
	if err != nil {
		logger.Error("Failed to find/create user: %v", err)
		http.Error(w, "Failed to process user", http.StatusInternalServerError)
		return
	}

	// Update last login
	_, err = database.GetPool().Exec(ctx, "UPDATE users SET last_login_at = $1 WHERE id = $2", time.Now(), user.ID)
	if err != nil {
		logger.Warn("Failed to update last login: %v", err)
	}

	// Generate tokens
	jwtService := getJWTService()
	if jwtService == nil {
		http.Error(w, "Authentication service not available", http.StatusInternalServerError)
		return
	}
	response, err := generateLoginResponseWithService(ctx, user, r, jwtService)
	if err != nil {
		logger.Error("Failed to generate tokens: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	logger.Info("User logged in via OIDC: %s (ID: %d)", user.Email, user.ID)

	// In a real app, you might redirect to frontend with tokens in URL params or cookies
	// For now, return JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Helper functions

type OIDCClaims struct {
	Email             string
	EmailVerified     bool
	Name              string
	PreferredUsername string
	Picture           string
	Sub               string
}

func findOrCreateOIDCUser(ctx context.Context, claims *struct {
	Email             string `json:"email"`
	EmailVerified     bool   `json:"email_verified"`
	Name              string `json:"name"`
	PreferredUsername string `json:"preferred_username"`
	Picture           string `json:"picture"`
	Sub               string `json:"sub"`
}) (*models.User, error) {
	provider := "oidc"

	// Try to find existing user by provider ID
	var user models.User
	err := database.GetPool().QueryRow(ctx,
		`SELECT id, email, username, provider, provider_id, full_name, avatar_url,
		is_active, is_admin, email_verified, last_login_at, created_at, updated_at
		FROM users WHERE provider = $1 AND provider_id = $2`,
		provider, claims.Sub).Scan(
		&user.ID, &user.Email, &user.Username, &user.Provider, &user.ProviderID,
		&user.FullName, &user.AvatarURL, &user.IsActive, &user.IsAdmin,
		&user.EmailVerified, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt)

	if err == nil {
		// User exists, update info if changed
		if user.Email != claims.Email || (user.FullName != "" && user.FullName != claims.Name) {
			_, err = database.GetPool().Exec(ctx,
				`UPDATE users SET email = $1, full_name = $2, avatar_url = $3, email_verified = $4
				WHERE id = $5`,
				claims.Email, claims.Name, claims.Picture, claims.EmailVerified, user.ID)
			if err != nil {
				logger.Warn("Failed to update user info: %v", err)
			}
		}
		return &user, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	// User doesn't exist, create new one
	// Generate username from email or preferred_username
	username := claims.PreferredUsername
	if username == "" {
		username = generateUsernameFromEmail(claims.Email)
	}

	var userID int
	err = database.GetPool().QueryRow(ctx,
		`INSERT INTO users (email, username, provider, provider_id, full_name, avatar_url, email_verified, password_hash, password_must_change)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NULL, false)
		RETURNING id`,
		claims.Email, username, provider, claims.Sub, claims.Name, claims.Picture, claims.EmailVerified).Scan(&userID)

	if err != nil {
		// Check if email already exists (user might have registered locally)
		err2 := database.GetPool().QueryRow(ctx,
			`SELECT id FROM users WHERE email = $1`, claims.Email).Scan(&userID)
		if err2 == nil {
			// Link OIDC to existing account
			_, err = database.GetPool().Exec(ctx,
				`UPDATE users SET provider = $1, provider_id = $2, avatar_url = $3
				WHERE id = $4`,
				provider, claims.Sub, claims.Picture, userID)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		// New user created - assign default 'user' role
		_, err = database.GetPool().Exec(ctx,
			`INSERT INTO user_roles (user_id, role_id)
			SELECT $1, id FROM roles WHERE name = 'user'
			ON CONFLICT DO NOTHING`,
			userID)
		if err != nil {
			logger.Warn("Failed to assign default role to OAuth user %d: %v", userID, err)
			// Continue - this is not a critical failure
		}
	}

	// Fetch created user
	return getUserByID(ctx, userID)
}

func generateOIDCState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func generateUsernameFromEmail(email string) string {
	// Extract part before @
	parts := strings.Split(email, "@")
	if len(parts) > 0 {
		return parts[0]
	}
	return "user"
}

func cleanupOIDCStates() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		for state, expiry := range oidcStateStore {
			if now.After(expiry) {
				delete(oidcStateStore, state)
			}
		}
	}
}
