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
