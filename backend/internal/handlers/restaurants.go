package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nomdb/backend/internal/database"
	"github.com/nomdb/backend/internal/logger"
	"github.com/nomdb/backend/internal/models"
)

func getFoodTypesForRestaurant(ctx context.Context, restaurantID int) ([]models.FoodType, error) {
	rows, err := database.GetPool().Query(ctx,
		`SELECT ft.id, ft.name, ft.created_at, ft.updated_at
		FROM food_types ft
		JOIN restaurant_food_types rft ON ft.id = rft.food_type_id
		WHERE rft.restaurant_id = $1
		ORDER BY ft.name`, restaurantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var foodTypes []models.FoodType
	for rows.Next() {
		var ft models.FoodType
		if err := rows.Scan(&ft.ID, &ft.Name, &ft.CreatedAt, &ft.UpdatedAt); err != nil {
			return nil, err
		}
		foodTypes = append(foodTypes, ft)
	}
	return foodTypes, nil
}

func getFoodTypesForSuggestion(ctx context.Context, suggestionID int) ([]models.FoodType, error) {
	rows, err := database.GetPool().Query(ctx,
		`SELECT ft.id, ft.name, ft.created_at, ft.updated_at
		FROM food_types ft
		JOIN suggestion_food_types sft ON ft.id = sft.food_type_id
		WHERE sft.suggestion_id = $1
		ORDER BY ft.name`, suggestionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var foodTypes []models.FoodType
	for rows.Next() {
		var ft models.FoodType
		if err := rows.Scan(&ft.ID, &ft.Name, &ft.CreatedAt, &ft.UpdatedAt); err != nil {
			return nil, err
		}
		foodTypes = append(foodTypes, ft)
	}
	return foodTypes, nil
}

func getFoodTypesForRestaurantsBatch(ctx context.Context, restaurantIDs []int) (map[int][]models.FoodType, error) {
	if len(restaurantIDs) == 0 {
		return make(map[int][]models.FoodType), nil
	}

	// Build dynamic query with placeholders
	placeholders := make([]string, len(restaurantIDs))
	args := make([]interface{}, len(restaurantIDs))
	for i, id := range restaurantIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT rft.restaurant_id, ft.id, ft.name, ft.created_at, ft.updated_at
		FROM food_types ft
		JOIN restaurant_food_types rft ON ft.id = rft.food_type_id
		WHERE rft.restaurant_id IN (%s)
		ORDER BY rft.restaurant_id, ft.name`, strings.Join(placeholders, ","))

	rows, err := database.GetPool().Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int][]models.FoodType)
	for rows.Next() {
		var restaurantID int
		var ft models.FoodType
		if err := rows.Scan(&restaurantID, &ft.ID, &ft.Name, &ft.CreatedAt, &ft.UpdatedAt); err != nil {
			return nil, err
		}
		result[restaurantID] = append(result[restaurantID], ft)
	}
	return result, nil
}

func getFoodTypesForSuggestionsBatch(ctx context.Context, suggestionIDs []int) (map[int][]models.FoodType, error) {
	if len(suggestionIDs) == 0 {
		return make(map[int][]models.FoodType), nil
	}

	// Build dynamic query with placeholders
	placeholders := make([]string, len(suggestionIDs))
	args := make([]interface{}, len(suggestionIDs))
	for i, id := range suggestionIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT sft.suggestion_id, ft.id, ft.name, ft.created_at, ft.updated_at
		FROM food_types ft
		JOIN suggestion_food_types sft ON ft.id = sft.food_type_id
		WHERE sft.suggestion_id IN (%s)
		ORDER BY sft.suggestion_id, ft.name`, strings.Join(placeholders, ","))

	rows, err := database.GetPool().Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int][]models.FoodType)
	for rows.Next() {
		var suggestionID int
		var ft models.FoodType
		if err := rows.Scan(&suggestionID, &ft.ID, &ft.Name, &ft.CreatedAt, &ft.UpdatedAt); err != nil {
			return nil, err
		}
		result[suggestionID] = append(result[suggestionID], ft)
	}
	return result, nil
}

func setFoodTypesForRestaurant(ctx context.Context, restaurantID int, foodTypeIDs []int) error {
	// Delete existing food types
	_, err := database.GetPool().Exec(ctx,
		"DELETE FROM restaurant_food_types WHERE restaurant_id = $1", restaurantID)
	if err != nil {
		return err
	}

	// Insert new food types
	for _, ftID := range foodTypeIDs {
		_, err := database.GetPool().Exec(ctx,
			"INSERT INTO restaurant_food_types (restaurant_id, food_type_id) VALUES ($1, $2)",
			restaurantID, ftID)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetRestaurants godoc
// @Summary List all restaurants
// @Description Get a list of all restaurants with optional filtering by category, food types, and location
// @Tags Restaurants
// @Accept json
// @Produce json
// @Param category_id query int false "Filter by category ID"
// @Param food_type_ids query string false "Filter by food type IDs (comma-separated)"
// @Param lat query number false "Latitude for distance filtering"
// @Param lng query number false "Longitude for distance filtering"
// @Param radius query number false "Radius in kilometers for distance filtering"
// @Success 200 {array} models.Restaurant "List of restaurants"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /restaurants [get]
func GetRestaurants(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := RequestContext(r)
	defer cancel()

	// Parse query parameters for filtering
	queryParams := r.URL.Query()
	searchQuery := queryParams.Get("q")        // search query for name/address
	categoryID := queryParams.Get("category_id")
	foodTypeIDs := queryParams.Get("food_type_ids") // comma-separated
	priceRange := queryParams.Get("price_range")    // max price range (1-4)
	minRating := queryParams.Get("min_rating")      // minimum average rating (1-5)
	sortBy := queryParams.Get("sort")               // sort: name, rating, date
	lat := queryParams.Get("lat")
	lng := queryParams.Get("lng")
	radius := queryParams.Get("radius") // in kilometers

	// Always include suggestions in the restaurant list
	includeSuggestions := true

	// Build dynamic query with filters using UNION to include both restaurants and suggestions
	var args []interface{}

	// Build restaurant query
	var restaurantConditions []string
	var restaurantArgs []interface{}
	restaurantArgIndex := 1

	// Search query filter
	if searchQuery != "" {
		restaurantConditions = append(restaurantConditions, fmt.Sprintf("(r.name ILIKE $%d OR r.address ILIKE $%d)", restaurantArgIndex, restaurantArgIndex))
		restaurantArgs = append(restaurantArgs, "%"+searchQuery+"%")
		restaurantArgIndex++
	}

	if categoryID != "" {
		if catID, err := strconv.Atoi(categoryID); err == nil {
			restaurantConditions = append(restaurantConditions, fmt.Sprintf("r.category_id = $%d", restaurantArgIndex))
			restaurantArgs = append(restaurantArgs, catID)
			restaurantArgIndex++
		}
	}

	// Price range filter
	if priceRange != "" {
		if pr, err := strconv.Atoi(priceRange); err == nil && pr >= 1 && pr <= 4 {
			restaurantConditions = append(restaurantConditions, fmt.Sprintf("(r.price_range IS NULL OR r.price_range <= $%d)", restaurantArgIndex))
			restaurantArgs = append(restaurantArgs, pr)
			restaurantArgIndex++
		}
	}

	if foodTypeIDs != "" {
		ftIDs := strings.Split(foodTypeIDs, ",")
		var validIDs []int
		for _, idStr := range ftIDs {
			if id, err := strconv.Atoi(strings.TrimSpace(idStr)); err == nil {
				validIDs = append(validIDs, id)
			}
		}
		if len(validIDs) > 0 {
			placeholders := make([]string, len(validIDs))
			for i, id := range validIDs {
				placeholders[i] = fmt.Sprintf("$%d", restaurantArgIndex)
				restaurantArgs = append(restaurantArgs, id)
				restaurantArgIndex++
			}
			restaurantConditions = append(restaurantConditions, fmt.Sprintf(`r.id IN (
				SELECT DISTINCT restaurant_id FROM restaurant_food_types
				WHERE food_type_id IN (%s)
			)`, strings.Join(placeholders, ",")))
		}
	}

	// Location/radius filter
	var distanceSelect string
	var distanceOrder string
	var hasDistance bool

	if lat != "" && lng != "" && radius != "" {
		latVal, latErr := strconv.ParseFloat(lat, 64)
		lngVal, lngErr := strconv.ParseFloat(lng, 64)
		radiusVal, radErr := strconv.ParseFloat(radius, 64)

		if latErr == nil && lngErr == nil && radErr == nil {
			hasDistance = true
			distanceSelect = fmt.Sprintf(`,
				(6371 * acos(
					cos(radians($%d)) * cos(radians(r.latitude)) *
					cos(radians(r.longitude) - radians($%d)) +
					sin(radians($%d)) * sin(radians(r.latitude))
				)) as distance`, restaurantArgIndex, restaurantArgIndex+1, restaurantArgIndex+2)
			restaurantArgs = append(restaurantArgs, latVal, lngVal, latVal)

			restaurantConditions = append(restaurantConditions, fmt.Sprintf(`r.latitude IS NOT NULL AND r.longitude IS NOT NULL AND
				(6371 * acos(
					cos(radians($%d)) * cos(radians(r.latitude)) *
					cos(radians(r.longitude) - radians($%d)) +
					sin(radians($%d)) * sin(radians(r.latitude))
				)) <= $%d`, restaurantArgIndex+3, restaurantArgIndex+4, restaurantArgIndex+5, restaurantArgIndex+6))
			restaurantArgs = append(restaurantArgs, latVal, lngVal, latVal, radiusVal)
			restaurantArgIndex += 7

			distanceOrder = "distance ASC,"
		}
	}

	restaurantWhereClause := ""
	if len(restaurantConditions) > 0 {
		restaurantWhereClause = "WHERE " + strings.Join(restaurantConditions, " AND ")
	}

	// Build HAVING clause for min_rating filter
	havingClause := ""
	if minRating != "" {
		if mr, err := strconv.ParseFloat(minRating, 64); err == nil && mr >= 1 && mr <= 5 {
			havingClause = fmt.Sprintf("HAVING COALESCE(AVG((rt.food_rating + rt.service_rating + rt.ambiance_rating) / 3.0), 0) >= %f", mr)
		}
	}

	restaurantQuery := fmt.Sprintf(`
		SELECT
			r.id, r.name, r.description, r.address, r.phone, r.website, r.latitude, r.longitude,
			r.google_place_id, r.category_id, NULL::integer as price_range, r.user_id as created_by, NULL::integer as updated_by, r.created_at, r.updated_at,
			c.id, c.name,
			COALESCE(AVG(rt.food_rating), 0) as avg_food,
			COALESCE(AVG(rt.service_rating), 0) as avg_service,
			COALESCE(AVG(rt.ambiance_rating), 0) as avg_ambiance,
			COUNT(rt.id) as rating_count,
			false as is_suggestion,
			NULL::integer as suggestion_id,
			NULL::text as status,
			NULL::text as notes,
			NULL::integer as user_id,
			cu.id, cu.username, cu.full_name, cu.avatar_url,
			NULL::integer as uu_id, NULL::text as uu_username, NULL::text as uu_full_name, NULL::text as uu_avatar_url
			%s
		FROM restaurants r
		LEFT JOIN categories c ON r.category_id = c.id
		LEFT JOIN ratings rt ON r.id = rt.restaurant_id
		LEFT JOIN users cu ON r.user_id = cu.id
		%s
		GROUP BY r.id, c.id, cu.id
		%s
	`, distanceSelect, restaurantWhereClause, havingClause)

	args = restaurantArgs

	// Build suggestion query if requested
	var finalQuery string
	if includeSuggestions {
		// Build suggestion conditions similar to restaurant conditions
		var suggestionConditions []string
		suggestionArgIndex := len(args) + 1

		if categoryID != "" {
			if catID, err := strconv.Atoi(categoryID); err == nil {
				suggestionConditions = append(suggestionConditions, fmt.Sprintf("s.suggested_category_id = $%d", suggestionArgIndex))
				args = append(args, catID)
				suggestionArgIndex++
			}
		}

		if foodTypeIDs != "" {
			ftIDs := strings.Split(foodTypeIDs, ",")
			var validIDs []int
			for _, idStr := range ftIDs {
				if id, err := strconv.Atoi(strings.TrimSpace(idStr)); err == nil {
					validIDs = append(validIDs, id)
				}
			}
			if len(validIDs) > 0 {
				placeholders := make([]string, len(validIDs))
				for i, id := range validIDs {
					placeholders[i] = fmt.Sprintf("$%d", suggestionArgIndex)
					args = append(args, id)
					suggestionArgIndex++
				}
				suggestionConditions = append(suggestionConditions, fmt.Sprintf(`s.id IN (
					SELECT DISTINCT suggestion_id FROM suggestion_food_types
					WHERE food_type_id IN (%s)
				)`, strings.Join(placeholders, ",")))
			}
		}

		// Add status filter to only show pending suggestions
		suggestionConditions = append(suggestionConditions, "s.status = 'pending'")

		var suggestionDistanceSelect string
		if hasDistance {
			latVal, _ := strconv.ParseFloat(lat, 64)
			lngVal, _ := strconv.ParseFloat(lng, 64)
			radiusVal, _ := strconv.ParseFloat(radius, 64)

			suggestionDistanceSelect = fmt.Sprintf(`,
				(6371 * acos(
					cos(radians($%d)) * cos(radians(s.latitude)) *
					cos(radians(s.longitude) - radians($%d)) +
					sin(radians($%d)) * sin(radians(s.latitude))
				)) as distance`, suggestionArgIndex, suggestionArgIndex+1, suggestionArgIndex+2)
			args = append(args, latVal, lngVal, latVal)

			suggestionConditions = append(suggestionConditions, fmt.Sprintf(`s.latitude IS NOT NULL AND s.longitude IS NOT NULL AND
				(6371 * acos(
					cos(radians($%d)) * cos(radians(s.latitude)) *
					cos(radians(s.longitude) - radians($%d)) +
					sin(radians($%d)) * sin(radians(s.latitude))
				)) <= $%d`, suggestionArgIndex+3, suggestionArgIndex+4, suggestionArgIndex+5, suggestionArgIndex+6))
			args = append(args, latVal, lngVal, latVal, radiusVal)
		}

		suggestionWhereClause := ""
		if len(suggestionConditions) > 0 {
			suggestionWhereClause = "WHERE " + strings.Join(suggestionConditions, " AND ")
		}

		suggestionQuery := fmt.Sprintf(`
			SELECT
				s.id, s.name, NULL::text as description, s.address, s.phone, s.website, s.latitude, s.longitude,
				s.google_place_id, s.suggested_category_id as category_id, NULL::integer as price_range, NULL::integer as created_by, NULL::integer as updated_by, s.created_at, s.updated_at,
				c.id, c.name,
				0.0 as avg_food,
				0.0 as avg_service,
				0.0 as avg_ambiance,
				0 as rating_count,
				true as is_suggestion,
				s.id as suggestion_id,
				s.status,
				s.notes,
				s.user_id,
				u.id, u.username, u.full_name, u.avatar_url,
				NULL::integer as uu_id, NULL::text as uu_username, NULL::text as uu_full_name, NULL::text as uu_avatar_url
				%s
			FROM restaurant_suggestions s
			LEFT JOIN categories c ON s.suggested_category_id = c.id
			LEFT JOIN users u ON s.user_id = u.id
			%s
		`, suggestionDistanceSelect, suggestionWhereClause)

		// Build ORDER BY clause
		orderByClause := "created_at DESC"
		switch sortBy {
		case "name":
			orderByClause = "name ASC"
		case "rating":
			orderByClause = "avg_food DESC, avg_service DESC, avg_ambiance DESC"
		case "date":
			orderByClause = "created_at DESC"
		}

		finalQuery = fmt.Sprintf(`
			SELECT * FROM (
				%s
				UNION ALL
				%s
			) combined
			ORDER BY %s %s
		`, restaurantQuery, suggestionQuery, distanceOrder, orderByClause)
	} else {
		// Build ORDER BY clause for single query
		orderByClause := "r.created_at DESC"
		switch sortBy {
		case "name":
			orderByClause = "r.name ASC"
		case "rating":
			orderByClause = "avg_food DESC, avg_service DESC, avg_ambiance DESC"
		case "date":
			orderByClause = "r.created_at DESC"
		}

		finalQuery = fmt.Sprintf(`
			%s
			ORDER BY %s %s
		`, restaurantQuery, distanceOrder, orderByClause)
	}

	rows, err := database.GetPool().Query(ctx, finalQuery, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	restaurants := []models.Restaurant{}
	restaurantIDs := []int{}
	suggestionIDs := []int{}

	for rows.Next() {
		var rest models.Restaurant
		var catID *int
		var catName *string
		var avgFood, avgService, avgAmbiance float64
		var ratingCount int
		var distance *float64
		var cuID, uuID *int
		var cuUsername, cuFullName, cuAvatarURL *string
		var uuUsername, uuFullName, uuAvatarURL *string

		var err error
		if hasDistance {
			err = rows.Scan(
				&rest.ID, &rest.Name, &rest.Description, &rest.Address, &rest.Phone, &rest.Website, &rest.Latitude, &rest.Longitude,
				&rest.GooglePlaceID, &rest.CategoryID, &rest.PriceRange, &rest.CreatedBy, &rest.UpdatedBy, &rest.CreatedAt, &rest.UpdatedAt,
				&catID, &catName,
				&avgFood, &avgService, &avgAmbiance, &ratingCount,
				&rest.IsSuggestion, &rest.SuggestionID, &rest.Status,
				&rest.Notes, &rest.UserID,
				&cuID, &cuUsername, &cuFullName, &cuAvatarURL,
				&uuID, &uuUsername, &uuFullName, &uuAvatarURL,
				&distance,
			)
		} else {
			err = rows.Scan(
				&rest.ID, &rest.Name, &rest.Description, &rest.Address, &rest.Phone, &rest.Website, &rest.Latitude, &rest.Longitude,
				&rest.GooglePlaceID, &rest.CategoryID, &rest.PriceRange, &rest.CreatedBy, &rest.UpdatedBy, &rest.CreatedAt, &rest.UpdatedAt,
				&catID, &catName,
				&avgFood, &avgService, &avgAmbiance, &ratingCount,
				&rest.IsSuggestion, &rest.SuggestionID, &rest.Status,
				&rest.Notes, &rest.UserID,
				&cuID, &cuUsername, &cuFullName, &cuAvatarURL,
				&uuID, &uuUsername, &uuFullName, &uuAvatarURL,
			)
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if distance != nil {
			rest.Distance = distance
		}

		if catID != nil && catName != nil {
			rest.Category = &models.Category{ID: *catID, Name: *catName}
		}

		if cuID != nil && cuUsername != nil {
			// For suggestions, this is the user who created the suggestion
			if rest.IsSuggestion {
				rest.User = &models.UserSummary{
					ID:        *cuID,
					Username:  *cuUsername,
					FullName:  cuFullName,
					AvatarURL: cuAvatarURL,
				}
			} else {
				rest.CreatedByUser = &models.UserSummary{
					ID:        *cuID,
					Username:  *cuUsername,
					FullName:  cuFullName,
					AvatarURL: cuAvatarURL,
				}
			}
		}

		if uuID != nil && uuUsername != nil {
			rest.UpdatedByUser = &models.UserSummary{
				ID:        *uuID,
				Username:  *uuUsername,
				FullName:  uuFullName,
				AvatarURL: uuAvatarURL,
			}
		}

		if ratingCount > 0 {
			overall := (avgFood + avgService + avgAmbiance) / 3
			rest.AvgRating = &models.AvgRating{
				Food:     avgFood,
				Service:  avgService,
				Ambiance: avgAmbiance,
				Overall:  overall,
				Count:    ratingCount,
			}
		}

		// Collect IDs for batch food type lookup
		if rest.IsSuggestion {
			suggestionIDs = append(suggestionIDs, rest.ID)
		} else {
			restaurantIDs = append(restaurantIDs, rest.ID)
		}

		restaurants = append(restaurants, rest)
	}

	// Batch fetch food types for all restaurants
	restaurantFoodTypes := make(map[int][]models.FoodType)
	if len(restaurantIDs) > 0 {
		foodTypeMap, err := getFoodTypesForRestaurantsBatch(ctx, restaurantIDs)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		restaurantFoodTypes = foodTypeMap
	}

	// Batch fetch food types for all suggestions
	suggestionFoodTypes := make(map[int][]models.FoodType)
	if len(suggestionIDs) > 0 {
		foodTypeMap, err := getFoodTypesForSuggestionsBatch(ctx, suggestionIDs)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		suggestionFoodTypes = foodTypeMap
	}

	// Assign food types to restaurants
	for i := range restaurants {
		if restaurants[i].IsSuggestion {
			restaurants[i].FoodTypes = suggestionFoodTypes[restaurants[i].ID]
		} else {
			restaurants[i].FoodTypes = restaurantFoodTypes[restaurants[i].ID]
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(restaurants)
}

// GetRestaurant godoc
// @Summary Get a restaurant by ID
// @Description Get detailed information about a specific restaurant including ratings and food types
// @Tags Restaurants
// @Accept json
// @Produce json
// @Param id path int true "Restaurant ID"
// @Success 200 {object} models.Restaurant "Restaurant details"
// @Failure 400 {object} map[string]string "Invalid restaurant ID"
// @Failure 404 {object} map[string]string "Restaurant not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /restaurants/{id} [get]
func GetRestaurant(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid restaurant ID", http.StatusBadRequest)
		return
	}

	ctx, cancel := RequestContext(r)
	defer cancel()

	// Try both restaurants and suggestions tables using UNION
	query := `
		SELECT * FROM (
			SELECT
				r.id, r.name, r.description, r.address, r.phone, r.website, r.latitude, r.longitude,
				r.google_place_id, r.category_id, r.user_id as created_by, NULL::integer as updated_by, r.created_at, r.updated_at,
				c.id, c.name,
				COALESCE(AVG(rt.food_rating), 0) as avg_food,
				COALESCE(AVG(rt.service_rating), 0) as avg_service,
				COALESCE(AVG(rt.ambiance_rating), 0) as avg_ambiance,
				COUNT(rt.id) as rating_count,
				false as is_suggestion,
				cu.id, cu.username, cu.full_name, cu.avatar_url,
				NULL::integer as uu_id, NULL::text as uu_username, NULL::text as uu_full_name, NULL::text as uu_avatar_url
			FROM restaurants r
			LEFT JOIN categories c ON r.category_id = c.id
			LEFT JOIN ratings rt ON r.id = rt.restaurant_id
			LEFT JOIN users cu ON r.user_id = cu.id
			WHERE r.id = $1
			GROUP BY r.id, c.id, cu.id

			UNION ALL

			SELECT
				s.id, s.name, NULL::text as description, s.address, s.phone, s.website, s.latitude, s.longitude,
				s.google_place_id, s.suggested_category_id as category_id, NULL::integer as created_by, NULL::integer as updated_by, s.created_at, s.updated_at,
				c.id, c.name,
				0.0 as avg_food,
				0.0 as avg_service,
				0.0 as avg_ambiance,
				0 as rating_count,
				true as is_suggestion,
				u.id, u.username, u.full_name, u.avatar_url,
				NULL::integer as uu_id, NULL::text as uu_username, NULL::text as uu_full_name, NULL::text as uu_avatar_url
			FROM restaurant_suggestions s
			LEFT JOIN categories c ON s.suggested_category_id = c.id
			LEFT JOIN users u ON s.user_id = u.id
			WHERE s.id = $1
		) combined
		LIMIT 1
	`

	var rest models.Restaurant
	var catID *int
	var catName *string
	var avgFood, avgService, avgAmbiance float64
	var ratingCount int
	var isSuggestion bool
	var cuID, uuID *int
	var cuUsername, cuFullName, cuAvatarURL *string
	var uuUsername, uuFullName, uuAvatarURL *string

	err = database.GetPool().QueryRow(ctx, query, id).Scan(
		&rest.ID, &rest.Name, &rest.Description, &rest.Address, &rest.Phone, &rest.Website, &rest.Latitude, &rest.Longitude,
		&rest.GooglePlaceID, &rest.CategoryID, &rest.CreatedBy, &rest.UpdatedBy, &rest.CreatedAt, &rest.UpdatedAt,
		&catID, &catName,
		&avgFood, &avgService, &avgAmbiance, &ratingCount,
		&isSuggestion,
		&cuID, &cuUsername, &cuFullName, &cuAvatarURL,
		&uuID, &uuUsername, &uuFullName, &uuAvatarURL,
	)
	if err != nil {
		http.Error(w, "Restaurant not found", http.StatusNotFound)
		return
	}

	rest.IsSuggestion = isSuggestion

	if catID != nil && catName != nil {
		rest.Category = &models.Category{ID: *catID, Name: *catName}
	}

	if cuID != nil && cuUsername != nil {
		rest.CreatedByUser = &models.UserSummary{
			ID:        *cuID,
			Username:  *cuUsername,
			FullName:  cuFullName,
			AvatarURL: cuAvatarURL,
		}
	}

	if uuID != nil && uuUsername != nil {
		rest.UpdatedByUser = &models.UserSummary{
			ID:        *uuID,
			Username:  *uuUsername,
			FullName:  uuFullName,
			AvatarURL: uuAvatarURL,
		}
	}

	// Get food types (check both restaurants and suggestions tables based on isSuggestion flag)
	var foodTypes []models.FoodType
	if isSuggestion {
		foodTypes, err = getFoodTypesForSuggestion(ctx, rest.ID)
	} else {
		foodTypes, err = getFoodTypesForRestaurant(ctx, rest.ID)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rest.FoodTypes = foodTypes

	if ratingCount > 0 {
		overall := (avgFood + avgService + avgAmbiance) / 3
		rest.AvgRating = &models.AvgRating{
			Food:     avgFood,
			Service:  avgService,
			Ambiance: avgAmbiance,
			Overall:  overall,
			Count:    ratingCount,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rest)
}

// CreateRestaurant godoc
// @Summary Create a new restaurant
// @Description Create a new restaurant with details and food types
// @Tags Restaurants
// @Accept json
// @Produce json
// @Param restaurant body models.CreateRestaurantRequest true "Restaurant creation request"
// @Success 201 {object} models.Restaurant "Created restaurant"
// @Failure 400 {object} map[string]string "Invalid request body"
// @Failure 409 {object} map[string]string "Restaurant already exists"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /restaurants [post]
func CreateRestaurant(w http.ResponseWriter, r *http.Request) {
	var req models.CreateRestaurantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	// Get user from context
	user, ok := GetUserFromContext(r)
	var userID *int
	if ok && user != nil {
		userID = &user.ID
	}

	ctx, cancel := RequestContext(r)
	defer cancel()

	var rest models.Restaurant
	err := database.GetPool().QueryRow(ctx,
		`INSERT INTO restaurants (name, description, address, phone, website, latitude, longitude, google_place_id, category_id, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, name, description, address, phone, website, latitude, longitude, google_place_id, category_id, created_by, updated_by, created_at, updated_at`,
		req.Name, req.Description, req.Address, req.Phone, req.Website, req.Latitude, req.Longitude, req.GooglePlaceID, req.CategoryID, userID, userID,
	).Scan(
		&rest.ID, &rest.Name, &rest.Description, &rest.Address, &rest.Phone, &rest.Website, &rest.Latitude, &rest.Longitude,
		&rest.GooglePlaceID, &rest.CategoryID, &rest.CreatedBy, &rest.UpdatedBy, &rest.CreatedAt, &rest.UpdatedAt,
	)
	if err != nil {
		// Check if it's a unique constraint violation
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == "23505" { // unique_violation
				logger.Warn("Duplicate restaurant creation attempt: %s", req.Name)
				if strings.Contains(pgErr.ConstraintName, "google_place_id") {
					http.Error(w, "A restaurant with this Google Place ID already exists", http.StatusConflict)
				} else if strings.Contains(pgErr.ConstraintName, "name_address") {
					http.Error(w, "A restaurant with this name and address already exists", http.StatusConflict)
				} else {
					http.Error(w, "This restaurant already exists", http.StatusConflict)
				}
				return
			}
		}
		logger.Error("Failed to create restaurant: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set food types
	if len(req.FoodTypeIDs) > 0 {
		if err := setFoodTypesForRestaurant(ctx, rest.ID, req.FoodTypeIDs); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		foodTypes, _ := getFoodTypesForRestaurant(ctx, rest.ID)
		rest.FoodTypes = foodTypes
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(rest)
}

// UpdateRestaurant godoc
// @Summary Update a restaurant
// @Description Update an existing restaurant's information
// @Tags Restaurants
// @Accept json
// @Produce json
// @Param id path int true "Restaurant ID"
// @Param restaurant body models.UpdateRestaurantRequest true "Restaurant update request"
// @Success 200 {object} models.Restaurant "Updated restaurant"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 404 {object} map[string]string "Restaurant not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /restaurants/{id} [put]
func UpdateRestaurant(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid restaurant ID", http.StatusBadRequest)
		return
	}

	var req models.UpdateRestaurantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx, cancel := RequestContext(r)
	defer cancel()

	var rest models.Restaurant
	err = database.GetPool().QueryRow(ctx,
		`UPDATE restaurants SET
			name = COALESCE($1, name),
			description = COALESCE($2, description),
			address = COALESCE($3, address),
			phone = COALESCE($4, phone),
			website = COALESCE($5, website),
			latitude = COALESCE($6, latitude),
			longitude = COALESCE($7, longitude),
			google_place_id = COALESCE($8, google_place_id),
			category_id = COALESCE($9, category_id),
			updated_at = NOW()
		WHERE id = $10
		RETURNING id, name, description, address, phone, website, latitude, longitude, google_place_id, category_id, user_id, created_at, updated_at`,
		req.Name, req.Description, req.Address, req.Phone, req.Website, req.Latitude, req.Longitude, req.GooglePlaceID, req.CategoryID, id,
	).Scan(
		&rest.ID, &rest.Name, &rest.Description, &rest.Address, &rest.Phone, &rest.Website, &rest.Latitude, &rest.Longitude,
		&rest.GooglePlaceID, &rest.CategoryID, &rest.CreatedBy, &rest.CreatedAt, &rest.UpdatedAt,
	)
	if err != nil {
		logger.Error("Failed to update restaurant ID %d: %v", id, err)
		http.Error(w, "Restaurant not found", http.StatusNotFound)
		return
	}

	// Update food types if provided
	if req.FoodTypeIDs != nil {
		if err := setFoodTypesForRestaurant(ctx, rest.ID, req.FoodTypeIDs); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	foodTypes, _ := getFoodTypesForRestaurant(ctx, rest.ID)
	rest.FoodTypes = foodTypes

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rest)
}

// DeleteRestaurant godoc
// @Summary Delete a restaurant
// @Description Delete a restaurant by ID
// @Tags Restaurants
// @Accept json
// @Produce json
// @Param id path int true "Restaurant ID"
// @Success 204 "Restaurant deleted successfully"
// @Failure 400 {object} map[string]string "Invalid restaurant ID"
// @Failure 404 {object} map[string]string "Restaurant not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /restaurants/{id} [delete]
func DeleteRestaurant(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := RequestContext(r)
	defer cancel()

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid restaurant ID", http.StatusBadRequest)
		return
	}

	result, err := database.GetPool().Exec(ctx,
		"DELETE FROM restaurants WHERE id = $1", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if result.RowsAffected() == 0 {
		http.Error(w, "Restaurant not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GlobalSearch godoc
// @Summary Global search for restaurants and suggestions
// @Description Search both restaurants and suggestions by name with pattern matching
// @Tags Search
// @Accept json
// @Produce json
// @Param q query string true "Search query string"
// @Success 200 {array} models.Restaurant "List of matching restaurants and suggestions"
// @Failure 400 {object} map[string]string "Query parameter 'q' is required"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /search [get]
func GlobalSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := RequestContext(r)
	defer cancel()
	searchPattern := "%" + strings.ToLower(query) + "%"

	// Search restaurants
	restaurantsQuery := `
		SELECT DISTINCT
			r.id, r.name, r.description, r.address, r.phone, r.website, r.latitude, r.longitude,
			r.google_place_id, r.category_id, r.created_at, r.updated_at,
			c.id, c.name,
			COALESCE(AVG(rat.food_rating), 0) as avg_food,
			COALESCE(AVG(rat.service_rating), 0) as avg_service,
			COALESCE(AVG(rat.ambiance_rating), 0) as avg_ambiance,
			COUNT(rat.id) as rating_count,
			false as is_suggestion,
			NULL::integer as suggestion_id,
			NULL::text as status
		FROM restaurants r
		LEFT JOIN categories c ON r.category_id = c.id
		LEFT JOIN ratings rat ON r.id = rat.restaurant_id
		WHERE LOWER(r.name) LIKE $1
		GROUP BY r.id, r.name, r.description, r.address, r.phone, r.website, r.latitude, r.longitude,
			r.google_place_id, r.category_id, r.created_at, r.updated_at, c.id, c.name

		UNION ALL

		SELECT
			NULL::integer, s.name, NULL::text, s.address, s.phone, s.website, s.latitude, s.longitude,
			s.google_place_id, s.suggested_category_id, s.created_at, s.updated_at,
			c.id, c.name,
			0::float, 0::float, 0::float, 0::integer,
			true as is_suggestion,
			s.id as suggestion_id,
			s.status
		FROM restaurant_suggestions s
		LEFT JOIN categories c ON s.suggested_category_id = c.id
		WHERE LOWER(s.name) LIKE $1
			AND s.status = 'pending'

		ORDER BY 2
		LIMIT 20
	`

	rows, err := database.GetPool().Query(ctx, restaurantsQuery, searchPattern)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	results := []models.Restaurant{}
	restaurantIDs := []int{}
	suggestionIDs := []int{}

	for rows.Next() {
		var rest models.Restaurant
		var restaurantID *int
		var catID *int
		var catName *string
		var avgFood, avgService, avgAmbiance float64
		var ratingCount int
		var isSuggestion bool
		var suggestionID *int
		var status *string

		err := rows.Scan(
			&restaurantID, &rest.Name, &rest.Description, &rest.Address, &rest.Phone, &rest.Website, &rest.Latitude, &rest.Longitude,
			&rest.GooglePlaceID, &rest.CategoryID, &rest.CreatedAt, &rest.UpdatedAt,
			&catID, &catName,
			&avgFood, &avgService, &avgAmbiance, &ratingCount,
			&isSuggestion, &suggestionID, &status,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Set restaurant ID if it's not null (i.e., it's a real restaurant, not a suggestion)
		if restaurantID != nil {
			rest.ID = *restaurantID
		}

		if catID != nil && catName != nil {
			rest.Category = &models.Category{ID: *catID, Name: *catName}
		}

		rest.IsSuggestion = isSuggestion
		rest.SuggestionID = suggestionID
		rest.Status = status

		if ratingCount > 0 {
			overall := (avgFood + avgService + avgAmbiance) / 3
			rest.AvgRating = &models.AvgRating{
				Food:     avgFood,
				Service:  avgService,
				Ambiance: avgAmbiance,
				Overall:  overall,
				Count:    ratingCount,
			}
		}

		// Collect IDs for batch food type lookup
		if isSuggestion && suggestionID != nil {
			suggestionIDs = append(suggestionIDs, *suggestionID)
		} else if rest.ID > 0 {
			restaurantIDs = append(restaurantIDs, rest.ID)
		}

		results = append(results, rest)
	}

	// Batch fetch food types for all restaurants
	restaurantFoodTypes := make(map[int][]models.FoodType)
	if len(restaurantIDs) > 0 {
		foodTypeMap, err := getFoodTypesForRestaurantsBatch(ctx, restaurantIDs)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		restaurantFoodTypes = foodTypeMap
	}

	// Batch fetch food types for all suggestions
	suggestionFoodTypes := make(map[int][]models.FoodType)
	if len(suggestionIDs) > 0 {
		foodTypeMap, err := getFoodTypesForSuggestionsBatch(ctx, suggestionIDs)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		suggestionFoodTypes = foodTypeMap
	}

	// Assign food types to results
	for i := range results {
		if results[i].IsSuggestion && results[i].SuggestionID != nil {
			results[i].FoodTypes = suggestionFoodTypes[*results[i].SuggestionID]
		} else if results[i].ID > 0 {
			results[i].FoodTypes = restaurantFoodTypes[results[i].ID]
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
