package models

import "time"

// User represents a user in the system
type User struct {
	ID                 int        `json:"id"`
	Email              string     `json:"email"`
	Username           string     `json:"username"`
	PasswordHash       *string    `json:"-"` // Never send to client
	Provider           string     `json:"provider"`
	ProviderID         *string    `json:"provider_id,omitempty"`
	FullName           *string    `json:"full_name"`
	AvatarURL          *string    `json:"avatar_url"`
	IsActive           bool       `json:"is_active"`
	IsAdmin            bool       `json:"is_admin"`
	EmailVerified      bool       `json:"email_verified"`
	PasswordMustChange bool       `json:"password_must_change"`
	LastLoginAt        *time.Time `json:"last_login_at"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// Session represents a user session with refresh token
type Session struct {
	ID           int       `json:"id"`
	UserID       int       `json:"user_id"`
	RefreshToken string    `json:"-"` // Never send to client
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
	LastUsedAt   time.Time `json:"last_used_at"`
	IPAddress    *string   `json:"ip_address,omitempty"`
	UserAgent    *string   `json:"user_agent,omitempty"`
}

// APIKey represents an API key for programmatic access
type APIKey struct {
	ID         int        `json:"id"`
	UserID     *int       `json:"user_id"`
	KeyHash    string     `json:"-"` // Never send to client
	Name       *string    `json:"name"`
	LastUsedAt *time.Time `json:"last_used_at"`
	ExpiresAt  *time.Time `json:"expires_at"`
	IsActive   bool       `json:"is_active"`
	CreatedAt  time.Time  `json:"created_at"`
}

// Auth Request/Response types

type RegisterRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
	FullName string `json:"full_name,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"` // seconds
	User         User   `json:"user"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type OAuthCallbackRequest struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// Context key for user in request context
type ContextKey string

const UserContextKey ContextKey = "user"
