package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/nomdb/backend/internal/config"
	"github.com/nomdb/backend/internal/database"
	"github.com/nomdb/backend/internal/handlers"
	"github.com/nomdb/backend/internal/logger"
	"github.com/nomdb/backend/internal/middleware"
	"github.com/nomdb/backend/internal/services"
	"github.com/rs/cors"
	httpSwagger "github.com/swaggo/http-swagger"
	"golang.org/x/time/rate"

	_ "github.com/nomdb/backend/docs" // Import generated docs
)

// @title The Nom Database API
// @version 1.0
// @description Restaurant rating and discovery API with Google Maps integration
// @description
// @description This API provides endpoints for managing restaurants, ratings, categories, and food types.
// @description It integrates with Google Maps for restaurant search and location data.
//
// @contact.name API Support
// @contact.url https://github.com/your-username/the-nom-database
// @contact.email support@nomdb.com
//
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
//
// @host localhost:8080
// @BasePath /api
//
// @schemes http https
//
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
//
// @tag.name Restaurants
// @tag.description Restaurant management endpoints
//
// @tag.name Ratings
// @tag.description Restaurant rating endpoints
//
// @tag.name Categories
// @tag.description Cultural category management
//
// @tag.name Food Types
// @tag.description Food type/cuisine management
//
// @tag.name Suggestions
// @tag.description Restaurant suggestion workflow
//
// @tag.name Google Maps
// @tag.description Google Places API integration
//
// @tag.name Photos
// @tag.description Menu photo management
//
// @tag.name Search
// @tag.description Global search functionality
//
// @tag.name Health
// @tag.description Health check endpoints
func main() {
	logger.Info("🚀 Starting The Nom Database server...")
	if logger.IsDebugMode() {
		logger.Debug("🐛 Debug mode enabled - detailed logging active")
	}

	// Load and validate configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Configuration error: %v", err)
	}

	// Connect to database
	if err := database.Connect(); err != nil {
		logger.Fatal("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Run database migrations
	databaseURL := os.Getenv("DATABASE_URL")
	migrationsPath := "/app/db/migrations_new"
	if _, err := os.Stat("./db/migrations_new"); err == nil {
		// Local development
		migrationsPath = "./db/migrations_new"
	}
	if err := database.RunMigrations(databaseURL, migrationsPath); err != nil {
		logger.Fatal("Failed to run migrations: %v", err)
	}

	// Initialize default admin user
	if err := database.InitDefaultAdmin(); err != nil {
		logger.Fatal("Failed to initialize default admin: %v", err)
	}

	// Initialize Google Maps service
	_ = services.NewGoogleMapsService()

	// Initialize S3 service (optional - falls back to local storage if not configured)
	if err := services.InitS3(); err != nil {
		logger.Debug("S3 initialization skipped: %v", err)
	}

	// Initialize authentication
	jwtSvc := handlers.InitAuthService()

	// Initialize OIDC (optional)
	if err := handlers.InitOIDC(); err != nil {
		logger.Warn("OIDC initialization skipped: %v", err)
	}

	// Initialize auth middleware
	middleware.InitAuthMiddleware(jwtSvc)

	// Create uploads directory
	uploadsDir := "./uploads"
	if err := os.MkdirAll(uploadsDir+"/menu_photos", 0755); err != nil {
		logger.Warn("Failed to create uploads directory: %v", err)
	} else {
		logger.Debug("📁 Uploads directory ready: %s", uploadsDir)
	}

	// Shared middleware chains for route groups
	optionalAuth := []func(http.Handler) http.Handler{
		middleware.OptionalAuthMiddleware,
		middleware.WithUserRoles,
	}

	healthHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if r.Method == "GET" {
			_, _ = w.Write([]byte("OK"))
		}
	}

	// Create router
	r := chi.NewRouter()

	r.Route("/api", func(api chi.Router) {
		// Static uploads (kept inside /api to match the mux PathPrefix route)
		api.Handle("/uploads/*", http.StripPrefix("/api/uploads/", http.FileServer(http.Dir(uploadsDir))))

		// Auth routes: public endpoints plus a protected group
		api.Route("/auth", func(auth chi.Router) {
			auth.Post("/register", handlers.Register)
			auth.Post("/login", handlers.Login)
			auth.Post("/refresh", handlers.RefreshToken)
			auth.Post("/logout", handlers.Logout)
			auth.Get("/oidc/login", handlers.OIDCLogin)
			auth.Get("/oidc/callback", handlers.OIDCCallback)
			auth.Post("/oidc/exchange", handlers.ExchangeOIDCCode)

			auth.Group(func(protected chi.Router) {
				protected.Use(middleware.AuthMiddleware)
				protected.Use(middleware.WithUserRoles) // Load roles and permissions
				protected.Get("/me", handlers.GetMe)
				protected.Post("/change-password", handlers.ChangePassword)
			})
		})

		// Categories - public reads, permission-guarded writes
		api.With(optionalAuth...).Get("/categories", handlers.GetCategories)
		api.With(optionalAuth...).Get("/categories/{id}", handlers.GetCategory)
		categoriesWrite := api.With(
			middleware.AuthMiddleware,
			middleware.WithUserRoles,
			middleware.RequireAnyPermission("categories.create", "categories.update", "categories.delete"),
		)
		categoriesWrite.Post("/categories", handlers.CreateCategory)
		categoriesWrite.Put("/categories/{id}", handlers.UpdateCategory)
		categoriesWrite.Delete("/categories/{id}", handlers.DeleteCategory)

		// Food Types - public reads, permission-guarded writes
		api.With(optionalAuth...).Get("/food-types", handlers.GetFoodTypes)
		api.With(optionalAuth...).Get("/food-types/{id}", handlers.GetFoodType)
		foodTypesWrite := api.With(
			middleware.AuthMiddleware,
			middleware.WithUserRoles,
			middleware.RequireAnyPermission("food_types.create", "food_types.update", "food_types.delete"),
		)
		foodTypesWrite.Post("/food-types", handlers.CreateFoodType)
		foodTypesWrite.Put("/food-types/{id}", handlers.UpdateFoodType)
		foodTypesWrite.Delete("/food-types/{id}", handlers.DeleteFoodType)

		// Restaurants: public reads (incl. nested public reads), authenticated writes.
		// The {id:[0-9]+} patterns keep the mux semantics: non-numeric or negative
		// IDs never reach the handlers and yield a router 404.
		api.Route("/restaurants", func(restaurants chi.Router) {
			restaurants.Group(func(public chi.Router) {
				public.Use(middleware.OptionalAuthMiddleware)
				public.Use(middleware.WithUserRoles)
				public.Get("/", handlers.GetRestaurants)
				public.Get("/paginated", handlers.GetRestaurantsPaginated)
				public.Get("/{id:[0-9]+}", handlers.GetRestaurant)
				public.Get("/{restaurantId}/ratings", handlers.GetRatings)
				public.Get("/{restaurantId}/photos", handlers.GetMenuPhotos)
				public.Get("/{restaurantId}/lists", handlers.GetRestaurantLists)
			})

			restaurants.Group(func(protected chi.Router) {
				protected.Use(middleware.AuthMiddleware)
				protected.Use(middleware.WithUserRoles)
				protected.With(middleware.RequirePermission("restaurants.create")).Post("/", handlers.CreateRestaurant)
				protected.Put("/{id:[0-9]+}", handlers.UpdateRestaurant)
				protected.Delete("/{id:[0-9]+}", handlers.DeleteRestaurant)
				protected.With(middleware.RequirePermission("photos.upload")).Post("/{restaurantId}/photos", handlers.UploadMenuPhoto)
			})
		})

		// Global Search (public)
		api.With(optionalAuth...).Get("/search", handlers.GlobalSearch)

		// Ratings (reads are nested under /restaurants above; writes require auth)
		api.Route("/ratings", func(ratings chi.Router) {
			ratings.Use(middleware.AuthMiddleware)
			ratings.Use(middleware.WithUserRoles)
			ratings.With(middleware.RequirePermission("ratings.create")).Post("/", handlers.CreateRating)
			// Update/Delete and votes: ownership checked in the handlers
			ratings.Put("/{id}", handlers.UpdateRating)
			ratings.Delete("/{id}", handlers.DeleteRating)
			ratings.Post("/{id}/vote", handlers.VoteOnReview)
			ratings.Delete("/{id}/vote", handlers.RemoveVote)
			ratings.Post("/{id}/photos", handlers.UploadReviewPhoto)
		})

		// Review photos (authenticated users only)
		api.Route("/review-photos", func(reviewPhotos chi.Router) {
			reviewPhotos.Use(middleware.AuthMiddleware)
			reviewPhotos.Use(middleware.WithUserRoles)
			reviewPhotos.Patch("/{id}", handlers.UpdateReviewPhotoCaption)
			reviewPhotos.Delete("/{id}", handlers.DeleteReviewPhoto)
		})

		// Public runtime config for the browser (Maps JS key, map ID)
		api.Get("/config", handlers.GetPublicConfig)

		// Google Maps (proxied through backend - public with rate limiting)
		api.With(optionalAuth...).Get("/places/search", handlers.SearchPlaces)
		api.With(optionalAuth...).Get("/places/{placeId}", handlers.GetPlaceDetails)
		api.With(optionalAuth...).Get("/geocode/cities", handlers.GeocodeCities)

		// Restaurant Suggestions (requires auth; permissions per action)
		api.Route("/suggestions", func(suggestions chi.Router) {
			suggestions.Use(middleware.AuthMiddleware)
			suggestions.Use(middleware.WithUserRoles)

			suggestionsRead := suggestions.With(middleware.RequirePermission("suggestions.read"))
			suggestionsRead.Get("/", handlers.GetSuggestions)
			suggestionsRead.Get("/{id}", handlers.GetSuggestion)

			suggestions.With(middleware.RequirePermission("suggestions.create")).Post("/", handlers.CreateSuggestion)

			moderate := suggestions.With(middleware.RequireAnyPermission("suggestions.approve", "suggestions.convert", "suggestions.reject"))
			moderate.Post("/{id}/approve", handlers.ApproveSuggestion)
			moderate.Post("/{id}/reject", handlers.RejectSuggestion)
			moderate.Patch("/{id}/status", handlers.UpdateSuggestionStatus)
			moderate.Post("/{id}/convert", handlers.ConvertSuggestion)
			moderate.Delete("/{id}", handlers.DeleteSuggestion)
		})

		// Update/Delete menu photos (ownership checked in handler)
		api.Route("/photos", func(photos chi.Router) {
			photos.Use(middleware.AuthMiddleware)
			photos.Use(middleware.WithUserRoles)
			photos.Patch("/{id}", handlers.UpdatePhotoCaption)
			photos.Delete("/{id}", handlers.DeleteMenuPhoto)
		})

		// Restaurant Lists (authenticated users only)
		api.Route("/lists", func(lists chi.Router) {
			lists.Use(middleware.AuthMiddleware)
			lists.Use(middleware.WithUserRoles)
			lists.Get("/", handlers.GetUserLists)
			lists.Post("/", handlers.CreateList)
			lists.Get("/{id}", handlers.GetList)
			lists.Put("/{id}", handlers.UpdateList)
			lists.Delete("/{id}", handlers.DeleteList)
			lists.Post("/{id}/restaurants", handlers.AddRestaurantToList)
			lists.Delete("/{id}/restaurants/{restaurantId}", handlers.RemoveRestaurantFromList)
		})

		// User Profile (public reads)
		api.With(optionalAuth...).Get("/users/{id}", handlers.GetUserProfile)
		api.With(optionalAuth...).Get("/users/{id}/reviews", handlers.GetUserReviews)

		// Update own profile (authenticated)
		api.Route("/user", func(user chi.Router) {
			user.Use(middleware.AuthMiddleware)
			user.Use(middleware.WithUserRoles)
			user.Put("/profile", handlers.UpdateUserProfile)
			user.Post("/avatar", handlers.UploadAvatar)
		})

		// Admin Routes (requires admin role)
		api.Route("/admin", func(admin chi.Router) {
			admin.Use(middleware.AuthMiddleware)
			admin.Use(middleware.WithUserRoles)
			admin.Use(middleware.RequireRole("admin"))

			// User Management
			admin.Get("/users", handlers.AdminListUsers)
			admin.Post("/users", handlers.AdminCreateUser)
			admin.Get("/users/{id}", handlers.AdminGetUser)
			admin.Put("/users/{id}", handlers.AdminUpdateUser)
			admin.Delete("/users/{id}", handlers.AdminDeleteUser)
			admin.Post("/users/{id}/roles", handlers.AdminAssignRole)
			admin.Delete("/users/{id}/roles/{roleId}", handlers.AdminRemoveRole)
			admin.Post("/users/{id}/reset-password", handlers.AdminResetPassword)

			// Role Management
			admin.Get("/roles", handlers.AdminListRoles)
			admin.Post("/roles", handlers.AdminCreateRole)
			admin.Get("/roles/{id}", handlers.AdminGetRole)
			admin.Put("/roles/{id}", handlers.AdminUpdateRole)
			admin.Delete("/roles/{id}", handlers.AdminDeleteRole)
			admin.Post("/roles/{id}/permissions", handlers.AdminAssignPermission)
			admin.Delete("/roles/{id}/permissions/{permissionId}", handlers.AdminRemovePermission)

			// Permission Management
			admin.Get("/permissions", handlers.AdminListPermissions)

			// System Statistics and Analytics
			admin.Get("/stats", handlers.AdminGetStatistics)
			admin.Get("/analytics/user-growth", handlers.AdminGetUserGrowth)
			admin.Get("/analytics/active-users", handlers.AdminGetActiveUsers)
			admin.Get("/analytics/popular-restaurants", handlers.AdminGetPopularRestaurants)

			// System Settings
			admin.Get("/settings", handlers.AdminGetSettings)

			// Audit Logs
			admin.Get("/audit-logs", handlers.AdminGetAuditLogs)

			// Content Moderation
			admin.Get("/ratings", handlers.AdminListRatings)
			admin.Delete("/ratings/{id}", handlers.AdminDeleteRating)
			admin.Get("/photos", handlers.AdminListPhotos)
			admin.Delete("/photos/{type}/{id}", handlers.AdminDeletePhoto)

			// Restaurant Moderation
			admin.Put("/restaurants/{id}", handlers.AdminUpdateRestaurant)
			admin.Delete("/restaurants/{id}", handlers.AdminDeleteRestaurant)
		})

		// Health check (support both GET and HEAD for Docker healthcheck)
		api.Get("/health", healthHandler)
		api.Head("/health", healthHandler)

		// Metrics endpoint
		api.Get("/metrics", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			stats := middleware.GetMetrics().GetStats()
			json.NewEncoder(w).Encode(stats)
		})

		// Serve the swagger.yaml file first
		api.Get("/swagger.yaml", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/x-yaml")
			http.ServeFile(w, r, "./docs/swagger.yaml")
		})

		// Redirect /docs to /docs/ for Swagger UI
		api.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/api/docs/", http.StatusMovedPermanently)
		})

		// Swagger UI - serve at /api/docs/ (must be after swagger.yaml)
		api.Handle("/docs/*", httpSwagger.Handler(
			httpSwagger.URL("/api/swagger.yaml"),
		))
	})

	// Initialize rate limiter
	// Allow 100 requests per minute per IP, with burst of 20
	rateLimiter := middleware.NewIPRateLimiter(rate.Every(time.Minute/100), 20)
	// Start cleanup task to prevent memory leaks (run every 10 minutes)
	rateLimiter.StartCleanupTask(10 * time.Minute)
	logger.Info("🔒 Rate limiting enabled: 100 req/min per IP, burst: 20")

	// CORS middleware - more restrictive configuration
	c := cors.New(cors.Options{
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "X-Requested-With"},
		AllowCredentials: true,
		MaxAge:           300, // Cache preflight for 5 minutes
	})
	logger.Info("🌐 CORS configured for origins: %v", cfg.AllowedOrigins)

	// Apply middleware chain (order matters)
	// Recovery -> RequestID -> Security headers -> Rate limiting -> Request validation -> Max bytes -> Sanitization -> Compression -> Logging -> CORS -> Router
	handler := middleware.RecoveryMiddleware(
		middleware.RequestIDMiddleware(
			middleware.SecurityHeadersMiddleware(
				middleware.RateLimitMiddleware(rateLimiter)(
					middleware.ValidateContentTypeMiddleware(
						middleware.MaxBytesMiddleware(10 * 1024 * 1024)( // 10MB max request size
							middleware.SanitizeInputMiddleware(
								middleware.CompressionMiddleware(
									middleware.LoggingMiddleware(
										c.Handler(r))))))))))

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.Info("🌐 Server listening on http://localhost:%s", port)
	logger.Info("📡 API available at http://localhost:%s/api", port)
	logger.Info("📚 Swagger UI available at http://localhost:%s/api/docs", port)
	logger.Info("🛡️  Security features enabled:")
	logger.Info("   ✓ Panic recovery and error handling")
	logger.Info("   ✓ Authentication mode: %s", cfg.AuthMode)
	logger.Info("   ✓ Rate limiting (100 req/min per IP)")
	logger.Info("   ✓ Request size limits (10MB max)")
	logger.Info("   ✓ Content-Type validation")
	logger.Info("   ✓ Input sanitization")
	logger.Info("   ✓ Security headers (XSS, clickjacking, MIME sniffing protection)")
	logger.Info("   ✓ CORS restrictions")
	logger.Info("✅ Server ready to accept connections")

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: handler,
		// Slowloris protection; full read/write timeouts are omitted on purpose
		// so large photo uploads on slow connections are not cut off.
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// Graceful shutdown on SIGINT/SIGTERM so deferred cleanup (DB close) runs
	shutdownCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("Server failed to start: %v", err)
		}
	}()

	<-shutdownCtx.Done()
	logger.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Graceful shutdown failed: %v", err)
	}
}
