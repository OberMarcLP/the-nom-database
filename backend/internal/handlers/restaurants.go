package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nomdb/backend/internal/database"
	"github.com/nomdb/backend/internal/logger"
	"github.com/nomdb/backend/internal/models"
	"github.com/nomdb/backend/internal/services"
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
	if err := rows.Err(); err != nil {
		return nil, err
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
	if err := rows.Err(); err != nil {
		return nil, err
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
	if err := rows.Err(); err != nil {
		return nil, err
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
	if err := rows.Err(); err != nil {
		return nil, err
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

	// Batch insert new food types
	if len(foodTypeIDs) > 0 {
		// Build batch insert query
		valueStrings := make([]string, len(foodTypeIDs))
		valueArgs := make([]interface{}, len(foodTypeIDs)*2)
		for i, ftID := range foodTypeIDs {
			valueStrings[i] = fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2)
			valueArgs[i*2] = restaurantID
			valueArgs[i*2+1] = ftID
		}
		query := fmt.Sprintf("INSERT INTO restaurant_food_types (restaurant_id, food_type_id) VALUES %s",
			strings.Join(valueStrings, ", "))
		_, err := database.GetPool().Exec(ctx, query, valueArgs...)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetRestaurants godoc
// @Summary List all restaurants
// @Description Get a list of all restaurants (including pending suggestions) with optional filtering, sorting and location search
// @Tags Restaurants
// @Accept json
// @Produce json
// @Param q query string false "Search in restaurant name and address"
// @Param category_id query int false "Filter by category ID"
// @Param food_type_ids query string false "Filter by food type IDs (comma-separated)"
// @Param price_range query int false "Maximum price range 1-4; restaurants without a price range always match"
// @Param min_rating query number false "Minimum average rating 1-5"
// @Param sort query string false "Sort order: name, rating or date (default: date)"
// @Param lat query number false "Latitude for distance filtering"
// @Param lng query number false "Longitude for distance filtering"
// @Param radius query number false "Radius in kilometers for distance filtering"
// @Success 200 {array} models.Restaurant "List of restaurants"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /restaurants [get]
func GetRestaurants(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := RequestContext(r)
	defer cancel()

	filters := parseRestaurantListFilters(r)
	query, args, dist := buildRestaurantListQuery(filters)

	rows, err := database.GetPool().Query(ctx, query, args...)
	if err != nil {
		logger.Error("request failed: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	restaurants := []models.Restaurant{}
	for rows.Next() {
		rest, err := scanRestaurantListRow(rows, dist.active)
		if err != nil {
			logger.Error("request failed: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		restaurants = append(restaurants, rest)
	}
	if err := rows.Err(); err != nil {
		logger.Error("request failed: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := attachRestaurantListFoodTypes(ctx, restaurants); err != nil {
		logger.Error("request failed: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(restaurants)
}

// restaurantListFilters holds the raw query parameters accepted by GetRestaurants.
type restaurantListFilters struct {
	searchQuery string // q: matched against name and address
	categoryID  string // category_id
	foodTypeIDs string // food_type_ids: comma-separated list
	priceRange  string // price_range: max price range (1-4)
	minRating   string // min_rating: minimum average rating (1-5)
	sortBy      string // sort: name, rating, date
	lat         string
	lng         string
	radius      string // in kilometers
}

// parseRestaurantListFilters extracts the supported filter parameters from the request.
func parseRestaurantListFilters(r *http.Request) restaurantListFilters {
	queryParams := r.URL.Query()
	return restaurantListFilters{
		searchQuery: queryParams.Get("q"),
		categoryID:  queryParams.Get("category_id"),
		foodTypeIDs: queryParams.Get("food_type_ids"),
		priceRange:  queryParams.Get("price_range"),
		minRating:   queryParams.Get("min_rating"),
		sortBy:      queryParams.Get("sort"),
		lat:         queryParams.Get("lat"),
		lng:         queryParams.Get("lng"),
		radius:      queryParams.Get("radius"),
	}
}

// parseFoodTypeIDList splits a comma-separated ID list and keeps the valid integers.
func parseFoodTypeIDList(commaSeparated string) []int {
	var validIDs []int
	for _, idStr := range strings.Split(commaSeparated, ",") {
		if id, err := strconv.Atoi(strings.TrimSpace(idStr)); err == nil {
			validIDs = append(validIDs, id)
		}
	}
	return validIDs
}

// distanceFilter carries the parsed location parameters shared by both UNION branches.
type distanceFilter struct {
	active bool
	lat    float64
	lng    float64
	radius float64
}

// parseDistanceFilter parses lat/lng/radius. The filter is only active when all
// three parameters are present and numeric; otherwise it is silently ignored.
func parseDistanceFilter(f restaurantListFilters) distanceFilter {
	if f.lat == "" || f.lng == "" || f.radius == "" {
		return distanceFilter{}
	}
	latVal, latErr := strconv.ParseFloat(f.lat, 64)
	lngVal, lngErr := strconv.ParseFloat(f.lng, 64)
	radiusVal, radErr := strconv.ParseFloat(f.radius, 64)
	if latErr != nil || lngErr != nil || radErr != nil {
		return distanceFilter{}
	}
	return distanceFilter{active: true, lat: latVal, lng: lngVal, radius: radiusVal}
}

// buildMinRatingHaving returns the HAVING clause and its argument for the
// min_rating filter, or an empty clause when the parameter is absent or out
// of range. argIndex is the positional-placeholder index to use.
func buildMinRatingHaving(minRating string, argIndex int) (string, []interface{}) {
	if minRating == "" {
		return "", nil
	}
	if mr, err := strconv.ParseFloat(minRating, 64); err == nil && mr >= 1 && mr <= 5 {
		clause := fmt.Sprintf("HAVING COALESCE(AVG((rt.food_rating + rt.service_rating + rt.ambiance_rating) / 3.0), 0) >= $%d", argIndex)
		return clause, []interface{}{mr}
	}
	return "", nil
}

// buildRestaurantBranch builds the restaurants half of the UNION query.
// Positional placeholders start at $1; the returned args match them in order.
func buildRestaurantBranch(f restaurantListFilters, dist distanceFilter) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	// Search query filter (name/address)
	if f.searchQuery != "" {
		conditions = append(conditions, fmt.Sprintf("(r.name ILIKE $%d OR r.address ILIKE $%d)", argIndex, argIndex))
		args = append(args, "%"+f.searchQuery+"%")
		argIndex++
	}

	if f.categoryID != "" {
		if catID, err := strconv.Atoi(f.categoryID); err == nil {
			conditions = append(conditions, fmt.Sprintf("r.category_id = $%d", argIndex))
			args = append(args, catID)
			argIndex++
		}
	}

	// Price range filter
	if f.priceRange != "" {
		if pr, err := strconv.Atoi(f.priceRange); err == nil && pr >= 1 && pr <= 4 {
			conditions = append(conditions, fmt.Sprintf("(r.price_range IS NULL OR r.price_range <= $%d)", argIndex))
			args = append(args, pr)
			argIndex++
		}
	}

	if validIDs := parseFoodTypeIDList(f.foodTypeIDs); len(validIDs) > 0 {
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

	// Location/radius filter
	distanceSelect := ""
	if dist.active {
		distanceSelect = fmt.Sprintf(`,
			(6371 * acos(
				cos(radians($%d)) * cos(radians(r.latitude)) *
				cos(radians(r.longitude) - radians($%d)) +
				sin(radians($%d)) * sin(radians(r.latitude))
			)) as distance`, argIndex, argIndex+1, argIndex+2)
		args = append(args, dist.lat, dist.lng, dist.lat)

		conditions = append(conditions, fmt.Sprintf(`r.latitude IS NOT NULL AND r.longitude IS NOT NULL AND
			(6371 * acos(
				cos(radians($%d)) * cos(radians(r.latitude)) *
				cos(radians(r.longitude) - radians($%d)) +
				sin(radians($%d)) * sin(radians(r.latitude))
			)) <= $%d`, argIndex+3, argIndex+4, argIndex+5, argIndex+6))
		args = append(args, dist.lat, dist.lng, dist.lat, dist.radius)
		argIndex += 7
	}

	havingClause, havingArgs := buildMinRatingHaving(f.minRating, argIndex)
	args = append(args, havingArgs...)

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	query := fmt.Sprintf(`
		SELECT
			r.id, r.name, r.description, r.address, r.phone, r.website, r.latitude, r.longitude,
			r.google_place_id, r.category_id, r.price_range, r.user_id as created_by, NULL::integer as updated_by, r.created_at, r.updated_at,
			c.id, c.name AS category_name,
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
	`, distanceSelect, whereClause, havingClause)

	return query, args
}

// buildSuggestionBranch builds the pending-suggestions half of the UNION query.
// argStart is the first free positional-placeholder index after the restaurant branch.
func buildSuggestionBranch(f restaurantListFilters, dist distanceFilter, argStart int) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	argIndex := argStart

	if f.categoryID != "" {
		if catID, err := strconv.Atoi(f.categoryID); err == nil {
			conditions = append(conditions, fmt.Sprintf("s.suggested_category_id = $%d", argIndex))
			args = append(args, catID)
			argIndex++
		}
	}

	if validIDs := parseFoodTypeIDList(f.foodTypeIDs); len(validIDs) > 0 {
		placeholders := make([]string, len(validIDs))
		for i, id := range validIDs {
			placeholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, id)
			argIndex++
		}
		conditions = append(conditions, fmt.Sprintf(`s.id IN (
			SELECT DISTINCT suggestion_id FROM suggestion_food_types
			WHERE food_type_id IN (%s)
		)`, strings.Join(placeholders, ",")))
	}

	// Only show pending suggestions
	conditions = append(conditions, "s.status = 'pending'")

	distanceSelect := ""
	if dist.active {
		distanceSelect = fmt.Sprintf(`,
			(6371 * acos(
				cos(radians($%d)) * cos(radians(s.latitude)) *
				cos(radians(s.longitude) - radians($%d)) +
				sin(radians($%d)) * sin(radians(s.latitude))
			)) as distance`, argIndex, argIndex+1, argIndex+2)
		args = append(args, dist.lat, dist.lng, dist.lat)

		conditions = append(conditions, fmt.Sprintf(`s.latitude IS NOT NULL AND s.longitude IS NOT NULL AND
			(6371 * acos(
				cos(radians($%d)) * cos(radians(s.latitude)) *
				cos(radians(s.longitude) - radians($%d)) +
				sin(radians($%d)) * sin(radians(s.latitude))
			)) <= $%d`, argIndex+3, argIndex+4, argIndex+5, argIndex+6))
		args = append(args, dist.lat, dist.lng, dist.lat, dist.radius)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	query := fmt.Sprintf(`
		SELECT
			s.id, s.name, NULL::text as description, s.address, s.phone, s.website, s.latitude, s.longitude,
			s.google_place_id, s.suggested_category_id as category_id, NULL::integer as price_range, NULL::integer as created_by, NULL::integer as updated_by, s.created_at, s.updated_at,
			c.id, c.name AS category_name,
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
	`, distanceSelect, whereClause)

	return query, args
}

// restaurantListOrderBy maps the sort parameter onto the ORDER BY expression,
// targeting the column aliases of the UNION subquery.
func restaurantListOrderBy(sortBy string) string {
	switch sortBy {
	case "name":
		return "name ASC"
	case "rating":
		return "avg_food DESC, avg_service DESC, avg_ambiance DESC"
	case "date":
		return "created_at DESC"
	}
	return "created_at DESC"
}

// buildRestaurantListQuery assembles the complete SQL statement and positional
// args for GetRestaurants and reports whether distance columns are selected.
func buildRestaurantListQuery(f restaurantListFilters) (string, []interface{}, distanceFilter) {
	dist := parseDistanceFilter(f)

	restaurantQuery, args := buildRestaurantBranch(f, dist)

	distanceOrder := ""
	if dist.active {
		distanceOrder = "distance ASC,"
	}

	// Suggestions are always part of the restaurant list.
	suggestionQuery, suggestionArgs := buildSuggestionBranch(f, dist, len(args)+1)
	args = append(args, suggestionArgs...)

	finalQuery := fmt.Sprintf(`
			SELECT * FROM (
				%s
				UNION ALL
				%s
			) combined
			ORDER BY %s %s
		`, restaurantQuery, suggestionQuery, distanceOrder, restaurantListOrderBy(f.sortBy))

	return finalQuery, args, dist
}

// scanRestaurantListRow scans a single row of the combined restaurant/suggestion
// query and assembles the nested category, user and rating structures.
// withDistance must match whether the query selects the distance column.
func scanRestaurantListRow(rows pgx.Rows, withDistance bool) (models.Restaurant, error) {
	var rest models.Restaurant
	var catID *int
	var catName *string
	var avgFood, avgService, avgAmbiance float64
	var ratingCount int
	var distance *float64
	var cuID, uuID *int
	var cuUsername, cuFullName, cuAvatarURL *string
	var uuUsername, uuFullName, uuAvatarURL *string

	targets := []interface{}{
		&rest.ID, &rest.Name, &rest.Description, &rest.Address, &rest.Phone, &rest.Website, &rest.Latitude, &rest.Longitude,
		&rest.GooglePlaceID, &rest.CategoryID, &rest.PriceRange, &rest.CreatedBy, &rest.UpdatedBy, &rest.CreatedAt, &rest.UpdatedAt,
		&catID, &catName,
		&avgFood, &avgService, &avgAmbiance, &ratingCount,
		&rest.IsSuggestion, &rest.SuggestionID, &rest.Status,
		&rest.Notes, &rest.UserID,
		&cuID, &cuUsername, &cuFullName, &cuAvatarURL,
		&uuID, &uuUsername, &uuFullName, &uuAvatarURL,
	}
	if withDistance {
		targets = append(targets, &distance)
	}
	if err := rows.Scan(targets...); err != nil {
		return models.Restaurant{}, err
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

	return rest, nil
}

// attachRestaurantListFoodTypes batch-loads the food types for all restaurants
// and pending suggestions in the result list and assigns them in place.
func attachRestaurantListFoodTypes(ctx context.Context, restaurants []models.Restaurant) error {
	restaurantIDs := []int{}
	suggestionIDs := []int{}
	for _, rest := range restaurants {
		if rest.IsSuggestion {
			suggestionIDs = append(suggestionIDs, rest.ID)
		} else {
			restaurantIDs = append(restaurantIDs, rest.ID)
		}
	}

	// Batch fetch food types for all restaurants
	restaurantFoodTypes := make(map[int][]models.FoodType)
	if len(restaurantIDs) > 0 {
		foodTypeMap, err := getFoodTypesForRestaurantsBatch(ctx, restaurantIDs)
		if err != nil {
			return err
		}
		restaurantFoodTypes = foodTypeMap
	}

	// Batch fetch food types for all suggestions
	suggestionFoodTypes := make(map[int][]models.FoodType)
	if len(suggestionIDs) > 0 {
		foodTypeMap, err := getFoodTypesForSuggestionsBatch(ctx, suggestionIDs)
		if err != nil {
			return err
		}
		suggestionFoodTypes = foodTypeMap
	}

	// Assign food types
	for i := range restaurants {
		if restaurants[i].IsSuggestion {
			restaurants[i].FoodTypes = suggestionFoodTypes[restaurants[i].ID]
		} else {
			restaurants[i].FoodTypes = restaurantFoodTypes[restaurants[i].ID]
		}
	}
	return nil
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
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
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
				r.google_place_id, r.category_id, r.price_range, r.user_id as created_by, NULL::integer as updated_by, r.created_at, r.updated_at,
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
				s.google_place_id, s.suggested_category_id as category_id, NULL::integer as price_range, NULL::integer as created_by, NULL::integer as updated_by, s.created_at, s.updated_at,
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
		&rest.GooglePlaceID, &rest.CategoryID, &rest.PriceRange, &rest.CreatedBy, &rest.UpdatedBy, &rest.CreatedAt, &rest.UpdatedAt,
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
		logger.Error("request failed: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
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

// validatePriceRange checks the optional price_range value (1 = $ to 4 = $$$$).
// nil means "no price range given" and is valid.
func validatePriceRange(priceRange *int) error {
	if priceRange == nil {
		return nil
	}
	if *priceRange < 1 || *priceRange > 4 {
		return errors.New("Price range must be between 1 and 4")
	}
	return nil
}

// CreateRestaurant godoc
// @Summary Create a new restaurant
// @Description Create a new restaurant with details and food types. price_range is optional (1 = $ to 4 = $$$$).
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

	if err := validatePriceRange(req.PriceRange); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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
		`INSERT INTO restaurants (name, description, address, phone, website, latitude, longitude, google_place_id, category_id, price_range, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, name, description, address, phone, website, latitude, longitude, google_place_id, category_id, price_range, created_by, updated_by, created_at, updated_at`,
		req.Name, req.Description, req.Address, req.Phone, req.Website, req.Latitude, req.Longitude, req.GooglePlaceID, req.CategoryID, req.PriceRange, userID, userID,
	).Scan(
		&rest.ID, &rest.Name, &rest.Description, &rest.Address, &rest.Phone, &rest.Website, &rest.Latitude, &rest.Longitude,
		&rest.GooglePlaceID, &rest.CategoryID, &rest.PriceRange, &rest.CreatedBy, &rest.UpdatedBy, &rest.CreatedAt, &rest.UpdatedAt,
	)
	if err != nil {
		// Check if it's a unique constraint violation
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
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
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Set food types
	if len(req.FoodTypeIDs) > 0 {
		if err := setFoodTypesForRestaurant(ctx, rest.ID, req.FoodTypeIDs); err != nil {
			logger.Error("request failed: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
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
// @Description Update an existing restaurant's information. price_range is optional (1 = $ to 4 = $$$$); omitted or null fields keep their current value.
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
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		http.Error(w, "Invalid restaurant ID", http.StatusBadRequest)
		return
	}

	var req models.UpdateRestaurantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := validatePriceRange(req.PriceRange); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := RequestContext(r)
	defer cancel()

	// The creator, admins, and roles with restaurants.update may modify a restaurant
	user, ok := GetUserFromContext(r)
	if !ok || user == nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}
	var createdBy *int
	if err := database.GetPool().QueryRow(ctx,
		"SELECT user_id FROM restaurants WHERE id = $1", id).Scan(&createdBy); err != nil {
		http.Error(w, "Restaurant not found", http.StatusNotFound)
		return
	}
	if !user.IsAdmin &&
		!services.HasPermission(user.Permissions, "restaurants.update") &&
		(createdBy == nil || *createdBy != user.ID) {
		http.Error(w, "You can only edit restaurants you created", http.StatusForbidden)
		return
	}

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
			price_range = $10, -- not COALESCE: the edit form sends full state, so absent means clear
			updated_at = NOW()
		WHERE id = $11
		RETURNING id, name, description, address, phone, website, latitude, longitude, google_place_id, category_id, price_range, user_id, created_at, updated_at`,
		req.Name, req.Description, req.Address, req.Phone, req.Website, req.Latitude, req.Longitude, req.GooglePlaceID, req.CategoryID, req.PriceRange, id,
	).Scan(
		&rest.ID, &rest.Name, &rest.Description, &rest.Address, &rest.Phone, &rest.Website, &rest.Latitude, &rest.Longitude,
		&rest.GooglePlaceID, &rest.CategoryID, &rest.PriceRange, &rest.CreatedBy, &rest.CreatedAt, &rest.UpdatedAt,
	)
	if err != nil {
		logger.Error("Failed to update restaurant ID %d: %v", id, err)
		http.Error(w, "Restaurant not found", http.StatusNotFound)
		return
	}

	// Update food types if provided
	if req.FoodTypeIDs != nil {
		if err := setFoodTypesForRestaurant(ctx, rest.ID, req.FoodTypeIDs); err != nil {
			logger.Error("request failed: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
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

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		http.Error(w, "Invalid restaurant ID", http.StatusBadRequest)
		return
	}

	// The creator, admins, and roles with restaurants.delete may delete a restaurant
	user, ok := GetUserFromContext(r)
	if !ok || user == nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}
	var createdBy *int
	if err := database.GetPool().QueryRow(ctx,
		"SELECT user_id FROM restaurants WHERE id = $1", id).Scan(&createdBy); err != nil {
		http.Error(w, "Restaurant not found", http.StatusNotFound)
		return
	}
	if !user.IsAdmin &&
		!services.HasPermission(user.Permissions, "restaurants.delete") &&
		(createdBy == nil || *createdBy != user.ID) {
		http.Error(w, "You can only delete restaurants you created", http.StatusForbidden)
		return
	}

	result, err := database.GetPool().Exec(ctx,
		"DELETE FROM restaurants WHERE id = $1", id)
	if err != nil {
		logger.Error("request failed: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
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
		logger.Error("request failed: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
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
			logger.Error("request failed: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
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
			logger.Error("request failed: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		restaurantFoodTypes = foodTypeMap
	}

	// Batch fetch food types for all suggestions
	suggestionFoodTypes := make(map[int][]models.FoodType)
	if len(suggestionIDs) > 0 {
		foodTypeMap, err := getFoodTypesForSuggestionsBatch(ctx, suggestionIDs)
		if err != nil {
			logger.Error("request failed: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
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
