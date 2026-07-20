package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/nomdb/backend/internal/database"
	"github.com/nomdb/backend/internal/logger"
	"github.com/nomdb/backend/internal/models"
)

// GetFoodTypes godoc
// @Summary List all food types
// @Description Get a list of all food types ordered by name
// @Tags Food Types
// @Accept json
// @Produce json
// @Success 200 {array} models.FoodType "List of food types"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /food-types [get]
func GetFoodTypes(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := RequestContext(r)
	defer cancel()

	rows, err := database.GetPool().Query(ctx,
		"SELECT id, name, created_at, updated_at FROM food_types ORDER BY name")
	if err != nil {
		logger.Error("request failed: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	foodTypes := []models.FoodType{}
	for rows.Next() {
		var ft models.FoodType
		if err := rows.Scan(&ft.ID, &ft.Name, &ft.CreatedAt, &ft.UpdatedAt); err != nil {
			logger.Error("request failed: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		foodTypes = append(foodTypes, ft)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=300")
	json.NewEncoder(w).Encode(foodTypes)
}

// GetFoodType godoc
// @Summary Get a food type by ID
// @Description Get detailed information about a specific food type
// @Tags Food Types
// @Accept json
// @Produce json
// @Param id path int true "Food Type ID"
// @Success 200 {object} models.FoodType "Food type details"
// @Failure 400 {object} map[string]string "Invalid food type ID"
// @Failure 404 {object} map[string]string "Food type not found"
// @Router /food-types/{id} [get]
func GetFoodType(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := RequestContext(r)
	defer cancel()

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil || id <= 0 {
		http.Error(w, "Invalid food type ID", http.StatusBadRequest)
		return
	}

	var ft models.FoodType
	err = database.GetPool().QueryRow(ctx,
		"SELECT id, name, created_at, updated_at FROM food_types WHERE id = $1", id).
		Scan(&ft.ID, &ft.Name, &ft.CreatedAt, &ft.UpdatedAt)
	if err != nil {
		http.Error(w, "Food type not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=300")
	json.NewEncoder(w).Encode(ft)
}

// CreateFoodType godoc
// @Summary Create a new food type
// @Description Create a new food type with the provided name
// @Tags Food Types
// @Accept json
// @Produce json
// @Param foodType body models.CreateFoodTypeRequest true "Food type creation request"
// @Success 201 {object} models.FoodType "Created food type"
// @Failure 400 {object} map[string]string "Invalid request body or name is required"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /food-types [post]
func CreateFoodType(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := RequestContext(r)
	defer cancel()

	var req models.CreateFoodTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	var ft models.FoodType
	err := database.GetPool().QueryRow(ctx,
		"INSERT INTO food_types (name) VALUES ($1) RETURNING id, name, created_at, updated_at",
		req.Name).Scan(&ft.ID, &ft.Name, &ft.CreatedAt, &ft.UpdatedAt)
	if err != nil {
		logger.Error("request failed: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ft)
}

// UpdateFoodType godoc
// @Summary Update a food type
// @Description Update an existing food type's name
// @Tags Food Types
// @Accept json
// @Produce json
// @Param id path int true "Food Type ID"
// @Param foodType body models.CreateFoodTypeRequest true "Food type update request"
// @Success 200 {object} models.FoodType "Updated food type"
// @Failure 400 {object} map[string]string "Invalid request or name is required"
// @Failure 404 {object} map[string]string "Food type not found"
// @Router /food-types/{id} [put]
func UpdateFoodType(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := RequestContext(r)
	defer cancel()

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil || id <= 0 {
		http.Error(w, "Invalid food type ID", http.StatusBadRequest)
		return
	}

	var req models.CreateFoodTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	var ft models.FoodType
	err = database.GetPool().QueryRow(ctx,
		"UPDATE food_types SET name = $1, updated_at = NOW() WHERE id = $2 RETURNING id, name, created_at, updated_at",
		req.Name, id).Scan(&ft.ID, &ft.Name, &ft.CreatedAt, &ft.UpdatedAt)
	if err != nil {
		http.Error(w, "Food type not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ft)
}

// DeleteFoodType godoc
// @Summary Delete a food type
// @Description Delete a food type by ID
// @Tags Food Types
// @Accept json
// @Produce json
// @Param id path int true "Food Type ID"
// @Success 204 "Food type deleted successfully"
// @Failure 400 {object} map[string]string "Invalid food type ID"
// @Failure 404 {object} map[string]string "Food type not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /food-types/{id} [delete]
func DeleteFoodType(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := RequestContext(r)
	defer cancel()

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil || id <= 0 {
		http.Error(w, "Invalid food type ID", http.StatusBadRequest)
		return
	}

	result, err := database.GetPool().Exec(ctx,
		"DELETE FROM food_types WHERE id = $1", id)
	if err != nil {
		logger.Error("request failed: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if result.RowsAffected() == 0 {
		http.Error(w, "Food type not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
