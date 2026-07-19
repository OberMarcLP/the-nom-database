package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/nomdb/backend/internal/database"
	"github.com/nomdb/backend/internal/logger"
	"github.com/nomdb/backend/internal/models"
)

// @Summary Get all ratings for moderation
// @Description Get paginated list of all ratings with filter options
// @Tags Admin
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param restaurant_id query int false "Filter by restaurant ID"
// @Param user_id query int false "Filter by user ID"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {string} string "Internal server error"
// @Security BearerAuth
// @Router /admin/ratings [get]
func AdminListRatings(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	restaurantIDStr := r.URL.Query().Get("restaurant_id")
	userIDStr := r.URL.Query().Get("user_id")

	offset := (page - 1) * limit
	ctx, cancel := RequestContext(r)
	defer cancel()

	// Build query with filters
	query := `
		SELECT r.id, r.restaurant_id, r.user_id, r.food_rating, r.service_rating,
		       r.ambiance_rating, r.comment, r.created_at,
		       rest.name as restaurant_name, u.username
		FROM ratings r
		JOIN restaurants rest ON r.restaurant_id = rest.id
		LEFT JOIN users u ON r.user_id = u.id
		WHERE 1=1
	`

	args := []interface{}{}
	argPos := 1

	if restaurantIDStr != "" {
		restaurantID, err := strconv.Atoi(restaurantIDStr)
		if err == nil {
			query += ` AND r.restaurant_id = $` + strconv.Itoa(argPos)
			args = append(args, restaurantID)
			argPos++
		}
	}

	if userIDStr != "" {
		userID, err := strconv.Atoi(userIDStr)
		if err == nil {
			query += ` AND r.user_id = $` + strconv.Itoa(argPos)
			args = append(args, userID)
			argPos++
		}
	}

	// Get total count
	countQuery := `SELECT COUNT(*) FROM ratings r WHERE 1=1`
	if restaurantIDStr != "" {
		countQuery += ` AND r.restaurant_id = $1`
	}
	if userIDStr != "" && restaurantIDStr == "" {
		countQuery += ` AND r.user_id = $1`
	} else if userIDStr != "" {
		countQuery += ` AND r.user_id = $2`
	}

	var total int
	err := database.GetPool().QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		logger.Error("Failed to count ratings: %v", err)
		http.Error(w, "Failed to count ratings", http.StatusInternalServerError)
		return
	}

	// Get paginated ratings
	query += ` ORDER BY r.created_at DESC LIMIT $` + strconv.Itoa(argPos) + ` OFFSET $` + strconv.Itoa(argPos+1)
	args = append(args, limit, offset)

	rows, err := database.GetPool().Query(ctx, query, args...)
	if err != nil {
		logger.Error("Failed to list ratings: %v", err)
		http.Error(w, "Failed to list ratings", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	ratings := []map[string]interface{}{}
	for rows.Next() {
		var id, restaurantID int
		var userID *int
		var foodRating, serviceRating, ambianceRating int
		var comment *string
		var createdAt time.Time
		var restaurantName string
		var username *string

		err := rows.Scan(&id, &restaurantID, &userID, &foodRating, &serviceRating,
			&ambianceRating, &comment, &createdAt, &restaurantName, &username)
		if err != nil {
			logger.Error("Failed to scan rating row: %v", err)
			continue
		}

		rating := map[string]interface{}{
			"id":              id,
			"restaurant_id":   restaurantID,
			"restaurant_name": restaurantName,
			"food_rating":     foodRating,
			"service_rating":  serviceRating,
			"ambiance_rating": ambianceRating,
			"created_at":      createdAt,
		}

		if userID != nil {
			rating["user_id"] = *userID
		}
		if username != nil {
			rating["username"] = *username
		}
		if comment != nil {
			rating["comment"] = *comment
		}

		ratings = append(ratings, rating)
	}

	response := map[string]interface{}{
		"ratings": ratings,
		"pagination": map[string]interface{}{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"totalPages": (total + limit - 1) / limit,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// @Summary Delete rating (moderation)
// @Description Admin deletes any rating
// @Tags Admin
// @Accept json
// @Produce json
// @Param id path int true "Rating ID"
// @Success 204 "Rating deleted successfully"
// @Failure 400 {string} string "Invalid rating ID"
// @Failure 500 {string} string "Internal server error"
// @Security BearerAuth
// @Router /admin/ratings/{id} [delete]
func AdminDeleteRating(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ratingID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid rating ID", http.StatusBadRequest)
		return
	}

	ctx, cancel := RequestContext(r)
	defer cancel()
	adminUser := r.Context().Value(models.UserContextKey).(*models.User)

	_, err = database.GetPool().Exec(ctx, "DELETE FROM ratings WHERE id = $1", ratingID)
	if err != nil {
		logger.Error("Failed to delete rating: %v", err)
		http.Error(w, "Failed to delete rating", http.StatusInternalServerError)
		return
	}

	CreateAuditLog(ctx, adminUser.ID, "delete_rating", "ratings", ratingID, map[string]string{
		"reason": "moderation",
	}, r)

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Get all photos for moderation
// @Description Get paginated list of all photos (menu and review)
// @Tags Admin
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param type query string false "Filter by type: menu or review"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {string} string "Internal server error"
// @Security BearerAuth
// @Router /admin/photos [get]
func AdminListPhotos(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	photoType := r.URL.Query().Get("type")
	offset := (page - 1) * limit
	ctx, cancel := RequestContext(r)
	defer cancel()

	photos := []map[string]interface{}{}
	var total int

	if photoType == "review" || photoType == "" {
		// Get review photos
		var reviewTotal int
		database.GetPool().QueryRow(ctx, "SELECT COUNT(*) FROM review_photos").Scan(&reviewTotal)
		total += reviewTotal

		if photoType == "review" || photoType == "" {
			query := `
				SELECT rp.id, rp.rating_id, rp.user_id, rp.filename, rp.photo_url, rp.caption,
				       rp.file_size, rp.created_at, u.username, r.restaurant_id, rest.name as restaurant_name
				FROM review_photos rp
				LEFT JOIN users u ON rp.user_id = u.id
				LEFT JOIN ratings r ON rp.rating_id = r.id
				LEFT JOIN restaurants rest ON r.restaurant_id = rest.id
				ORDER BY rp.created_at DESC
				LIMIT $1 OFFSET $2
			`

			rows, err := database.GetPool().Query(ctx, query, limit, offset)
			if err != nil {
				logger.Error("Failed to list review photos: %v", err)
				http.Error(w, "Failed to list photos", http.StatusInternalServerError)
				return
			}
			defer rows.Close()

			for rows.Next() {
				var id, ratingID int
				var userID *int
				var restaurantID *int
				var filename, photoURL string
				var caption *string
				var fileSize *int
				var createdAt time.Time
				var username *string
				var restaurantName *string

				err := rows.Scan(&id, &ratingID, &userID, &filename, &photoURL, &caption, &fileSize, &createdAt, &username, &restaurantID, &restaurantName)
				if err != nil {
					logger.Error("Failed to scan review photo: %v", err)
					continue
				}

				photo := map[string]interface{}{
					"id":         id,
					"type":       "review",
					"rating_id":  ratingID,
					"filename":   photoURL, // Use photo_url as filename for consistency with frontend
					"created_at": createdAt,
				}

				if userID != nil {
					photo["user_id"] = *userID
				}
				if username != nil {
					photo["username"] = *username
				}
				if caption != nil {
					photo["caption"] = *caption
				}
				if fileSize != nil {
					photo["file_size"] = *fileSize
				}
				if restaurantID != nil {
					photo["restaurant_id"] = *restaurantID
				}
				if restaurantName != nil {
					photo["restaurant_name"] = *restaurantName
				}

				photos = append(photos, photo)
			}
		}
	}

	if photoType == "menu" || photoType == "" {
		// Get menu photos
		var menuTotal int
		database.GetPool().QueryRow(ctx, "SELECT COUNT(*) FROM menu_photos").Scan(&menuTotal)
		total += menuTotal

		if photoType == "menu" || photoType == "" {
			query := `
				SELECT mp.id, mp.restaurant_id, mp.user_id, mp.filename, mp.caption,
				       mp.file_size, mp.created_at, r.name as restaurant_name, u.username
				FROM menu_photos mp
				JOIN restaurants r ON mp.restaurant_id = r.id
				LEFT JOIN users u ON mp.user_id = u.id
				ORDER BY mp.created_at DESC
				LIMIT $1 OFFSET $2
			`

			rows, err := database.GetPool().Query(ctx, query, limit, offset)
			if err != nil {
				logger.Error("Failed to list menu photos: %v", err)
				http.Error(w, "Failed to list photos", http.StatusInternalServerError)
				return
			}
			defer rows.Close()

			for rows.Next() {
				var id, restaurantID int
				var userID *int
				var filename, caption string
				var fileSize *int
				var createdAt time.Time
				var restaurantName string
				var username *string

				err := rows.Scan(&id, &restaurantID, &userID, &filename, &caption,
					&fileSize, &createdAt, &restaurantName, &username)
				if err != nil {
					logger.Error("Failed to scan menu photo: %v", err)
					continue
				}

				photo := map[string]interface{}{
					"id":              id,
					"type":            "menu",
					"restaurant_id":   restaurantID,
					"restaurant_name": restaurantName,
					"filename":        filename,
					"caption":         caption,
					"created_at":      createdAt,
				}

				if userID != nil {
					photo["user_id"] = *userID
				}
				if username != nil {
					photo["username"] = *username
				}
				if fileSize != nil {
					photo["file_size"] = *fileSize
				}

				photos = append(photos, photo)
			}
		}
	}

	response := map[string]interface{}{
		"photos": photos,
		"pagination": map[string]interface{}{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"totalPages": (total + limit - 1) / limit,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// @Summary Delete photo (moderation)
// @Description Admin deletes any photo
// @Tags Admin
// @Accept json
// @Produce json
// @Param type path string true "Photo type: menu or review"
// @Param id path int true "Photo ID"
// @Success 204 "Photo deleted successfully"
// @Failure 400 {string} string "Invalid request"
// @Failure 500 {string} string "Internal server error"
// @Security BearerAuth
// @Router /admin/photos/{type}/{id} [delete]
func AdminDeletePhoto(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	photoType := vars["type"]
	photoID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid photo ID", http.StatusBadRequest)
		return
	}

	if photoType != "menu" && photoType != "review" {
		http.Error(w, "Invalid photo type", http.StatusBadRequest)
		return
	}

	ctx, cancel := RequestContext(r)
	defer cancel()
	adminUser := r.Context().Value(models.UserContextKey).(*models.User)

	// Get photo filename before deleting
	var filename string
	var table string
	if photoType == "menu" {
		table = "menu_photos"
	} else {
		table = "review_photos"
	}

	err = database.GetPool().QueryRow(ctx,
		`SELECT filename FROM `+table+` WHERE id = $1`, photoID).Scan(&filename)

	if err == sql.ErrNoRows {
		http.Error(w, "Photo not found", http.StatusNotFound)
		return
	}
	if err != nil {
		logger.Error("Failed to get photo: %v", err)
		http.Error(w, "Failed to get photo", http.StatusInternalServerError)
		return
	}

	// Delete from database
	_, err = database.GetPool().Exec(ctx, `DELETE FROM `+table+` WHERE id = $1`, photoID)
	if err != nil {
		logger.Error("Failed to delete photo from database: %v", err)
		http.Error(w, "Failed to delete photo", http.StatusInternalServerError)
		return
	}

	// Delete file from filesystem
	uploadsDir := "./uploads"
	photoPath := filepath.Join(uploadsDir, filename)
	if err := os.Remove(photoPath); err != nil {
		logger.Warn("Failed to delete photo file %s: %v", photoPath, err)
	}

	CreateAuditLog(ctx, adminUser.ID, "delete_photo", table, photoID, map[string]string{
		"type":     photoType,
		"filename": filename,
		"reason":   "moderation",
	}, r)

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Update restaurant (moderation)
// @Description Admin updates any restaurant (override ownership)
// @Tags Admin
// @Accept json
// @Produce json
// @Param id path int true "Restaurant ID"
// @Param request body map[string]interface{} true "Update fields"
// @Success 200 {object} models.Restaurant
// @Failure 400 {string} string "Invalid request"
// @Failure 500 {string} string "Internal server error"
// @Security BearerAuth
// @Router /admin/restaurants/{id} [put]
func AdminUpdateRestaurant(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	restaurantID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid restaurant ID", http.StatusBadRequest)
		return
	}

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx, cancel := RequestContext(r)
	defer cancel()
	adminUser := r.Context().Value(models.UserContextKey).(*models.User)

	// Build dynamic update query (simplified version - add more fields as needed)
	updates := []string{"updated_at = NOW()"}
	args := []interface{}{restaurantID}
	argPos := 2

	if name, ok := req["name"].(string); ok {
		updates = append(updates, "name = $"+strconv.Itoa(argPos))
		args = append(args, name)
		argPos++
	}

	if description, ok := req["description"].(string); ok {
		updates = append(updates, "description = $"+strconv.Itoa(argPos))
		args = append(args, description)
		argPos++
	}

	if address, ok := req["address"].(string); ok {
		updates = append(updates, "address = $"+strconv.Itoa(argPos))
		args = append(args, address)
		argPos++
	}

	query := "UPDATE restaurants SET " + strings.Join(updates, ", ") + " WHERE id = $1"
	_, err = database.GetPool().Exec(ctx, query, args...)

	if err != nil {
		logger.Error("Failed to update restaurant: %v", err)
		http.Error(w, "Failed to update restaurant", http.StatusInternalServerError)
		return
	}

	CreateAuditLog(ctx, adminUser.ID, "update_restaurant", "restaurants", restaurantID, req, r)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Restaurant updated successfully",
	})
}

// @Summary Delete restaurant (moderation)
// @Description Admin deletes any restaurant (override ownership)
// @Tags Admin
// @Accept json
// @Produce json
// @Param id path int true "Restaurant ID"
// @Success 204 "Restaurant deleted successfully"
// @Failure 400 {string} string "Invalid restaurant ID"
// @Failure 500 {string} string "Internal server error"
// @Security BearerAuth
// @Router /admin/restaurants/{id} [delete]
func AdminDeleteRestaurant(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	restaurantID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid restaurant ID", http.StatusBadRequest)
		return
	}

	ctx, cancel := RequestContext(r)
	defer cancel()
	adminUser := r.Context().Value(models.UserContextKey).(*models.User)

	_, err = database.GetPool().Exec(ctx, "DELETE FROM restaurants WHERE id = $1", restaurantID)
	if err != nil {
		logger.Error("Failed to delete restaurant: %v", err)
		http.Error(w, "Failed to delete restaurant", http.StatusInternalServerError)
		return
	}

	CreateAuditLog(ctx, adminUser.ID, "delete_restaurant", "restaurants", restaurantID, nil, r)

	w.WriteHeader(http.StatusNoContent)
}
