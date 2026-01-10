package main

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
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

	// Create router
	r := mux.NewRouter()

	// Create uploads directory and serve static files
	uploadsDir := "./uploads"
	if err := os.MkdirAll(uploadsDir+"/menu_photos", 0755); err != nil {
		logger.Warn("Failed to create uploads directory: %v", err)
	} else {
		logger.Debug("📁 Uploads directory ready: %s", uploadsDir)
	}
	r.PathPrefix("/api/uploads/").Handler(
		http.StripPrefix("/api/uploads/", http.FileServer(http.Dir(uploadsDir))))

	// API routes
	api := r.PathPrefix("/api").Subrouter()

	// Public auth routes (no authentication required)
	api.HandleFunc("/auth/register", handlers.Register).Methods("POST")
	api.HandleFunc("/auth/login", handlers.Login).Methods("POST")
	api.HandleFunc("/auth/refresh", handlers.RefreshToken).Methods("POST")
	api.HandleFunc("/auth/logout", handlers.Logout).Methods("POST")
	api.HandleFunc("/auth/oidc/login", handlers.OIDCLogin).Methods("GET")
	api.HandleFunc("/auth/oidc/callback", handlers.OIDCCallback).Methods("GET")

	// Protected auth routes (authentication required)
	authRoutes := api.PathPrefix("/auth").Subrouter()
	authRoutes.Use(middleware.AuthMiddleware)
	authRoutes.Use(middleware.WithUserRoles) // Load roles and permissions
	authRoutes.HandleFunc("/me", handlers.GetMe).Methods("GET")
	authRoutes.HandleFunc("/change-password", handlers.ChangePassword).Methods("POST")

	// Categories - Public GET routes
	api.Handle("/categories", middleware.OptionalAuthMiddleware(middleware.WithUserRoles(http.HandlerFunc(handlers.GetCategories)))).Methods("GET")
	api.Handle("/categories/{id}", middleware.OptionalAuthMiddleware(middleware.WithUserRoles(http.HandlerFunc(handlers.GetCategory)))).Methods("GET")

	// Categories - Protected POST/PUT/DELETE routes
	api.Handle("/categories",
		middleware.AuthMiddleware(
			middleware.WithUserRoles(
				middleware.RequireAnyPermission("categories.create", "categories.update", "categories.delete")(
					http.HandlerFunc(handlers.CreateCategory))))).Methods("POST")
	api.Handle("/categories/{id}",
		middleware.AuthMiddleware(
			middleware.WithUserRoles(
				middleware.RequireAnyPermission("categories.create", "categories.update", "categories.delete")(
					http.HandlerFunc(handlers.UpdateCategory))))).Methods("PUT")
	api.Handle("/categories/{id}",
		middleware.AuthMiddleware(
			middleware.WithUserRoles(
				middleware.RequireAnyPermission("categories.create", "categories.update", "categories.delete")(
					http.HandlerFunc(handlers.DeleteCategory))))).Methods("DELETE")

	// Food Types - Public GET routes
	api.Handle("/food-types", middleware.OptionalAuthMiddleware(middleware.WithUserRoles(http.HandlerFunc(handlers.GetFoodTypes)))).Methods("GET")
	api.Handle("/food-types/{id}", middleware.OptionalAuthMiddleware(middleware.WithUserRoles(http.HandlerFunc(handlers.GetFoodType)))).Methods("GET")

	// Food Types - Protected POST/PUT/DELETE routes
	api.Handle("/food-types",
		middleware.AuthMiddleware(
			middleware.WithUserRoles(
				middleware.RequireAnyPermission("food_types.create", "food_types.update", "food_types.delete")(
					http.HandlerFunc(handlers.CreateFoodType))))).Methods("POST")
	api.Handle("/food-types/{id}",
		middleware.AuthMiddleware(
			middleware.WithUserRoles(
				middleware.RequireAnyPermission("food_types.create", "food_types.update", "food_types.delete")(
					http.HandlerFunc(handlers.UpdateFoodType))))).Methods("PUT")
	api.Handle("/food-types/{id}",
		middleware.AuthMiddleware(
			middleware.WithUserRoles(
				middleware.RequireAnyPermission("food_types.create", "food_types.update", "food_types.delete")(
					http.HandlerFunc(handlers.DeleteFoodType))))).Methods("DELETE")

	// Restaurants - Public GET routes
	restaurantsGet := api.PathPrefix("/restaurants").Subrouter()
	restaurantsGet.Use(middleware.OptionalAuthMiddleware)
	restaurantsGet.Use(middleware.WithUserRoles)
	restaurantsGet.HandleFunc("", handlers.GetRestaurants).Methods("GET")
	restaurantsGet.HandleFunc("/paginated", handlers.GetRestaurantsPaginated).Methods("GET")
	restaurantsGet.HandleFunc("/{id:[0-9]+}", handlers.GetRestaurant).Methods("GET")

	// Restaurants - Protected POST route
	restaurantsPost := api.PathPrefix("/restaurants").Subrouter()
	restaurantsPost.Use(middleware.AuthMiddleware)
	restaurantsPost.Use(middleware.WithUserRoles)
	restaurantsPost.Use(middleware.RequirePermission("restaurants.create"))
	restaurantsPost.HandleFunc("", handlers.CreateRestaurant).Methods("POST")

	// Restaurants - Protected PUT/DELETE routes
	restaurantsPutDelete := api.PathPrefix("/restaurants").Subrouter()
	restaurantsPutDelete.Use(middleware.AuthMiddleware)
	restaurantsPutDelete.Use(middleware.WithUserRoles)
	restaurantsPutDelete.HandleFunc("/{id:[0-9]+}", handlers.UpdateRestaurant).Methods("PUT")
	restaurantsPutDelete.HandleFunc("/{id:[0-9]+}", handlers.DeleteRestaurant).Methods("DELETE")

	// Global Search (public)
	api.Handle("/search", middleware.OptionalAuthMiddleware(middleware.WithUserRoles(http.HandlerFunc(handlers.GlobalSearch)))).Methods("GET")

	// Ratings (read public, write requires auth)
	api.Handle("/restaurants/{restaurantId}/ratings", middleware.OptionalAuthMiddleware(middleware.WithUserRoles(http.HandlerFunc(handlers.GetRatings)))).Methods("GET")

	ratingsProtected := api.PathPrefix("/ratings").Subrouter()
	ratingsProtected.Use(middleware.AuthMiddleware)
	ratingsProtected.Use(middleware.WithUserRoles)
	ratingsProtected.Use(middleware.RequirePermission("ratings.create"))
	ratingsProtected.HandleFunc("", handlers.CreateRating).Methods("POST")

	// Delete ratings (ownership checked in handler)
	ratingsDelete := api.PathPrefix("/ratings").Subrouter()
	ratingsDelete.Use(middleware.AuthMiddleware)
	ratingsDelete.Use(middleware.WithUserRoles)
	ratingsDelete.HandleFunc("/{id}", handlers.DeleteRating).Methods("DELETE")

	// Vote on reviews (authenticated users only)
	ratingsVote := api.PathPrefix("/ratings").Subrouter()
	ratingsVote.Use(middleware.AuthMiddleware)
	ratingsVote.Use(middleware.WithUserRoles)
	ratingsVote.HandleFunc("/{id}/vote", handlers.VoteOnReview).Methods("POST")
	ratingsVote.HandleFunc("/{id}/vote", handlers.RemoveVote).Methods("DELETE")

	// Review photos upload (separate route to avoid conflicts)
	ratingsPhotos := api.PathPrefix("").Subrouter()
	ratingsPhotos.Use(middleware.AuthMiddleware)
	ratingsPhotos.Use(middleware.WithUserRoles)
	ratingsPhotos.HandleFunc("/ratings/{id}/photos", handlers.UploadReviewPhoto).Methods("POST")

	// Review photos (authenticated users only)
	reviewPhotos := api.PathPrefix("/review-photos").Subrouter()
	reviewPhotos.Use(middleware.AuthMiddleware)
	reviewPhotos.Use(middleware.WithUserRoles)
	reviewPhotos.HandleFunc("/{id}", handlers.DeleteReviewPhoto).Methods("DELETE")

	// Google Maps (proxied through backend - public with rate limiting)
	api.Handle("/places/search", middleware.OptionalAuthMiddleware(middleware.WithUserRoles(http.HandlerFunc(handlers.SearchPlaces)))).Methods("GET")
	api.Handle("/places/{placeId}", middleware.OptionalAuthMiddleware(middleware.WithUserRoles(http.HandlerFunc(handlers.GetPlaceDetails)))).Methods("GET")
	api.Handle("/geocode/cities", middleware.OptionalAuthMiddleware(middleware.WithUserRoles(http.HandlerFunc(handlers.GeocodeCities)))).Methods("GET")

	// Restaurant Suggestions (requires auth)
	suggestionsProtected := api.PathPrefix("/suggestions").Subrouter()
	suggestionsProtected.Use(middleware.AuthMiddleware)
	suggestionsProtected.Use(middleware.WithUserRoles)
	suggestionsProtected.Use(middleware.RequirePermission("suggestions.read"))
	suggestionsProtected.HandleFunc("", handlers.GetSuggestions).Methods("GET")
	suggestionsProtected.HandleFunc("/{id}", handlers.GetSuggestion).Methods("GET")

	// Create suggestions
	suggestionsCreate := api.PathPrefix("/suggestions").Subrouter()
	suggestionsCreate.Use(middleware.AuthMiddleware)
	suggestionsCreate.Use(middleware.WithUserRoles)
	suggestionsCreate.Use(middleware.RequirePermission("suggestions.create"))
	suggestionsCreate.HandleFunc("", handlers.CreateSuggestion).Methods("POST")

	// Approve/Convert/Reject suggestions (moderator/admin only)
	suggestionsModerate := api.PathPrefix("/suggestions").Subrouter()
	suggestionsModerate.Use(middleware.AuthMiddleware)
	suggestionsModerate.Use(middleware.WithUserRoles)
	suggestionsModerate.Use(middleware.RequireAnyPermission("suggestions.approve", "suggestions.convert", "suggestions.reject"))
	suggestionsModerate.HandleFunc("/{id}/approve", handlers.ApproveSuggestion).Methods("POST")
	suggestionsModerate.HandleFunc("/{id}/reject", handlers.RejectSuggestion).Methods("POST")
	suggestionsModerate.HandleFunc("/{id}/status", handlers.UpdateSuggestionStatus).Methods("PATCH")
	suggestionsModerate.HandleFunc("/{id}/convert", handlers.ConvertSuggestion).Methods("POST")
	suggestionsModerate.HandleFunc("/{id}", handlers.DeleteSuggestion).Methods("DELETE")

	// Menu Photos (read public, write requires auth)
	api.Handle("/restaurants/{restaurantId}/photos", middleware.OptionalAuthMiddleware(middleware.WithUserRoles(http.HandlerFunc(handlers.GetMenuPhotos)))).Methods("GET")

	photosProtected := api.PathPrefix("").Subrouter()
	photosProtected.Use(middleware.AuthMiddleware)
	photosProtected.Use(middleware.WithUserRoles)
	photosProtected.Use(middleware.RequirePermission("photos.upload"))
	photosProtected.HandleFunc("/restaurants/{restaurantId}/photos", handlers.UploadMenuPhoto).Methods("POST")

	// Update/Delete photos (ownership checked in handler)
	photosModify := api.PathPrefix("/photos").Subrouter()
	photosModify.Use(middleware.AuthMiddleware)
	photosModify.Use(middleware.WithUserRoles)
	photosModify.HandleFunc("/{id}", handlers.UpdatePhotoCaption).Methods("PATCH")
	photosModify.HandleFunc("/{id}", handlers.DeleteMenuPhoto).Methods("DELETE")

	// Restaurant Lists (authenticated users only)
	listsProtected := api.PathPrefix("/lists").Subrouter()
	listsProtected.Use(middleware.AuthMiddleware)
	listsProtected.Use(middleware.WithUserRoles)
	listsProtected.HandleFunc("", handlers.GetUserLists).Methods("GET")
	listsProtected.HandleFunc("", handlers.CreateList).Methods("POST")
	listsProtected.HandleFunc("/{id}", handlers.GetList).Methods("GET")
	listsProtected.HandleFunc("/{id}", handlers.UpdateList).Methods("PUT")
	listsProtected.HandleFunc("/{id}", handlers.DeleteList).Methods("DELETE")
	listsProtected.HandleFunc("/{id}/restaurants", handlers.AddRestaurantToList).Methods("POST")
	listsProtected.HandleFunc("/{id}/restaurants/{restaurantId}", handlers.RemoveRestaurantFromList).Methods("DELETE")

	// Get lists for a specific restaurant
	api.Handle("/restaurants/{restaurantId}/lists", middleware.OptionalAuthMiddleware(middleware.WithUserRoles(http.HandlerFunc(handlers.GetRestaurantLists)))).Methods("GET")

	// User Profile
	api.Handle("/users/{id}", middleware.OptionalAuthMiddleware(middleware.WithUserRoles(http.HandlerFunc(handlers.GetUserProfile)))).Methods("GET")
	api.Handle("/users/{id}/reviews", middleware.OptionalAuthMiddleware(middleware.WithUserRoles(http.HandlerFunc(handlers.GetUserReviews)))).Methods("GET")

	// Update profile (authenticated)
	userProtected := api.PathPrefix("/user").Subrouter()
	userProtected.Use(middleware.AuthMiddleware)
	userProtected.HandleFunc("/profile", handlers.UpdateUserProfile).Methods("PUT")

	// Admin Routes (requires admin permissions)
	adminRoutes := api.PathPrefix("/admin").Subrouter()
	adminRoutes.Use(middleware.AuthMiddleware)
	adminRoutes.Use(middleware.WithUserRoles)
	adminRoutes.Use(middleware.RequireRole("admin"))

	// User Management
	adminRoutes.HandleFunc("/users", handlers.AdminListUsers).Methods("GET")
	adminRoutes.HandleFunc("/users", handlers.AdminCreateUser).Methods("POST")
	adminRoutes.HandleFunc("/users/{id}", handlers.AdminGetUser).Methods("GET")
	adminRoutes.HandleFunc("/users/{id}", handlers.AdminUpdateUser).Methods("PUT")
	adminRoutes.HandleFunc("/users/{id}", handlers.AdminDeleteUser).Methods("DELETE")
	adminRoutes.HandleFunc("/users/{id}/roles", handlers.AdminAssignRole).Methods("POST")
	adminRoutes.HandleFunc("/users/{id}/roles/{roleId}", handlers.AdminRemoveRole).Methods("DELETE")
	adminRoutes.HandleFunc("/users/{id}/reset-password", handlers.AdminResetPassword).Methods("POST")

	// Role Management
	adminRoutes.HandleFunc("/roles", handlers.AdminListRoles).Methods("GET")
	adminRoutes.HandleFunc("/roles", handlers.AdminCreateRole).Methods("POST")
	adminRoutes.HandleFunc("/roles/{id}", handlers.AdminGetRole).Methods("GET")
	adminRoutes.HandleFunc("/roles/{id}", handlers.AdminUpdateRole).Methods("PUT")
	adminRoutes.HandleFunc("/roles/{id}", handlers.AdminDeleteRole).Methods("DELETE")
	adminRoutes.HandleFunc("/roles/{id}/permissions", handlers.AdminAssignPermission).Methods("POST")
	adminRoutes.HandleFunc("/roles/{id}/permissions/{permissionId}", handlers.AdminRemovePermission).Methods("DELETE")

	// Permission Management
	adminRoutes.HandleFunc("/permissions", handlers.AdminListPermissions).Methods("GET")

	// System Statistics and Analytics
	adminRoutes.HandleFunc("/stats", handlers.AdminGetStatistics).Methods("GET")
	adminRoutes.HandleFunc("/analytics/user-growth", handlers.AdminGetUserGrowth).Methods("GET")
	adminRoutes.HandleFunc("/analytics/active-users", handlers.AdminGetActiveUsers).Methods("GET")
	adminRoutes.HandleFunc("/analytics/popular-restaurants", handlers.AdminGetPopularRestaurants).Methods("GET")

	// System Settings
	adminRoutes.HandleFunc("/settings", handlers.AdminGetSettings).Methods("GET")

	// Audit Logs
	adminRoutes.HandleFunc("/audit-logs", handlers.AdminGetAuditLogs).Methods("GET")

	// Content Moderation
	adminRoutes.HandleFunc("/ratings", handlers.AdminListRatings).Methods("GET")
	adminRoutes.HandleFunc("/ratings/{id}", handlers.AdminDeleteRating).Methods("DELETE")
	adminRoutes.HandleFunc("/photos", handlers.AdminListPhotos).Methods("GET")
	adminRoutes.HandleFunc("/photos/{type}/{id}", handlers.AdminDeletePhoto).Methods("DELETE")

	// Restaurant Moderation
	adminRoutes.HandleFunc("/restaurants/{id}", handlers.AdminUpdateRestaurant).Methods("PUT")
	adminRoutes.HandleFunc("/restaurants/{id}", handlers.AdminDeleteRestaurant).Methods("DELETE")

	// Health check (support both GET and HEAD for Docker healthcheck)
	api.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if r.Method == "GET" {
			_, _ = w.Write([]byte("OK"))
		}
	}).Methods("GET", "HEAD")

	// Metrics endpoint
	api.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		stats := middleware.GetMetrics().GetStats()
		json.NewEncoder(w).Encode(stats)
	}).Methods("GET")

	// Serve the swagger.yaml file first
	api.HandleFunc("/swagger.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-yaml")
		http.ServeFile(w, r, "./docs/swagger.yaml")
	}).Methods("GET")

	// Redirect /docs to /docs/ for Swagger UI
	api.HandleFunc("/docs", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/api/docs/", http.StatusMovedPermanently)
	}).Methods("GET")

	// Swagger UI - serve at /api/docs/ (must be after swagger.yaml)
	api.PathPrefix("/docs/").Handler(httpSwagger.Handler(
		httpSwagger.URL("/api/swagger.yaml"),
	))

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

	if err := http.ListenAndServe(":"+port, handler); err != nil {
		logger.Fatal("Server failed to start: %v", err)
	}
}
