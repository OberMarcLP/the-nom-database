package database

import (
	"context"
	"fmt"
	"os"
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
			password = "nomdb_secret"
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

	// Optimize connection pool settings for performance
	config.MaxConns = 25                              // Maximum number of connections
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

// InitDefaultAdmin ensures the default admin user exists with password set
func InitDefaultAdmin() error {
	ctx := context.Background()

	// Check if admin user exists
	var userID int
	var hasPassword bool
	err := pool.QueryRow(ctx,
		`SELECT id, password_hash IS NOT NULL FROM users WHERE email = 'admin@nomdb.local' AND username = 'admin'`).
		Scan(&userID, &hasPassword)

	// Hash default password "admin"
	passwordHash, err2 := auth.HashPassword("admin", nil)
	if err2 != nil {
		return fmt.Errorf("failed to hash default password: %w", err2)
	}

	// Generate Gravatar URL for admin
	gravatarService := services.NewGravatarService()
	avatarURL := gravatarService.GetAvatarURL("admin@nomdb.local", 256)

	if err == pgx.ErrNoRows {
		// Admin user doesn't exist - create it
		logger.Info("Creating default admin user (username: admin, password: admin)")

		err = pool.QueryRow(ctx,
			`INSERT INTO users (username, email, password_hash, full_name, avatar_url, is_active, is_admin, email_verified, password_must_change, provider)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			RETURNING id`,
			"admin", "admin@nomdb.local", passwordHash, "Administrator", avatarURL, true, true, true, true, "local").
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

	// If admin already has password, skip
	if hasPassword {
		logger.Debug("Default admin user already has password set")
		return nil
	}

	// Update admin user with password and Gravatar
	_, err = pool.Exec(ctx,
		`UPDATE users SET password_hash = $1, avatar_url = $2, password_must_change = true WHERE email = 'admin@nomdb.local' AND username = 'admin'`,
		passwordHash, avatarURL)

	if err != nil {
		return fmt.Errorf("failed to set default admin password: %w", err)
	}

	logger.Info("🔑 Default admin user initialized (username: admin, password: admin)")
	logger.Warn("⚠️  IMPORTANT: Change the default admin password on first login!")

	return nil
}
