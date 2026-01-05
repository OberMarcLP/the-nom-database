package services

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"
)

// GravatarService provides methods for generating Gravatar URLs
type GravatarService struct{}

// NewGravatarService creates a new Gravatar service instance
func NewGravatarService() *GravatarService {
	return &GravatarService{}
}

// GetAvatarURL generates a Gravatar URL from an email address
// Returns a Gravatar URL with default fallback to identicon
//
// Parameters:
//   - email: The user's email address
//   - size: Avatar size in pixels (default: 200, max: 2048)
//
// Returns a URL string pointing to the Gravatar image
func (s *GravatarService) GetAvatarURL(email string, size int) string {
	if size == 0 {
		size = 200
	}
	if size > 2048 {
		size = 2048
	}

	// Normalize email: trim whitespace and lowercase
	email = strings.TrimSpace(strings.ToLower(email))

	// Generate MD5 hash of email
	hash := md5.Sum([]byte(email))
	emailHash := hex.EncodeToString(hash[:])

	// Generate Gravatar URL with identicon as default fallback
	// d=identicon: generates a unique geometric pattern based on email hash
	return fmt.Sprintf("https://www.gravatar.com/avatar/%s?s=%d&d=identicon", emailHash, size)
}

// GetAvatarURLWithOptions generates a Gravatar URL with custom options
//
// Parameters:
//   - email: The user's email address
//   - size: Avatar size in pixels (default: 200, max: 2048)
//   - defaultImage: Fallback image type (identicon, monsterid, wavatar, retro, robohash, blank)
//
// Returns a URL string pointing to the Gravatar image
func (s *GravatarService) GetAvatarURLWithOptions(email string, size int, defaultImage string) string {
	if size == 0 {
		size = 200
	}
	if size > 2048 {
		size = 2048
	}

	// Validate default image option
	validDefaults := map[string]bool{
		"identicon": true,
		"monsterid": true,
		"wavatar":   true,
		"retro":     true,
		"robohash":  true,
		"blank":     true,
		"404":       true, // Returns 404 if no Gravatar exists
	}

	if !validDefaults[defaultImage] {
		defaultImage = "identicon"
	}

	// Normalize email
	email = strings.TrimSpace(strings.ToLower(email))

	// Generate MD5 hash
	hash := md5.Sum([]byte(email))
	emailHash := hex.EncodeToString(hash[:])

	return fmt.Sprintf("https://www.gravatar.com/avatar/%s?s=%d&d=%s", emailHash, size, defaultImage)
}
