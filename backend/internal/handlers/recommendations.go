package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/nomdb/backend/internal/database"
	"github.com/nomdb/backend/internal/models"
)

// GetSimilarRestaurants returns restaurants similar to the given restaurant
// based on shared categories and food types
func GetSimilarRestaurants(w http.ResponseWriter, r *http.Request) {
	// Get restaurant ID from query params
	restaurantIDStr := r.URL.Query().Get("restaurant_id")
	if restaurantIDStr == "" {
		http.Error(w, "restaurant_id is required", http.StatusBadRequest)
		return
	}

	restaurantID, err := strconv.Atoi(restaurantIDStr)
	if err != nil {
		http.Error(w, "Invalid restaurant_id", http.StatusBadRequest)
		return
	}

	// Get limit from query params (default 6)
	limitStr := r.URL.Query().Get("limit")
	limit := 6
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	ctx := context.Background()

	// Find similar restaurants based on category and food types
	query := `
		WITH target_restaurant AS (
			SELECT category_id, ARRAY_AGG(food_type_id) as food_type_ids
			FROM restaurants r
			LEFT JOIN restaurant_food_types rft ON r.id = rft.restaurant_id
			WHERE r.id = $1
			GROUP BY r.id, category_id
		)
		SELECT DISTINCT
			r.id,
			r.name,
			r.address,
			r.category_id,
			r.price_range,
			r.google_place_id,
			r.latitude,
			r.longitude,
			r.created_at,
			r.updated_at,
			COALESCE(AVG(rat.food_rating), 0) as avg_food_rating,
			COALESCE(AVG(rat.service_rating), 0) as avg_service_rating,
			COALESCE(AVG(rat.ambiance_rating), 0) as avg_ambiance_rating,
			COUNT(rat.id) as rating_count,
			c.id as category_id,
			c.name as category_name,
			(
				SELECT COUNT(*)
				FROM restaurant_food_types rft2
				WHERE rft2.restaurant_id = r.id
				AND rft2.food_type_id = ANY((SELECT food_type_ids FROM target_restaurant))
			) as matching_food_types
		FROM restaurants r
		CROSS JOIN target_restaurant tr
		LEFT JOIN ratings rat ON r.id = rat.restaurant_id
		LEFT JOIN categories c ON r.category_id = c.id
		WHERE r.id != $1
		AND (
			r.category_id = tr.category_id
			OR EXISTS (
				SELECT 1 FROM restaurant_food_types rft
				WHERE rft.restaurant_id = r.id
				AND rft.food_type_id = ANY(tr.food_type_ids)
			)
		)
		GROUP BY r.id, r.name, r.address, r.category_id, r.price_range,
			r.google_place_id, r.latitude, r.longitude, r.created_at, r.updated_at,
			c.id, c.name
		ORDER BY matching_food_types DESC, AVG(rat.food_rating) DESC
		LIMIT $2
	`

	rows, err := database.GetPool().Query(ctx, query, restaurantID, limit)
	if err != nil {
		http.Error(w, "Failed to fetch similar restaurants", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	restaurants := []models.Restaurant{}
	for rows.Next() {
		var restaurant models.Restaurant
		var category models.Category
		var matchingFoodTypes int
		var avgFoodRating, avgServiceRating, avgAmbianceRating float64
		var ratingCount int

		err := rows.Scan(
			&restaurant.ID, &restaurant.Name, &restaurant.Address,
			&restaurant.CategoryID, &restaurant.PriceRange,
			&restaurant.GooglePlaceID, &restaurant.Latitude, &restaurant.Longitude,
			&restaurant.CreatedAt, &restaurant.UpdatedAt,
			&avgFoodRating, &avgServiceRating,
			&avgAmbianceRating, &ratingCount,
			&category.ID, &category.Name,
			&matchingFoodTypes,
		)
		if err != nil {
			continue
		}

		restaurant.Category = &category
		restaurant.AvgRating = &models.AvgRating{
			Food:     avgFoodRating,
			Service:  avgServiceRating,
			Ambiance: avgAmbianceRating,
			Overall:  (avgFoodRating + avgServiceRating + avgAmbianceRating) / 3.0,
			Count:    ratingCount,
		}
		restaurants = append(restaurants, restaurant)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(restaurants)
}

// GetPopularRestaurants returns the most popular restaurants based on ratings
func GetPopularRestaurants(w http.ResponseWriter, r *http.Request) {
	// Get limit from query params (default 10)
	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// Get minimum ratings count (default 3)
	minRatingsStr := r.URL.Query().Get("min_ratings")
	minRatings := 3
	if minRatingsStr != "" {
		if mr, err := strconv.Atoi(minRatingsStr); err == nil && mr > 0 {
			minRatings = mr
		}
	}

	ctx := context.Background()

	query := `
		SELECT
			r.id,
			r.name,
			r.address,
			r.category_id,
			r.price_range,
			r.google_place_id,
			r.latitude,
			r.longitude,
			r.created_at,
			r.updated_at,
			COALESCE(AVG(rat.food_rating), 0) as avg_food_rating,
			COALESCE(AVG(rat.service_rating), 0) as avg_service_rating,
			COALESCE(AVG(rat.ambiance_rating), 0) as avg_ambiance_rating,
			COUNT(rat.id) as rating_count,
			c.id as category_id,
			c.name as category_name
		FROM restaurants r
		LEFT JOIN ratings rat ON r.id = rat.restaurant_id
		LEFT JOIN categories c ON r.category_id = c.id
		GROUP BY r.id, r.name, r.address, r.category_id, r.price_range,
			r.google_place_id, r.latitude, r.longitude, r.created_at, r.updated_at,
			c.id, c.name
		HAVING COUNT(rat.id) >= $1
		ORDER BY AVG((rat.food_rating + rat.service_rating + rat.ambiance_rating) / 3.0) DESC,
			COUNT(rat.id) DESC
		LIMIT $2
	`

	rows, err := database.GetPool().Query(ctx, query, minRatings, limit)
	if err != nil {
		http.Error(w, "Failed to fetch popular restaurants", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	restaurants := []models.Restaurant{}
	for rows.Next() {
		var restaurant models.Restaurant
		var category models.Category
		var avgFoodRating, avgServiceRating, avgAmbianceRating float64
		var ratingCount int

		err := rows.Scan(
			&restaurant.ID, &restaurant.Name, &restaurant.Address,
			&restaurant.CategoryID, &restaurant.PriceRange,
			&restaurant.GooglePlaceID, &restaurant.Latitude, &restaurant.Longitude,
			&restaurant.CreatedAt, &restaurant.UpdatedAt,
			&avgFoodRating, &avgServiceRating,
			&avgAmbianceRating, &ratingCount,
			&category.ID, &category.Name,
		)
		if err != nil {
			continue
		}

		restaurant.Category = &category
		restaurant.AvgRating = &models.AvgRating{
			Food:     avgFoodRating,
			Service:  avgServiceRating,
			Ambiance: avgAmbianceRating,
			Overall:  (avgFoodRating + avgServiceRating + avgAmbianceRating) / 3.0,
			Count:    ratingCount,
		}
		restaurants = append(restaurants, restaurant)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(restaurants)
}

// GetRecentRestaurants returns the most recently added restaurants
func GetRecentRestaurants(w http.ResponseWriter, r *http.Request) {
	// Get limit from query params (default 10)
	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	ctx := context.Background()

	query := `
		SELECT
			r.id,
			r.name,
			r.address,
			r.category_id,
			r.price_range,
			r.google_place_id,
			r.latitude,
			r.longitude,
			r.created_at,
			r.updated_at,
			COALESCE(AVG(rat.food_rating), 0) as avg_food_rating,
			COALESCE(AVG(rat.service_rating), 0) as avg_service_rating,
			COALESCE(AVG(rat.ambiance_rating), 0) as avg_ambiance_rating,
			COUNT(rat.id) as rating_count,
			c.id as category_id,
			c.name as category_name
		FROM restaurants r
		LEFT JOIN ratings rat ON r.id = rat.restaurant_id
		LEFT JOIN categories c ON r.category_id = c.id
		GROUP BY r.id, r.name, r.address, r.category_id, r.price_range,
			r.google_place_id, r.latitude, r.longitude, r.created_at, r.updated_at,
			c.id, c.name
		ORDER BY r.created_at DESC
		LIMIT $1
	`

	rows, err := database.GetPool().Query(ctx, query, limit)
	if err != nil {
		http.Error(w, "Failed to fetch recent restaurants", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	restaurants := []models.Restaurant{}
	for rows.Next() {
		var restaurant models.Restaurant
		var category models.Category
		var avgFoodRating, avgServiceRating, avgAmbianceRating float64
		var ratingCount int

		err := rows.Scan(
			&restaurant.ID, &restaurant.Name, &restaurant.Address,
			&restaurant.CategoryID, &restaurant.PriceRange,
			&restaurant.GooglePlaceID, &restaurant.Latitude, &restaurant.Longitude,
			&restaurant.CreatedAt, &restaurant.UpdatedAt,
			&avgFoodRating, &avgServiceRating,
			&avgAmbianceRating, &ratingCount,
			&category.ID, &category.Name,
		)
		if err != nil {
			continue
		}

		restaurant.Category = &category
		restaurant.AvgRating = &models.AvgRating{
			Food:     avgFoodRating,
			Service:  avgServiceRating,
			Ambiance: avgAmbianceRating,
			Overall:  (avgFoodRating + avgServiceRating + avgAmbianceRating) / 3.0,
			Count:    ratingCount,
		}
		restaurants = append(restaurants, restaurant)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(restaurants)
}

// GetPersonalizedRecommendations returns personalized recommendations based on user's review history
func GetPersonalizedRecommendations(w http.ResponseWriter, r *http.Request) {
	user, ok := GetUserFromContext(r)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get limit from query params (default 10)
	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	ctx := context.Background()

	// Find restaurants similar to those the user has rated highly
	query := `
		WITH user_preferences AS (
			SELECT
				r.category_id,
				ARRAY_AGG(DISTINCT rft.food_type_id) as preferred_food_types,
				AVG((rat.food_rating + rat.service_rating + rat.ambiance_rating) / 3.0) as avg_user_rating
			FROM ratings rat
			JOIN restaurants r ON rat.restaurant_id = r.id
			LEFT JOIN restaurant_food_types rft ON r.id = rft.restaurant_id
			WHERE rat.user_id = $1
			AND (rat.food_rating + rat.service_rating + rat.ambiance_rating) / 3.0 >= 4.0
			GROUP BY r.category_id
		),
		user_reviewed_restaurants AS (
			SELECT restaurant_id FROM ratings WHERE user_id = $1
		)
		SELECT DISTINCT
			r.id,
			r.name,
			r.address,
			r.category_id,
			r.price_range,
			r.google_place_id,
			r.latitude,
			r.longitude,
			r.created_at,
			r.updated_at,
			COALESCE(AVG(rat.food_rating), 0) as avg_food_rating,
			COALESCE(AVG(rat.service_rating), 0) as avg_service_rating,
			COALESCE(AVG(rat.ambiance_rating), 0) as avg_ambiance_rating,
			COUNT(rat.id) as rating_count,
			c.id as category_id,
			c.name as category_name
		FROM restaurants r
		JOIN user_preferences up ON r.category_id = up.category_id
		LEFT JOIN ratings rat ON r.id = rat.restaurant_id
		LEFT JOIN categories c ON r.category_id = c.id
		LEFT JOIN restaurant_food_types rft ON r.id = rft.restaurant_id
		WHERE r.id NOT IN (SELECT restaurant_id FROM user_reviewed_restaurants)
		AND (
			up.preferred_food_types IS NULL
			OR rft.food_type_id = ANY(up.preferred_food_types)
		)
		GROUP BY r.id, r.name, r.address, r.category_id, r.price_range,
			r.google_place_id, r.latitude, r.longitude, r.created_at, r.updated_at,
			c.id, c.name
		ORDER BY AVG(rat.food_rating) DESC, COUNT(rat.id) DESC
		LIMIT $2
	`

	rows, err := database.GetPool().Query(ctx, query, user.ID, limit)
	if err != nil {
		http.Error(w, "Failed to fetch personalized recommendations", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	restaurants := []models.Restaurant{}
	for rows.Next() {
		var restaurant models.Restaurant
		var category models.Category
		var avgFoodRating, avgServiceRating, avgAmbianceRating float64
		var ratingCount int

		err := rows.Scan(
			&restaurant.ID, &restaurant.Name, &restaurant.Address,
			&restaurant.CategoryID, &restaurant.PriceRange,
			&restaurant.GooglePlaceID, &restaurant.Latitude, &restaurant.Longitude,
			&restaurant.CreatedAt, &restaurant.UpdatedAt,
			&avgFoodRating, &avgServiceRating,
			&avgAmbianceRating, &ratingCount,
			&category.ID, &category.Name,
		)
		if err != nil {
			continue
		}

		restaurant.Category = &category
		restaurant.AvgRating = &models.AvgRating{
			Food:     avgFoodRating,
			Service:  avgServiceRating,
			Ambiance: avgAmbianceRating,
			Overall:  (avgFoodRating + avgServiceRating + avgAmbianceRating) / 3.0,
			Count:    ratingCount,
		}
		restaurants = append(restaurants, restaurant)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(restaurants)
}
