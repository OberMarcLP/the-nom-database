package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/nomdb/backend/internal/logger"
	"github.com/nomdb/backend/internal/services"
)

var mapsService = services.NewGoogleMapsService()

// @Summary Search for places
// @Description Search for places using Google Maps Places API
// @Tags Google Maps
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Success 200 {array} models.GooglePlaceResult "List of matching places"
// @Failure 400 {object} map[string]string "Missing query parameter"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /places/search [get]
func SearchPlaces(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	results, err := mapsService.SearchPlaces(query)
	if err != nil {
		logger.Error("request failed: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// @Summary Geocode cities
// @Description Geocode city names to get coordinates using Google Maps Geocoding API
// @Tags Google Maps
// @Accept json
// @Produce json
// @Param q query string true "City name to geocode"
// @Success 200 {array} models.GooglePlaceResult "List of geocoded cities"
// @Failure 400 {object} map[string]string "Missing query parameter"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /geocode/cities [get]
func GeocodeCities(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	results, err := mapsService.GeocodeCities(query)
	if err != nil {
		logger.Error("request failed: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// @Summary Get place details
// @Description Get detailed information about a place using Google Maps Place Details API
// @Tags Google Maps
// @Accept json
// @Produce json
// @Param placeId path string true "Google Place ID"
// @Success 200 {object} models.GooglePlaceResult "Place details"
// @Failure 400 {object} map[string]string "Missing place ID"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /places/{placeId} [get]
func GetPlaceDetails(w http.ResponseWriter, r *http.Request) {
	placeID := chi.URLParam(r, "placeId")
	if placeID == "" {
		http.Error(w, "Place ID is required", http.StatusBadRequest)
		return
	}

	result, err := mapsService.GetPlaceDetails(placeID)
	if err != nil {
		logger.Error("request failed: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
