package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/nomdb/backend/internal/logger"
	"github.com/nomdb/backend/internal/models"
)

// oidcPendingLogin is a login response parked under a one-time code so the
// tokens never travel through a URL (history, proxy logs, referrers).
type oidcPendingLogin struct {
	response  models.LoginResponse
	expiresAt time.Time
}

var (
	oidcCodeStore = make(map[string]oidcPendingLogin) // In production, use Redis
	oidcCodeMu    sync.Mutex
)

// issueOIDCLoginCode stores the login response under a random one-time code
// and returns the frontend redirect URL carrying only that code.
func issueOIDCLoginCode(response *models.LoginResponse) (string, error) {
	code, err := generateOIDCState()
	if err != nil {
		return "", err
	}

	now := time.Now()
	oidcCodeMu.Lock()
	// Opportunistic cleanup of expired entries
	for k, v := range oidcCodeStore {
		if now.After(v.expiresAt) {
			delete(oidcCodeStore, k)
		}
	}
	oidcCodeStore[code] = oidcPendingLogin{response: *response, expiresAt: now.Add(60 * time.Second)}
	oidcCodeMu.Unlock()

	frontendURL := strings.TrimRight(os.Getenv("FRONTEND_URL"), "/")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}
	return fmt.Sprintf("%s/auth/callback?code=%s", frontendURL, code), nil
}

// ExchangeOIDCCode godoc
// @Summary Exchange OIDC login code
// @Description Exchange the one-time code from the OIDC callback redirect for tokens
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} models.LoginResponse
// @Failure 400 {string} string "Invalid request"
// @Failure 401 {string} string "Invalid or expired code"
// @Router /auth/oidc/exchange [post]
func ExchangeOIDCCode(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Code == "" {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	oidcCodeMu.Lock()
	pending, exists := oidcCodeStore[req.Code]
	delete(oidcCodeStore, req.Code) // strictly one-time use
	oidcCodeMu.Unlock()

	if !exists || time.Now().After(pending.expiresAt) {
		http.Error(w, "Invalid or expired code", http.StatusUnauthorized)
		return
	}

	// Refresh token travels only via httpOnly cookie, never in the body
	if svc := getJWTService(); svc != nil {
		setRefreshCookie(w, r, pending.response.RefreshToken, svc.GetRefreshTokenDuration())
	}
	pending.response.RefreshToken = ""

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(pending.response); err != nil {
		logger.Error("Failed to encode response: %v", err)
	}
}
