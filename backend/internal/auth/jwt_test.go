package auth

import (
	"encoding/base64"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/nomdb/backend/internal/models"
)

const testJWTSecret = "unit-test-secret-key"

func newTestJWTService(secret string, accessTokenDuration time.Duration) *JWTService {
	return NewJWTService(secret, accessTokenDuration, 7*24*time.Hour)
}

func testUser() *models.User {
	return &models.User{
		ID:       42,
		Email:    "nom@example.com",
		Username: "nomuser",
		IsAdmin:  true,
	}
}

func TestJWTService_GenerateAndValidateAccessToken(t *testing.T) {
	t.Parallel()

	svc := newTestJWTService(testJWTSecret, 15*time.Minute)
	user := testUser()

	token, err := svc.GenerateAccessToken(user)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}
	if parts := strings.Split(token, "."); len(parts) != 3 {
		t.Fatalf("token has %d segments, want 3: %q", len(parts), token)
	}

	claims, err := svc.ValidateAccessToken(token)
	if err != nil {
		t.Fatalf("ValidateAccessToken() error = %v", err)
	}

	if claims.UserID != user.ID {
		t.Errorf("UserID = %d, want %d", claims.UserID, user.ID)
	}
	if claims.Email != user.Email {
		t.Errorf("Email = %q, want %q", claims.Email, user.Email)
	}
	if claims.Username != user.Username {
		t.Errorf("Username = %q, want %q", claims.Username, user.Username)
	}
	if claims.IsAdmin != user.IsAdmin {
		t.Errorf("IsAdmin = %v, want %v", claims.IsAdmin, user.IsAdmin)
	}
	if claims.Issuer != "nomdb" {
		t.Errorf("Issuer = %q, want %q", claims.Issuer, "nomdb")
	}
	if claims.Subject != "42" {
		t.Errorf("Subject = %q, want %q", claims.Subject, "42")
	}
	if claims.ExpiresAt == nil {
		t.Fatal("ExpiresAt is nil, want it set")
	}
	if remaining := time.Until(claims.ExpiresAt.Time); remaining <= 0 || remaining > 15*time.Minute {
		t.Errorf("token expires in %v, want within (0, 15m]", remaining)
	}
}

func TestJWTService_ExpiredTokenRejected(t *testing.T) {
	t.Parallel()

	// A negative duration produces a token that expired in the past.
	svc := newTestJWTService(testJWTSecret, -1*time.Minute)

	token, err := svc.GenerateAccessToken(testUser())
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	claims, err := svc.ValidateAccessToken(token)
	if claims != nil {
		t.Error("ValidateAccessToken(expired) returned claims, want nil")
	}
	if !errors.Is(err, ErrExpiredToken) {
		t.Errorf("ValidateAccessToken(expired) error = %v, want ErrExpiredToken", err)
	}
}

func TestJWTService_WrongSecretRejected(t *testing.T) {
	t.Parallel()

	issuer := newTestJWTService("secret-a", 15*time.Minute)
	verifier := newTestJWTService("secret-b", 15*time.Minute)

	token, err := issuer.GenerateAccessToken(testUser())
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	claims, err := verifier.ValidateAccessToken(token)
	if claims != nil {
		t.Error("ValidateAccessToken(wrong secret) returned claims, want nil")
	}
	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("ValidateAccessToken(wrong secret) error = %v, want ErrInvalidToken", err)
	}
}

func TestJWTService_TamperedSignatureRejected(t *testing.T) {
	t.Parallel()

	svc := newTestJWTService(testJWTSecret, 15*time.Minute)
	token, err := svc.GenerateAccessToken(testUser())
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	parts := strings.Split(token, ".")
	parts[2] = base64.RawURLEncoding.EncodeToString(make([]byte, 32)) // all-zero signature
	tampered := strings.Join(parts, ".")

	claims, err := svc.ValidateAccessToken(tampered)
	if claims != nil {
		t.Error("ValidateAccessToken(tampered signature) returned claims, want nil")
	}
	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("ValidateAccessToken(tampered signature) error = %v, want ErrInvalidToken", err)
	}
}

func TestJWTService_TamperedPayloadRejected(t *testing.T) {
	t.Parallel()

	svc := newTestJWTService(testJWTSecret, 15*time.Minute)

	tokenA, err := svc.GenerateAccessToken(&models.User{ID: 1, Email: "a@example.com", Username: "a"})
	if err != nil {
		t.Fatalf("GenerateAccessToken(user A) error = %v", err)
	}
	tokenB, err := svc.GenerateAccessToken(&models.User{ID: 2, Email: "b@example.com", Username: "b", IsAdmin: true})
	if err != nil {
		t.Fatalf("GenerateAccessToken(user B) error = %v", err)
	}

	partsA := strings.Split(tokenA, ".")
	partsB := strings.Split(tokenB, ".")

	// Payload of user B combined with signature of user A must not validate.
	forged := partsA[0] + "." + partsB[1] + "." + partsA[2]

	claims, err := svc.ValidateAccessToken(forged)
	if claims != nil {
		t.Error("ValidateAccessToken(forged payload) returned claims, want nil")
	}
	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("ValidateAccessToken(forged payload) error = %v, want ErrInvalidToken", err)
	}
}

func TestJWTService_GarbageTokensRejected(t *testing.T) {
	t.Parallel()

	svc := newTestJWTService(testJWTSecret, 15*time.Minute)

	tests := []struct {
		name  string
		token string
	}{
		{"empty string", ""},
		{"three garbage segments", "a.b.c"},
		{"not base64", "!!!.@@@.###"},
		{"single segment", "justonesegment"},
		{"only dots", ".."},
		{"unicode garbage", "🍜.🍣.🍰"},
		{"truncated real-looking token", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			claims, err := svc.ValidateAccessToken(tt.token)
			if claims != nil {
				t.Errorf("ValidateAccessToken(%q) returned claims, want nil", tt.token)
			}
			if !errors.Is(err, ErrInvalidToken) {
				t.Errorf("ValidateAccessToken(%q) error = %v, want ErrInvalidToken", tt.token, err)
			}
		})
	}
}

func TestJWTService_AlgNoneTokenRejected(t *testing.T) {
	t.Parallel()

	svc := newTestJWTService(testJWTSecret, 15*time.Minute)

	claims := &Claims{
		UserID:   42,
		Email:    "nom@example.com",
		Username: "nomuser",
		IsAdmin:  true,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "nomdb",
		},
	}

	unsigned := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	token, err := unsigned.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("failed to build alg=none token: %v", err)
	}

	got, err := svc.ValidateAccessToken(token)
	if got != nil {
		t.Error("ValidateAccessToken(alg=none) returned claims, want nil")
	}
	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("ValidateAccessToken(alg=none) error = %v, want ErrInvalidToken", err)
	}
}

func TestJWTService_NotYetValidTokenRejected(t *testing.T) {
	t.Parallel()

	svc := newTestJWTService(testJWTSecret, 15*time.Minute)

	// Correctly signed with the same secret, but nbf lies in the future.
	claims := &Claims{
		UserID: 42,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			NotBefore: jwt.NewNumericDate(time.Now().Add(30 * time.Minute)),
			Issuer:    "nomdb",
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tok.SignedString([]byte(testJWTSecret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	got, err := svc.ValidateAccessToken(token)
	if got != nil {
		t.Error("ValidateAccessToken(nbf in future) returned claims, want nil")
	}
	if err == nil {
		t.Fatal("ValidateAccessToken(nbf in future) error = nil, want an error")
	}
	if errors.Is(err, ErrExpiredToken) {
		t.Errorf("ValidateAccessToken(nbf in future) error = %v, want a non-expiry rejection", err)
	}
}

func TestJWTService_GenerateRefreshToken(t *testing.T) {
	t.Parallel()

	svc := newTestJWTService(testJWTSecret, 15*time.Minute)

	first, err := svc.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken() error = %v", err)
	}
	if first == "" {
		t.Fatal("GenerateRefreshToken() returned an empty token")
	}

	raw, err := base64.URLEncoding.DecodeString(first)
	if err != nil {
		t.Fatalf("refresh token is not URL-safe base64: %v", err)
	}
	if len(raw) != 32 {
		t.Errorf("refresh token decodes to %d bytes, want 32", len(raw))
	}

	second, err := svc.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken() second call error = %v", err)
	}
	if first == second {
		t.Error("two refresh tokens are identical, want random tokens")
	}
}

func TestJWTService_Durations(t *testing.T) {
	t.Parallel()

	access := 15 * time.Minute
	refresh := 7 * 24 * time.Hour
	svc := NewJWTService(testJWTSecret, access, refresh)

	if got := svc.GetAccessTokenDuration(); got != access {
		t.Errorf("GetAccessTokenDuration() = %v, want %v", got, access)
	}
	if got := svc.GetRefreshTokenDuration(); got != refresh {
		t.Errorf("GetRefreshTokenDuration() = %v, want %v", got, refresh)
	}
}
