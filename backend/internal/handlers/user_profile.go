package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/nomdb/backend/internal/database"
	"github.com/nomdb/backend/internal/logger"
	"github.com/nomdb/backend/internal/models"
	"github.com/nomdb/backend/internal/services"
)

// GetUserProfile returns a user's profile with stats
func GetUserProfile(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := RequestContext(r)
	defer cancel()

	vars := mux.Vars(r)
	userIDStr := vars["id"]

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Get user basic info
	var user models.User
	err = database.GetPool().QueryRow(ctx, `
		SELECT id, username, email, full_name, avatar_url, created_at, updated_at
		FROM users WHERE id = $1
	`, userID).Scan(
		&user.ID, &user.Username, &user.Email, &user.FullName,
		&user.AvatarURL, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Email is PII and the login identifier - only the owner or admins see it
	if requester, ok := GetUserFromContext(r); !ok || requester == nil ||
		(requester.ID != user.ID && !requester.IsAdmin) {
		user.Email = ""
	}

	// Get user stats
	var stats struct {
		TotalReviews      int     `json:"total_reviews"`
		TotalRestaurants  int     `json:"total_restaurants"`
		TotalLists        int     `json:"total_lists"`
		AvgFoodRating     float64 `json:"avg_food_rating"`
		AvgServiceRating  float64 `json:"avg_service_rating"`
		AvgAmbianceRating float64 `json:"avg_ambiance_rating"`
	}

	// Get review counts and averages
	err = database.GetPool().QueryRow(ctx, `
		SELECT
			COUNT(*) as total_reviews,
			COUNT(DISTINCT restaurant_id) as total_restaurants,
			COALESCE(AVG(food_rating), 0) as avg_food,
			COALESCE(AVG(service_rating), 0) as avg_service,
			COALESCE(AVG(ambiance_rating), 0) as avg_ambiance
		FROM ratings
		WHERE user_id = $1
	`, userID).Scan(
		&stats.TotalReviews, &stats.TotalRestaurants,
		&stats.AvgFoodRating, &stats.AvgServiceRating, &stats.AvgAmbianceRating,
	)
	if err != nil {
		http.Error(w, "Failed to fetch user stats", http.StatusInternalServerError)
		return
	}

	// Get list count
	err = database.GetPool().QueryRow(ctx, `
		SELECT COUNT(*) FROM restaurant_lists WHERE user_id = $1
	`, userID).Scan(&stats.TotalLists)
	if err != nil {
		http.Error(w, "Failed to fetch list count", http.StatusInternalServerError)
		return
	}

	response := struct {
		User  models.User `json:"user"`
		Stats interface{} `json:"stats"`
	}{
		User:  user,
		Stats: stats,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetUserReviews returns all reviews by a user
func GetUserReviews(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := RequestContext(r)
	defer cancel()

	vars := mux.Vars(r)
	userIDStr := vars["id"]

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	query := `
		SELECT
			r.id, r.restaurant_id, r.user_id, r.food_rating, r.service_rating,
			r.ambiance_rating, r.comment, r.helpful_count, r.not_helpful_count,
			r.created_at, r.updated_at,
			rest.id, rest.name, rest.address,
			u.id, u.username, u.full_name, u.avatar_url
		FROM ratings r
		JOIN restaurants rest ON r.restaurant_id = rest.id
		LEFT JOIN users u ON r.user_id = u.id
		WHERE r.user_id = $1
		ORDER BY r.created_at DESC
	`

	rows, err := database.GetPool().Query(ctx, query, userID)
	if err != nil {
		http.Error(w, "Failed to fetch reviews", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	reviews := []models.Rating{}
	ratingIDs := []int{}
	for rows.Next() {
		var rating models.Rating
		var restaurant models.Restaurant
		var user models.UserSummary

		err := rows.Scan(
			&rating.ID, &rating.RestaurantID, &rating.UserID, &rating.FoodRating,
			&rating.ServiceRating, &rating.AmbianceRating, &rating.Comment,
			&rating.HelpfulCount, &rating.NotHelpfulCount, &rating.CreatedAt, &rating.UpdatedAt,
			&restaurant.ID, &restaurant.Name, &restaurant.Address,
			&user.ID, &user.Username, &user.FullName, &user.AvatarURL,
		)
		if err != nil {
			continue
		}

		rating.Restaurant = &restaurant
		rating.User = &user
		ratingIDs = append(ratingIDs, rating.ID)
		reviews = append(reviews, rating)
	}

	photosByRating, err := getReviewPhotosByRatingIDs(ctx, ratingIDs)
	if err != nil {
		http.Error(w, "Failed to fetch review photos", http.StatusInternalServerError)
		return
	}
	for i := range reviews {
		reviews[i].Photos = photosByRating[reviews[i].ID]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reviews)
}

// UpdateUserProfile updates user profile information
func UpdateUserProfile(w http.ResponseWriter, r *http.Request) {
	user, ok := GetUserFromContext(r)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Username *string `json:"username"`
		FullName *string `json:"full_name"`
		Email    *string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx, cancel := RequestContext(r)
	defer cancel()

	// Build dynamic update query
	updates := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.Username != nil {
		// OIDC users cannot change username
		if user.Provider == "oidc" {
			http.Error(w, "OIDC users cannot change username", http.StatusBadRequest)
			return
		}
		updates = append(updates, "username = $"+strconv.Itoa(argIndex))
		args = append(args, *req.Username)
		argIndex++
	}
	if req.FullName != nil {
		updates = append(updates, "full_name = $"+strconv.Itoa(argIndex))
		args = append(args, *req.FullName)
		argIndex++
	}
	if req.Email != nil {
		updates = append(updates, "email = $"+strconv.Itoa(argIndex))
		args = append(args, *req.Email)
		argIndex++
	}

	if len(updates) == 0 {
		http.Error(w, "No fields to update", http.StatusBadRequest)
		return
	}

	// Add user ID as last parameter
	args = append(args, user.ID)

	query := "UPDATE users SET " + strings.Join(updates, ", ") +
		", updated_at = CURRENT_TIMESTAMP WHERE id = $" + strconv.Itoa(argIndex) +
		" RETURNING id, username, email, full_name, avatar_url, created_at, updated_at"

	var updatedUser models.User
	err := database.GetPool().QueryRow(ctx, query, args...).Scan(
		&updatedUser.ID, &updatedUser.Username, &updatedUser.Email,
		&updatedUser.FullName, &updatedUser.AvatarURL,
		&updatedUser.CreatedAt, &updatedUser.UpdatedAt,
	)
	if err != nil {
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedUser)
}

const (
	maxAvatarSize = 5 << 20 // 5MB
)

// @Summary Upload user avatar
// @Description Upload a profile picture for the authenticated user (JPEG, PNG, or WebP, max 5MB)
// @Tags User
// @Accept multipart/form-data
// @Produce json
// @Param avatar formData file true "Avatar image file"
// @Success 200 {object} map[string]string "Avatar URL"
// @Failure 400 {object} map[string]string "Invalid request or file"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /user/avatar [post]
// @Security BearerAuth
func UploadAvatar(w http.ResponseWriter, r *http.Request) {
	uploadsDir := "./uploads"

	// Get authenticated user from context
	user, ok := r.Context().Value(models.UserContextKey).(*models.User)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse multipart form
	r.Body = http.MaxBytesReader(w, r.Body, maxAvatarSize)
	if err := r.ParseMultipartForm(maxAvatarSize); err != nil {
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}

	// Get file
	file, header, err := r.FormFile("avatar")
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file size
	if header.Size > maxAvatarSize {
		http.Error(w, fmt.Sprintf("File too large. Maximum size is %d MB", maxAvatarSize/(1<<20)), http.StatusBadRequest)
		return
	}

	// Validate file type by reading magic bytes (more secure than Content-Type header)
	magicBytes := make([]byte, 12)
	n, err := file.Read(magicBytes)
	if err != nil || n < 3 {
		http.Error(w, "Failed to read file for validation", http.StatusBadRequest)
		return
	}
	// Seek back to the beginning for processing
	if _, err := file.Seek(0, 0); err != nil {
		http.Error(w, "Failed to process file", http.StatusInternalServerError)
		return
	}

	// Validate magic bytes match allowed image types
	detectedType, valid := validateImageMagicBytes(magicBytes[:n])
	if !valid {
		http.Error(w, "Invalid image file. Only JPEG, PNG, and WebP images are allowed", http.StatusBadRequest)
		return
	}

	// Log if Content-Type doesn't match detected type (potential spoofing attempt)
	contentType := header.Header.Get("Content-Type")
	if contentType != detectedType {
		logger.Warn("Content-Type mismatch in avatar upload: header=%s, detected=%s", contentType, detectedType)
	}

	// Process image (resize to square, compress)
	imageProcessor := services.NewImageProcessor()
	processedImage, _, err := imageProcessor.ProcessUpload(file, header.Filename)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to process image: %v", err), http.StatusBadRequest)
		return
	}

	// Generate unique filename
	filename := fmt.Sprintf("avatar_%d_%s.jpg", user.ID, uuid.New().String())

	ctx, cancel := RequestContext(r)
	defer cancel()
	s3Service := services.GetS3Service()
	var avatarURL string

	if s3Service != nil {
		// Upload to S3
		s3Key := fmt.Sprintf("avatars/%s", filename)
		_, err = s3Service.UploadFile(ctx, s3Key, bytes.NewReader(processedImage), "image/jpeg")
		if err != nil {
			logger.Error("Failed to upload avatar to S3: %v", err)
			http.Error(w, "Failed to upload file", http.StatusInternalServerError)
			return
		}

		// Generate presigned URL (valid for 1 year since avatars don't change often)
		avatarURL, err = s3Service.GetPresignedURL(ctx, s3Key, 24*365*time.Hour)
		if err != nil {
			logger.Error("Failed to generate presigned URL: %v", err)
			http.Error(w, "Failed to generate URL", http.StatusInternalServerError)
			return
		}
	} else {
		// Use local storage (for development)
		logger.Warn("S3 not configured, using local storage for avatars")

		// Ensure avatars directory exists
		avatarsDir := filepath.Join(uploadsDir, "avatars")
		if err := os.MkdirAll(avatarsDir, 0755); err != nil {
			logger.Error("Failed to create avatars directory: %v", err)
			http.Error(w, "Failed to save file", http.StatusInternalServerError)
			return
		}

		// Save avatar to local storage
		filePath := filepath.Join(avatarsDir, filename)
		if err := os.WriteFile(filePath, processedImage, 0644); err != nil {
			logger.Error("Failed to save avatar file: %v", err)
			http.Error(w, "Failed to save file", http.StatusInternalServerError)
			return
		}

		avatarURL = fmt.Sprintf("/api/uploads/avatars/%s", filename)
		logger.Debug("Avatar saved to local storage: %s", filePath)
	}

	// Update user's avatar_url in database
	_, err = database.GetPool().Exec(ctx,
		`UPDATE users SET avatar_url = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`,
		avatarURL, user.ID)
	if err != nil {
		logger.Error("Failed to update user avatar: %v", err)
		http.Error(w, "Failed to update avatar", http.StatusInternalServerError)
		return
	}

	logger.Info("User %d uploaded new avatar: %s", user.ID, avatarURL)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"avatar_url": avatarURL,
		"message":    "Avatar uploaded successfully",
	})
}
