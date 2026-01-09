package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/nomdb/backend/internal/database"
	"github.com/nomdb/backend/internal/logger"
)

// @Summary Get system statistics
// @Description Get comprehensive system statistics for admin dashboard
// @Tags Admin
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {string} string "Internal server error"
// @Security BearerAuth
// @Router /admin/stats [get]
func AdminGetStatistics(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	stats := make(map[string]interface{})

	// Get user statistics
	var totalUsers, activeUsers, verifiedUsers int
	database.GetPool().QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&totalUsers)
	database.GetPool().QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE is_active = true").Scan(&activeUsers)
	database.GetPool().QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE email_verified = true").Scan(&verifiedUsers)

	// Get user growth (last 30 days)
	var newUsersLast30Days int
	database.GetPool().QueryRow(ctx,
		"SELECT COUNT(*) FROM users WHERE created_at > NOW() - INTERVAL '30 days'").Scan(&newUsersLast30Days)

	stats["users"] = map[string]interface{}{
		"total":              totalUsers,
		"active":             activeUsers,
		"verified":           verifiedUsers,
		"new_last_30_days":   newUsersLast30Days,
	}

	// Get restaurant statistics
	var totalRestaurants, pendingSuggestions, approvedSuggestions, rejectedSuggestions int
	database.GetPool().QueryRow(ctx, "SELECT COUNT(*) FROM restaurants").Scan(&totalRestaurants)
	database.GetPool().QueryRow(ctx, "SELECT COUNT(*) FROM restaurant_suggestions WHERE status = 'pending'").Scan(&pendingSuggestions)
	database.GetPool().QueryRow(ctx, "SELECT COUNT(*) FROM restaurant_suggestions WHERE status = 'approved'").Scan(&approvedSuggestions)
	database.GetPool().QueryRow(ctx, "SELECT COUNT(*) FROM restaurant_suggestions WHERE status = 'rejected'").Scan(&rejectedSuggestions)

	stats["restaurants"] = map[string]interface{}{
		"total":               totalRestaurants,
		"pending_suggestions": pendingSuggestions,
		"approved_suggestions": approvedSuggestions,
		"rejected_suggestions": rejectedSuggestions,
	}

	// Get rating statistics
	var totalRatings int
	var avgFoodRating, avgServiceRating, avgAmbianceRating float64
	database.GetPool().QueryRow(ctx, "SELECT COUNT(*) FROM ratings").Scan(&totalRatings)
	database.GetPool().QueryRow(ctx, "SELECT COALESCE(AVG(food_rating), 0) FROM ratings").Scan(&avgFoodRating)
	database.GetPool().QueryRow(ctx, "SELECT COALESCE(AVG(service_rating), 0) FROM ratings").Scan(&avgServiceRating)
	database.GetPool().QueryRow(ctx, "SELECT COALESCE(AVG(ambiance_rating), 0) FROM ratings").Scan(&avgAmbianceRating)

	stats["ratings"] = map[string]interface{}{
		"total":           totalRatings,
		"avg_food":        avgFoodRating,
		"avg_service":     avgServiceRating,
		"avg_ambiance":    avgAmbianceRating,
	}

	// Get content statistics
	var totalMenuPhotos, totalReviewPhotos, totalLists int
	database.GetPool().QueryRow(ctx, "SELECT COUNT(*) FROM menu_photos").Scan(&totalMenuPhotos)
	database.GetPool().QueryRow(ctx, "SELECT COUNT(*) FROM review_photos").Scan(&totalReviewPhotos)
	database.GetPool().QueryRow(ctx, "SELECT COUNT(*) FROM restaurant_lists").Scan(&totalLists)

	stats["content"] = map[string]interface{}{
		"menu_photos":   totalMenuPhotos,
		"review_photos": totalReviewPhotos,
		"lists":         totalLists,
	}

	// Get activity statistics
	var totalActivities int
	database.GetPool().QueryRow(ctx, "SELECT COUNT(*) FROM activities WHERE created_at > NOW() - INTERVAL '7 days'").Scan(&totalActivities)

	stats["activity"] = map[string]interface{}{
		"last_7_days": totalActivities,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// @Summary Get user growth analytics
// @Description Get user growth data over time
// @Tags Admin
// @Accept json
// @Produce json
// @Param days query int false "Number of days to retrieve" default(30)
// @Success 200 {array} map[string]interface{}
// @Failure 500 {string} string "Internal server error"
// @Security BearerAuth
// @Router /admin/analytics/user-growth [get]
func AdminGetUserGrowth(w http.ResponseWriter, r *http.Request) {
	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	if days < 1 || days > 365 {
		days = 30
	}

	ctx := context.Background()

	query := `
		SELECT DATE(created_at) as date, COUNT(*) as count
		FROM users
		WHERE created_at > NOW() - INTERVAL '` + strconv.Itoa(days) + ` days'
		GROUP BY DATE(created_at)
		ORDER BY date ASC
	`

	rows, err := database.GetPool().Query(ctx, query)
	if err != nil {
		logger.Error("Failed to get user growth: %v", err)
		http.Error(w, "Failed to get user growth data", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var data []map[string]interface{}
	for rows.Next() {
		var date time.Time
		var count int
		if err := rows.Scan(&date, &count); err != nil {
			logger.Error("Failed to scan user growth row: %v", err)
			continue
		}
		data = append(data, map[string]interface{}{
			"date":  date.Format("2006-01-02"),
			"count": count,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// @Summary Get most active users
// @Description Get list of most active users by various metrics
// @Tags Admin
// @Accept json
// @Produce json
// @Param limit query int false "Number of users to retrieve" default(10)
// @Success 200 {array} map[string]interface{}
// @Failure 500 {string} string "Internal server error"
// @Security BearerAuth
// @Router /admin/analytics/active-users [get]
func AdminGetActiveUsers(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 10
	}

	ctx := context.Background()

	query := `
		SELECT u.id, u.username, u.email, u.avatar_url,
		       COUNT(DISTINCT ra.id) as rating_count,
		       COUNT(DISTINCT res.id) as restaurant_count,
		       COUNT(DISTINCT rl.id) as list_count
		FROM users u
		LEFT JOIN ratings ra ON u.id = ra.user_id
		LEFT JOIN restaurants res ON u.id = res.user_id
		LEFT JOIN restaurant_lists rl ON u.id = rl.user_id
		WHERE u.is_active = true
		GROUP BY u.id, u.username, u.email, u.avatar_url
		ORDER BY (COUNT(DISTINCT ra.id) + COUNT(DISTINCT res.id) + COUNT(DISTINCT rl.id)) DESC
		LIMIT $1
	`

	rows, err := database.GetPool().Query(ctx, query, limit)
	if err != nil {
		logger.Error("Failed to get active users: %v", err)
		http.Error(w, "Failed to get active users", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []map[string]interface{}
	for rows.Next() {
		var id int
		var username, email string
		var avatarURL *string
		var ratingCount, restaurantCount, listCount int

		if err := rows.Scan(&id, &username, &email, &avatarURL, &ratingCount, &restaurantCount, &listCount); err != nil {
			logger.Error("Failed to scan active user row: %v", err)
			continue
		}

		user := map[string]interface{}{
			"id":               id,
			"username":         username,
			"email":            email,
			"rating_count":     ratingCount,
			"restaurant_count": restaurantCount,
			"list_count":       listCount,
			"total_activity":   ratingCount + restaurantCount + listCount,
		}

		if avatarURL != nil {
			user["avatar_url"] = *avatarURL
		}

		users = append(users, user)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// @Summary Get popular restaurants
// @Description Get list of most popular restaurants by rating count
// @Tags Admin
// @Accept json
// @Produce json
// @Param limit query int false "Number of restaurants to retrieve" default(10)
// @Success 200 {array} map[string]interface{}
// @Failure 500 {string} string "Internal server error"
// @Security BearerAuth
// @Router /admin/analytics/popular-restaurants [get]
func AdminGetPopularRestaurants(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 10
	}

	ctx := context.Background()

	query := `
		SELECT r.id, r.name, r.address, c.name as category,
		       COUNT(ra.id) as rating_count,
		       COALESCE(AVG((ra.food_rating + ra.service_rating + ra.ambiance_rating) / 3.0), 0) as avg_rating
		FROM restaurants r
		LEFT JOIN categories c ON r.category_id = c.id
		LEFT JOIN ratings ra ON r.id = ra.restaurant_id
		GROUP BY r.id, r.name, r.address, c.name
		ORDER BY rating_count DESC, avg_rating DESC
		LIMIT $1
	`

	rows, err := database.GetPool().Query(ctx, query, limit)
	if err != nil {
		logger.Error("Failed to get popular restaurants: %v", err)
		http.Error(w, "Failed to get popular restaurants", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var restaurants []map[string]interface{}
	for rows.Next() {
		var id int
		var name string
		var address, category *string
		var ratingCount int
		var avgRating float64

		if err := rows.Scan(&id, &name, &address, &category, &ratingCount, &avgRating); err != nil {
			logger.Error("Failed to scan popular restaurant row: %v", err)
			continue
		}

		restaurant := map[string]interface{}{
			"id":           id,
			"name":         name,
			"rating_count": ratingCount,
			"avg_rating":   avgRating,
		}

		if address != nil {
			restaurant["address"] = *address
		}
		if category != nil {
			restaurant["category"] = *category
		}

		restaurants = append(restaurants, restaurant)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(restaurants)
}

// @Summary Get audit logs
// @Description Get paginated audit logs with optional filtering
// @Tags Admin
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(50)
// @Param user_id query int false "Filter by user ID"
// @Param action query string false "Filter by action"
// @Param resource_type query string false "Filter by resource type"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {string} string "Internal server error"
// @Security BearerAuth
// @Router /admin/audit-logs [get]
func AdminGetAuditLogs(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 50
	}

	userIDStr := r.URL.Query().Get("user_id")
	action := r.URL.Query().Get("action")
	resourceType := r.URL.Query().Get("resource_type")

	offset := (page - 1) * limit
	ctx := context.Background()

	// Build query with filters
	query := `
		SELECT al.id, al.user_id, u.username, al.action, al.resource_type,
		       al.resource_id, al.details, al.ip_address, al.user_agent, al.created_at
		FROM audit_logs al
		LEFT JOIN users u ON al.user_id = u.id
		WHERE 1=1
	`

	args := []interface{}{}
	argPos := 1

	if userIDStr != "" {
		userID, err := strconv.Atoi(userIDStr)
		if err == nil {
			query += ` AND al.user_id = $` + strconv.Itoa(argPos)
			args = append(args, userID)
			argPos++
		}
	}

	if action != "" {
		query += ` AND al.action = $` + strconv.Itoa(argPos)
		args = append(args, action)
		argPos++
	}

	if resourceType != "" {
		query += ` AND al.resource_type = $` + strconv.Itoa(argPos)
		args = append(args, resourceType)
		argPos++
	}

	// Get total count
	countQuery := `SELECT COUNT(*) FROM audit_logs al WHERE 1=1`
	if userIDStr != "" {
		countQuery += ` AND al.user_id = $1`
	}
	if action != "" {
		countQuery += ` AND al.action = $` + strconv.Itoa(len(args))
	}
	if resourceType != "" {
		countQuery += ` AND al.resource_type = $` + strconv.Itoa(len(args))
	}

	var total int
	err := database.GetPool().QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		logger.Error("Failed to count audit logs: %v", err)
		http.Error(w, "Failed to count audit logs", http.StatusInternalServerError)
		return
	}

	// Get paginated logs
	query += ` ORDER BY al.created_at DESC LIMIT $` + strconv.Itoa(argPos) + ` OFFSET $` + strconv.Itoa(argPos+1)
	args = append(args, limit, offset)

	rows, err := database.GetPool().Query(ctx, query, args...)
	if err != nil {
		logger.Error("Failed to get audit logs: %v", err)
		http.Error(w, "Failed to get audit logs", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var logs []map[string]interface{}
	for rows.Next() {
		var id, resourceID int
		var userID *int
		var username *string
		var action, resourceType string
		var details, ipAddress, userAgent *string
		var createdAt time.Time

		err := rows.Scan(&id, &userID, &username, &action, &resourceType,
			&resourceID, &details, &ipAddress, &userAgent, &createdAt)
		if err != nil {
			logger.Error("Failed to scan audit log row: %v", err)
			continue
		}

		log := map[string]interface{}{
			"id":            id,
			"action":        action,
			"resource_type": resourceType,
			"resource_id":   resourceID,
			"created_at":    createdAt,
		}

		if userID != nil {
			log["user_id"] = *userID
		}
		if username != nil {
			log["username"] = *username
		}
		if details != nil {
			log["details"] = *details
		}
		if ipAddress != nil {
			log["ip_address"] = *ipAddress
		}
		if userAgent != nil {
			log["user_agent"] = *userAgent
		}

		logs = append(logs, log)
	}

	response := map[string]interface{}{
		"logs": logs,
		"pagination": map[string]interface{}{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"totalPages": (total + limit - 1) / limit,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// @Summary Get system settings
// @Description Get current system settings and configuration
// @Tags Admin
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /admin/settings [get]
func AdminGetSettings(w http.ResponseWriter, r *http.Request) {
	settings := map[string]interface{}{
		"auth_mode":       getEnvOrDefault("AUTH_MODE", "both"),
		"oidc_enabled":    getEnvOrDefault("OIDC_ISSUER_URL", "") != "",
		"s3_enabled":      getEnvOrDefault("S3_BUCKET_NAME", "") != "",
		"debug_mode":      getEnvOrDefault("DEBUG", "false") == "true",
		"allowed_origins": getEnvOrDefault("ALLOWED_ORIGINS", "http://localhost:3000"),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}

func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
