package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nomdb/backend/internal/database"
	"github.com/nomdb/backend/internal/logger"
	"github.com/nomdb/backend/internal/models"
)

// Helper functions for suggestion food types
func setFoodTypesForSuggestion(ctx context.Context, suggestionID int, foodTypeIDs []int) error {
	// Delete existing food types
	_, err := database.GetPool().Exec(ctx,
		"DELETE FROM suggestion_food_types WHERE suggestion_id = $1", suggestionID)
	if err != nil {
		return err
	}

	// Insert new food types
	for _, ftID := range foodTypeIDs {
		_, err := database.GetPool().Exec(ctx,
			"INSERT INTO suggestion_food_types (suggestion_id, food_type_id) VALUES ($1, $2)",
			suggestionID, ftID)
		if err != nil {
			return err
		}
	}
	return nil
}

// @Summary List all restaurant suggestions
// @Description Get a list of all restaurant suggestions with optional status filter
// @Tags Suggestions
// @Accept json
// @Produce json
// @Param status query string false "Filter by status (pending, approved, tested, rejected)"
// @Success 200 {array} models.RestaurantSuggestion "List of suggestions"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /suggestions [get]
func GetSuggestions(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	statusFilter := r.URL.Query().Get("status")

	var query string
	var args []interface{}

	if statusFilter != "" {
		query = `
			SELECT
				s.id, s.name, s.address, s.phone, s.website, s.latitude, s.longitude,
				s.google_place_id, s.suggested_category_id, s.notes, s.status,
				s.user_id, s.created_at, s.updated_at,
				c.id, c.name,
				u.id, u.username, u.full_name, u.avatar_url
			FROM restaurant_suggestions s
			LEFT JOIN categories c ON s.suggested_category_id = c.id
			LEFT JOIN users u ON s.user_id = u.id
			WHERE s.status = $1
			ORDER BY s.created_at DESC
		`
		args = append(args, statusFilter)
	} else {
		query = `
			SELECT
				s.id, s.name, s.address, s.phone, s.website, s.latitude, s.longitude,
				s.google_place_id, s.suggested_category_id, s.notes, s.status,
				s.user_id, s.created_at, s.updated_at,
				c.id, c.name,
				u.id, u.username, u.full_name, u.avatar_url
			FROM restaurant_suggestions s
			LEFT JOIN categories c ON s.suggested_category_id = c.id
			LEFT JOIN users u ON s.user_id = u.id
			ORDER BY s.created_at DESC
		`
	}

	rows, err := database.GetPool().Query(ctx, query, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	suggestions := []models.RestaurantSuggestion{}
	for rows.Next() {
		var sug models.RestaurantSuggestion
		var catID *int
		var catName *string
		var userID *int
		var username *string
		var fullName *string
		var avatarURL *string

		if err := rows.Scan(
			&sug.ID, &sug.Name, &sug.Address, &sug.Phone, &sug.Website, &sug.Latitude, &sug.Longitude,
			&sug.GooglePlaceID, &sug.SuggestedCategoryID, &sug.Notes, &sug.Status,
			&sug.UserID, &sug.CreatedAt, &sug.UpdatedAt,
			&catID, &catName,
			&userID, &username, &fullName, &avatarURL,
		); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if catID != nil && catName != nil {
			sug.Category = &models.Category{ID: *catID, Name: *catName}
		}

		if userID != nil && username != nil {
			sug.User = &models.UserSummary{
				ID:        *userID,
				Username:  *username,
				FullName:  fullName,
				AvatarURL: avatarURL,
			}
		}

		// Get food types
		foodTypes, err := getFoodTypesForSuggestion(ctx, sug.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		sug.FoodTypes = foodTypes

		suggestions = append(suggestions, sug)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(suggestions)
}

// @Summary Get a suggestion by ID
// @Description Get detailed information about a specific restaurant suggestion
// @Tags Suggestions
// @Accept json
// @Produce json
// @Param id path int true "Suggestion ID"
// @Success 200 {object} models.RestaurantSuggestion "Suggestion details"
// @Failure 400 {object} map[string]string "Invalid suggestion ID"
// @Failure 404 {object} map[string]string "Suggestion not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /suggestions/{id} [get]
func GetSuggestion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid suggestion ID", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	query := `
		SELECT
			s.id, s.name, s.address, s.phone, s.website, s.latitude, s.longitude,
			s.google_place_id, s.suggested_category_id, s.notes, s.status,
			s.user_id, s.created_at, s.updated_at,
			c.id, c.name,
			u.id, u.username, u.full_name, u.avatar_url
		FROM restaurant_suggestions s
		LEFT JOIN categories c ON s.suggested_category_id = c.id
		LEFT JOIN users u ON s.user_id = u.id
		WHERE s.id = $1
	`

	var sug models.RestaurantSuggestion
	var catID *int
	var catName *string
	var userID *int
	var username *string
	var fullName *string
	var avatarURL *string

	err = database.GetPool().QueryRow(ctx, query, id).Scan(
		&sug.ID, &sug.Name, &sug.Address, &sug.Phone, &sug.Website, &sug.Latitude, &sug.Longitude,
		&sug.GooglePlaceID, &sug.SuggestedCategoryID, &sug.Notes, &sug.Status,
		&sug.UserID, &sug.CreatedAt, &sug.UpdatedAt,
		&catID, &catName,
		&userID, &username, &fullName, &avatarURL,
	)
	if err != nil {
		http.Error(w, "Suggestion not found", http.StatusNotFound)
		return
	}

	if catID != nil && catName != nil {
		sug.Category = &models.Category{ID: *catID, Name: *catName}
	}

	if userID != nil && username != nil {
		sug.User = &models.UserSummary{
			ID:        *userID,
			Username:  *username,
			FullName:  fullName,
			AvatarURL: avatarURL,
		}
	}

	// Get food types
	foodTypes, err := getFoodTypesForSuggestion(ctx, sug.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sug.FoodTypes = foodTypes

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sug)
}

// @Summary Create a new restaurant suggestion
// @Description Create a new restaurant suggestion with details and optional food types
// @Tags Suggestions
// @Accept json
// @Produce json
// @Param suggestion body models.CreateSuggestionRequest true "Suggestion creation request"
// @Success 201 {object} models.RestaurantSuggestion "Created suggestion"
// @Failure 400 {object} map[string]string "Invalid request body"
// @Failure 409 {object} map[string]string "Suggestion already exists or restaurant exists"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /suggestions [post]
func CreateSuggestion(w http.ResponseWriter, r *http.Request) {
	var req models.CreateSuggestionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	// Get user from context
	user, ok := GetUserFromContext(r)
	var userID *int
	if ok && user != nil {
		userID = &user.ID
	}

	// Check if restaurant already exists in the restaurants table
	var existingRestaurantID int
	var checkQuery string
	var checkArgs []interface{}

	if req.GooglePlaceID != nil && *req.GooglePlaceID != "" {
		// Check by Google Place ID first
		checkQuery = "SELECT id FROM restaurants WHERE google_place_id = $1"
		checkArgs = []interface{}{*req.GooglePlaceID}
	} else if req.Address != nil && *req.Address != "" {
		// Check by name and address combination
		checkQuery = "SELECT id FROM restaurants WHERE LOWER(name) = LOWER($1) AND LOWER(address) = LOWER($2)"
		checkArgs = []interface{}{req.Name, *req.Address}
	}

	if checkQuery != "" {
		err := database.GetPool().QueryRow(ctx, checkQuery, checkArgs...).Scan(&existingRestaurantID)
		if err == nil {
			// Restaurant already exists
			logger.Warn("Attempt to create suggestion for existing restaurant: %s (ID: %d)", req.Name, existingRestaurantID)
			http.Error(w, "This restaurant already exists in the database. Please search for it instead.", http.StatusConflict)
			return
		}
		// If error is "no rows", that's fine - restaurant doesn't exist
	}

	var sug models.RestaurantSuggestion
	err := database.GetPool().QueryRow(ctx,
		`INSERT INTO restaurant_suggestions (name, address, phone, website, latitude, longitude, google_place_id, suggested_category_id, notes, user_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, name, address, phone, website, latitude, longitude, google_place_id, suggested_category_id, notes, status, user_id, created_at, updated_at`,
		req.Name, req.Address, req.Phone, req.Website, req.Latitude, req.Longitude, req.GooglePlaceID, req.SuggestedCategoryID, req.Notes, userID,
	).Scan(
		&sug.ID, &sug.Name, &sug.Address, &sug.Phone, &sug.Website, &sug.Latitude, &sug.Longitude,
		&sug.GooglePlaceID, &sug.SuggestedCategoryID, &sug.Notes, &sug.Status, &sug.UserID, &sug.CreatedAt, &sug.UpdatedAt,
	)
	if err != nil {
		// Check if it's a unique constraint violation
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == "23505" { // unique_violation
				logger.Warn("Duplicate suggestion creation attempt: %s", req.Name)
				if strings.Contains(pgErr.ConstraintName, "google_place_id") {
					http.Error(w, "A suggestion for this restaurant (Google Place ID) already exists", http.StatusConflict)
				} else if strings.Contains(pgErr.ConstraintName, "name_address") {
					http.Error(w, "A suggestion for this restaurant (name and address) already exists", http.StatusConflict)
				} else {
					http.Error(w, "This suggestion already exists", http.StatusConflict)
				}
				return
			}
		}
		logger.Error("Failed to create suggestion: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set food types
	if len(req.FoodTypeIDs) > 0 {
		if err := setFoodTypesForSuggestion(ctx, sug.ID, req.FoodTypeIDs); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		foodTypes, err := getFoodTypesForSuggestion(ctx, sug.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		sug.FoodTypes = foodTypes
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(sug)
}

// @Summary Update suggestion status
// @Description Update the status of a restaurant suggestion (pending, approved, tested, rejected)
// @Tags Suggestions
// @Accept json
// @Produce json
// @Param id path int true "Suggestion ID"
// @Param status body models.UpdateSuggestionStatusRequest true "Status update request"
// @Success 200 {object} models.RestaurantSuggestion "Updated suggestion"
// @Failure 400 {object} map[string]string "Invalid request or status"
// @Failure 404 {object} map[string]string "Suggestion not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /suggestions/{id}/status [patch]
func UpdateSuggestionStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid suggestion ID", http.StatusBadRequest)
		return
	}

	var req models.UpdateSuggestionStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate status
	validStatuses := map[string]bool{"pending": true, "approved": true, "tested": true, "rejected": true}
	if !validStatuses[req.Status] {
		http.Error(w, "Invalid status", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	var sug models.RestaurantSuggestion
	err = database.GetPool().QueryRow(ctx,
		`UPDATE restaurant_suggestions SET status = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING id, name, address, phone, website, latitude, longitude, google_place_id, suggested_category_id, notes, status, created_at, updated_at`,
		req.Status, id,
	).Scan(
		&sug.ID, &sug.Name, &sug.Address, &sug.Phone, &sug.Website, &sug.Latitude, &sug.Longitude,
		&sug.GooglePlaceID, &sug.SuggestedCategoryID, &sug.Notes, &sug.Status, &sug.CreatedAt, &sug.UpdatedAt,
	)
	if err != nil {
		http.Error(w, "Suggestion not found", http.StatusNotFound)
		return
	}

	foodTypes, err := getFoodTypesForSuggestion(ctx, sug.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sug.FoodTypes = foodTypes

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sug)
}

// @Summary Convert suggestion to restaurant
// @Description Convert a restaurant suggestion to a permanent restaurant with initial ratings
// @Tags Suggestions
// @Accept json
// @Produce json
// @Param id path int true "Suggestion ID"
// @Param conversion body models.ConvertSuggestionRequest true "Conversion request with ratings"
// @Success 200 {object} map[string]interface{} "Conversion result with restaurant_id"
// @Failure 400 {object} map[string]string "Invalid request or ratings"
// @Failure 404 {object} map[string]string "Suggestion not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /suggestions/{id}/convert [post]
func ConvertSuggestion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid suggestion ID", http.StatusBadRequest)
		return
	}

	var req models.ConvertSuggestionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	// Get user from context
	user, ok := GetUserFromContext(r)
	var userID *int
	if ok && user != nil {
		userID = &user.ID
	}

	// Get the suggestion
	var sug models.RestaurantSuggestion
	err = database.GetPool().QueryRow(ctx,
		`SELECT id, name, address, phone, website, latitude, longitude, google_place_id, suggested_category_id, status
		FROM restaurant_suggestions WHERE id = $1`, id,
	).Scan(
		&sug.ID, &sug.Name, &sug.Address, &sug.Phone, &sug.Website, &sug.Latitude, &sug.Longitude,
		&sug.GooglePlaceID, &sug.SuggestedCategoryID, &sug.Status,
	)
	if err != nil {
		http.Error(w, "Suggestion not found", http.StatusNotFound)
		return
	}

	// Determine category (override if provided, otherwise use suggested)
	categoryID := sug.SuggestedCategoryID
	if req.CategoryID != nil {
		categoryID = req.CategoryID
	}

	// Validate ratings
	if req.FoodRating < 1 || req.FoodRating > 5 ||
		req.ServiceRating < 1 || req.ServiceRating > 5 ||
		req.AmbianceRating < 1 || req.AmbianceRating > 5 {
		http.Error(w, "Ratings must be between 1 and 5", http.StatusBadRequest)
		return
	}

	// Create restaurant
	var restaurantID int
	err = database.GetPool().QueryRow(ctx,
		`INSERT INTO restaurants (name, description, address, phone, website, latitude, longitude, google_place_id, category_id, user_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id`,
		sug.Name, req.Description, sug.Address, sug.Phone, sug.Website, sug.Latitude, sug.Longitude, sug.GooglePlaceID, categoryID, userID,
	).Scan(&restaurantID)
	if err != nil {
		logger.Error("Failed to create restaurant from suggestion: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Copy food types from suggestion to restaurant
	foodTypes, err := getFoodTypesForSuggestion(ctx, sug.ID)
	if err != nil {
		logger.Warn("Failed to get food types for suggestion %d: %v", sug.ID, err)
	}
	if len(foodTypes) > 0 {
		var foodTypeIDs []int
		for _, ft := range foodTypes {
			foodTypeIDs = append(foodTypeIDs, ft.ID)
		}
		for _, ftID := range foodTypeIDs {
			_, err := database.GetPool().Exec(ctx,
				"INSERT INTO restaurant_food_types (restaurant_id, food_type_id) VALUES ($1, $2)",
				restaurantID, ftID)
			if err != nil {
				// Non-fatal, continue
				continue
			}
		}
	}

	// Create initial rating from the conversion
	_, err = database.GetPool().Exec(ctx,
		`INSERT INTO ratings (restaurant_id, food_rating, service_rating, ambiance_rating, comment, user_id)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		restaurantID, req.FoodRating, req.ServiceRating, req.AmbianceRating, req.Comment, userID,
	)
	if err != nil {
		logger.Warn("Failed to create initial rating for restaurant %d: %v", restaurantID, err)
	}

	// Delete the suggestion after successful conversion
	_, err = database.GetPool().Exec(ctx,
		"DELETE FROM restaurant_suggestions WHERE id = $1", id)
	if err != nil {
		// Non-fatal, but log it
		logger.Warn("Failed to delete converted suggestion %d: %v", id, err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"restaurant_id": restaurantID,
		"message":       "Suggestion converted to restaurant successfully",
	})
}

// @Summary Approve a suggestion
// @Description Approve a restaurant suggestion and convert it to a restaurant
// @Tags Suggestions
// @Accept json
// @Produce json
// @Param id path int true "Suggestion ID"
// @Success 200 {object} map[string]interface{} "Approval result with restaurant_id"
// @Failure 400 {object} map[string]string "Invalid suggestion ID"
// @Failure 404 {object} map[string]string "Suggestion not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /suggestions/{id}/approve [post]
func ApproveSuggestion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid suggestion ID", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	// Get user from context
	user, ok := GetUserFromContext(r)
	var userID *int
	if ok && user != nil {
		userID = &user.ID
	}

	// Get the suggestion
	var sug models.RestaurantSuggestion
	err = database.GetPool().QueryRow(ctx,
		`SELECT id, name, address, phone, website, latitude, longitude, google_place_id, suggested_category_id, status
		FROM restaurant_suggestions WHERE id = $1`, id,
	).Scan(
		&sug.ID, &sug.Name, &sug.Address, &sug.Phone, &sug.Website, &sug.Latitude, &sug.Longitude,
		&sug.GooglePlaceID, &sug.SuggestedCategoryID, &sug.Status,
	)
	if err != nil {
		http.Error(w, "Suggestion not found", http.StatusNotFound)
		return
	}

	// Create restaurant
	var restaurantID int
	err = database.GetPool().QueryRow(ctx,
		`INSERT INTO restaurants (name, address, phone, website, latitude, longitude, google_place_id, category_id, user_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id`,
		sug.Name, sug.Address, sug.Phone, sug.Website, sug.Latitude, sug.Longitude, sug.GooglePlaceID, sug.SuggestedCategoryID, userID,
	).Scan(&restaurantID)
	if err != nil {
		logger.Error("Failed to create restaurant from suggestion: %v", err)
		http.Error(w, "Failed to create restaurant", http.StatusInternalServerError)
		return
	}

	// Copy food types from suggestion to restaurant
	foodTypes, err := getFoodTypesForSuggestion(ctx, sug.ID)
	if err == nil && len(foodTypes) > 0 {
		for _, ft := range foodTypes {
			_, _ = database.GetPool().Exec(ctx,
				"INSERT INTO restaurant_food_types (restaurant_id, food_type_id) VALUES ($1, $2)",
				restaurantID, ft.ID)
		}
	}

	// Update suggestion status to approved
	_, err = database.GetPool().Exec(ctx,
		"UPDATE restaurant_suggestions SET status = 'approved', updated_at = NOW() WHERE id = $1", id)
	if err != nil {
		logger.Warn("Failed to update suggestion status: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"restaurant_id": restaurantID,
		"message":       "Suggestion approved and converted to restaurant successfully",
	})
}

// @Summary Reject a suggestion
// @Description Reject a restaurant suggestion
// @Tags Suggestions
// @Accept json
// @Produce json
// @Param id path int true "Suggestion ID"
// @Success 200 {object} map[string]interface{} "Rejection result"
// @Failure 400 {object} map[string]string "Invalid suggestion ID"
// @Failure 404 {object} map[string]string "Suggestion not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /suggestions/{id}/reject [post]
func RejectSuggestion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid suggestion ID", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	// Update suggestion status to rejected
	result, err := database.GetPool().Exec(ctx,
		"UPDATE restaurant_suggestions SET status = 'rejected', updated_at = NOW() WHERE id = $1", id)
	if err != nil {
		logger.Error("Failed to reject suggestion: %v", err)
		http.Error(w, "Failed to reject suggestion", http.StatusInternalServerError)
		return
	}

	if result.RowsAffected() == 0 {
		http.Error(w, "Suggestion not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Suggestion rejected successfully",
	})
}

// @Summary Delete a suggestion
// @Description Delete a restaurant suggestion by ID
// @Tags Suggestions
// @Accept json
// @Produce json
// @Param id path int true "Suggestion ID"
// @Success 204 "Suggestion deleted successfully"
// @Failure 400 {object} map[string]string "Invalid suggestion ID"
// @Failure 404 {object} map[string]string "Suggestion not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /suggestions/{id} [delete]
func DeleteSuggestion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid suggestion ID", http.StatusBadRequest)
		return
	}

	result, err := database.GetPool().Exec(context.Background(),
		"DELETE FROM restaurant_suggestions WHERE id = $1", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if result.RowsAffected() == 0 {
		http.Error(w, "Suggestion not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
