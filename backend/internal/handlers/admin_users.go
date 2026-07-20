package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/nomdb/backend/internal/auth"
	"github.com/nomdb/backend/internal/database"
	"github.com/nomdb/backend/internal/logger"
	"github.com/nomdb/backend/internal/models"
	"github.com/nomdb/backend/internal/services"
)

// CreateAuditLog creates an audit log entry for admin actions
func CreateAuditLog(ctx context.Context, userID int, action, resourceType string, resourceID int, details interface{}, r *http.Request) {
	detailsJSON, _ := json.Marshal(details)

	ipAddress := r.Header.Get("X-Forwarded-For")
	if ipAddress == "" {
		ipAddress = r.Header.Get("X-Real-IP")
	}
	if ipAddress == "" {
		ipAddress = r.RemoteAddr
	}

	userAgent := r.Header.Get("User-Agent")

	_, err := database.GetPool().Exec(ctx,
		`INSERT INTO audit_logs (user_id, action, resource_type, resource_id, details, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		userID, action, resourceType, resourceID, detailsJSON, ipAddress, userAgent)

	if err != nil {
		logger.Error("Failed to create audit log: %v", err)
	}
}

// @Summary List all users
// @Description Get paginated list of all users with optional filtering
// @Tags Admin
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param search query string false "Search by username, email, or full name"
// @Param role query string false "Filter by role name"
// @Param is_active query bool false "Filter by active status"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Forbidden"
// @Failure 500 {string} string "Internal server error"
// @Security BearerAuth
// @Router /admin/users [get]
func AdminListUsers(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	search := r.URL.Query().Get("search")
	roleFilter := r.URL.Query().Get("role")
	isActiveStr := r.URL.Query().Get("is_active")

	offset := (page - 1) * limit
	ctx, cancel := RequestContext(r)
	defer cancel()

	// Build query with filters
	query := `
		SELECT DISTINCT u.id, u.username, u.email, u.full_name, u.avatar_url, u.bio,
		       u.is_active, u.is_admin, u.email_verified, u.password_must_change,
		       u.last_login_at, u.created_at, u.updated_at
		FROM users u
		LEFT JOIN user_roles ur ON u.id = ur.user_id
		LEFT JOIN roles ro ON ur.role_id = ro.id
		WHERE 1=1
	`

	args := []interface{}{}
	argPos := 1

	if search != "" {
		query += ` AND (u.username ILIKE $` + strconv.Itoa(argPos) +
			` OR u.email ILIKE $` + strconv.Itoa(argPos) +
			` OR u.full_name ILIKE $` + strconv.Itoa(argPos) + `)`
		args = append(args, "%"+search+"%")
		argPos++
	}

	if roleFilter != "" {
		query += ` AND ro.name = $` + strconv.Itoa(argPos)
		args = append(args, roleFilter)
		argPos++
	}

	if isActiveStr != "" {
		isActive, err := strconv.ParseBool(isActiveStr)
		if err == nil {
			query += ` AND u.is_active = $` + strconv.Itoa(argPos)
			args = append(args, isActive)
			argPos++
		}
	}

	// Get total count
	countQuery := `SELECT COUNT(DISTINCT u.id) FROM users u
		LEFT JOIN user_roles ur ON u.id = ur.user_id
		LEFT JOIN roles ro ON ur.role_id = ro.id
		WHERE 1=1`

	if search != "" {
		countQuery += ` AND (u.username ILIKE $1 OR u.email ILIKE $1 OR u.full_name ILIKE $1)`
	}
	if roleFilter != "" {
		countQuery += ` AND ro.name = $` + strconv.Itoa(len(args))
	}
	if isActiveStr != "" {
		countQuery += ` AND u.is_active = $` + strconv.Itoa(len(args))
	}

	var total int
	err := database.GetPool().QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		logger.Error("Failed to count users: %v", err)
		http.Error(w, "Failed to count users", http.StatusInternalServerError)
		return
	}

	// Get paginated users
	query += ` ORDER BY u.created_at DESC LIMIT $` + strconv.Itoa(argPos) + ` OFFSET $` + strconv.Itoa(argPos+1)
	args = append(args, limit, offset)

	rows, err := database.GetPool().Query(ctx, query, args...)
	if err != nil {
		logger.Error("Failed to list users: %v", err)
		http.Error(w, "Failed to list users", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	users := []models.User{}
	for rows.Next() {
		var user models.User
		var createdAt, updatedAt time.Time
		var fullName, avatarURL, bio sql.NullString
		var lastLogin *time.Time

		err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &fullName, &avatarURL, &bio,
			&user.IsActive, &user.IsAdmin, &user.EmailVerified, &user.PasswordMustChange,
			&lastLogin, &createdAt, &updatedAt,
		)
		if err != nil {
			logger.Error("Failed to scan user: %v", err)
			continue
		}

		if fullName.Valid {
			user.FullName = fullName.String
		}
		if avatarURL.Valid {
			user.AvatarURL = avatarURL.String
		}
		if bio.Valid {
			user.Bio = bio.String
		}
		if lastLogin != nil {
			user.LastLoginAt = lastLogin
		}
		user.CreatedAt = createdAt
		user.UpdatedAt = updatedAt

		// Load roles for each user
		user.Roles, _ = services.GetUserRoles(ctx, database.GetPool(), user.ID)

		users = append(users, user)
	}

	response := map[string]interface{}{
		"users": users,
		"pagination": map[string]interface{}{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"totalPages": (total + limit - 1) / limit,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// @Summary Get user by ID
// @Description Get detailed information about a specific user
// @Tags Admin
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} models.User
// @Failure 400 {string} string "Invalid user ID"
// @Failure 404 {string} string "User not found"
// @Failure 500 {string} string "Internal server error"
// @Security BearerAuth
// @Router /admin/users/{id} [get]
func AdminGetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["id"])
	if err != nil || userID <= 0 {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	ctx, cancel := RequestContext(r)
	defer cancel()
	var user models.User
	var lastLogin *time.Time
	var createdAt, updatedAt time.Time
	var fullName, avatarURL, bio, oidcSubject, oidcProvider sql.NullString

	err = database.GetPool().QueryRow(ctx,
		`SELECT id, username, email, full_name, avatar_url, bio, is_active, is_admin,
		        email_verified, password_must_change, oidc_subject, oidc_provider,
		        last_login_at, created_at, updated_at
		FROM users WHERE id = $1`, userID).Scan(
		&user.ID, &user.Username, &user.Email, &fullName, &avatarURL, &bio,
		&user.IsActive, &user.IsAdmin, &user.EmailVerified, &user.PasswordMustChange,
		&oidcSubject, &oidcProvider, &lastLogin, &createdAt, &updatedAt,
	)

	if err == sql.ErrNoRows {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	if err != nil {
		logger.Error("Failed to get user: %v", err)
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		return
	}

	if fullName.Valid {
		user.FullName = fullName.String
	}
	if avatarURL.Valid {
		user.AvatarURL = avatarURL.String
	}
	if bio.Valid {
		user.Bio = bio.String
	}
	if oidcSubject.Valid {
		user.OIDCSubject = oidcSubject.String
	}
	if oidcProvider.Valid {
		user.OIDCProvider = oidcProvider.String
	}
	if lastLogin != nil {
		user.LastLoginAt = lastLogin
	}
	user.CreatedAt = createdAt
	user.UpdatedAt = updatedAt

	// Load roles and permissions
	user.Roles, _ = services.GetUserRoles(ctx, database.GetPool(), user.ID)
	permissions, _ := services.GetUserPermissions(ctx, database.GetPool(), user.ID)
	user.Permissions = permissions

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// @Summary Create a new user
// @Description Admin creates a new user account
// @Tags Admin
// @Accept json
// @Produce json
// @Param request body models.RegisterRequest true "User details"
// @Success 201 {object} models.User
// @Failure 400 {string} string "Invalid request"
// @Failure 409 {string} string "User already exists"
// @Failure 500 {string} string "Internal server error"
// @Security BearerAuth
// @Router /admin/users [post]
func AdminCreateUser(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Email == "" || req.Username == "" || req.Password == "" {
		http.Error(w, "Email, username, and password are required", http.StatusBadRequest)
		return
	}

	if !isValidEmail(req.Email) {
		http.Error(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	if len(req.Password) < 8 {
		http.Error(w, "Password must be at least 8 characters", http.StatusBadRequest)
		return
	}

	// Hash password
	passwordHash, err := auth.HashPassword(req.Password, nil)
	if err != nil {
		logger.Error("Failed to hash password: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Generate Gravatar URL
	gravatarService := services.NewGravatarService()
	avatarURL := gravatarService.GetAvatarURL(req.Email, 256)

	ctx, cancel := RequestContext(r)
	defer cancel()
	var userID int
	err = database.GetPool().QueryRow(ctx,
		`INSERT INTO users (email, username, password_hash, full_name, avatar_url, email_verified, password_must_change)
		VALUES ($1, $2, $3, $4, $5, false, true)
		RETURNING id`,
		req.Email, req.Username, passwordHash, req.FullName, avatarURL).Scan(&userID)

	if err != nil {
		if isUniqueViolation(err) {
			http.Error(w, "User with this email or username already exists", http.StatusConflict)
			return
		}
		logger.Error("Failed to create user: %v", err)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Assign default "user" role
	_, err = database.GetPool().Exec(ctx,
		`INSERT INTO user_roles (user_id, role_id)
		SELECT $1, id FROM roles WHERE name = 'user'`,
		userID)
	if err != nil {
		logger.Error("Failed to assign default role: %v", err)
	}

	// Get the created user
	adminUser := r.Context().Value(models.UserContextKey).(*models.User)
	CreateAuditLog(ctx, adminUser.ID, "create_user", "users", userID, map[string]interface{}{
		"username": req.Username,
		"email":    req.Email,
	}, r)

	var user models.User
	var createdAt, updatedAt time.Time
	err = database.GetPool().QueryRow(ctx,
		`SELECT id, username, email, full_name, avatar_url, is_active, is_admin,
		        email_verified, password_must_change, created_at, updated_at
		FROM users WHERE id = $1`, userID).Scan(
		&user.ID, &user.Username, &user.Email, &user.FullName, &user.AvatarURL,
		&user.IsActive, &user.IsAdmin, &user.EmailVerified, &user.PasswordMustChange,
		&createdAt, &updatedAt,
	)

	if err != nil {
		logger.Error("Failed to get created user: %v", err)
		http.Error(w, "User created but failed to retrieve", http.StatusInternalServerError)
		return
	}

	user.CreatedAt = createdAt
	user.UpdatedAt = updatedAt
	user.Roles, _ = services.GetUserRoles(ctx, database.GetPool(), user.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// @Summary Update user
// @Description Admin updates user information
// @Tags Admin
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param request body map[string]interface{} true "Update fields"
// @Success 200 {object} models.User
// @Failure 400 {string} string "Invalid request"
// @Failure 404 {string} string "User not found"
// @Failure 500 {string} string "Internal server error"
// @Security BearerAuth
// @Router /admin/users/{id} [put]
func AdminUpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["id"])
	if err != nil || userID <= 0 {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx, cancel := RequestContext(r)
	defer cancel()

	// Build dynamic update query
	updates := []string{}
	args := []interface{}{userID}
	argPos := 2

	if email, ok := req["email"].(string); ok && email != "" {
		if !isValidEmail(email) {
			http.Error(w, "Invalid email format", http.StatusBadRequest)
			return
		}
		updates = append(updates, "email = $"+strconv.Itoa(argPos))
		args = append(args, email)
		argPos++
	}

	if username, ok := req["username"].(string); ok && username != "" {
		updates = append(updates, "username = $"+strconv.Itoa(argPos))
		args = append(args, username)
		argPos++
	}

	if fullName, ok := req["full_name"].(string); ok {
		updates = append(updates, "full_name = $"+strconv.Itoa(argPos))
		args = append(args, fullName)
		argPos++
	}

	if bio, ok := req["bio"].(string); ok {
		updates = append(updates, "bio = $"+strconv.Itoa(argPos))
		args = append(args, bio)
		argPos++
	}

	if isActive, ok := req["is_active"].(bool); ok {
		updates = append(updates, "is_active = $"+strconv.Itoa(argPos))
		args = append(args, isActive)
		argPos++
	}

	if emailVerified, ok := req["email_verified"].(bool); ok {
		updates = append(updates, "email_verified = $"+strconv.Itoa(argPos))
		args = append(args, emailVerified)
		argPos++
	}

	if passwordMustChange, ok := req["password_must_change"].(bool); ok {
		updates = append(updates, "password_must_change = $"+strconv.Itoa(argPos))
		args = append(args, passwordMustChange)
		argPos++
	}

	if len(updates) == 0 {
		http.Error(w, "No fields to update", http.StatusBadRequest)
		return
	}

	updates = append(updates, "updated_at = NOW()")

	query := "UPDATE users SET " + strings.Join(updates, ", ") + " WHERE id = $1"
	_, err = database.GetPool().Exec(ctx, query, args...)

	if err != nil {
		if isUniqueViolation(err) {
			http.Error(w, "Email or username already exists", http.StatusConflict)
			return
		}
		logger.Error("Failed to update user: %v", err)
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}

	adminUser := r.Context().Value(models.UserContextKey).(*models.User)
	CreateAuditLog(ctx, adminUser.ID, "update_user", "users", userID, req, r)

	// Return updated user
	var user models.User
	var createdAt, updatedAt time.Time
	var fullName, avatarURL, bio sql.NullString
	var lastLogin *time.Time

	err = database.GetPool().QueryRow(ctx,
		`SELECT id, username, email, full_name, avatar_url, bio, is_active, is_admin,
		        email_verified, password_must_change, last_login_at, created_at, updated_at
		FROM users WHERE id = $1`, userID).Scan(
		&user.ID, &user.Username, &user.Email, &fullName, &avatarURL, &bio,
		&user.IsActive, &user.IsAdmin, &user.EmailVerified, &user.PasswordMustChange,
		&lastLogin, &createdAt, &updatedAt,
	)

	if err != nil {
		logger.Error("Failed to get updated user: %v", err)
		http.Error(w, "User updated but failed to retrieve", http.StatusInternalServerError)
		return
	}

	if fullName.Valid {
		user.FullName = fullName.String
	}
	if avatarURL.Valid {
		user.AvatarURL = avatarURL.String
	}
	if bio.Valid {
		user.Bio = bio.String
	}
	if lastLogin != nil {
		user.LastLoginAt = lastLogin
	}
	user.CreatedAt = createdAt
	user.UpdatedAt = updatedAt
	user.Roles, _ = services.GetUserRoles(ctx, database.GetPool(), user.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// @Summary Delete user
// @Description Admin deletes a user account
// @Tags Admin
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 204 "User deleted successfully"
// @Failure 400 {string} string "Invalid user ID or cannot delete self"
// @Failure 404 {string} string "User not found"
// @Failure 500 {string} string "Internal server error"
// @Security BearerAuth
// @Router /admin/users/{id} [delete]
func AdminDeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["id"])
	if err != nil || userID <= 0 {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	adminUser := r.Context().Value(models.UserContextKey).(*models.User)
	if adminUser.ID == userID {
		http.Error(w, "Cannot delete your own account", http.StatusBadRequest)
		return
	}

	ctx, cancel := RequestContext(r)
	defer cancel()

	// Check if user exists
	var exists bool
	err = database.GetPool().QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", userID).Scan(&exists)
	if err != nil || !exists {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Delete user (CASCADE will handle related records)
	_, err = database.GetPool().Exec(ctx, "DELETE FROM users WHERE id = $1", userID)
	if err != nil {
		logger.Error("Failed to delete user: %v", err)
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	CreateAuditLog(ctx, adminUser.ID, "delete_user", "users", userID, nil, r)

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Assign role to user
// @Description Admin assigns a role to a user
// @Tags Admin
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param request body map[string]interface{} true "Role assignment"
// @Success 200 {object} models.User
// @Failure 400 {string} string "Invalid request"
// @Failure 404 {string} string "User or role not found"
// @Failure 500 {string} string "Internal server error"
// @Security BearerAuth
// @Router /admin/users/{id}/roles [post]
func AdminAssignRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["id"])
	if err != nil || userID <= 0 {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var req struct {
		RoleID int `json:"role_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx, cancel := RequestContext(r)
	defer cancel()
	adminUser := r.Context().Value(models.UserContextKey).(*models.User)

	// Insert user role
	_, err = database.GetPool().Exec(ctx,
		`INSERT INTO user_roles (user_id, role_id, assigned_by)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, role_id) DO NOTHING`,
		userID, req.RoleID, adminUser.ID)

	if err != nil {
		logger.Error("Failed to assign role: %v", err)
		http.Error(w, "Failed to assign role", http.StatusInternalServerError)
		return
	}

	CreateAuditLog(ctx, adminUser.ID, "assign_role", "users", userID, map[string]interface{}{
		"role_id": req.RoleID,
	}, r)

	// Return updated user
	AdminGetUser(w, r)
}

// @Summary Remove role from user
// @Description Admin removes a role from a user
// @Tags Admin
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param roleId path int true "Role ID"
// @Success 200 {object} models.User
// @Failure 400 {string} string "Invalid request"
// @Failure 500 {string} string "Internal server error"
// @Security BearerAuth
// @Router /admin/users/{id}/roles/{roleId} [delete]
func AdminRemoveRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["id"])
	if err != nil || userID <= 0 {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	roleID, err := strconv.Atoi(vars["roleId"])
	if err != nil || roleID <= 0 {
		http.Error(w, "Invalid role ID", http.StatusBadRequest)
		return
	}

	ctx, cancel := RequestContext(r)
	defer cancel()
	adminUser := r.Context().Value(models.UserContextKey).(*models.User)

	_, err = database.GetPool().Exec(ctx,
		`DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2`,
		userID, roleID)

	if err != nil {
		logger.Error("Failed to remove role: %v", err)
		http.Error(w, "Failed to remove role", http.StatusInternalServerError)
		return
	}

	CreateAuditLog(ctx, adminUser.ID, "remove_role", "users", userID, map[string]interface{}{
		"role_id": roleID,
	}, r)

	// Return updated user
	AdminGetUser(w, r)
}

// @Summary Reset user password
// @Description Admin resets a user's password
// @Tags Admin
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param request body map[string]string true "New password"
// @Success 200 {string} string "Password reset successfully"
// @Failure 400 {string} string "Invalid request"
// @Failure 404 {string} string "User not found"
// @Failure 500 {string} string "Internal server error"
// @Security BearerAuth
// @Router /admin/users/{id}/reset-password [post]
func AdminResetPassword(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["id"])
	if err != nil || userID <= 0 {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var req struct {
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.NewPassword) < 8 {
		http.Error(w, "Password must be at least 8 characters", http.StatusBadRequest)
		return
	}

	passwordHash, err := auth.HashPassword(req.NewPassword, nil)
	if err != nil {
		logger.Error("Failed to hash password: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	ctx, cancel := RequestContext(r)
	defer cancel()
	_, err = database.GetPool().Exec(ctx,
		`UPDATE users SET password_hash = $1, password_must_change = true, updated_at = NOW()
		WHERE id = $2`,
		passwordHash, userID)

	if err != nil {
		logger.Error("Failed to reset password: %v", err)
		http.Error(w, "Failed to reset password", http.StatusInternalServerError)
		return
	}

	adminUser := r.Context().Value(models.UserContextKey).(*models.User)
	CreateAuditLog(ctx, adminUser.ID, "reset_password", "users", userID, nil, r)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Password reset successfully. User must change password on next login.",
	})
}

// Helper function to check for unique violations
func isUniqueViolation(err error) bool {
	return strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint")
}
