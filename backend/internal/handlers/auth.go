package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/nomdb/backend/internal/auth"
	"github.com/nomdb/backend/internal/database"
	"github.com/nomdb/backend/internal/logger"
	"github.com/nomdb/backend/internal/models"
	"github.com/nomdb/backend/internal/services"
)

// InitAuthService initializes the JWT service and returns it
func InitAuthService() *auth.JWTService {
	secretKey := os.Getenv("JWT_SECRET_KEY")
	if secretKey == "" {
		authMode := os.Getenv("AUTH_MODE")
		if authMode == "local" || authMode == "both" || authMode == "" {
			logger.Fatal("JWT_SECRET_KEY environment variable is required for local/both auth mode")
		}
		logger.Warn("JWT_SECRET_KEY not set - JWT authentication unavailable")
		return nil
	}

	accessTokenDuration := 15 * time.Minute
	refreshTokenDuration := 7 * 24 * time.Hour

	jwtService := auth.NewJWTService(secretKey, accessTokenDuration, refreshTokenDuration)
	auth.SetGlobalJWTService(jwtService)
	logger.Info("🔐 JWT service initialized (access: %v, refresh: %v)", accessTokenDuration, refreshTokenDuration)
	return jwtService
}

func getJWTService() *auth.JWTService {
	return auth.GetGlobalJWTService()
}

// @Summary Register a new user
// @Description Create a new user account with email and password
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body models.RegisterRequest true "Registration details"
// @Success 201 {object} models.LoginResponse
// @Failure 400 {string} string "Invalid request"
// @Failure 409 {string} string "User already exists"
// @Failure 500 {string} string "Internal server error"
// @Router /auth/register [post]
func Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Email == "" || req.Username == "" || req.Password == "" {
		http.Error(w, "Email, username, and password are required", http.StatusBadRequest)
		return
	}

	// Basic email validation
	if !isValidEmail(req.Email) {
		http.Error(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	// Password strength check
	if len(req.Password) < 8 {
		http.Error(w, "Password must be at least 8 characters", http.StatusBadRequest)
		return
	}

	// Hash password
	passwordHash, err := auth.HashPassword(req.Password, nil)
	if err != nil {
		logger.Error("Failed to hash password: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Generate Gravatar URL
	gravatarService := services.NewGravatarService()
	avatarURL := gravatarService.GetAvatarURL(req.Email, 256)

	// Create user
	ctx, cancel := RequestContext(r)
	defer cancel()
	var userID int
	err = database.GetPool().QueryRow(ctx,
		`INSERT INTO users (email, username, password_hash, provider, full_name, avatar_url, email_verified, password_must_change)
		VALUES ($1, $2, $3, 'local', $4, $5, false, false)
		RETURNING id`,
		req.Email, req.Username, passwordHash, req.FullName, avatarURL).Scan(&userID)

	if err != nil {
		if isDuplicateKeyError(err) {
			http.Error(w, "User with this email or username already exists", http.StatusConflict)
			return
		}
		logger.Error("Failed to create user: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Assign default 'user' role to new user
	_, err = database.GetPool().Exec(ctx,
		`INSERT INTO user_roles (user_id, role_id)
		SELECT $1, id FROM roles WHERE name = 'user'
		ON CONFLICT DO NOTHING`,
		userID)
	if err != nil {
		logger.Warn("Failed to assign default role to user %d: %v", userID, err)
		// Continue - this is not a critical failure
	}

	// Fetch created user
	user, err := getUserByID(ctx, userID)
	if err != nil {
		logger.Error("Failed to fetch created user: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
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

	logger.Info("User registered: %s (ID: %d)", user.Email, user.ID)

	// Refresh token travels only via httpOnly cookie, never in the body
	setRefreshCookie(w, r, response.RefreshToken, jwtService.GetRefreshTokenDuration())
	response.RefreshToken = ""

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Failed to encode response: %v", err)
	}
}

// @Summary Login
// @Description Login with email and password
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "Login credentials"
// @Success 200 {object} models.LoginResponse
// @Failure 400 {string} string "Invalid request"
// @Failure 401 {string} string "Invalid credentials"
// @Failure 500 {string} string "Internal server error"
// @Router /auth/login [post]
func Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	ctx, cancel := RequestContext(r)
	defer cancel()

	// Fetch user
	user, err := getUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		logger.Error("Failed to fetch user: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Check if user is active
	if !user.IsActive {
		http.Error(w, "Account is disabled", http.StatusUnauthorized)
		return
	}

	// Verify password
	if user.PasswordHash == nil {
		http.Error(w, "Invalid credentials - OAuth user", http.StatusUnauthorized)
		return
	}

	valid, err := auth.VerifyPassword(req.Password, *user.PasswordHash)
	if err != nil {
		logger.Error("Failed to verify password: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if !valid {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
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

	logger.Info("User logged in: %s (ID: %d)", user.Email, user.ID)

	// Refresh token travels only via httpOnly cookie, never in the body
	setRefreshCookie(w, r, response.RefreshToken, jwtService.GetRefreshTokenDuration())
	response.RefreshToken = ""

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Failed to encode response: %v", err)
	}
}

// @Summary Refresh token
// @Description Get a new access token using a refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body models.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} models.LoginResponse
// @Failure 400 {string} string "Invalid request"
// @Failure 401 {string} string "Invalid or expired refresh token"
// @Failure 500 {string} string "Internal server error"
// @Router /auth/refresh [post]
func RefreshToken(w http.ResponseWriter, r *http.Request) {
	// Body is optional - the token normally arrives via httpOnly cookie
	var req models.RefreshTokenRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	refreshToken, fromCookie := refreshTokenFromRequest(r, req.RefreshToken)
	if refreshToken == "" {
		http.Error(w, "Refresh token is required", http.StatusBadRequest)
		return
	}
	if fromCookie && !checkAuthOrigin(r) {
		http.Error(w, "Origin not allowed", http.StatusForbidden)
		return
	}

	ctx, cancel := RequestContext(r)
	defer cancel()

	// Fetch session
	var session models.Session
	var userID int
	err := database.GetPool().QueryRow(ctx,
		`SELECT id, user_id, expires_at FROM sessions WHERE refresh_token = $1`,
		hashRefreshToken(refreshToken)).Scan(&session.ID, &userID, &session.ExpiresAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			clearRefreshCookie(w, r)
			http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
			return
		}
		logger.Error("Failed to fetch session: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Check if expired
	if session.ExpiresAt.Before(time.Now()) {
		// Delete expired session
		if _, err := database.GetPool().Exec(ctx, "DELETE FROM sessions WHERE id = $1", session.ID); err != nil {
			logger.Warn("Failed to delete expired session: %v", err)
		}
		clearRefreshCookie(w, r)
		http.Error(w, "Refresh token expired", http.StatusUnauthorized)
		return
	}

	// Update last used
	_, err = database.GetPool().Exec(ctx,
		"UPDATE sessions SET last_used_at = $1 WHERE id = $2",
		time.Now(), session.ID)
	if err != nil {
		logger.Warn("Failed to update session last_used_at: %v", err)
	}

	// Fetch user
	user, err := getUserByID(ctx, userID)
	if err != nil {
		logger.Error("Failed to fetch user: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Check if user is active
	if !user.IsActive {
		http.Error(w, "Account is disabled", http.StatusUnauthorized)
		return
	}

	// Generate new access token and rotate the refresh token
	jwtService := getJWTService()
	if jwtService == nil {
		http.Error(w, "Authentication service not available", http.StatusInternalServerError)
		return
	}
	accessToken, err := jwtService.GenerateAccessToken(user)
	if err != nil {
		logger.Error("Failed to generate access token: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Rotation: a stolen refresh token becomes worthless after its first
	// legitimate use; the absolute session expiry stays unchanged.
	newRefreshToken, err := jwtService.GenerateRefreshToken()
	if err != nil {
		logger.Error("Failed to generate refresh token: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if _, err := database.GetPool().Exec(ctx,
		"UPDATE sessions SET refresh_token = $1 WHERE id = $2",
		hashRefreshToken(newRefreshToken), session.ID); err != nil {
		logger.Error("Failed to rotate refresh token: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Rotate the cookie alongside the session; the body stays token-free
	setRefreshCookie(w, r, newRefreshToken, jwtService.GetRefreshTokenDuration())

	response := models.LoginResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int(jwtService.GetAccessTokenDuration().Seconds()),
		User:        *user,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Failed to encode response: %v", err)
	}
}

// @Summary Logout
// @Description Invalidate refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body models.RefreshTokenRequest true "Refresh token to invalidate"
// @Success 200 {string} string "Logged out successfully"
// @Failure 400 {string} string "Invalid request"
// @Router /auth/logout [post]
func Logout(w http.ResponseWriter, r *http.Request) {
	// Body is optional - the token normally arrives via httpOnly cookie
	var req models.RefreshTokenRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	refreshToken, fromCookie := refreshTokenFromRequest(r, req.RefreshToken)
	if fromCookie && !checkAuthOrigin(r) {
		http.Error(w, "Origin not allowed", http.StatusForbidden)
		return
	}
	if refreshToken != "" {
		ctx, cancel := RequestContext(r)
		defer cancel()
		_, err := database.GetPool().Exec(ctx, "DELETE FROM sessions WHERE refresh_token = $1", hashRefreshToken(refreshToken))
		if err != nil {
			logger.Warn("Failed to delete session: %v", err)
		}
	}
	clearRefreshCookie(w, r)

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Get current user
// @Description Get the currently authenticated user's information
// @Tags Auth
// @Produce json
// @Success 200 {object} models.User
// @Failure 401 {string} string "Unauthorized"
// @Security BearerAuth
// @Router /auth/me [get]
func GetMe(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(models.UserContextKey).(*models.User)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(user); err != nil {
		logger.Error("Failed to encode response: %v", err)
	}
}

// Helper functions

func getUserByID(ctx context.Context, userID int) (*models.User, error) {
	var user models.User
	err := database.GetPool().QueryRow(ctx,
		`SELECT id, email, username, password_hash, provider, provider_id, full_name, avatar_url,
		is_active, is_admin, email_verified, password_must_change, last_login_at, created_at, updated_at
		FROM users WHERE id = $1`, userID).Scan(
		&user.ID, &user.Email, &user.Username, &user.PasswordHash, &user.Provider, &user.ProviderID,
		&user.FullName, &user.AvatarURL, &user.IsActive, &user.IsAdmin, &user.EmailVerified,
		&user.PasswordMustChange, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// getUserByEmail resolves a login identifier - an email address or a
// username (only used by the login flow).
func getUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := database.GetPool().QueryRow(ctx,
		`SELECT id, email, username, password_hash, provider, provider_id, full_name, avatar_url,
		is_active, is_admin, email_verified, password_must_change, last_login_at, created_at, updated_at
		FROM users WHERE email = $1 OR username = $1`, email).Scan(
		&user.ID, &user.Email, &user.Username, &user.PasswordHash, &user.Provider, &user.ProviderID,
		&user.FullName, &user.AvatarURL, &user.IsActive, &user.IsAdmin, &user.EmailVerified,
		&user.PasswordMustChange, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func generateLoginResponseWithService(ctx context.Context, user *models.User, r *http.Request, jwtSvc *auth.JWTService) (*models.LoginResponse, error) {
	// Generate access token
	accessToken, err := jwtSvc.GenerateAccessToken(user)
	if err != nil {
		return nil, err
	}

	// Generate refresh token
	refreshToken, err := jwtSvc.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	// Store session
	expiresAt := time.Now().Add(jwtSvc.GetRefreshTokenDuration())
	ipAddress := r.RemoteAddr
	userAgent := r.UserAgent()

	// Only the token digest is persisted - see hashRefreshToken
	_, err = database.GetPool().Exec(ctx,
		`INSERT INTO sessions (user_id, refresh_token, expires_at, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5)`,
		user.ID, hashRefreshToken(refreshToken), expiresAt, ipAddress, userAgent)
	if err != nil {
		return nil, err
	}

	return &models.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(jwtSvc.GetAccessTokenDuration().Seconds()),
		User:         *user,
	}, nil
}

// @Summary Change password
// @Description Change the current user's password
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body models.ChangePasswordRequest true "Password change details"
// @Success 200 {string} string "Password changed successfully"
// @Failure 400 {string} string "Invalid request"
// @Failure 401 {string} string "Unauthorized or invalid old password"
// @Failure 500 {string} string "Internal server error"
// @Security ApiKeyAuth
// @Router /auth/change-password [post]
func ChangePassword(w http.ResponseWriter, r *http.Request) {
	// Get user from context (set by auth middleware)
	user, ok := GetUserFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.OldPassword == "" || req.NewPassword == "" {
		http.Error(w, "Old password and new password are required", http.StatusBadRequest)
		return
	}

	// Password strength check
	if len(req.NewPassword) < 8 {
		http.Error(w, "New password must be at least 8 characters", http.StatusBadRequest)
		return
	}

	// OAuth users cannot change password
	if user.PasswordHash == nil {
		http.Error(w, "OAuth users cannot change password", http.StatusBadRequest)
		return
	}

	// Verify old password
	valid, err := auth.VerifyPassword(req.OldPassword, *user.PasswordHash)
	if err != nil {
		logger.Error("Failed to verify password: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if !valid {
		http.Error(w, "Invalid old password", http.StatusUnauthorized)
		return
	}

	// Hash new password
	newPasswordHash, err := auth.HashPassword(req.NewPassword, nil)
	if err != nil {
		logger.Error("Failed to hash new password: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Update password and clear password_must_change flag
	ctx, cancel := RequestContext(r)
	defer cancel()
	_, err = database.GetPool().Exec(ctx,
		`UPDATE users SET password_hash = $1, password_must_change = false, updated_at = CURRENT_TIMESTAMP WHERE id = $2`,
		newPasswordHash, user.ID)
	if err != nil {
		logger.Error("Failed to update password: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	logger.Info("Password changed for user: %s (ID: %d)", user.Email, user.ID)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"message": "Password changed successfully"}); err != nil {
		logger.Error("Failed to encode response: %v", err)
	}
}

func isValidEmail(email string) bool {
	// Basic email validation - you might want to use a more robust library
	return len(email) > 3 && strings.Contains(email, "@") && strings.Contains(email, ".")
}

func isDuplicateKeyError(err error) bool {
	// Check for PostgreSQL duplicate key error
	return err != nil && (strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint"))
}

// GetUserFromContext extracts user from request context
func GetUserFromContext(r *http.Request) (*models.User, bool) {
	user, ok := r.Context().Value(models.UserContextKey).(*models.User)
	return user, ok
}

// GetUserIDFromPath extracts user ID from URL path
func GetUserIDFromPath(r *http.Request) (int, error) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	if idStr == "" {
		return 0, errors.New("missing user ID")
	}
	return strconv.Atoi(idStr)
}
