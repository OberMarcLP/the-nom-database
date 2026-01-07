package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/nomdb/backend/internal/database"
	"github.com/nomdb/backend/internal/models"
	"github.com/nomdb/backend/internal/services"
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

	// Get current user ID if authenticated
	var currentUserID *int
	if user, ok := r.Context().Value(models.UserContextKey).(*models.User); ok && user != nil {
		currentUserID = &user.ID
	}

	rows, err := database.GetPool().Query(context.Background(),
		`SELECT r.id, r.restaurant_id, r.food_rating, r.service_rating, r.ambiance_rating,
			r.comment, r.user_id, r.created_at, r.updated_at, r.helpful_count, r.not_helpful_count,
			u.id, u.username, u.full_name, u.avatar_url,
			CASE WHEN $2::integer IS NOT NULL THEN rv.vote_type ELSE NULL END as user_vote
		FROM ratings r
		LEFT JOIN users u ON r.user_id = u.id
		LEFT JOIN review_votes rv ON r.id = rv.rating_id AND rv.user_id = $2
		WHERE r.restaurant_id = $1
		ORDER BY r.created_at DESC`, restaurantID, currentUserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	ratings := []models.Rating{}
	for rows.Next() {
		var rt models.Rating
		var userID *int
		var username *string
		var fullName *string
		var avatarURL *string

		if err := rows.Scan(&rt.ID, &rt.RestaurantID, &rt.FoodRating, &rt.ServiceRating, &rt.AmbianceRating,
			&rt.Comment, &rt.UserID, &rt.CreatedAt, &rt.UpdatedAt, &rt.HelpfulCount, &rt.NotHelpfulCount,
			&userID, &username, &fullName, &avatarURL, &rt.UserVote); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if userID != nil && username != nil {
			rt.User = &models.UserSummary{
				ID:        *userID,
				Username:  *username,
				FullName:  fullName,
				AvatarURL: avatarURL,
			}
		}

		// Fetch review photos for this rating
		photoRows, err := database.GetPool().Query(context.Background(),
			`SELECT id, rating_id, photo_url, caption, display_order, created_at
			FROM review_photos
			WHERE rating_id = $1
			ORDER BY display_order, created_at`, rt.ID)
		if err == nil {
			defer photoRows.Close()
			photos := []models.ReviewPhoto{}
			for photoRows.Next() {
				var photo models.ReviewPhoto
				if err := photoRows.Scan(&photo.ID, &photo.RatingID, &photo.PhotoURL, &photo.Caption, &photo.DisplayOrder, &photo.CreatedAt); err == nil {
					photos = append(photos, photo)
				}
			}
			rt.Photos = photos
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
		RETURNING id, restaurant_id, food_rating, service_rating, ambiance_rating, comment, user_id, created_at, updated_at, helpful_count, not_helpful_count`,
		req.RestaurantID, req.FoodRating, req.ServiceRating, req.AmbianceRating, req.Comment, userID,
	).Scan(&rt.ID, &rt.RestaurantID, &rt.FoodRating, &rt.ServiceRating, &rt.AmbianceRating, &rt.Comment, &rt.UserID, &rt.CreatedAt, &rt.UpdatedAt, &rt.HelpfulCount, &rt.NotHelpfulCount)
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

// VoteOnReview godoc
// @Summary Vote on a review
// @Description Vote a review as helpful or not helpful
// @Tags Ratings
// @Accept json
// @Produce json
// @Param id path int true "Rating ID"
// @Param vote body map[string]string true "Vote type: helpful or not_helpful"
// @Success 200 {object} models.Rating "Updated rating with vote counts"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Rating not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /ratings/{id}/vote [post]
func VoteOnReview(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ratingID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid rating ID", http.StatusBadRequest)
		return
	}

	// Get user from context
	user, ok := GetUserFromContext(r)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		VoteType string `json:"vote_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.VoteType != "helpful" && req.VoteType != "not_helpful" {
		http.Error(w, "Vote type must be 'helpful' or 'not_helpful'", http.StatusBadRequest)
		return
	}

	// Check if rating exists
	var exists bool
	err = database.GetPool().QueryRow(context.Background(),
		"SELECT EXISTS(SELECT 1 FROM ratings WHERE id = $1)", ratingID).Scan(&exists)
	if err != nil || !exists {
		http.Error(w, "Rating not found", http.StatusNotFound)
		return
	}

	// Upsert vote (insert or update if exists)
	_, err = database.GetPool().Exec(context.Background(),
		`INSERT INTO review_votes (rating_id, user_id, vote_type)
		VALUES ($1, $2, $3)
		ON CONFLICT (rating_id, user_id)
		DO UPDATE SET vote_type = $3, created_at = CURRENT_TIMESTAMP`,
		ratingID, user.ID, req.VoteType)
	if err != nil {
		log.Printf("ERROR: Failed to insert/update vote - ratingID=%d, userID=%d, voteType=%s, error=%v",
			ratingID, user.ID, req.VoteType, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return updated rating with vote counts
	var rt models.Rating
	err = database.GetPool().QueryRow(context.Background(),
		`SELECT id, restaurant_id, food_rating, service_rating, ambiance_rating,
			comment, user_id, created_at, updated_at, helpful_count, not_helpful_count
		FROM ratings WHERE id = $1`, ratingID).Scan(
		&rt.ID, &rt.RestaurantID, &rt.FoodRating, &rt.ServiceRating, &rt.AmbianceRating,
		&rt.Comment, &rt.UserID, &rt.CreatedAt, &rt.UpdatedAt, &rt.HelpfulCount, &rt.NotHelpfulCount)
	if err != nil {
		log.Printf("ERROR: Failed to fetch updated rating - ratingID=%d, error=%v", ratingID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rt.UserVote = &req.VoteType

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rt)
}

// RemoveVote godoc
// @Summary Remove vote from a review
// @Description Remove your vote from a review
// @Tags Ratings
// @Accept json
// @Produce json
// @Param id path int true "Rating ID"
// @Success 200 {object} models.Rating "Updated rating without vote"
// @Failure 400 {object} map[string]string "Invalid rating ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /ratings/{id}/vote [delete]
func RemoveVote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ratingID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid rating ID", http.StatusBadRequest)
		return
	}

	// Get user from context
	user, ok := GetUserFromContext(r)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Delete vote
	_, err = database.GetPool().Exec(context.Background(),
		"DELETE FROM review_votes WHERE rating_id = $1 AND user_id = $2",
		ratingID, user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return updated rating
	var rt models.Rating
	err = database.GetPool().QueryRow(context.Background(),
		`SELECT id, restaurant_id, food_rating, service_rating, ambiance_rating,
			comment, user_id, created_at, updated_at, helpful_count, not_helpful_count
		FROM ratings WHERE id = $1`, ratingID).Scan(
		&rt.ID, &rt.RestaurantID, &rt.FoodRating, &rt.ServiceRating, &rt.AmbianceRating,
		&rt.Comment, &rt.UserID, &rt.CreatedAt, &rt.UpdatedAt, &rt.HelpfulCount, &rt.NotHelpfulCount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rt)
}

// UploadReviewPhoto godoc
// @Summary Upload a photo to a review
// @Description Upload a photo to attach to a review/rating
// @Tags Ratings
// @Accept multipart/form-data
// @Produce json
// @Param id path int true "Rating ID"
// @Param photo formData file true "Photo file"
// @Param caption formData string false "Photo caption"
// @Success 201 {object} models.ReviewPhoto "Photo uploaded successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /ratings/{id}/photos [post]
func UploadReviewPhoto(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ratingID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid rating ID", http.StatusBadRequest)
		return
	}

	// Get user from context
	user, ok := GetUserFromContext(r)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Verify the rating belongs to this user
	var ownerID int
	err = database.GetPool().QueryRow(context.Background(),
		"SELECT user_id FROM ratings WHERE id = $1", ratingID).Scan(&ownerID)
	if err != nil {
		http.Error(w, "Rating not found", http.StatusNotFound)
		return
	}
	if ownerID != user.ID {
		http.Error(w, "You can only upload photos to your own reviews", http.StatusForbidden)
		return
	}

	// Parse multipart form
	err = r.ParseMultipartForm(10 << 20) // 10MB max
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("photo")
	if err != nil {
		http.Error(w, "Photo file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	caption := r.FormValue("caption")

	// Process image (resize, compress, generate thumbnail)
	imageProcessor := services.NewImageProcessor()
	fullImage, _, err := imageProcessor.ProcessUpload(file, handler.Filename)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to process image: %v", err), http.StatusBadRequest)
		return
	}

	// Generate unique filename
	filename := uuid.New().String() + ".jpg"

	ctx := context.Background()
	s3Service := services.GetS3Service()
	var photoURL string
	var s3Err error

	// Try S3 first if configured
	if s3Service != nil {
		s3Key := fmt.Sprintf("review_photos/%s", filename)
		_, s3Err = s3Service.UploadFile(ctx, s3Key, bytes.NewReader(fullImage), "image/jpeg")
		if s3Err == nil {
			// S3 upload succeeded, get presigned URL
			photoURL, err = s3Service.GetPresignedURL(ctx, s3Key, time.Hour)
			if err != nil {
				log.Printf("WARN: S3 upload succeeded but presigned URL failed, falling back to local: %v", err)
				s3Err = err // Treat as S3 failure
			}
		} else {
			log.Printf("WARN: S3 upload failed, falling back to local storage: %v", s3Err)
		}
	}

	// Fallback to local storage if S3 is not configured or failed
	if s3Service == nil || s3Err != nil {
		localPath := filepath.Join("./uploads/review_photos", filename)
		if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
			http.Error(w, "Failed to create upload directory", http.StatusInternalServerError)
			return
		}
		if err := os.WriteFile(localPath, fullImage, 0644); err != nil {
			http.Error(w, "Failed to save file locally", http.StatusInternalServerError)
			return
		}
		photoURL = fmt.Sprintf("/api/uploads/review_photos/%s", filename)
		log.Printf("INFO: Review photo saved locally - ratingID=%d, path=%s", ratingID, localPath)
	}

	// Get current max display_order for this rating
	var maxOrder int
	err = database.GetPool().QueryRow(context.Background(),
		"SELECT COALESCE(MAX(display_order), -1) FROM review_photos WHERE rating_id = $1", ratingID).Scan(&maxOrder)
	if err != nil {
		log.Printf("ERROR: Failed to get max display_order - ratingID=%d, error=%v", ratingID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Insert photo record
	var photo models.ReviewPhoto
	err = database.GetPool().QueryRow(context.Background(),
		`INSERT INTO review_photos (rating_id, photo_url, caption, display_order)
		VALUES ($1, $2, $3, $4)
		RETURNING id, rating_id, photo_url, caption, display_order, created_at`,
		ratingID, photoURL, caption, maxOrder+1).Scan(
		&photo.ID, &photo.RatingID, &photo.PhotoURL, &photo.Caption, &photo.DisplayOrder, &photo.CreatedAt)
	if err != nil {
		log.Printf("ERROR: Failed to insert review photo - ratingID=%d, error=%v", ratingID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(photo)
}

// DeleteReviewPhoto godoc
// @Summary Delete a review photo
// @Description Delete a photo from a review
// @Tags Ratings
// @Accept json
// @Produce json
// @Param id path int true "Photo ID"
// @Success 204 "Photo deleted successfully"
// @Failure 400 {object} map[string]string "Invalid photo ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Photo not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /review-photos/{id} [delete]
func DeleteReviewPhoto(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	photoID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid photo ID", http.StatusBadRequest)
		return
	}

	// Get user from context
	user, ok := GetUserFromContext(r)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Verify the photo belongs to a rating owned by this user
	var ownerID int
	var photoURL string
	err = database.GetPool().QueryRow(context.Background(),
		`SELECT r.user_id, rp.photo_url FROM review_photos rp
		JOIN ratings r ON rp.rating_id = r.id
		WHERE rp.id = $1`, photoID).Scan(&ownerID, &photoURL)
	if err != nil {
		http.Error(w, "Photo not found", http.StatusNotFound)
		return
	}
	if ownerID != user.ID {
		http.Error(w, "You can only delete photos from your own reviews", http.StatusForbidden)
		return
	}

	// Delete from database
	result, err := database.GetPool().Exec(context.Background(),
		"DELETE FROM review_photos WHERE id = $1", photoID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if result.RowsAffected() == 0 {
		http.Error(w, "Photo not found", http.StatusNotFound)
		return
	}

	// TODO: Delete from S3 if needed (optional cleanup)
	// services.DeleteFromS3(photoURL)

	w.WriteHeader(http.StatusNoContent)
}
