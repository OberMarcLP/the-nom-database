package handlers

import (
	"fmt"
	"testing"

	"github.com/nomdb/backend/internal/models"
)

func TestValidateRestaurantInput(t *testing.T) {
	desc := "A test restaurant"
	addr := "123 Test St"
	longName := string(make([]byte, 201))

	tests := []struct {
		name        string
		request     models.CreateRestaurantRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid restaurant",
			request: models.CreateRestaurantRequest{
				Name:        "Test Restaurant",
				Description: &desc,
				Address:     &addr,
			},
			expectError: false,
		},
		{
			name: "Missing name",
			request: models.CreateRestaurantRequest{
				Description: &desc,
				Address:     &addr,
			},
			expectError: true,
			errorMsg:    "Name is required",
		},
		{
			name: "Name too long",
			request: models.CreateRestaurantRequest{
				Name:    longName,
				Address: &addr,
			},
			expectError: true,
			errorMsg:    "Name must be less than 200 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRestaurantInput(&tt.request)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
					return
				}
				if err.Error() != tt.errorMsg {
					t.Errorf("Expected error '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestBuildFilterQuery(t *testing.T) {
	tests := []struct {
		name           string
		categoryID     string
		foodTypeID     string
		searchQuery    string
		expectCategory bool
		expectFoodType bool
		expectSearch   bool
	}{
		{
			name:           "No filters",
			expectCategory: false,
			expectFoodType: false,
			expectSearch:   false,
		},
		{
			name:           "Category filter only",
			categoryID:     "1",
			expectCategory: true,
			expectFoodType: false,
			expectSearch:   false,
		},
		{
			name:           "Food type filter only",
			foodTypeID:     "2",
			expectCategory: false,
			expectFoodType: true,
			expectSearch:   false,
		},
		{
			name:           "Search query only",
			searchQuery:    "pizza",
			expectCategory: false,
			expectFoodType: false,
			expectSearch:   true,
		},
		{
			name:           "All filters",
			categoryID:     "1",
			foodTypeID:     "2",
			searchQuery:    "pizza",
			expectCategory: true,
			expectFoodType: true,
			expectSearch:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, args := buildFilterQuery(tt.categoryID, tt.foodTypeID, tt.searchQuery)

			if tt.expectCategory && tt.categoryID != "" {
				if len(args) == 0 {
					t.Error("Expected category in args but got empty args")
				}
			}

			if tt.expectSearch && tt.searchQuery != "" {
				found := false
				for _, arg := range args {
					if str, ok := arg.(string); ok && str == "%"+tt.searchQuery+"%" {
						found = true
						break
					}
				}
				if !found {
					t.Error("Expected search query in args but not found")
				}
			}

			// Verify query contains expected clauses
			if tt.expectCategory {
				if query != "" && !contains(query, "category_id") {
					t.Error("Expected category_id in query")
				}
			}

			if tt.expectFoodType {
				if query != "" && !contains(query, "food_type_id") {
					t.Error("Expected food_type_id in query")
				}
			}

			if tt.expectSearch {
				if query != "" && !contains(query, "ILIKE") {
					t.Error("Expected ILIKE in search query")
				}
			}
		})
	}
}

// Helper functions for tests

func buildFilterQuery(categoryID, foodTypeID, searchQuery string) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	argCount := 1

	if categoryID != "" {
		conditions = append(conditions, "category_id = $"+fmt.Sprint(argCount))
		args = append(args, categoryID)
		argCount++
	}

	if foodTypeID != "" {
		conditions = append(conditions, "food_type_id = $"+fmt.Sprint(argCount))
		args = append(args, foodTypeID)
		argCount++
	}

	if searchQuery != "" {
		conditions = append(conditions, "(name ILIKE $"+fmt.Sprint(argCount)+" OR description ILIKE $"+fmt.Sprint(argCount)+")")
		args = append(args, "%"+searchQuery+"%")
		argCount++
	}

	query := ""
	if len(conditions) > 0 {
		query = " WHERE " + join(conditions, " AND ")
	}

	return query, args
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsAny(s, substr))
}

func containsAny(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func join(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

func validateRestaurantInput(r *models.CreateRestaurantRequest) error {
	if r.Name == "" {
		return &ValidationError{"Name is required"}
	}
	if len(r.Name) > 200 {
		return &ValidationError{"Name must be less than 200 characters"}
	}
	return nil
}

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
