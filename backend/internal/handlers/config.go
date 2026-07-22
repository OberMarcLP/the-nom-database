package handlers

import (
	"encoding/json"
	"net/http"
	"os"
)

// GetPublicConfig godoc
// @Summary Get public runtime configuration
// @Description Returns non-secret configuration the browser needs (Maps JS API key, map ID). Keys served here must be restricted by HTTP referrer in the Google Cloud console.
// @Tags Config
// @Produce json
// @Success 200 {object} map[string]string
// @Router /config [get]
func GetPublicConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"google_maps_api_key": os.Getenv("GOOGLE_MAPS_API_KEY"),
		"google_maps_map_id":  os.Getenv("GOOGLE_MAPS_MAP_ID"),
	})
}
