package handlers

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/nomdb/backend/internal/database"
	"github.com/nomdb/backend/internal/logger"
	"github.com/nomdb/backend/internal/services"
)

// menuPhotoKeys returns the storage keys of a menu photo and its thumbnail.
// The thumbnail shares the photo's UUID (<uuid>.jpg -> <uuid>_thumb.jpg);
// thumbnails uploaded before that convention existed stay orphaned.
func menuPhotoKeys(filename string) []string {
	base := strings.TrimSuffix(filename, ".jpg")
	return []string{
		"menu_photos/" + filename,
		"menu_photos/thumbnails/" + base + "_thumb.jpg",
	}
}

// reviewPhotoKey returns the storage key of a review photo.
func reviewPhotoKey(filename string) string {
	return "review_photos/" + filename
}

// collectRestaurantPhotoKeys gathers the storage keys of every photo that
// belongs to a restaurant. Must run BEFORE the restaurant row is deleted -
// the ON DELETE CASCADE wipes the rows the keys come from.
func collectRestaurantPhotoKeys(ctx context.Context, restaurantID int) []string {
	var keys []string

	menuRows, err := database.GetPool().Query(ctx,
		"SELECT filename FROM menu_photos WHERE restaurant_id = $1", restaurantID)
	if err == nil {
		for menuRows.Next() {
			var filename string
			if menuRows.Scan(&filename) == nil && filename != "" {
				keys = append(keys, menuPhotoKeys(filename)...)
			}
		}
		menuRows.Close()
	}

	reviewRows, err := database.GetPool().Query(ctx,
		`SELECT rp.filename FROM review_photos rp
		JOIN ratings r ON rp.rating_id = r.id
		WHERE r.restaurant_id = $1`, restaurantID)
	if err == nil {
		for reviewRows.Next() {
			var filename string
			if reviewRows.Scan(&filename) == nil && filename != "" {
				keys = append(keys, reviewPhotoKey(filename))
			}
		}
		reviewRows.Close()
	}

	return keys
}

// collectRatingPhotoKeys gathers the storage keys of a rating's photos.
// Must run before the rating row is deleted (CASCADE).
func collectRatingPhotoKeys(ctx context.Context, ratingID int) []string {
	var keys []string
	rows, err := database.GetPool().Query(ctx,
		"SELECT filename FROM review_photos WHERE rating_id = $1", ratingID)
	if err == nil {
		for rows.Next() {
			var filename string
			if rows.Scan(&filename) == nil && filename != "" {
				keys = append(keys, reviewPhotoKey(filename))
			}
		}
		rows.Close()
	}
	return keys
}

// deleteStoredPhotos removes photo objects from S3 (or the local uploads
// fallback) best-effort: failures are logged and never fail the request.
func deleteStoredPhotos(ctx context.Context, keys []string) {
	if len(keys) == 0 {
		return
	}
	s3Service := services.GetS3Service()
	for _, key := range keys {
		if s3Service != nil {
			if err := s3Service.DeleteFile(ctx, key); err != nil {
				logger.Warn("Failed to delete %s from S3: %v", key, err)
			}
		} else if err := os.Remove(filepath.Join(uploadsDir, filepath.FromSlash(key))); err != nil && !os.IsNotExist(err) {
			logger.Warn("Failed to delete local file %s: %v", key, err)
		}
	}
}
