package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/nomdb/backend/internal/database"
	"github.com/nomdb/backend/internal/models"
)

// GetRatings godoc
// @Summary Get ratings for a restaurant
// @Description Get all ratings for a specific restaurant
// @Tags Ratings
// @Accept json
// @Produce json
// @Param restaurantId path int true "Restaurant ID"
// @Success 200 {array} models.Rating "List of ratings"
// @Failure 400 {object} map[string]string "Invalid restaurant ID"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /restaurants/{restaurantId}/ratings [get]
func GetRatings(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	restaurantID, err := strconv.Atoi(vars["restaurantId"])
	if err != nil {
		http.Error(w, "Invalid restaurant ID", http.StatusBadRequest)
		return
	}

	rows, err := database.GetPool().Query(context.Background(),
		`SELECT id, restaurant_id, food_rating, service_rating, ambiance_rating, comment, created_at
		FROM ratings WHERE restaurant_id = $1 ORDER BY created_at DESC`, restaurantID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	ratings := []models.Rating{}
	for rows.Next() {
		var rt models.Rating
		if err := rows.Scan(&rt.ID, &rt.RestaurantID, &rt.FoodRating, &rt.ServiceRating, &rt.AmbianceRating, &rt.Comment, &rt.CreatedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		ratings = append(ratings, rt)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ratings)
}

// CreateRating godoc
// @Summary Create a new rating
// @Description Create a new rating for a restaurant
// @Tags Ratings
// @Accept json
// @Produce json
// @Param rating body models.CreateRatingRequest true "Rating creation request"
// @Success 201 {object} models.Rating "Created rating"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /ratings [post]
func CreateRating(w http.ResponseWriter, r *http.Request) {
	var req models.CreateRatingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.RestaurantID == 0 {
		http.Error(w, "Restaurant ID is required", http.StatusBadRequest)
		return
	}

	if req.FoodRating < 1 || req.FoodRating > 5 ||
		req.ServiceRating < 1 || req.ServiceRating > 5 ||
		req.AmbianceRating < 1 || req.AmbianceRating > 5 {
		http.Error(w, "Ratings must be between 1 and 5", http.StatusBadRequest)
		return
	}

	// Check if restaurant exists
	var exists bool
	err := database.GetPool().QueryRow(context.Background(),
		"SELECT EXISTS(SELECT 1 FROM restaurants WHERE id = $1)", req.RestaurantID).Scan(&exists)
	if err != nil || !exists {
		http.Error(w, "Restaurant not found", http.StatusNotFound)
		return
	}

	// Get user from context
	user, ok := GetUserFromContext(r)
	var userID *int
	if ok && user != nil {
		userID = &user.ID
	}

	var rt models.Rating
	err = database.GetPool().QueryRow(context.Background(),
		`INSERT INTO ratings (restaurant_id, food_rating, service_rating, ambiance_rating, comment, user_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, restaurant_id, food_rating, service_rating, ambiance_rating, comment, user_id, created_at`,
		req.RestaurantID, req.FoodRating, req.ServiceRating, req.AmbianceRating, req.Comment, userID,
	).Scan(&rt.ID, &rt.RestaurantID, &rt.FoodRating, &rt.ServiceRating, &rt.AmbianceRating, &rt.Comment, &rt.UserID, &rt.CreatedAt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(rt)
}

// DeleteRating godoc
// @Summary Delete a rating
// @Description Delete a rating by ID
// @Tags Ratings
// @Accept json
// @Produce json
// @Param id path int true "Rating ID"
// @Success 204 "Rating deleted successfully"
// @Failure 400 {object} map[string]string "Invalid rating ID"
// @Failure 404 {object} map[string]string "Rating not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /ratings/{id} [delete]
func DeleteRating(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid rating ID", http.StatusBadRequest)
		return
	}

	result, err := database.GetPool().Exec(context.Background(),
		"DELETE FROM ratings WHERE id = $1", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if result.RowsAffected() == 0 {
		http.Error(w, "Rating not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
