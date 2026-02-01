package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/nomdb/backend/internal/database"
	"github.com/nomdb/backend/internal/logger"
	"github.com/nomdb/backend/internal/models"
)

// @Summary List all roles
// @Description Get list of all roles with their permissions
// @Tags Admin
// @Accept json
// @Produce json
// @Success 200 {array} models.Role
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Forbidden"
// @Failure 500 {string} string "Internal server error"
// @Security BearerAuth
// @Router /admin/roles [get]
func AdminListRoles(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := RequestContext(r)
	defer cancel()

	query := `
		SELECT r.id, r.name, r.description, r.is_system, r.created_at, r.updated_at
		FROM roles r
		ORDER BY r.name
	`

	rows, err := database.GetPool().Query(ctx, query)
	if err != nil {
		logger.Error("Failed to list roles: %v", err)
		http.Error(w, "Failed to list roles", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	roles := []models.Role{}
	for rows.Next() {
		var role models.Role
		err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.IsSystem, &role.CreatedAt, &role.UpdatedAt)
		if err != nil {
			logger.Error("Failed to scan role: %v", err)
			continue
		}

		// Load permissions for this role
		permQuery := `
			SELECT p.id, p.name, p.resource, p.action, p.description, p.created_at, p.updated_at
			FROM permissions p
			JOIN role_permissions rp ON p.id = rp.permission_id
			WHERE rp.role_id = $1
			ORDER BY p.resource, p.action
		`

		permRows, err := database.GetPool().Query(ctx, permQuery, role.ID)
		if err != nil {
			logger.Error("Failed to get permissions for role %d: %v", role.ID, err)
			continue
		}

		role.Permissions = []models.Permission{}
		for permRows.Next() {
			var perm models.Permission
			var desc sql.NullString
			err := permRows.Scan(&perm.ID, &perm.Name, &perm.Resource, &perm.Action, &desc, &perm.CreatedAt, &perm.UpdatedAt)
			if err != nil {
				logger.Error("Failed to scan permission: %v", err)
				continue
			}
			if desc.Valid {
				perm.Description = &desc.String
			}
			role.Permissions = append(role.Permissions, perm)
		}
		permRows.Close()

		roles = append(roles, role)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(roles)
}

// @Summary Get role by ID
// @Description Get detailed information about a specific role
// @Tags Admin
// @Accept json
// @Produce json
// @Param id path int true "Role ID"
// @Success 200 {object} models.Role
// @Failure 400 {string} string "Invalid role ID"
// @Failure 404 {string} string "Role not found"
// @Failure 500 {string} string "Internal server error"
// @Security BearerAuth
// @Router /admin/roles/{id} [get]
func AdminGetRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roleID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid role ID", http.StatusBadRequest)
		return
	}

	ctx, cancel := RequestContext(r)
	defer cancel()
	var role models.Role
	var desc sql.NullString

	err = database.GetPool().QueryRow(ctx,
		`SELECT id, name, description, is_system, created_at, updated_at
		FROM roles WHERE id = $1`, roleID).Scan(
		&role.ID, &role.Name, &desc, &role.IsSystem, &role.CreatedAt, &role.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		http.Error(w, "Role not found", http.StatusNotFound)
		return
	}
	if err != nil {
		logger.Error("Failed to get role: %v", err)
		http.Error(w, "Failed to get role", http.StatusInternalServerError)
		return
	}

	if desc.Valid {
		role.Description = &desc.String
	}

	// Load permissions for this role
	permQuery := `
		SELECT p.id, p.name, p.resource, p.action, p.description, p.created_at, p.updated_at
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1
		ORDER BY p.resource, p.action
	`

	permRows, err := database.GetPool().Query(ctx, permQuery, role.ID)
	if err != nil {
		logger.Error("Failed to get permissions for role: %v", err)
		http.Error(w, "Failed to load permissions", http.StatusInternalServerError)
		return
	}
	defer permRows.Close()

	role.Permissions = []models.Permission{}
	for permRows.Next() {
		var perm models.Permission
		var permDesc sql.NullString
		err := permRows.Scan(&perm.ID, &perm.Name, &perm.Resource, &perm.Action, &permDesc, &perm.CreatedAt, &perm.UpdatedAt)
		if err != nil {
			logger.Error("Failed to scan permission: %v", err)
			continue
		}
		if permDesc.Valid {
			perm.Description = &permDesc.String
		}
		role.Permissions = append(role.Permissions, perm)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(role)
}

// @Summary Create a new role
// @Description Admin creates a new role
// @Tags Admin
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "Role details"
// @Success 201 {object} models.Role
// @Failure 400 {string} string "Invalid request"
// @Failure 409 {string} string "Role already exists"
// @Failure 500 {string} string "Internal server error"
// @Security BearerAuth
// @Router /admin/roles [post]
func AdminCreateRole(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Role name is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := RequestContext(r)
	defer cancel()
	var roleID int
	err := database.GetPool().QueryRow(ctx,
		`INSERT INTO roles (name, description, is_system)
		VALUES ($1, $2, false)
		RETURNING id`,
		req.Name, req.Description).Scan(&roleID)

	if err != nil {
		if isUniqueViolation(err) {
			http.Error(w, "Role with this name already exists", http.StatusConflict)
			return
		}
		logger.Error("Failed to create role: %v", err)
		http.Error(w, "Failed to create role", http.StatusInternalServerError)
		return
	}

	adminUser := r.Context().Value(models.UserContextKey).(*models.User)
	CreateAuditLog(ctx, adminUser.ID, "create_role", "roles", roleID, map[string]interface{}{
		"name":        req.Name,
		"description": req.Description,
	}, r)

	// Return created role
	var role models.Role
	var desc sql.NullString
	err = database.GetPool().QueryRow(ctx,
		`SELECT id, name, description, is_system, created_at, updated_at
		FROM roles WHERE id = $1`, roleID).Scan(
		&role.ID, &role.Name, &desc, &role.IsSystem, &role.CreatedAt, &role.UpdatedAt,
	)

	if err != nil {
		logger.Error("Failed to get created role: %v", err)
		http.Error(w, "Role created but failed to retrieve", http.StatusInternalServerError)
		return
	}

	if desc.Valid {
		role.Description = &desc.String
	}
	role.Permissions = []models.Permission{}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(role)
}

// @Summary Update role
// @Description Admin updates role information
// @Tags Admin
// @Accept json
// @Produce json
// @Param id path int true "Role ID"
// @Param request body map[string]interface{} true "Update fields"
// @Success 200 {object} models.Role
// @Failure 400 {string} string "Invalid request or cannot update system role"
// @Failure 404 {string} string "Role not found"
// @Failure 500 {string} string "Internal server error"
// @Security BearerAuth
// @Router /admin/roles/{id} [put]
func AdminUpdateRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roleID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid role ID", http.StatusBadRequest)
		return
	}

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx, cancel := RequestContext(r)
	defer cancel()

	// Check if role is system role
	var isSystem bool
	err = database.GetPool().QueryRow(ctx, "SELECT is_system FROM roles WHERE id = $1", roleID).Scan(&isSystem)
	if err == sql.ErrNoRows {
		http.Error(w, "Role not found", http.StatusNotFound)
		return
	}
	if err != nil {
		logger.Error("Failed to check role: %v", err)
		http.Error(w, "Failed to check role", http.StatusInternalServerError)
		return
	}

	if isSystem {
		http.Error(w, "Cannot update system role", http.StatusBadRequest)
		return
	}

	// Build dynamic update query
	updates := []string{}
	args := []interface{}{roleID}
	argPos := 2

	if name, ok := req["name"].(string); ok && name != "" {
		updates = append(updates, "name = $"+strconv.Itoa(argPos))
		args = append(args, name)
		argPos++
	}

	if description, ok := req["description"].(string); ok {
		updates = append(updates, "description = $"+strconv.Itoa(argPos))
		args = append(args, description)
		argPos++
	}

	if len(updates) == 0 {
		http.Error(w, "No fields to update", http.StatusBadRequest)
		return
	}

	updates = append(updates, "updated_at = NOW()")

	query := "UPDATE roles SET " + strings.Join(updates, ", ") + " WHERE id = $1"
	_, err = database.GetPool().Exec(ctx, query, args...)

	if err != nil {
		if isUniqueViolation(err) {
			http.Error(w, "Role name already exists", http.StatusConflict)
			return
		}
		logger.Error("Failed to update role: %v", err)
		http.Error(w, "Failed to update role", http.StatusInternalServerError)
		return
	}

	adminUser := r.Context().Value(models.UserContextKey).(*models.User)
	CreateAuditLog(ctx, adminUser.ID, "update_role", "roles", roleID, req, r)

	// Return updated role
	AdminGetRole(w, r)
}

// @Summary Delete role
// @Description Admin deletes a role
// @Tags Admin
// @Accept json
// @Produce json
// @Param id path int true "Role ID"
// @Success 204 "Role deleted successfully"
// @Failure 400 {string} string "Invalid role ID or cannot delete system role"
// @Failure 404 {string} string "Role not found"
// @Failure 500 {string} string "Internal server error"
// @Security BearerAuth
// @Router /admin/roles/{id} [delete]
func AdminDeleteRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roleID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid role ID", http.StatusBadRequest)
		return
	}

	ctx, cancel := RequestContext(r)
	defer cancel()

	// Check if role is system role
	var isSystem bool
	err = database.GetPool().QueryRow(ctx, "SELECT is_system FROM roles WHERE id = $1", roleID).Scan(&isSystem)
	if err == sql.ErrNoRows {
		http.Error(w, "Role not found", http.StatusNotFound)
		return
	}
	if err != nil {
		logger.Error("Failed to check role: %v", err)
		http.Error(w, "Failed to check role", http.StatusInternalServerError)
		return
	}

	if isSystem {
		http.Error(w, "Cannot delete system role", http.StatusBadRequest)
		return
	}

	// Delete role (CASCADE will handle related records)
	_, err = database.GetPool().Exec(ctx, "DELETE FROM roles WHERE id = $1", roleID)
	if err != nil {
		logger.Error("Failed to delete role: %v", err)
		http.Error(w, "Failed to delete role", http.StatusInternalServerError)
		return
	}

	adminUser := r.Context().Value(models.UserContextKey).(*models.User)
	CreateAuditLog(ctx, adminUser.ID, "delete_role", "roles", roleID, nil, r)

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Assign permission to role
// @Description Admin assigns a permission to a role
// @Tags Admin
// @Accept json
// @Produce json
// @Param id path int true "Role ID"
// @Param request body map[string]int true "Permission assignment"
// @Success 200 {object} models.Role
// @Failure 400 {string} string "Invalid request"
// @Failure 404 {string} string "Role or permission not found"
// @Failure 500 {string} string "Internal server error"
// @Security BearerAuth
// @Router /admin/roles/{id}/permissions [post]
func AdminAssignPermission(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roleID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid role ID", http.StatusBadRequest)
		return
	}

	var req struct {
		PermissionID int `json:"permission_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx, cancel := RequestContext(r)
	defer cancel()
	adminUser := r.Context().Value(models.UserContextKey).(*models.User)

	// Insert role permission
	_, err = database.GetPool().Exec(ctx,
		`INSERT INTO role_permissions (role_id, permission_id)
		VALUES ($1, $2)
		ON CONFLICT (role_id, permission_id) DO NOTHING`,
		roleID, req.PermissionID)

	if err != nil {
		logger.Error("Failed to assign permission: %v", err)
		http.Error(w, "Failed to assign permission", http.StatusInternalServerError)
		return
	}

	CreateAuditLog(ctx, adminUser.ID, "assign_permission", "roles", roleID, map[string]interface{}{
		"permission_id": req.PermissionID,
	}, r)

	// Return updated role
	AdminGetRole(w, r)
}

// @Summary Remove permission from role
// @Description Admin removes a permission from a role
// @Tags Admin
// @Accept json
// @Produce json
// @Param id path int true "Role ID"
// @Param permissionId path int true "Permission ID"
// @Success 200 {object} models.Role
// @Failure 400 {string} string "Invalid request"
// @Failure 500 {string} string "Internal server error"
// @Security BearerAuth
// @Router /admin/roles/{id}/permissions/{permissionId} [delete]
func AdminRemovePermission(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roleID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid role ID", http.StatusBadRequest)
		return
	}

	permissionID, err := strconv.Atoi(vars["permissionId"])
	if err != nil {
		http.Error(w, "Invalid permission ID", http.StatusBadRequest)
		return
	}

	ctx, cancel := RequestContext(r)
	defer cancel()
	adminUser := r.Context().Value(models.UserContextKey).(*models.User)

	_, err = database.GetPool().Exec(ctx,
		`DELETE FROM role_permissions WHERE role_id = $1 AND permission_id = $2`,
		roleID, permissionID)

	if err != nil {
		logger.Error("Failed to remove permission: %v", err)
		http.Error(w, "Failed to remove permission", http.StatusInternalServerError)
		return
	}

	CreateAuditLog(ctx, adminUser.ID, "remove_permission", "roles", roleID, map[string]interface{}{
		"permission_id": permissionID,
	}, r)

	// Return updated role
	AdminGetRole(w, r)
}

// @Summary List all permissions
// @Description Get list of all available permissions
// @Tags Admin
// @Accept json
// @Produce json
// @Success 200 {array} models.Permission
// @Failure 500 {string} string "Internal server error"
// @Security BearerAuth
// @Router /admin/permissions [get]
func AdminListPermissions(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := RequestContext(r)
	defer cancel()

	query := `
		SELECT id, name, resource, action, description, created_at, updated_at
		FROM permissions
		ORDER BY resource, action
	`

	rows, err := database.GetPool().Query(ctx, query)
	if err != nil {
		logger.Error("Failed to list permissions: %v", err)
		http.Error(w, "Failed to list permissions", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	permissions := []models.Permission{}
	for rows.Next() {
		var perm models.Permission
		var desc sql.NullString
		err := rows.Scan(&perm.ID, &perm.Name, &perm.Resource, &perm.Action, &desc, &perm.CreatedAt, &perm.UpdatedAt)
		if err != nil {
			logger.Error("Failed to scan permission: %v", err)
			continue
		}
		if desc.Valid {
			perm.Description = &desc.String
		}
		permissions = append(permissions, perm)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(permissions)
}
