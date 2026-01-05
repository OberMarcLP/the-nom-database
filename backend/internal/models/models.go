package models

import "time"

// UserSummary represents basic user information for display purposes
type UserSummary struct {
	ID        int     `json:"id"`
	Username  string  `json:"username"`
	FullName  *string `json:"full_name"`
	AvatarURL *string `json:"avatar_url"`
}

type Category struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type FoodType struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Restaurant struct {
	ID            int          `json:"id"`
	Name          string       `json:"name"`
	Description   *string      `json:"description"`
	Address       *string      `json:"address"`
	Phone         *string      `json:"phone"`
	Website       *string      `json:"website"`
	Latitude      *float64     `json:"latitude"`
	Longitude     *float64     `json:"longitude"`
	GooglePlaceID *string      `json:"google_place_id"`
	CategoryID    *int         `json:"category_id"`
	Category      *Category    `json:"category,omitempty"`
	FoodTypes     []FoodType   `json:"food_types,omitempty"`
	AvgRating     *AvgRating   `json:"avg_rating,omitempty"`
	Distance      *float64     `json:"distance,omitempty"` // Distance in km from search location
	IsSuggestion  bool         `json:"is_suggestion"`      // Indicates if this is from suggestions table
	SuggestionID  *int         `json:"suggestion_id,omitempty"`
	Status        *string      `json:"status,omitempty"`    // For suggestions: pending, approved, tested, rejected
	CreatedBy     *int         `json:"created_by,omitempty"` // User ID who created this restaurant
	UpdatedBy     *int         `json:"updated_by,omitempty"` // User ID who last updated this restaurant
	CreatedByUser *UserSummary `json:"created_by_user,omitempty"`
	UpdatedByUser *UserSummary `json:"updated_by_user,omitempty"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
}

type Rating struct {
	ID             int          `json:"id"`
	RestaurantID   int          `json:"restaurant_id"`
	UserID         *int         `json:"user_id,omitempty"`
	User           *UserSummary `json:"user,omitempty"`
	FoodRating     int          `json:"food_rating"`
	ServiceRating  int          `json:"service_rating"`
	AmbianceRating int          `json:"ambiance_rating"`
	Comment        *string      `json:"comment"`
	CreatedAt      time.Time    `json:"created_at"`
}

type AvgRating struct {
	Food     float64 `json:"food"`
	Service  float64 `json:"service"`
	Ambiance float64 `json:"ambiance"`
	Overall  float64 `json:"overall"`
	Count    int     `json:"count"`
}

// Request/Response types
type CreateRestaurantRequest struct {
	Name          string   `json:"name"`
	Description   *string  `json:"description"`
	Address       *string  `json:"address"`
	Phone         *string  `json:"phone"`
	Website       *string  `json:"website"`
	Latitude      *float64 `json:"latitude"`
	Longitude     *float64 `json:"longitude"`
	GooglePlaceID *string  `json:"google_place_id"`
	CategoryID    *int     `json:"category_id"`
	FoodTypeIDs   []int    `json:"food_type_ids"`
}

type UpdateRestaurantRequest struct {
	Name          *string  `json:"name"`
	Description   *string  `json:"description"`
	Address       *string  `json:"address"`
	Phone         *string  `json:"phone"`
	Website       *string  `json:"website"`
	Latitude      *float64 `json:"latitude"`
	Longitude     *float64 `json:"longitude"`
	GooglePlaceID *string  `json:"google_place_id"`
	CategoryID    *int     `json:"category_id"`
	FoodTypeIDs   []int    `json:"food_type_ids"`
}

type CreateRatingRequest struct {
	RestaurantID   int     `json:"restaurant_id"`
	FoodRating     int     `json:"food_rating"`
	ServiceRating  int     `json:"service_rating"`
	AmbianceRating int     `json:"ambiance_rating"`
	Comment        *string `json:"comment"`
}

type CreateCategoryRequest struct {
	Name string `json:"name"`
}

type CreateFoodTypeRequest struct {
	Name string `json:"name"`
}

// Pagination types
type PaginationParams struct {
	Limit  int    `json:"limit"`
	Cursor string `json:"cursor,omitempty"`
}

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	NextCursor *string     `json:"next_cursor,omitempty"`
	HasMore    bool        `json:"has_more"`
	Total      *int        `json:"total,omitempty"`
}

type GooglePlaceResult struct{
	PlaceID   string  `json:"place_id"`
	Name      string  `json:"name"`
	Address   string  `json:"address"`
	Phone     string  `json:"phone,omitempty"`
	Website   string  `json:"website,omitempty"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// Restaurant Suggestion System
type RestaurantSuggestion struct {
	ID                  int          `json:"id"`
	Name                string       `json:"name"`
	Address             *string      `json:"address"`
	Phone               *string      `json:"phone"`
	Website             *string      `json:"website"`
	Latitude            *float64     `json:"latitude"`
	Longitude           *float64     `json:"longitude"`
	GooglePlaceID       *string      `json:"google_place_id"`
	SuggestedCategoryID *int         `json:"suggested_category_id"`
	Category            *Category    `json:"category,omitempty"`
	FoodTypes           []FoodType   `json:"food_types,omitempty"`
	Notes               *string      `json:"notes"`
	Status              string       `json:"status"`
	UserID              *int         `json:"user_id,omitempty"`
	User                *UserSummary `json:"user,omitempty"`
	CreatedAt           time.Time    `json:"created_at"`
	UpdatedAt           time.Time    `json:"updated_at"`
}

type CreateSuggestionRequest struct {
	Name                string   `json:"name"`
	Address             *string  `json:"address"`
	Phone               *string  `json:"phone"`
	Website             *string  `json:"website"`
	Latitude            *float64 `json:"latitude"`
	Longitude           *float64 `json:"longitude"`
	GooglePlaceID       *string  `json:"google_place_id"`
	SuggestedCategoryID *int     `json:"suggested_category_id"`
	FoodTypeIDs         []int    `json:"food_type_ids"`
	Notes               *string  `json:"notes"`
}

type UpdateSuggestionStatusRequest struct {
	Status string `json:"status"`
}

type ConvertSuggestionRequest struct {
	Description    *string `json:"description"`
	CategoryID     *int    `json:"category_id"`
	FoodRating     int     `json:"food_rating"`
	ServiceRating  int     `json:"service_rating"`
	AmbianceRating int     `json:"ambiance_rating"`
	Comment        *string `json:"comment"`
}

// Menu Photos
type MenuPhoto struct {
	ID               int       `json:"id"`
	RestaurantID     int       `json:"restaurant_id"`
	Filename         string    `json:"filename"`
	OriginalFilename *string   `json:"original_filename"`
	Caption          string    `json:"caption"`
	FileSize         *int      `json:"file_size"`
	MimeType         *string   `json:"mime_type"`
	URL              string    `json:"url"` // Computed field
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type UploadPhotoResponse struct {
	Photo MenuPhoto `json:"photo"`
}
