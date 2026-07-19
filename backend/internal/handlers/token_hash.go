package handlers

import (
	"crypto/sha256"
	"encoding/hex"
)

// hashRefreshToken returns the SHA-256 hex digest of a refresh token.
// Only the digest is stored in the sessions table so a database leak
// does not expose usable refresh tokens.
func hashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
