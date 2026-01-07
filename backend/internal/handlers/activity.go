package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/nomdb/backend/internal/database"
	"github.com/nomdb/backend/internal/models"
)

// ActivityItem represents a single activity in the feed
type ActivityItem struct {
	Type       string              `json:"type"` // "review" or "restaurant"
	Timestamp  string              `json:"timestamp"`
	User       *models.UserSummary `json:"user,omitempty"`
	Restaurant *struct {
		ID      int     `json:"id"`
		Name    string  `json:"name"`
		Address *string `json:"address,omitempty"`
	} `json:"restaurant,omitempty"`
	Review *struct {
		ID              int     `json:"id"`
		FoodRating      int     `json:"food_rating"`
		ServiceRating   int     `json:"service_rating"`
		AmbianceRating  int     `json:"ambiance_rating"`
		Comment         *string `json:"comment,omitempty"`
	} `json:"review,omitempty"`
}

// GetActivityFeed returns recent activity including reviews and new restaurants
func GetActivityFeed(w http.ResponseWriter, r *http.Request) {
	// Get limit from query params (default 20, max 50)
	limitStr := r.URL.Query().Get("limit")
	limit := 20
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	ctx := context.Background()

	// Query to get recent reviews and restaurants combined
	query := `
		WITH recent_reviews AS (
			SELECT
				'review' as type,
				rat.created_at as timestamp,
				u.id as user_id,
				u.username as user_username,
				u.full_name as user_full_name,
				u.avatar_url as user_avatar_url,
				r.id as restaurant_id,
				r.name as restaurant_name,
				r.address as restaurant_address,
				rat.id as review_id,
				rat.food_rating,
				rat.service_rating,
				rat.ambiance_rating,
				rat.comment
			FROM ratings rat
			JOIN restaurants r ON rat.restaurant_id = r.id
			LEFT JOIN users u ON rat.user_id = u.id
			ORDER BY rat.created_at DESC
			LIMIT $1
		),
		recent_restaurants AS (
			SELECT
				'restaurant' as type,
				r.created_at as timestamp,
				cu.id as user_id,
				cu.username as user_username,
				cu.full_name as user_full_name,
				cu.avatar_url as user_avatar_url,
				r.id as restaurant_id,
				r.name as restaurant_name,
				r.address as restaurant_address,
				NULL::integer as review_id,
				NULL::integer as food_rating,
				NULL::integer as service_rating,
				NULL::integer as ambiance_rating,
				NULL::text as comment
			FROM restaurants r
			LEFT JOIN users cu ON r.created_by = cu.id
			ORDER BY r.created_at DESC
			LIMIT $1
		)
		SELECT * FROM (
			SELECT * FROM recent_reviews
			UNION ALL
			SELECT * FROM recent_restaurants
		) combined
		ORDER BY timestamp DESC
		LIMIT $1
	`

	rows, err := database.GetPool().Query(ctx, query, limit)
	if err != nil {
		http.Error(w, "Failed to fetch activity feed", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	activities := []ActivityItem{}
	for rows.Next() {
		var activity ActivityItem
		var userID *int
		var userUsername, userFullName, userAvatarURL *string
		var restaurantID int
		var restaurantName string
		var restaurantAddress *string
		var reviewID, foodRating, serviceRating, ambianceRating *int
		var comment *string

		err := rows.Scan(
			&activity.Type,
			&activity.Timestamp,
			&userID,
			&userUsername,
			&userFullName,
			&userAvatarURL,
			&restaurantID,
			&restaurantName,
			&restaurantAddress,
			&reviewID,
			&foodRating,
			&serviceRating,
			&ambianceRating,
			&comment,
		)
		if err != nil {
			continue
		}

		// Build user summary if user exists
		if userID != nil && userUsername != nil {
			activity.User = &models.UserSummary{
				ID:        *userID,
				Username:  *userUsername,
				FullName:  userFullName,
				AvatarURL: userAvatarURL,
			}
		}

		// Build restaurant info
		activity.Restaurant = &struct {
			ID      int     `json:"id"`
			Name    string  `json:"name"`
			Address *string `json:"address,omitempty"`
		}{
			ID:      restaurantID,
			Name:    restaurantName,
			Address: restaurantAddress,
		}

		// Build review info if this is a review activity
		if activity.Type == "review" && reviewID != nil {
			activity.Review = &struct {
				ID              int     `json:"id"`
				FoodRating      int     `json:"food_rating"`
				ServiceRating   int     `json:"service_rating"`
				AmbianceRating  int     `json:"ambiance_rating"`
				Comment         *string `json:"comment,omitempty"`
			}{
				ID:              *reviewID,
				FoodRating:      *foodRating,
				ServiceRating:   *serviceRating,
				AmbianceRating:  *ambianceRating,
				Comment:         comment,
			}
		}

		activities = append(activities, activity)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(activities)
}
