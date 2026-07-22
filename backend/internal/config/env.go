package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/nomdb/backend/internal/logger"
)

// Config holds all configuration for the application
type Config struct {
	// Database
	DatabaseURL string

	// Google Maps
	GoogleMapsAPIKey string

	// Authentication
	AuthMode         string
	JWTSecretKey     string
	OIDCIssuerURL    string
	OIDCClientID     string
	OIDCClientSecret string
	OIDCRedirectURL  string

	// AWS S3
	AWSAccessKeyID     string
	AWSSecretAccessKey string
	AWSRegion          string
	S3BucketName       string
	S3Endpoint         string

	// Server
	Port           string
	AllowedOrigins []string
	Debug          bool
}

// Load loads and validates environment variables
func Load() (*Config, error) {
	cfg := &Config{
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		GoogleMapsAPIKey:   os.Getenv("GOOGLE_MAPS_API_KEY"),
		AuthMode:           getEnvOrDefault("AUTH_MODE", "local"),
		JWTSecretKey:       os.Getenv("JWT_SECRET_KEY"),
		OIDCIssuerURL:      os.Getenv("OIDC_ISSUER_URL"),
		OIDCClientID:       os.Getenv("OIDC_CLIENT_ID"),
		OIDCClientSecret:   os.Getenv("OIDC_CLIENT_SECRET"),
		OIDCRedirectURL:    os.Getenv("OIDC_REDIRECT_URL"),
		AWSAccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
		AWSSecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
		AWSRegion:          os.Getenv("AWS_REGION"),
		S3BucketName:       os.Getenv("S3_BUCKET_NAME"),
		S3Endpoint:         os.Getenv("S3_ENDPOINT"),
		Port:               getEnvOrDefault("PORT", "8080"),
		Debug:              os.Getenv("DEBUG") == "true",
	}

	// Parse allowed origins
	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	if allowedOrigins != "" {
		cfg.AllowedOrigins = splitAndTrim(allowedOrigins, ",")
	} else {
		cfg.AllowedOrigins = []string{"http://localhost:3000", "http://localhost:5173"}
	}

	// Validate required variables
	var errors []string

	if cfg.DatabaseURL == "" {
		errors = append(errors, "DATABASE_URL is required")
	}

	// Validate auth mode
	validAuthModes := []string{"none", "local", "oauth", "both"}
	if !contains(validAuthModes, cfg.AuthMode) {
		errors = append(errors, fmt.Sprintf("AUTH_MODE must be one of: %v", validAuthModes))
	}

	// Validate auth-specific requirements
	if cfg.AuthMode == "local" || cfg.AuthMode == "both" {
		if cfg.JWTSecretKey == "" {
			errors = append(errors, "JWT_SECRET_KEY is required for local/both auth mode")
		} else if len(cfg.JWTSecretKey) < 32 {
			logger.Warn("⚠️  JWT_SECRET_KEY is less than 32 characters - consider using a longer key for better security")
		}
	}

	if cfg.AuthMode == "oauth" || cfg.AuthMode == "both" {
		if cfg.OIDCIssuerURL == "" {
			errors = append(errors, "OIDC_ISSUER_URL is required for oauth/both auth mode")
		}
		if cfg.OIDCClientID == "" {
			errors = append(errors, "OIDC_CLIENT_ID is required for oauth/both auth mode")
		}
		if cfg.OIDCClientSecret == "" {
			errors = append(errors, "OIDC_CLIENT_SECRET is required for oauth/both auth mode")
		}
	}

	// Warn about optional but recommended variables
	if cfg.GoogleMapsAPIKey == "" {
		logger.Warn("⚠️  GOOGLE_MAPS_API_KEY not set - Google Maps features will be unavailable")
	}

	if cfg.AuthMode == "none" {
		logger.Warn("⚠️  Authentication is DISABLED (AUTH_MODE=none) - only use for testing!")
	}

	// Return errors if any
	if len(errors) > 0 {
		return nil, fmt.Errorf("configuration validation failed:\n  - %s", strings.Join(errors, "\n  - "))
	}

	return cfg, nil
}

// Helper functions

func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func splitAndTrim(s, sep string) []string {
	parts := strings.Split(s, sep)
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
