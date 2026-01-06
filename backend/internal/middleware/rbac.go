package middleware

import (
	"context"
	"net/http"

	"github.com/nomdb/backend/internal/database"
	"github.com/nomdb/backend/internal/logger"
	"github.com/nomdb/backend/internal/models"
	"github.com/nomdb/backend/internal/services"
)

// RequirePermission middleware checks if the authenticated user has a specific permission
func RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user from context (set by AuthMiddleware)
			user, ok := r.Context().Value(models.UserContextKey).(*models.User)
			if !ok || user == nil {
				logger.Warn("User not authenticated for permission check: %s", permission)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Check if user has the required permission
			if !services.HasPermission(user.Permissions, permission) {
				logger.Warn("User %d (@%s) denied access - missing permission: %s", user.ID, user.Username, permission)
				http.Error(w, "Forbidden: insufficient permissions", http.StatusForbidden)
				return
			}

			logger.Debug("User %d (@%s) granted access with permission: %s", user.ID, user.Username, permission)
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyPermission middleware checks if the authenticated user has any of the specified permissions
func RequireAnyPermission(permissions ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := r.Context().Value(models.UserContextKey).(*models.User)
			if !ok || user == nil {
				logger.Warn("User not authenticated for permission check: %v", permissions)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if !services.HasAnyPermission(user.Permissions, permissions) {
				logger.Warn("User %d (@%s) denied access - missing any of: %v", user.ID, user.Username, permissions)
				http.Error(w, "Forbidden: insufficient permissions", http.StatusForbidden)
				return
			}

			logger.Debug("User %d (@%s) granted access with permissions: %v", user.ID, user.Username, permissions)
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAllPermissions middleware checks if the authenticated user has all of the specified permissions
func RequireAllPermissions(permissions ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := r.Context().Value(models.UserContextKey).(*models.User)
			if !ok || user == nil {
				logger.Warn("User not authenticated for permission check: %v", permissions)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if !services.HasAllPermissions(user.Permissions, permissions) {
				logger.Warn("User %d (@%s) denied access - missing all of: %v", user.ID, user.Username, permissions)
				http.Error(w, "Forbidden: insufficient permissions", http.StatusForbidden)
				return
			}

			logger.Debug("User %d (@%s) granted access with all permissions: %v", user.ID, user.Username, permissions)
			next.ServeHTTP(w, r)
		})
	}
}

// RequireRole middleware checks if the authenticated user has a specific role
func RequireRole(roleName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := r.Context().Value(models.UserContextKey).(*models.User)
			if !ok || user == nil {
				logger.Warn("User not authenticated for role check: %s", roleName)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if !services.HasRole(user.Roles, roleName) {
				logger.Warn("User %d (@%s) denied access - missing role: %s", user.ID, user.Username, roleName)
				http.Error(w, "Forbidden: insufficient permissions", http.StatusForbidden)
				return
			}

			logger.Debug("User %d (@%s) granted access with role: %s", user.ID, user.Username, roleName)
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAdmin middleware checks if the authenticated user has the admin role
func RequireAdmin(next http.Handler) http.Handler {
	return RequireRole("admin")(next)
}

// RequireModerator middleware checks if the authenticated user has admin or moderator role
func RequireModerator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(models.UserContextKey).(*models.User)
		if !ok || user == nil {
			logger.Warn("User not authenticated for moderator check")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if !services.IsAdmin(user.Roles) && !services.IsModerator(user.Roles) {
			logger.Warn("User %d (@%s) denied access - not a moderator or admin", user.ID, user.Username)
			http.Error(w, "Forbidden: insufficient permissions", http.StatusForbidden)
			return
		}

		logger.Debug("User %d (@%s) granted moderator access", user.ID, user.Username)
		next.ServeHTTP(w, r)
	})
}

// WithUserRoles middleware loads user roles and permissions into the user object in context
func WithUserRoles(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(models.UserContextKey).(*models.User)
		if ok && user != nil {
			// Load roles and permissions for this user
			db := database.GetPool()
			roles, permissions, err := services.GetUserRolesAndPermissions(r.Context(), db, user.ID)
			if err != nil {
				logger.Error("Failed to load user roles/permissions for user %d: %v", user.ID, err)
				// Continue without roles/permissions rather than failing the request
			} else {
				user.Roles = roles
				user.Permissions = permissions

				// Update user in context
				ctx := context.WithValue(r.Context(), models.UserContextKey, user)
				r = r.WithContext(ctx)

				logger.Debug("Loaded %d roles and %d permissions for user %d (@%s)",
					len(roles), len(permissions), user.ID, user.Username)
			}
		}

		next.ServeHTTP(w, r)
	})
}
