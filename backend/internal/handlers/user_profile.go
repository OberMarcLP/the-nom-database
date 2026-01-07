package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/nomdb/backend/internal/database"
	"github.com/nomdb/backend/internal/models"
)

// GetUserProfile returns a user's profile with stats
func GetUserProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["id"]

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	// Get user basic info
	var user models.User
	err = database.GetPool().QueryRow(ctx, `
		SELECT id, username, email, full_name, avatar_url, created_at, updated_at
		FROM users WHERE id = $1
	`, userID).Scan(
		&user.ID, &user.Username, &user.Email, &user.FullName,
		&user.AvatarURL, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Get user stats
	var stats struct {
		TotalReviews      int     `json:"total_reviews"`
		TotalRestaurants  int     `json:"total_restaurants"`
		TotalLists        int     `json:"total_lists"`
		AvgFoodRating     float64 `json:"avg_food_rating"`
		AvgServiceRating  float64 `json:"avg_service_rating"`
		AvgAmbianceRating float64 `json:"avg_ambiance_rating"`
	}

	// Get review counts and averages
	err = database.GetPool().QueryRow(ctx, `
		SELECT
			COUNT(*) as total_reviews,
			COUNT(DISTINCT restaurant_id) as total_restaurants,
			COALESCE(AVG(food_rating), 0) as avg_food,
			COALESCE(AVG(service_rating), 0) as avg_service,
			COALESCE(AVG(ambiance_rating), 0) as avg_ambiance
		FROM ratings
		WHERE user_id = $1
	`, userID).Scan(
		&stats.TotalReviews, &stats.TotalRestaurants,
		&stats.AvgFoodRating, &stats.AvgServiceRating, &stats.AvgAmbianceRating,
	)
	if err != nil {
		http.Error(w, "Failed to fetch user stats", http.StatusInternalServerError)
		return
	}

	// Get list count
	err = database.GetPool().QueryRow(ctx, `
		SELECT COUNT(*) FROM restaurant_lists WHERE user_id = $1
	`, userID).Scan(&stats.TotalLists)
	if err != nil {
		http.Error(w, "Failed to fetch list count", http.StatusInternalServerError)
		return
	}

	response := struct {
		User  models.User `json:"user"`
		Stats interface{} `json:"stats"`
	}{
		User:  user,
		Stats: stats,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetUserReviews returns all reviews by a user
func GetUserReviews(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["id"]

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	query := `
		SELECT
			r.id, r.restaurant_id, r.user_id, r.food_rating, r.service_rating,
			r.ambiance_rating, r.comment, r.helpful_count, r.not_helpful_count,
			r.created_at, r.updated_at,
			rest.id, rest.name, rest.address,
			u.id, u.username, u.full_name, u.avatar_url
		FROM ratings r
		JOIN restaurants rest ON r.restaurant_id = rest.id
		LEFT JOIN users u ON r.user_id = u.id
		WHERE r.user_id = $1
		ORDER BY r.created_at DESC
	`

	rows, err := database.GetPool().Query(ctx, query, userID)
	if err != nil {
		http.Error(w, "Failed to fetch reviews", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	reviews := []models.Rating{}
	for rows.Next() {
		var rating models.Rating
		var restaurant models.Restaurant
		var user models.UserSummary

		err := rows.Scan(
			&rating.ID, &rating.RestaurantID, &rating.UserID, &rating.FoodRating,
			&rating.ServiceRating, &rating.AmbianceRating, &rating.Comment,
			&rating.HelpfulCount, &rating.NotHelpfulCount, &rating.CreatedAt, &rating.UpdatedAt,
			&restaurant.ID, &restaurant.Name, &restaurant.Address,
			&user.ID, &user.Username, &user.FullName, &user.AvatarURL,
		)
		if err != nil {
			continue
		}

		// Attach restaurant info (limited)
		rating.User = &user
		reviews = append(reviews, rating)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reviews)
}

// UpdateUserProfile updates user profile information
func UpdateUserProfile(w http.ResponseWriter, r *http.Request) {
	user, ok := GetUserFromContext(r)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Username *string `json:"username"`
		FullName *string `json:"full_name"`
		Email    *string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	// Build dynamic update query
	updates := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.Username != nil {
		updates = append(updates, "username = $"+strconv.Itoa(argIndex))
		args = append(args, *req.Username)
		argIndex++
	}
	if req.FullName != nil {
		updates = append(updates, "full_name = $"+strconv.Itoa(argIndex))
		args = append(args, *req.FullName)
		argIndex++
	}
	if req.Email != nil {
		updates = append(updates, "email = $"+strconv.Itoa(argIndex))
		args = append(args, *req.Email)
		argIndex++
	}

	if len(updates) == 0 {
		http.Error(w, "No fields to update", http.StatusBadRequest)
		return
	}

	// Add user ID as last parameter
	args = append(args, user.ID)

	query := "UPDATE users SET " + strings.Join(updates, ", ") +
		", updated_at = CURRENT_TIMESTAMP WHERE id = $" + strconv.Itoa(argIndex) +
		" RETURNING id, username, email, full_name, avatar_url, created_at, updated_at"

	var updatedUser models.User
	err := database.GetPool().QueryRow(ctx, query, args...).Scan(
		&updatedUser.ID, &updatedUser.Username, &updatedUser.Email,
		&updatedUser.FullName, &updatedUser.AvatarURL,
		&updatedUser.CreatedAt, &updatedUser.UpdatedAt,
	)
	if err != nil {
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedUser)
}
