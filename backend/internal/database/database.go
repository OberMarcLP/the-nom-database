package database

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nomdb/backend/internal/auth"
	"github.com/nomdb/backend/internal/logger"
	"github.com/nomdb/backend/internal/services"
)

var pool *pgxpool.Pool

func Connect() error {
	logger.Info("🔌 Connecting to database...")

	// Try to get DATABASE_URL first, otherwise construct from individual components
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Build connection string from individual components
		host := os.Getenv("DB_HOST")
		port := os.Getenv("DB_PORT")
		user := os.Getenv("DB_USER")
		password := os.Getenv("DB_PASSWORD")
		dbname := os.Getenv("DB_NAME")
		sslmode := os.Getenv("DB_SSLMODE")

		// Set defaults
		if host == "" {
			host = "localhost"
		}
		if port == "" {
			port = "5432"
		}
		if user == "" {
			user = "nomdb"
		}
		if password == "" {
			// Check if we're in production mode
			if os.Getenv("GO_ENV") == "production" || os.Getenv("ENV") == "production" {
				logger.Error("❌ DB_PASSWORD is required in production mode")
				return fmt.Errorf("DB_PASSWORD environment variable is required in production")
			}
			// Only use default password in development
			password = "nomdb_secret"
			logger.Warn("⚠️  Using default database password - DO NOT use in production!")
		}
		if dbname == "" {
			dbname = "nomdb"
		}
		if sslmode == "" {
			sslmode = "disable"
		}

		logger.Debug("Building connection string from components: host=%s port=%s user=%s dbname=%s sslmode=%s",
			host, port, user, dbname, sslmode)

		dbURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
			user, password, host, port, dbname, sslmode)
	} else {
		logger.Debug("Using DATABASE_URL from environment")
	}

	// Parse the connection string and configure pool settings
	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		logger.Error("❌ Failed to parse database URL: %v", err)
		return fmt.Errorf("unable to parse database URL: %w", err)
	}

	// Optimize connection pool settings for performance (DB_MAX_CONNS optional, default 25)
	maxConns := 25
	if v := os.Getenv("DB_MAX_CONNS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			maxConns = n
		}
	}
	config.MaxConns = int32(maxConns)                 // Maximum number of connections
	config.MinConns = 5                               // Minimum number of idle connections
	config.MaxConnLifetime = time.Hour                // Max connection lifetime (1 hour)
	config.MaxConnIdleTime = 30 * time.Minute         // Max idle time (30 minutes)
	config.HealthCheckPeriod = time.Minute            // Health check every minute

	logger.Debug("Pool configuration: MaxConns=%d, MinConns=%d, MaxConnLifetime=%ds, MaxConnIdleTime=%ds",
		config.MaxConns, config.MinConns, int(config.MaxConnLifetime.Seconds()), int(config.MaxConnIdleTime.Seconds()))

	pool, err = pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		logger.Error("❌ Failed to create database connection pool: %v", err)
		return fmt.Errorf("unable to connect to database: %w", err)
	}

	logger.Debug("Testing database connection...")
	if err := pool.Ping(context.Background()); err != nil {
		logger.Error("❌ Database ping failed: %v", err)
		return fmt.Errorf("unable to ping database: %w", err)
	}

	logger.Info("✅ Database connected successfully")
	return nil
}

func GetPool() *pgxpool.Pool {
	return pool
}

func Close() {
	if pool != nil {
		logger.Info("🔌 Closing database connection...")
		pool.Close()
		logger.Info("✅ Database connection closed")
	}
}

// generateSecurePassword creates a random password for initial admin setup
func generateSecurePassword(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

// InitDefaultAdmin ensures the default admin user exists with password set
func InitDefaultAdmin() error {
	ctx := context.Background()

	// Check if admin user exists
	var userID int
	var hasPassword bool
	err := pool.QueryRow(ctx,
		`SELECT id, password_hash IS NOT NULL FROM users WHERE email = 'admin@nomdb.local' AND username = 'admin'`).
		Scan(&userID, &hasPassword)

	// Check for admin password from environment or generate a secure one
	adminPassword := os.Getenv("ADMIN_DEFAULT_PASSWORD")
	explicitPassword := adminPassword != ""
	if adminPassword == "" {
		// Generate a random secure password
		var genErr error
		adminPassword, genErr = generateSecurePassword(16)
		if genErr != nil {
			return fmt.Errorf("failed to generate secure password: %w", genErr)
		}
	}

	// Hash the password
	passwordHash, err2 := auth.HashPassword(adminPassword, nil)
	if err2 != nil {
		return fmt.Errorf("failed to hash default password: %w", err2)
	}

	// Generate Gravatar URL for admin
	gravatarService := services.NewGravatarService()
	avatarURL := gravatarService.GetAvatarURL("admin@nomdb.local", 256)

	if err == pgx.ErrNoRows {
		// Admin user doesn't exist - create it
		logger.Info("Creating default admin user (username: admin)")
		if explicitPassword {
			logger.Info("🔑 Admin password taken from ADMIN_DEFAULT_PASSWORD")
		} else {
			logger.Warn("⚠️  Generated admin password: %s", adminPassword)
			logger.Warn("⚠️  IMPORTANT: Save this password and change it immediately after first login!")
		}

		err = pool.QueryRow(ctx,
			`INSERT INTO users (username, email, password_hash, full_name, avatar_url, is_active, is_admin, email_verified, password_must_change, provider)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			RETURNING id`,
			"admin", "admin@nomdb.local", passwordHash, "Administrator", avatarURL, true, true, true, !explicitPassword, "local").
			Scan(&userID)

		if err != nil {
			return fmt.Errorf("failed to create admin user: %w", err)
		}

		// Assign admin role
		_, err = pool.Exec(ctx,
			`INSERT INTO user_roles (user_id, role_id)
			SELECT $1, id FROM roles WHERE name = 'admin'`,
			userID)

		if err != nil {
			return fmt.Errorf("failed to assign admin role: %w", err)
		}

		logger.Info("✅ Default admin user created successfully")
		return nil
	}

	if err != nil {
		return fmt.Errorf("failed to check admin user: %w", err)
	}

	// If admin already has a password, keep it - unless an explicit
	// ADMIN_DEFAULT_PASSWORD is configured, which always wins so a known
	// password can be enforced from the environment (dev/ops recovery).
	if hasPassword && !explicitPassword {
		logger.Debug("Default admin user already has password set")
		return nil
	}

	// Update admin user with password and Gravatar
	_, err = pool.Exec(ctx,
		`UPDATE users SET password_hash = $1, avatar_url = $2, password_must_change = $3 WHERE email = 'admin@nomdb.local' AND username = 'admin'`,
		passwordHash, avatarURL, !explicitPassword)

	if err != nil {
		return fmt.Errorf("failed to set default admin password: %w", err)
	}

	logger.Info("🔑 Default admin user password reset")
	if explicitPassword {
		logger.Info("🔑 Admin password taken from ADMIN_DEFAULT_PASSWORD")
	} else {
		logger.Warn("⚠️  Generated admin password: %s", adminPassword)
		logger.Warn("⚠️  IMPORTANT: Save this password and change it immediately after first login!")
	}

	return nil
}
