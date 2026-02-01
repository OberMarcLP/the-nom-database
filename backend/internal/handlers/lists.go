package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/nomdb/backend/internal/database"
	"github.com/nomdb/backend/internal/models"
)

// GetUserLists returns all lists for the authenticated user
func GetUserLists(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := RequestContext(r)
	defer cancel()

	user, ok := GetUserFromContext(r)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	query := `
		SELECT
			l.id, l.user_id, l.name, l.description, l.is_public, l.created_at, l.updated_at,
			COUNT(lr.id) as restaurant_count
		FROM restaurant_lists l
		LEFT JOIN restaurant_list_items lr ON l.id = lr.list_id
		WHERE l.user_id = $1
		GROUP BY l.id
		ORDER BY l.updated_at DESC
	`

	rows, err := database.GetPool().Query(ctx, query, user.ID)
	if err != nil {
		log.Printf("ERROR: Failed to fetch lists: %v", err)
		http.Error(w, "Failed to fetch lists", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	lists := []models.RestaurantList{}
	for rows.Next() {
		var list models.RestaurantList
		err := rows.Scan(
			&list.ID, &list.UserID, &list.Name, &list.Description, &list.IsPublic,
			&list.CreatedAt, &list.UpdatedAt, &list.RestaurantCount,
		)
		if err != nil {
			log.Printf("ERROR: Failed to scan list: %v", err)
			continue
		}
		lists = append(lists, list)
	}
	if err := rows.Err(); err != nil {
		log.Printf("ERROR: Rows iteration error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lists)
}

// GetList returns a single list with its restaurants
func GetList(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := RequestContext(r)
	defer cancel()

	vars := mux.Vars(r)
	listID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid list ID", http.StatusBadRequest)
		return
	}

	user, _ := GetUserFromContext(r)

	// Get list details
	var list models.RestaurantList
	query := `
		SELECT id, user_id, name, description, is_public, created_at, updated_at
		FROM restaurant_lists
		WHERE id = $1
	`

	err = database.GetPool().QueryRow(ctx, query, listID).Scan(
		&list.ID, &list.UserID, &list.Name, &list.Description, &list.IsPublic,
		&list.CreatedAt, &list.UpdatedAt,
	)
	if err != nil {
		log.Printf("ERROR: Failed to fetch list: %v", err)
		http.Error(w, "List not found", http.StatusNotFound)
		return
	}

	// Check if user has access to this list
	if !list.IsPublic && (user == nil || user.ID != list.UserID) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Get restaurants in the list
	restaurantsQuery := `
		SELECT
			lr.id, lr.list_id, lr.restaurant_id, lr.notes, lr.added_at,
			r.id, r.name, r.description, r.address, r.phone, r.website,
			r.latitude, r.longitude, r.google_place_id, r.category_id,
			r.created_at, r.updated_at,
			c.id, c.name
		FROM restaurant_list_items lr
		JOIN restaurants r ON lr.restaurant_id = r.id
		LEFT JOIN categories c ON r.category_id = c.id
		WHERE lr.list_id = $1
		ORDER BY lr.added_at DESC
	`

	rows, err := database.GetPool().Query(ctx, restaurantsQuery, listID)
	if err != nil {
		log.Printf("ERROR: Failed to fetch list restaurants: %v", err)
		http.Error(w, "Failed to fetch list restaurants", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	restaurants := []models.ListRestaurant{}
	for rows.Next() {
		var lr models.ListRestaurant
		var restaurant models.Restaurant
		var categoryID *int
		var categoryName *string

		err := rows.Scan(
			&lr.ID, &lr.ListID, &lr.RestaurantID, &lr.Notes, &lr.AddedAt,
			&restaurant.ID, &restaurant.Name, &restaurant.Description, &restaurant.Address,
			&restaurant.Phone, &restaurant.Website, &restaurant.Latitude, &restaurant.Longitude,
			&restaurant.GooglePlaceID, &categoryID, &restaurant.CreatedAt, &restaurant.UpdatedAt,
			&categoryID, &categoryName,
		)
		if err != nil {
			log.Printf("ERROR: Failed to scan list restaurant: %v", err)
			continue
		}

		if categoryID != nil && categoryName != nil {
			restaurant.Category = &models.Category{
				ID:   *categoryID,
				Name: *categoryName,
			}
		}
		lr.Restaurant = &restaurant
		restaurants = append(restaurants, lr)
	}

	response := struct {
		List        models.RestaurantList    `json:"list"`
		Restaurants []models.ListRestaurant  `json:"restaurants"`
	}{
		List:        list,
		Restaurants: restaurants,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CreateList creates a new restaurant list
func CreateList(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := RequestContext(r)
	defer cancel()

	user, ok := GetUserFromContext(r)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.CreateListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "List name is required", http.StatusBadRequest)
		return
	}

	query := `
		INSERT INTO restaurant_lists (user_id, name, description, is_public)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, name, description, is_public, created_at, updated_at
	`

	var list models.RestaurantList
	err := database.GetPool().QueryRow(
		ctx, query,
		user.ID, req.Name, req.Description, req.IsPublic,
	).Scan(
		&list.ID, &list.UserID, &list.Name, &list.Description, &list.IsPublic,
		&list.CreatedAt, &list.UpdatedAt,
	)
	if err != nil {
		log.Printf("ERROR: Failed to create list: %v", err)
		http.Error(w, "Failed to create list", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(list)
}

// UpdateList updates an existing restaurant list
func UpdateList(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := RequestContext(r)
	defer cancel()

	user, ok := GetUserFromContext(r)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	listID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid list ID", http.StatusBadRequest)
		return
	}

	// Verify ownership
	var ownerID int
	err = database.GetPool().QueryRow(ctx,
		"SELECT user_id FROM restaurant_lists WHERE id = $1", listID).Scan(&ownerID)
	if err != nil {
		http.Error(w, "List not found", http.StatusNotFound)
		return
	}
	if ownerID != user.ID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var req models.UpdateListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	query := `
		UPDATE restaurant_lists
		SET name = COALESCE($1, name),
		    description = COALESCE($2, description),
		    is_public = COALESCE($3, is_public)
		WHERE id = $4
		RETURNING id, user_id, name, description, is_public, created_at, updated_at
	`

	var list models.RestaurantList
	err = database.GetPool().QueryRow(
		ctx, query,
		req.Name, req.Description, req.IsPublic, listID,
	).Scan(
		&list.ID, &list.UserID, &list.Name, &list.Description, &list.IsPublic,
		&list.CreatedAt, &list.UpdatedAt,
	)
	if err != nil {
		log.Printf("ERROR: Failed to update list: %v", err)
		http.Error(w, "Failed to update list", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

// DeleteList deletes a restaurant list
func DeleteList(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := RequestContext(r)
	defer cancel()

	user, ok := GetUserFromContext(r)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	listID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid list ID", http.StatusBadRequest)
		return
	}

	// Verify ownership
	var ownerID int
	err = database.GetPool().QueryRow(ctx,
		"SELECT user_id FROM restaurant_lists WHERE id = $1", listID).Scan(&ownerID)
	if err != nil {
		http.Error(w, "List not found", http.StatusNotFound)
		return
	}
	if ownerID != user.ID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	_, err = database.GetPool().Exec(ctx,
		"DELETE FROM restaurant_lists WHERE id = $1", listID)
	if err != nil {
		log.Printf("ERROR: Failed to delete list: %v", err)
		http.Error(w, "Failed to delete list", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AddRestaurantToList adds a restaurant to a list
func AddRestaurantToList(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := RequestContext(r)
	defer cancel()

	user, ok := GetUserFromContext(r)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	listID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid list ID", http.StatusBadRequest)
		return
	}

	// Verify ownership
	var ownerID int
	err = database.GetPool().QueryRow(ctx,
		"SELECT user_id FROM restaurant_lists WHERE id = $1", listID).Scan(&ownerID)
	if err != nil {
		http.Error(w, "List not found", http.StatusNotFound)
		return
	}
	if ownerID != user.ID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var req models.AddRestaurantToListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	query := `
		INSERT INTO restaurant_list_items (list_id, restaurant_id, notes)
		VALUES ($1, $2, $3)
		ON CONFLICT (list_id, restaurant_id) DO UPDATE SET notes = $3
		RETURNING id, list_id, restaurant_id, notes, added_at
	`

	var lr models.ListRestaurant
	err = database.GetPool().QueryRow(
		ctx, query,
		listID, req.RestaurantID, req.Notes,
	).Scan(&lr.ID, &lr.ListID, &lr.RestaurantID, &lr.Notes, &lr.AddedAt)
	if err != nil {
		log.Printf("ERROR: Failed to add restaurant to list: %v", err)
		http.Error(w, "Failed to add restaurant to list", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(lr)
}

// RemoveRestaurantFromList removes a restaurant from a list
func RemoveRestaurantFromList(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := RequestContext(r)
	defer cancel()

	user, ok := GetUserFromContext(r)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	listID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid list ID", http.StatusBadRequest)
		return
	}

	restaurantID, err := strconv.Atoi(vars["restaurantId"])
	if err != nil {
		http.Error(w, "Invalid restaurant ID", http.StatusBadRequest)
		return
	}

	// Verify ownership
	var ownerID int
	err = database.GetPool().QueryRow(ctx,
		"SELECT user_id FROM restaurant_lists WHERE id = $1", listID).Scan(&ownerID)
	if err != nil {
		http.Error(w, "List not found", http.StatusNotFound)
		return
	}
	if ownerID != user.ID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	_, err = database.GetPool().Exec(ctx,
		"DELETE FROM restaurant_list_items WHERE list_id = $1 AND restaurant_id = $2",
		listID, restaurantID)
	if err != nil {
		log.Printf("ERROR: Failed to remove restaurant from list: %v", err)
		http.Error(w, "Failed to remove restaurant from list", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetRestaurantLists returns all lists that contain a specific restaurant for the current user
func GetRestaurantLists(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := RequestContext(r)
	defer cancel()

	user, ok := GetUserFromContext(r)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	restaurantID, err := strconv.Atoi(vars["restaurantId"])
	if err != nil {
		http.Error(w, "Invalid restaurant ID", http.StatusBadRequest)
		return
	}

	query := `
		SELECT
			l.id, l.user_id, l.name, l.description, l.is_public, l.created_at, l.updated_at,
			EXISTS(SELECT 1 FROM restaurant_list_items WHERE list_id = l.id AND restaurant_id = $2) as contains_restaurant
		FROM restaurant_lists l
		WHERE l.user_id = $1
		ORDER BY l.name
	`

	rows, err := database.GetPool().Query(ctx, query, user.ID, restaurantID)
	if err != nil {
		log.Printf("ERROR: Failed to fetch restaurant lists: %v", err)
		http.Error(w, "Failed to fetch lists", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type ListWithStatus struct {
		models.RestaurantList
		ContainsRestaurant bool `json:"contains_restaurant"`
	}

	lists := []ListWithStatus{}
	for rows.Next() {
		var list ListWithStatus
		err := rows.Scan(
			&list.ID, &list.UserID, &list.Name, &list.Description, &list.IsPublic,
			&list.CreatedAt, &list.UpdatedAt, &list.ContainsRestaurant,
		)
		if err != nil {
			log.Printf("ERROR: Failed to scan list: %v", err)
			continue
		}
		lists = append(lists, list)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lists)
}
