package services

import (
	"strings"
	"testing"
)

func TestGravatarService_GetAvatarURL(t *testing.T) {
	service := NewGravatarService()

	tests := []struct {
		name          string
		email         string
		size          int
		expectedHash  string
		expectedSize  int
		expectedInURL string
	}{
		{
			name:          "Standard email with default size",
			email:         "user@example.com",
			size:          0,
			expectedHash:  "b58996c504c5638798eb6b511e6f49af", // MD5 of user@example.com
			expectedSize:  200,
			expectedInURL: "d=identicon",
		},
		{
			name:          "Email with uppercase should be normalized",
			email:         "User@Example.COM",
			size:          100,
			expectedHash:  "b58996c504c5638798eb6b511e6f49af", // Same as lowercase
			expectedSize:  100,
			expectedInURL: "s=100",
		},
		{
			name:          "Email with whitespace should be trimmed",
			email:         "  user@example.com  ",
			size:          256,
			expectedHash:  "b58996c504c5638798eb6b511e6f49af",
			expectedSize:  256,
			expectedInURL: "s=256",
		},
		{
			name:          "Size over 2048 should be capped",
			email:         "test@test.com",
			size:          3000,
			expectedSize:  2048,
			expectedInURL: "s=2048",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := service.GetAvatarURL(tt.email, tt.size)

			// Check if URL contains expected hash
			if !strings.Contains(url, tt.expectedHash) {
				t.Errorf("Expected URL to contain hash %s, got %s", tt.expectedHash, url)
			}

			// Check if URL contains expected substring
			if !strings.Contains(url, tt.expectedInURL) {
				t.Errorf("Expected URL to contain %s, got %s", tt.expectedInURL, url)
			}

			// Verify URL format
			if !strings.HasPrefix(url, "https://www.gravatar.com/avatar/") {
				t.Errorf("URL should start with gravatar.com/avatar/, got %s", url)
			}
		})
	}
}

func TestGravatarService_GetAvatarURLWithOptions(t *testing.T) {
	service := NewGravatarService()

	tests := []struct {
		name         string
		email        string
		size         int
		defaultImage string
		expectedInURL string
	}{
		{
			name:         "Retro style default image",
			email:        "user@example.com",
			size:         200,
			defaultImage: "retro",
			expectedInURL: "d=retro",
		},
		{
			name:         "Robohash default image",
			email:        "user@example.com",
			size:         200,
			defaultImage: "robohash",
			expectedInURL: "d=robohash",
		},
		{
			name:         "Invalid default falls back to identicon",
			email:        "user@example.com",
			size:         200,
			defaultImage: "invalid",
			expectedInURL: "d=identicon",
		},
		{
			name:         "404 option for checking if Gravatar exists",
			email:        "user@example.com",
			size:         200,
			defaultImage: "404",
			expectedInURL: "d=404",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := service.GetAvatarURLWithOptions(tt.email, tt.size, tt.defaultImage)

			if !strings.Contains(url, tt.expectedInURL) {
				t.Errorf("Expected URL to contain %s, got %s", tt.expectedInURL, url)
			}
		})
	}
}

func TestGravatarService_ConsistentHashing(t *testing.T) {
	service := NewGravatarService()

	// Same email should always produce same hash
	email := "test@example.com"
	url1 := service.GetAvatarURL(email, 200)
	url2 := service.GetAvatarURL(email, 200)

	if url1 != url2 {
		t.Errorf("Same email should produce same URL. Got %s and %s", url1, url2)
	}

	// Different emails should produce different hashes
	url3 := service.GetAvatarURL("different@example.com", 200)
	if url1 == url3 {
		t.Errorf("Different emails should produce different URLs")
	}
}
