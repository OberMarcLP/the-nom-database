package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// All tests in this file exercise validation paths that return before any
// database access, so they run without a database connection.

func TestGetRatings_InvalidRestaurantID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		restaurantID string
	}{
		{"non-numeric id", "abc"},
		{"missing id", ""},
		{"float id", "1.5"},
		{"numeric with suffix", "12abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rr := httptest.NewRecorder()
			r := newRequestWithVars(http.MethodGet, "/api/restaurants/"+tt.restaurantID+"/ratings", nil,
				map[string]string{"restaurantId": tt.restaurantID})

			GetRatings(rr, r)

			if rr.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
			}
			if !strings.Contains(rr.Body.String(), "Invalid restaurant ID") {
				t.Errorf("body = %q, want it to mention the invalid restaurant ID", rr.Body.String())
			}
		})
	}
}

func TestCreateRating_ValidationErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		body     string
		wantMsg  string
		wantCode int
	}{
		{"malformed json", `{"restaurant_id":`, "Invalid request body", http.StatusBadRequest},
		{"empty body", ``, "Invalid request body", http.StatusBadRequest},
		{"wrong field type", `{"restaurant_id":"one"}`, "Invalid request body", http.StatusBadRequest},
		{"missing restaurant id", `{"food_rating":3,"service_rating":3,"ambiance_rating":3}`, "Restaurant ID is required", http.StatusBadRequest},
		{"food rating below range", `{"restaurant_id":1,"food_rating":0,"service_rating":3,"ambiance_rating":3}`, "Ratings must be between 1 and 5", http.StatusBadRequest},
		{"food rating above range", `{"restaurant_id":1,"food_rating":6,"service_rating":3,"ambiance_rating":3}`, "Ratings must be between 1 and 5", http.StatusBadRequest},
		{"service rating below range", `{"restaurant_id":1,"food_rating":3,"service_rating":0,"ambiance_rating":3}`, "Ratings must be between 1 and 5", http.StatusBadRequest},
		{"ambiance rating above range", `{"restaurant_id":1,"food_rating":3,"service_rating":3,"ambiance_rating":6}`, "Ratings must be between 1 and 5", http.StatusBadRequest},
		{"all ratings missing", `{"restaurant_id":1}`, "Ratings must be between 1 and 5", http.StatusBadRequest},
		{"negative ratings", `{"restaurant_id":1,"food_rating":-1,"service_rating":-1,"ambiance_rating":-1}`, "Ratings must be between 1 and 5", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rr := httptest.NewRecorder()
			r := newJSONRequestWithVars(http.MethodPost, "/api/ratings", tt.body, nil)

			CreateRating(rr, r)

			if rr.Code != tt.wantCode {
				t.Errorf("status = %d, want %d", rr.Code, tt.wantCode)
			}
			if !strings.Contains(rr.Body.String(), tt.wantMsg) {
				t.Errorf("body = %q, want it to contain %q", rr.Body.String(), tt.wantMsg)
			}
		})
	}
}

func TestUpdateRating_InvalidInput(t *testing.T) {
	t.Parallel()

	t.Run("non-numeric rating id", func(t *testing.T) {
		t.Parallel()

		rr := httptest.NewRecorder()
		r := newJSONRequestWithVars(http.MethodPut, "/api/ratings/abc", `{}`, map[string]string{"id": "abc"})

		UpdateRating(rr, r)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
		}
		if !strings.Contains(rr.Body.String(), "Invalid rating ID") {
			t.Errorf("body = %q, want it to mention the invalid rating ID", rr.Body.String())
		}
	})

	t.Run("malformed body", func(t *testing.T) {
		t.Parallel()

		rr := httptest.NewRecorder()
		r := newJSONRequestWithVars(http.MethodPut, "/api/ratings/1", `{"food_rating":`, map[string]string{"id": "1"})

		UpdateRating(rr, r)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
		}
		if !strings.Contains(rr.Body.String(), "Invalid request body") {
			t.Errorf("body = %q, want it to mention the invalid body", rr.Body.String())
		}
	})
}

func TestDeleteRating_InvalidInput(t *testing.T) {
	t.Parallel()

	t.Run("non-numeric rating id", func(t *testing.T) {
		t.Parallel()

		rr := httptest.NewRecorder()
		r := newRequestWithVars(http.MethodDelete, "/api/ratings/abc", nil, map[string]string{"id": "abc"})

		DeleteRating(rr, r)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		t.Parallel()

		rr := httptest.NewRecorder()
		r := newRequestWithVars(http.MethodDelete, "/api/ratings/1", nil, map[string]string{"id": "1"})

		DeleteRating(rr, r)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
		}
	})
}

func TestVoteOnReview_InvalidInput(t *testing.T) {
	t.Parallel()

	t.Run("non-numeric rating id", func(t *testing.T) {
		t.Parallel()

		rr := httptest.NewRecorder()
		r := newJSONRequestWithVars(http.MethodPost, "/api/ratings/abc/vote", `{"is_helpful":true}`,
			map[string]string{"id": "abc"})

		VoteOnReview(rr, r)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		t.Parallel()

		rr := httptest.NewRecorder()
		r := newJSONRequestWithVars(http.MethodPost, "/api/ratings/1/vote", `{"is_helpful":true}`,
			map[string]string{"id": "1"})

		VoteOnReview(rr, r)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
		}
	})
}

func TestRemoveVote_InvalidInput(t *testing.T) {
	t.Parallel()

	t.Run("non-numeric rating id", func(t *testing.T) {
		t.Parallel()

		rr := httptest.NewRecorder()
		r := newRequestWithVars(http.MethodDelete, "/api/ratings/abc/vote", nil, map[string]string{"id": "abc"})

		RemoveVote(rr, r)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		t.Parallel()

		rr := httptest.NewRecorder()
		r := newRequestWithVars(http.MethodDelete, "/api/ratings/1/vote", nil, map[string]string{"id": "1"})

		RemoveVote(rr, r)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
		}
	})
}

func TestUploadReviewPhoto_InvalidInput(t *testing.T) {
	t.Parallel()

	t.Run("non-numeric rating id", func(t *testing.T) {
		t.Parallel()

		rr := httptest.NewRecorder()
		r := newRequestWithVars(http.MethodPost, "/api/ratings/abc/photos", nil, map[string]string{"id": "abc"})

		UploadReviewPhoto(rr, r)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		t.Parallel()

		rr := httptest.NewRecorder()
		r := newRequestWithVars(http.MethodPost, "/api/ratings/1/photos", nil, map[string]string{"id": "1"})

		UploadReviewPhoto(rr, r)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
		}
	})
}

func TestDeleteReviewPhoto_InvalidInput(t *testing.T) {
	t.Parallel()

	t.Run("non-numeric photo id", func(t *testing.T) {
		t.Parallel()

		rr := httptest.NewRecorder()
		r := newRequestWithVars(http.MethodDelete, "/api/review-photos/abc", nil, map[string]string{"id": "abc"})

		DeleteReviewPhoto(rr, r)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
		}
		if !strings.Contains(rr.Body.String(), "Invalid photo ID") {
			t.Errorf("body = %q, want it to mention the invalid photo ID", rr.Body.String())
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		t.Parallel()

		rr := httptest.NewRecorder()
		r := newRequestWithVars(http.MethodDelete, "/api/review-photos/1", nil, map[string]string{"id": "1"})

		DeleteReviewPhoto(rr, r)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
		}
	})
}
