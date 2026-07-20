package auth

import (
	"testing"
	"time"
)

// TestGlobalJWTService_SetAndGet exercises the package-level service registry.
// It mutates global state, so it deliberately does not run in parallel and
// restores the previous value on cleanup.
func TestGlobalJWTService_SetAndGet(t *testing.T) {
	original := GetGlobalJWTService()
	t.Cleanup(func() { SetGlobalJWTService(original) })

	svc := NewJWTService("global-service-secret", time.Minute, time.Hour)

	SetGlobalJWTService(svc)
	if got := GetGlobalJWTService(); got != svc {
		t.Errorf("GetGlobalJWTService() = %p, want the instance passed to SetGlobalJWTService (%p)", got, svc)
	}

	SetGlobalJWTService(nil)
	if got := GetGlobalJWTService(); got != nil {
		t.Errorf("GetGlobalJWTService() after SetGlobalJWTService(nil) = %p, want nil", got)
	}
}
