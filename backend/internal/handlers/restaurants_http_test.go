package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// These tests cover the validation and authentication guard paths of the
// restaurant handlers, all of which return before any database access.

func TestGetRestaurant_InvalidID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		id   string
	}{
		{"non-numeric id", "abc"},
		{"missing id", ""},
		{"numeric with suffix", "12abc"},
		{"float id", "3.14"},
		{"zero id", "0"},
		{"negative id", "-5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rr := httptest.NewRecorder()
			r := newRequestWithVars(http.MethodGet, "/api/restaurants/"+tt.id, nil, map[string]string{"id": tt.id})

			GetRestaurant(rr, r)

			if rr.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
			}
			if !strings.Contains(rr.Body.String(), "Invalid restaurant ID") {
				t.Errorf("body = %q, want it to mention the invalid restaurant ID", rr.Body.String())
			}
		})
	}
}

func TestCreateRestaurant_ValidationErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		body    string
		wantMsg string
	}{
		{"malformed json", `{"name":`, "Invalid request body"},
		{"empty body", ``, "Invalid request body"},
		{"wrong field type", `{"name":123}`, "Invalid request body"},
		{"missing name", `{}`, "Name is required"},
		{"empty name", `{"name":""}`, "Name is required"},
		{"price range zero", `{"name":"Nom","price_range":0}`, "Price range must be between 1 and 4"},
		{"price range above maximum", `{"name":"Nom","price_range":5}`, "Price range must be between 1 and 4"},
		{"price range negative", `{"name":"Nom","price_range":-1}`, "Price range must be between 1 and 4"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rr := httptest.NewRecorder()
			r := newJSONRequestWithVars(http.MethodPost, "/api/restaurants", tt.body, nil)

			CreateRestaurant(rr, r)

			if rr.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
			}
			if !strings.Contains(rr.Body.String(), tt.wantMsg) {
				t.Errorf("body = %q, want it to contain %q", rr.Body.String(), tt.wantMsg)
			}
		})
	}
}

func TestUpdateRestaurant_GuardPaths(t *testing.T) {
	t.Parallel()

	t.Run("invalid restaurant id", func(t *testing.T) {
		t.Parallel()

		for _, id := range []string{"abc", "0", "-5"} {
			rr := httptest.NewRecorder()
			r := newJSONRequestWithVars(http.MethodPut, "/api/restaurants/"+id, `{}`, map[string]string{"id": id})

			UpdateRestaurant(rr, r)

			if rr.Code != http.StatusBadRequest {
				t.Errorf("id %q: status = %d, want %d", id, rr.Code, http.StatusBadRequest)
			}
			if !strings.Contains(rr.Body.String(), "Invalid restaurant ID") {
				t.Errorf("id %q: body = %q, want it to mention the invalid restaurant ID", id, rr.Body.String())
			}
		}
	})

	t.Run("malformed body", func(t *testing.T) {
		t.Parallel()

		rr := httptest.NewRecorder()
		r := newJSONRequestWithVars(http.MethodPut, "/api/restaurants/1", `{"name":`, map[string]string{"id": "1"})

		UpdateRestaurant(rr, r)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
		}
		if !strings.Contains(rr.Body.String(), "Invalid request body") {
			t.Errorf("body = %q, want it to mention the invalid body", rr.Body.String())
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		t.Parallel()

		rr := httptest.NewRecorder()
		r := newJSONRequestWithVars(http.MethodPut, "/api/restaurants/1", `{"name":"New Name"}`, map[string]string{"id": "1"})

		UpdateRestaurant(rr, r)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
		}
		if !strings.Contains(rr.Body.String(), "Authentication required") {
			t.Errorf("body = %q, want it to mention required authentication", rr.Body.String())
		}
	})

	t.Run("invalid price range rejected before auth check", func(t *testing.T) {
		t.Parallel()

		for _, body := range []string{`{"price_range":0}`, `{"price_range":5}`, `{"price_range":-1}`} {
			rr := httptest.NewRecorder()
			r := newJSONRequestWithVars(http.MethodPut, "/api/restaurants/1", body, map[string]string{"id": "1"})

			UpdateRestaurant(rr, r)

			if rr.Code != http.StatusBadRequest {
				t.Errorf("body %s: status = %d, want %d", body, rr.Code, http.StatusBadRequest)
			}
			if !strings.Contains(rr.Body.String(), "Price range must be between 1 and 4") {
				t.Errorf("body %s: response = %q, want the price range message", body, rr.Body.String())
			}
		}
	})

	t.Run("valid or absent price range passes validation to auth guard", func(t *testing.T) {
		t.Parallel()

		// 401 (not 400) proves these bodies clear the price_range validation
		// and reach the next guard; DB access is never hit without a user.
		for _, body := range []string{`{"price_range":1}`, `{"price_range":4}`, `{}`} {
			rr := httptest.NewRecorder()
			r := newJSONRequestWithVars(http.MethodPut, "/api/restaurants/1", body, map[string]string{"id": "1"})

			UpdateRestaurant(rr, r)

			if rr.Code != http.StatusUnauthorized {
				t.Errorf("body %s: status = %d, want %d", body, rr.Code, http.StatusUnauthorized)
			}
		}
	})
}

func TestValidatePriceRange(t *testing.T) {
	t.Parallel()

	intPtr := func(v int) *int { return &v }

	t.Run("valid values", func(t *testing.T) {
		t.Parallel()

		for _, pr := range []*int{nil, intPtr(1), intPtr(2), intPtr(3), intPtr(4)} {
			if err := validatePriceRange(pr); err != nil {
				t.Errorf("validatePriceRange(%v) = %v, want nil", pr, err)
			}
		}
	})

	t.Run("out of range values", func(t *testing.T) {
		t.Parallel()

		for _, v := range []int{0, 5, -1, 100} {
			err := validatePriceRange(intPtr(v))
			if err == nil {
				t.Errorf("validatePriceRange(%d) = nil, want an error", v)
				continue
			}
			if err.Error() != "Price range must be between 1 and 4" {
				t.Errorf("validatePriceRange(%d) message = %q, want the contract message", v, err.Error())
			}
		}
	})
}

func TestDeleteRestaurant_GuardPaths(t *testing.T) {
	t.Parallel()

	t.Run("invalid restaurant id", func(t *testing.T) {
		t.Parallel()

		for _, id := range []string{"xyz", "0", "-5"} {
			rr := httptest.NewRecorder()
			r := newRequestWithVars(http.MethodDelete, "/api/restaurants/"+id, nil, map[string]string{"id": id})

			DeleteRestaurant(rr, r)

			if rr.Code != http.StatusBadRequest {
				t.Errorf("id %q: status = %d, want %d", id, rr.Code, http.StatusBadRequest)
			}
			if !strings.Contains(rr.Body.String(), "Invalid restaurant ID") {
				t.Errorf("id %q: body = %q, want it to mention the invalid restaurant ID", id, rr.Body.String())
			}
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		t.Parallel()

		rr := httptest.NewRecorder()
		r := newRequestWithVars(http.MethodDelete, "/api/restaurants/1", nil, map[string]string{"id": "1"})

		DeleteRestaurant(rr, r)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
		}
		if !strings.Contains(rr.Body.String(), "Authentication required") {
			t.Errorf("body = %q, want it to mention required authentication", rr.Body.String())
		}
	})
}
