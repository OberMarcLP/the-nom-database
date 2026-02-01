package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/nomdb/backend/internal/database"
	"github.com/nomdb/backend/internal/logger"
	"github.com/nomdb/backend/internal/models"
	"github.com/nomdb/backend/internal/services"
)

const (
	maxUploadSize    = 5 << 20 // 5MB
	uploadsDir       = "./uploads/menu_photos"
	thumbnailsSubdir = "thumbnails"
)

// Image magic bytes for validation
var imageMagicBytes = map[string][][]byte{
	"image/jpeg": {
		{0xFF, 0xD8, 0xFF}, // JPEG
	},
	"image/png": {
		{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, // PNG
	},
	"image/webp": {
		// WebP starts with RIFF....WEBP
		{0x52, 0x49, 0x46, 0x46}, // RIFF header (WebP uses RIFF container)
	},
}

// validateImageMagicBytes checks if the file content matches allowed image types
// Returns the detected mime type and whether validation passed
func validateImageMagicBytes(data []byte) (string, bool) {
	for mimeType, signatures := range imageMagicBytes {
		for _, sig := range signatures {
			if len(data) >= len(sig) && bytes.HasPrefix(data, sig) {
				// Special check for WebP - need to verify WEBP marker at offset 8
				if mimeType == "image/webp" {
					if len(data) >= 12 {
						webpMarker := data[8:12]
						if string(webpMarker) == "WEBP" {
							return mimeType, true
						}
					}
					continue
				}
				return mimeType, true
			}
		}
	}
	return "", false
}

func init() {
	// Create uploads directory if it doesn't exist (fallback for local storage)
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		logger.Warn("Failed to create uploads directory: %v", err)
	}
}

// @Summary Get menu photos for a restaurant
// @Description Retrieve all menu photos for a specific restaurant with presigned URLs
// @Tags Photos
// @Accept json
// @Produce json
// @Param restaurantId path int true "Restaurant ID"
// @Success 200 {array} models.MenuPhoto "List of menu photos"
// @Failure 400 {object} map[string]string "Invalid restaurant ID"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /restaurants/{restaurantId}/photos [get]
func GetMenuPhotos(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	restaurantID, err := strconv.Atoi(vars["restaurantId"])
	if err != nil {
		http.Error(w, "Invalid restaurant ID", http.StatusBadRequest)
		return
	}

	ctx, cancel := RequestContext(r)
	defer cancel()
	rows, err := database.GetPool().Query(ctx,
		`SELECT id, restaurant_id, filename, original_filename, caption, file_size, mime_type, created_at, updated_at
		FROM menu_photos
		WHERE restaurant_id = $1
		ORDER BY created_at DESC`, restaurantID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	photos := []models.MenuPhoto{}
	s3Service := services.GetS3Service()

	for rows.Next() {
		var photo models.MenuPhoto
		if err := rows.Scan(
			&photo.ID, &photo.RestaurantID, &photo.Filename, &photo.OriginalFilename,
			&photo.Caption, &photo.FileSize, &photo.MimeType, &photo.CreatedAt, &photo.UpdatedAt,
		); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Generate URL based on storage type
		if s3Service != nil {
			// Generate presigned URL for S3 (valid for 1 hour)
			presignedURL, err := s3Service.GetPresignedURL(ctx, fmt.Sprintf("menu_photos/%s", photo.Filename), time.Hour)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to generate URL: %v", err), http.StatusInternalServerError)
				return
			}
			photo.URL = presignedURL
		} else {
			// Use local file URL
			photo.URL = fmt.Sprintf("/api/uploads/menu_photos/%s", photo.Filename)
		}

		photos = append(photos, photo)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(photos)
}

// @Summary Upload a menu photo
// @Description Upload a menu photo for a restaurant (JPEG, PNG, or WebP, max 5MB)
// @Tags Photos
// @Accept multipart/form-data
// @Produce json
// @Param restaurantId path int true "Restaurant ID"
// @Param photo formData file true "Menu photo file"
// @Param caption formData string true "Photo caption"
// @Success 201 {object} models.UploadPhotoResponse "Uploaded photo details"
// @Failure 400 {object} map[string]string "Invalid request or file"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /restaurants/{restaurantId}/photos [post]
func UploadMenuPhoto(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	restaurantID, err := strconv.Atoi(vars["restaurantId"])
	if err != nil {
		http.Error(w, "Invalid restaurant ID", http.StatusBadRequest)
		return
	}

	// Parse multipart form
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}

	// Get caption
	caption := r.FormValue("caption")
	if caption == "" {
		http.Error(w, "Caption is required", http.StatusBadRequest)
		return
	}

	// Get file
	file, header, err := r.FormFile("photo")
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file size
	if header.Size > maxUploadSize {
		http.Error(w, fmt.Sprintf("File too large. Maximum size is %d MB", maxUploadSize/(1<<20)), http.StatusBadRequest)
		return
	}

	// Validate file type by reading magic bytes (more secure than Content-Type header)
	// Read the first 12 bytes for magic byte detection
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
		logger.Warn("Content-Type mismatch: header=%s, detected=%s", contentType, detectedType)
	}

	// Process image (resize, compress, generate thumbnail)
	imageProcessor := services.NewImageProcessor()
	fullImage, thumbnail, err := imageProcessor.ProcessUpload(file, header.Filename)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to process image: %v", err), http.StatusBadRequest)
		return
	}

	// Generate unique filename (always use .jpg extension after processing)
	filename := uuid.New().String() + ".jpg"
	thumbnailFilename := uuid.New().String() + "_thumb.jpg"

	ctx, cancel := RequestContext(r)
	defer cancel()
	s3Service := services.GetS3Service()
	var fileSize int64 = int64(len(fullImage))
	var photoURL string

	if s3Service != nil {
		// Upload full image to S3
		s3Key := fmt.Sprintf("menu_photos/%s", filename)
		_, err = s3Service.UploadFile(ctx, s3Key, bytes.NewReader(fullImage), "image/jpeg")
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to upload file to S3: %v", err), http.StatusInternalServerError)
			return
		}

		// Upload thumbnail to S3
		s3ThumbKey := fmt.Sprintf("menu_photos/thumbnails/%s", thumbnailFilename)
		_, err = s3Service.UploadFile(ctx, s3ThumbKey, bytes.NewReader(thumbnail), "image/jpeg")
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to upload thumbnail to S3: %v", err), http.StatusInternalServerError)
			return
		}

		// Generate presigned URL for immediate response
		photoURL, err = s3Service.GetPresignedURL(ctx, s3Key, time.Hour)
		if err != nil {
			http.Error(w, "Failed to generate URL", http.StatusInternalServerError)
			return
		}
	} else {
		// Fallback to local storage
		filePath := filepath.Join(uploadsDir, filename)
		if err := os.WriteFile(filePath, fullImage, 0644); err != nil {
			http.Error(w, "Failed to save file", http.StatusInternalServerError)
			return
		}

		// Save thumbnail
		thumbnailDir := filepath.Join(uploadsDir, thumbnailsSubdir)
		if err := os.MkdirAll(thumbnailDir, 0755); err != nil {
			logger.Warn("Failed to create thumbnail directory: %v", err)
		}
		thumbnailPath := filepath.Join(thumbnailDir, thumbnailFilename)
		if err := os.WriteFile(thumbnailPath, thumbnail, 0644); err != nil {
			http.Error(w, "Failed to save thumbnail", http.StatusInternalServerError)
			return
		}

		photoURL = fmt.Sprintf("/api/uploads/menu_photos/%s", filename)
	}

	// Save to database (always use image/jpeg as mime type after processing)
	var photo models.MenuPhoto
	err = database.GetPool().QueryRow(ctx,
		`INSERT INTO menu_photos (restaurant_id, filename, original_filename, caption, file_size, mime_type)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, restaurant_id, filename, original_filename, caption, file_size, mime_type, created_at, updated_at`,
		restaurantID, filename, header.Filename, caption, int(fileSize), "image/jpeg",
	).Scan(
		&photo.ID, &photo.RestaurantID, &photo.Filename, &photo.OriginalFilename,
		&photo.Caption, &photo.FileSize, &photo.MimeType, &photo.CreatedAt, &photo.UpdatedAt,
	)
	if err != nil {
		// Clean up uploaded file on database error
		if s3Service != nil {
			if delErr := s3Service.DeleteFile(ctx, fmt.Sprintf("menu_photos/%s", filename)); delErr != nil {
				logger.Warn("Failed to delete file after database error: %v", delErr)
			}
		} else {
			if delErr := os.Remove(filepath.Join(uploadsDir, filename)); delErr != nil {
				logger.Warn("Failed to delete file after database error: %v", delErr)
			}
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set URL
	photo.URL = photoURL

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(models.UploadPhotoResponse{Photo: photo})
}

// @Summary Update photo caption
// @Description Update the caption of a menu photo
// @Tags Photos
// @Accept json
// @Produce json
// @Param id path int true "Photo ID"
// @Param caption body object{caption=string} true "Caption update request"
// @Success 200 {object} models.MenuPhoto "Updated photo"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 404 {object} map[string]string "Photo not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /photos/{id} [put]
func UpdatePhotoCaption(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid photo ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Caption string `json:"caption"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Caption == "" {
		http.Error(w, "Caption is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := RequestContext(r)
	defer cancel()
	var photo models.MenuPhoto
	err = database.GetPool().QueryRow(ctx,
		`UPDATE menu_photos SET caption = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING id, restaurant_id, filename, original_filename, caption, file_size, mime_type, created_at, updated_at`,
		req.Caption, id,
	).Scan(
		&photo.ID, &photo.RestaurantID, &photo.Filename, &photo.OriginalFilename,
		&photo.Caption, &photo.FileSize, &photo.MimeType, &photo.CreatedAt, &photo.UpdatedAt,
	)
	if err != nil {
		http.Error(w, "Photo not found", http.StatusNotFound)
		return
	}

	photo.URL = fmt.Sprintf("/api/uploads/menu_photos/%s", photo.Filename)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(photo)
}

// @Summary Delete a menu photo
// @Description Delete a menu photo by ID
// @Tags Photos
// @Accept json
// @Produce json
// @Param id path int true "Photo ID"
// @Success 204 "Photo deleted successfully"
// @Failure 400 {object} map[string]string "Invalid photo ID"
// @Failure 404 {object} map[string]string "Photo not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /photos/{id} [delete]
func DeleteMenuPhoto(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid photo ID", http.StatusBadRequest)
		return
	}

	ctx, cancel := RequestContext(r)
	defer cancel()

	// Get filename before deleting from DB
	var filename string
	err = database.GetPool().QueryRow(ctx,
		"SELECT filename FROM menu_photos WHERE id = $1", id).Scan(&filename)
	if err != nil {
		http.Error(w, "Photo not found", http.StatusNotFound)
		return
	}

	// Delete from database
	result, err := database.GetPool().Exec(ctx,
		"DELETE FROM menu_photos WHERE id = $1", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if result.RowsAffected() == 0 {
		http.Error(w, "Photo not found", http.StatusNotFound)
		return
	}

	// Delete file from storage (non-fatal if fails)
	s3Service := services.GetS3Service()
	if s3Service != nil {
		// Delete from S3
		if delErr := s3Service.DeleteFile(ctx, fmt.Sprintf("menu_photos/%s", filename)); delErr != nil {
			logger.Warn("Failed to delete file from S3: %v", delErr)
		}
	} else {
		// Delete from local disk
		filePath := filepath.Join(uploadsDir, filename)
		if delErr := os.Remove(filePath); delErr != nil {
			logger.Warn("Failed to delete file from disk: %v", delErr)
		}
	}

	w.WriteHeader(http.StatusNoContent)
}
