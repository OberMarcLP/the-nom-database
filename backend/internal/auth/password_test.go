package auth

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"golang.org/x/crypto/argon2"
)

// testArgon2Params returns deliberately cheap parameters so tests stay fast.
// Security-grade parameters are covered by TestHashPassword_NilParamsUsesDefaults.
func testArgon2Params() *Argon2Params {
	return &Argon2Params{
		Memory:      1024, // 1 MiB
		Iterations:  1,
		Parallelism: 1,
		SaltLength:  16,
		KeyLength:   32,
	}
}

func TestHashPassword_VerifyRoundtrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		password string
	}{
		{"simple ascii password", "s3cret-Passw0rd!"},
		{"empty password", ""},
		{"unicode password", "pässwörtli-🍜-細麺"},
		{"whitespace password", "  spaces and\ttabs  "},
		{"long password", strings.Repeat("a", 512)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			hash, err := HashPassword(tt.password, testArgon2Params())
			if err != nil {
				t.Fatalf("HashPassword() error = %v", err)
			}
			if hash == "" {
				t.Fatal("HashPassword() returned an empty hash")
			}

			ok, err := VerifyPassword(tt.password, hash)
			if err != nil {
				t.Fatalf("VerifyPassword(correct password) error = %v", err)
			}
			if !ok {
				t.Error("VerifyPassword(correct password) = false, want true")
			}

			ok, err = VerifyPassword(tt.password+"x", hash)
			if err != nil {
				t.Fatalf("VerifyPassword(wrong password) error = %v", err)
			}
			if ok {
				t.Error("VerifyPassword(wrong password) = true, want false")
			}
		})
	}
}

func TestHashPassword_SaltMakesHashesUnique(t *testing.T) {
	t.Parallel()

	const password = "same-password-every-time"

	first, err := HashPassword(password, testArgon2Params())
	if err != nil {
		t.Fatalf("HashPassword() first call error = %v", err)
	}
	second, err := HashPassword(password, testArgon2Params())
	if err != nil {
		t.Fatalf("HashPassword() second call error = %v", err)
	}

	if first == second {
		t.Error("two hashes of the same password are identical, want different salts to produce different hashes")
	}

	for i, hash := range []string{first, second} {
		ok, err := VerifyPassword(password, hash)
		if err != nil {
			t.Fatalf("VerifyPassword() hash #%d error = %v", i+1, err)
		}
		if !ok {
			t.Errorf("VerifyPassword() hash #%d = false, want true", i+1)
		}
	}
}

func TestHashPassword_EncodedFormat(t *testing.T) {
	t.Parallel()

	hash, err := HashPassword("format-check", testArgon2Params())
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	wantPrefix := fmt.Sprintf("$argon2id$v=%d$m=1024,t=1,p=1$", argon2.Version)
	if !strings.HasPrefix(hash, wantPrefix) {
		t.Errorf("hash prefix = %q, want prefix %q", hash, wantPrefix)
	}
	if parts := strings.Split(hash, "$"); len(parts) != 6 {
		t.Errorf("hash has %d $-separated parts, want 6: %q", len(parts), hash)
	}
}

func TestHashPassword_NilParamsUsesDefaults(t *testing.T) {
	t.Parallel()

	hash, err := HashPassword("uses-default-params", nil)
	if err != nil {
		t.Fatalf("HashPassword(nil params) error = %v", err)
	}

	wantPrefix := fmt.Sprintf("$argon2id$v=%d$m=65536,t=3,p=2$", argon2.Version)
	if !strings.HasPrefix(hash, wantPrefix) {
		t.Errorf("hash prefix = %q, want default-parameter prefix %q", hash, wantPrefix)
	}

	ok, err := VerifyPassword("uses-default-params", hash)
	if err != nil {
		t.Fatalf("VerifyPassword() error = %v", err)
	}
	if !ok {
		t.Error("VerifyPassword() = false, want true for default-parameter hash")
	}
}

func TestDefaultArgon2Params(t *testing.T) {
	t.Parallel()

	p := DefaultArgon2Params()
	if p.Memory != 64*1024 {
		t.Errorf("Memory = %d, want %d", p.Memory, 64*1024)
	}
	if p.Iterations != 3 {
		t.Errorf("Iterations = %d, want 3", p.Iterations)
	}
	if p.Parallelism != 2 {
		t.Errorf("Parallelism = %d, want 2", p.Parallelism)
	}
	if p.SaltLength != 16 {
		t.Errorf("SaltLength = %d, want 16", p.SaltLength)
	}
	if p.KeyLength != 32 {
		t.Errorf("KeyLength = %d, want 32", p.KeyLength)
	}
}

func TestVerifyPassword_InvalidHashRejected(t *testing.T) {
	t.Parallel()

	// "c2FsdA" / "aGFzaA" are valid raw-base64 segments ("salt" / "hash").
	tests := []struct {
		name    string
		hash    string
		wantErr error
	}{
		{"empty hash", "", ErrInvalidHash},
		{"plain text", "not-a-hash-at-all", ErrInvalidHash},
		{"too few segments", "$argon2id$v=19$m=1024,t=1,p=1$c2FsdA", ErrInvalidHash},
		{"too many segments", "$argon2id$v=19$m=1024,t=1,p=1$c2FsdA$aGFzaA$extra", ErrInvalidHash},
		{"wrong algorithm", "$argon2i$v=19$m=1024,t=1,p=1$c2FsdA$aGFzaA", ErrInvalidHash},
		{"bcrypt style hash", "$2a$10$abcdefghijklmnopqrstuv", ErrInvalidHash},
		{"incompatible version", "$argon2id$v=18$m=1024,t=1,p=1$c2FsdA$aGFzaA", ErrIncompatibleVersion},
		{"malformed version segment", "$argon2id$version$m=1024,t=1,p=1$c2FsdA$aGFzaA", ErrInvalidHash},
		{"malformed params segment", "$argon2id$v=19$garbage$c2FsdA$aGFzaA", ErrInvalidHash},
		{"invalid base64 salt", "$argon2id$v=19$m=1024,t=1,p=1$!!!!$aGFzaA", ErrInvalidHash},
		{"invalid base64 hash", "$argon2id$v=19$m=1024,t=1,p=1$c2FsdA$!!!!", ErrInvalidHash},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ok, err := VerifyPassword("any-password", tt.hash)
			if ok {
				t.Errorf("VerifyPassword(%q) = true, want false", tt.hash)
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("VerifyPassword(%q) error = %v, want %v", tt.hash, err, tt.wantErr)
			}
		})
	}
}

func TestVerifyPassword_TamperedHashRejected(t *testing.T) {
	t.Parallel()

	hash, err := HashPassword("original-password", testArgon2Params())
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	parts := strings.Split(hash, "$")
	if len(parts) != 6 {
		t.Fatalf("unexpected hash layout: %q", hash)
	}

	// Flip the first character of the digest segment to a different, still
	// valid base64 character so the string decodes but no longer matches.
	digest := parts[5]
	replacement := byte('A')
	if digest[0] == 'A' {
		replacement = 'B'
	}
	parts[5] = string(replacement) + digest[1:]
	tampered := strings.Join(parts, "$")

	ok, err := VerifyPassword("original-password", tampered)
	if err != nil {
		t.Fatalf("VerifyPassword(tampered) unexpected error = %v", err)
	}
	if ok {
		t.Error("VerifyPassword(tampered digest) = true, want false")
	}
}

func TestVerifyPassword_TamperedParamsRejected(t *testing.T) {
	t.Parallel()

	hash, err := HashPassword("original-password", testArgon2Params())
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	// Doubling the memory parameter must change the derived key.
	tampered := strings.Replace(hash, "$m=1024,", "$m=2048,", 1)
	if tampered == hash {
		t.Fatalf("failed to tamper with params of %q", hash)
	}

	ok, err := VerifyPassword("original-password", tampered)
	if err != nil {
		t.Fatalf("VerifyPassword(tampered params) unexpected error = %v", err)
	}
	if ok {
		t.Error("VerifyPassword(tampered params) = true, want false")
	}
}

func TestVerifyPassword_TruncatedHashRejected(t *testing.T) {
	t.Parallel()

	hash, err := HashPassword("original-password", testArgon2Params())
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	truncated := hash[:len(hash)/2]
	ok, err := VerifyPassword("original-password", truncated)
	if ok {
		t.Errorf("VerifyPassword(truncated) = true, want false")
	}
	if !errors.Is(err, ErrInvalidHash) {
		t.Errorf("VerifyPassword(truncated) error = %v, want ErrInvalidHash", err)
	}
}
