package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/nomdb/backend/internal/database"
	"github.com/nomdb/backend/internal/logger"
	"github.com/nomdb/backend/internal/models"
)

// GetRestaurantsPaginated godoc
// @Summary Get paginated list of restaurants
// @Description Get restaurants with cursor-based pagination and optional filtering by category, food types, and search query
// @Tags Restaurants
// @Accept json
// @Produce json
// @Param cursor query string false "Pagination cursor (encoded last ID)"
// @Param limit query int false "Number of items per page (default 20, max 100)"
// @Param category_id query int false "Filter by category ID"
// @Param food_type_ids query string false "Filter by food type IDs (comma-separated)"
// @Param q query string false "Search query for name or description"
// @Success 200 {object} models.PaginatedResponse "Paginated list of restaurants"
// @Failure 400 {object} map[string]string "Invalid cursor or parameters"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /restaurants/paginated [get]
func GetRestaurantsPaginated(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := RequestContext(r)
	defer cancel()

	// Parse pagination parameters
	pagination := ParsePaginationParams(r)

	// Decode cursor to get last ID
	lastID, err := DecodeCursor(pagination.Cursor)
	if err != nil {
		logger.Error("Invalid cursor: %v", err)
		http.Error(w, "Invalid cursor", http.StatusBadRequest)
		return
	}

	// Parse query parameters for filtering
	queryParams := r.URL.Query()
	categoryID := queryParams.Get("category_id")
	foodTypeIDs := queryParams.Get("food_type_ids")
	searchQuery := queryParams.Get("q")

	// Build query with filters and pagination
	var conditions []string
	var args []interface{}
	argIndex := 1

	// Cursor-based pagination - newest first (r.id DESC); next page = older items (r.id < lastID)
	if lastID > 0 {
		conditions = append(conditions, fmt.Sprintf("r.id < $%d", argIndex))
		args = append(args, lastID)
		argIndex++
	}

	// Category filter
	if categoryID != "" {
		if catID, parseErr := strconv.Atoi(categoryID); parseErr == nil {
			conditions = append(conditions, fmt.Sprintf("r.category_id = $%d", argIndex))
			args = append(args, catID)
			argIndex++
		}
	}

	// Food type filter
	if foodTypeIDs != "" {
		ftIDs := strings.Split(foodTypeIDs, ",")
		var validIDs []int
		for _, idStr := range ftIDs {
			if id, parseErr := strconv.Atoi(strings.TrimSpace(idStr)); parseErr == nil {
				validIDs = append(validIDs, id)
			}
		}
		if len(validIDs) > 0 {
			placeholders := make([]string, len(validIDs))
			for i, id := range validIDs {
				placeholders[i] = fmt.Sprintf("$%d", argIndex)
				args = append(args, id)
				argIndex++
			}
			conditions = append(conditions, fmt.Sprintf(`r.id IN (
				SELECT DISTINCT restaurant_id FROM restaurant_food_types
				WHERE food_type_id IN (%s)
			)`, strings.Join(placeholders, ",")))
		}
	}

	// Search query filter (ILIKE works without migration 000008; use search_vector for full-text after running that migration)
	if searchQuery != "" {
		conditions = append(conditions, fmt.Sprintf("(r.name ILIKE $%d OR r.description ILIKE $%d)", argIndex, argIndex))
		args = append(args, "%"+searchQuery+"%")
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Fetch one more than limit to determine if there are more results
	fetchLimit := pagination.Limit + 1

	query := fmt.Sprintf(`
		SELECT
			r.id, r.name, r.description, r.address, r.phone, r.website, r.latitude, r.longitude,
			r.google_place_id, r.category_id, r.created_at, r.updated_at,
			c.id, c.name,
			COALESCE(AVG(rt.food_rating), 0) as avg_food,
			COALESCE(AVG(rt.service_rating), 0) as avg_service,
			COALESCE(AVG(rt.ambiance_rating), 0) as avg_ambiance,
			COUNT(rt.id) as rating_count
		FROM restaurants r
		LEFT JOIN categories c ON r.category_id = c.id
		LEFT JOIN ratings rt ON r.id = rt.restaurant_id
		%s
		GROUP BY r.id, c.id
		ORDER BY r.id DESC
		LIMIT $%d
	`, whereClause, argIndex)

	args = append(args, fetchLimit)

	rows, err := database.GetPool().Query(ctx, query, args...)
	if err != nil {
		logger.Error("Failed to query restaurants: %v", err)
		http.Error(w, "Failed to fetch restaurants", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var restaurants []models.Restaurant
	var restaurantIDs []int
	var lastFetchedID int

	for rows.Next() {
		var restaurant models.Restaurant
		var categoryID *int
		var categoryName *string
		var avgFood, avgService, avgAmbiance float64
		var ratingCount int

		err := rows.Scan(
			&restaurant.ID, &restaurant.Name, &restaurant.Description, &restaurant.Address,
			&restaurant.Phone, &restaurant.Website, &restaurant.Latitude, &restaurant.Longitude,
			&restaurant.GooglePlaceID, &restaurant.CategoryID, &restaurant.CreatedAt, &restaurant.UpdatedAt,
			&categoryID, &categoryName,
			&avgFood, &avgService, &avgAmbiance, &ratingCount,
		)
		if err != nil {
			logger.Error("Failed to scan restaurant: %v", err)
			http.Error(w, "Failed to process restaurants", http.StatusInternalServerError)
			return
		}

		if categoryID != nil && categoryName != nil {
			restaurant.Category = &models.Category{
				ID:   *categoryID,
				Name: *categoryName,
			}
		}

		if ratingCount > 0 {
			overall := (avgFood + avgService + avgAmbiance) / 3
			restaurant.AvgRating = &models.AvgRating{
				Food:     avgFood,
				Service:  avgService,
				Ambiance: avgAmbiance,
				Overall:  overall,
				Count:    ratingCount,
			}
		}

		restaurants = append(restaurants, restaurant)
		restaurantIDs = append(restaurantIDs, restaurant.ID)
		lastFetchedID = restaurant.ID
	}

	// Check if there are more results
	hasMore := len(restaurants) > pagination.Limit
	if hasMore {
		// Remove the extra item
		restaurants = restaurants[:pagination.Limit]
		restaurantIDs = restaurantIDs[:pagination.Limit]
	}

	// Fetch food types for all restaurants in batch
	if len(restaurantIDs) > 0 {
		foodTypesMap, err := getFoodTypesForRestaurantsBatch(ctx, restaurantIDs)
		if err != nil {
			logger.Error("Failed to fetch food types: %v", err)
			// Continue without food types rather than failing entirely
		} else {
			for i := range restaurants {
				if foodTypes, exists := foodTypesMap[restaurants[i].ID]; exists {
					restaurants[i].FoodTypes = foodTypes
				}
			}
		}
	}

	// Build paginated response
	var nextCursor *string
	if hasMore && lastFetchedID > 0 {
		cursor := EncodeCursor(lastFetchedID)
		nextCursor = &cursor
	}

	response := models.PaginatedResponse{
		Data:       restaurants,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "private, max-age=60")
	json.NewEncoder(w).Encode(response)
}
