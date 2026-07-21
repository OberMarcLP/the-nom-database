package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/nomdb/backend/internal/database"
	"github.com/nomdb/backend/internal/logger"
	"github.com/nomdb/backend/internal/models"
)

// GetCategories godoc
// @Summary List all categories
// @Description Get a list of all cultural categories
// @Tags Categories
// @Accept json
// @Produce json
// @Success 200 {array} models.Category "List of categories"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /categories [get]
func GetCategories(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := RequestContext(r)
	defer cancel()

	rows, err := database.GetPool().Query(ctx,
		"SELECT id, name, created_at, updated_at FROM categories ORDER BY name")
	if err != nil {
		logger.Error("request failed: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	categories := []models.Category{}
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.CreatedAt, &c.UpdatedAt); err != nil {
			logger.Error("request failed: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		categories = append(categories, c)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=300")
	if err := json.NewEncoder(w).Encode(categories); err != nil {
		logger.Error("Failed to encode response: %v", err)
	}
}

// GetCategory godoc
// @Summary Get a category by ID
// @Description Get detailed information about a specific category
// @Tags Categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {object} models.Category "Category details"
// @Failure 400 {object} map[string]string "Invalid category ID"
// @Failure 404 {object} map[string]string "Category not found"
// @Router /categories/{id} [get]
func GetCategory(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := RequestContext(r)
	defer cancel()

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	var c models.Category
	err = database.GetPool().QueryRow(ctx,
		"SELECT id, name, created_at, updated_at FROM categories WHERE id = $1", id).
		Scan(&c.ID, &c.Name, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		http.Error(w, "Category not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=300")
	if err := json.NewEncoder(w).Encode(c); err != nil {
		logger.Error("Failed to encode response: %v", err)
	}
}

// CreateCategory godoc
// @Summary Create a new category
// @Description Create a new cultural category with the provided name
// @Tags Categories
// @Accept json
// @Produce json
// @Param category body models.CreateCategoryRequest true "Category creation request"
// @Success 201 {object} models.Category "Created category"
// @Failure 400 {object} map[string]string "Invalid request body or name is required"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /categories [post]
func CreateCategory(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := RequestContext(r)
	defer cancel()

	var req models.CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	var c models.Category
	err := database.GetPool().QueryRow(ctx,
		"INSERT INTO categories (name) VALUES ($1) RETURNING id, name, created_at, updated_at",
		req.Name).Scan(&c.ID, &c.Name, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		logger.Error("request failed: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(c)
}

// UpdateCategory godoc
// @Summary Update a category
// @Description Update an existing category's name
// @Tags Categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Param category body models.CreateCategoryRequest true "Category update request"
// @Success 200 {object} models.Category "Updated category"
// @Failure 400 {object} map[string]string "Invalid request or name is required"
// @Failure 404 {object} map[string]string "Category not found"
// @Router /categories/{id} [put]
func UpdateCategory(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := RequestContext(r)
	defer cancel()

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	var req models.CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	var c models.Category
	err = database.GetPool().QueryRow(ctx,
		"UPDATE categories SET name = $1, updated_at = NOW() WHERE id = $2 RETURNING id, name, created_at, updated_at",
		req.Name, id).Scan(&c.ID, &c.Name, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		http.Error(w, "Category not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(c)
}

// DeleteCategory godoc
// @Summary Delete a category
// @Description Delete a category by ID
// @Tags Categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Success 204 "Category deleted successfully"
// @Failure 400 {object} map[string]string "Invalid category ID"
// @Failure 404 {object} map[string]string "Category not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /categories/{id} [delete]
func DeleteCategory(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := RequestContext(r)
	defer cancel()

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	result, err := database.GetPool().Exec(ctx,
		"DELETE FROM categories WHERE id = $1", id)
	if err != nil {
		logger.Error("request failed: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if result.RowsAffected() == 0 {
		http.Error(w, "Category not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
