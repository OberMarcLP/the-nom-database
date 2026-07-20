package handlers

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParsePaginationParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		query      string
		wantLimit  int
		wantCursor string
	}{
		{"defaults", "", DefaultPageLimit, ""},
		{"custom limit", "limit=50", 50, ""},
		{"limit at maximum", "limit=100", MaxPageLimit, ""},
		{"limit above maximum is capped", "limit=1000", MaxPageLimit, ""},
		{"zero limit falls back to default", "limit=0", DefaultPageLimit, ""},
		{"negative limit falls back to default", "limit=-5", DefaultPageLimit, ""},
		{"non-numeric limit falls back to default", "limit=abc", DefaultPageLimit, ""},
		{"cursor is passed through", "cursor=Nw", DefaultPageLimit, "Nw"},
		{"limit and cursor combined", "limit=7&cursor=abc", 7, "abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := httptest.NewRequest(http.MethodGet, "/api/restaurants?"+tt.query, nil)
			params := ParsePaginationParams(r)

			if params.Limit != tt.wantLimit {
				t.Errorf("Limit = %d, want %d", params.Limit, tt.wantLimit)
			}
			if params.Cursor != tt.wantCursor {
				t.Errorf("Cursor = %q, want %q", params.Cursor, tt.wantCursor)
			}
		})
	}
}

func TestCursorEncodeDecodeRoundtrip(t *testing.T) {
	t.Parallel()

	for _, id := range []int{0, 1, 42, 999999} {
		encoded := EncodeCursor(id)
		decoded, err := DecodeCursor(encoded)
		if err != nil {
			t.Errorf("DecodeCursor(EncodeCursor(%d)) error = %v", id, err)
			continue
		}
		if decoded != id {
			t.Errorf("DecodeCursor(EncodeCursor(%d)) = %d, want %d", id, decoded, id)
		}
	}
}

func TestDecodeCursor_EdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("empty cursor decodes to zero without error", func(t *testing.T) {
		t.Parallel()

		id, err := DecodeCursor("")
		if err != nil {
			t.Fatalf("error = %v, want nil", err)
		}
		if id != 0 {
			t.Errorf("id = %d, want 0", id)
		}
	})

	t.Run("invalid base64 is rejected", func(t *testing.T) {
		t.Parallel()

		if _, err := DecodeCursor("!!!not-base64!!!"); err == nil {
			t.Error("error = nil, want an error for invalid base64")
		}
	})

	t.Run("base64 of a non-number is rejected", func(t *testing.T) {
		t.Parallel()

		cursor := base64.StdEncoding.EncodeToString([]byte("abc"))
		if _, err := DecodeCursor(cursor); err == nil {
			t.Error("error = nil, want an error for a non-numeric cursor payload")
		}
	})
}

func TestBuildPaginatedResponse(t *testing.T) {
	t.Parallel()

	data := []string{"one", "two"}

	t.Run("with next page", func(t *testing.T) {
		t.Parallel()

		resp := BuildPaginatedResponse(data, true, 5)
		if !resp.HasMore {
			t.Error("HasMore = false, want true")
		}
		if resp.NextCursor == nil {
			t.Fatal("NextCursor = nil, want a cursor")
		}
		if want := EncodeCursor(5); *resp.NextCursor != want {
			t.Errorf("NextCursor = %q, want %q", *resp.NextCursor, want)
		}
	})

	t.Run("last page has no cursor", func(t *testing.T) {
		t.Parallel()

		resp := BuildPaginatedResponse(data, false, 0)
		if resp.HasMore {
			t.Error("HasMore = true, want false")
		}
		if resp.NextCursor != nil {
			t.Errorf("NextCursor = %q, want nil", *resp.NextCursor)
		}
	})

	t.Run("has more but invalid next id yields no cursor", func(t *testing.T) {
		t.Parallel()

		resp := BuildPaginatedResponse(data, true, 0)
		if resp.NextCursor != nil {
			t.Errorf("NextCursor = %q, want nil for nextID 0", *resp.NextCursor)
		}
	})
}
