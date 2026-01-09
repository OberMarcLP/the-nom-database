package services

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nomdb/backend/internal/models"
)

// GetUserRolesAndPermissions fetches all roles and permissions for a user
func GetUserRolesAndPermissions(ctx context.Context, db *pgxpool.Pool, userID int) ([]models.Role, []string, error) {

	// Query to get roles with their permissions
	query := `
		SELECT DISTINCT
			r.id, r.name, r.description, r.created_at, r.updated_at,
			p.id, p.name, p.description, p.resource, p.action, p.created_at
		FROM user_roles ur
		JOIN roles r ON ur.role_id = r.id
		JOIN role_permissions rp ON r.id = rp.role_id
		JOIN permissions p ON rp.permission_id = p.id
		WHERE ur.user_id = $1
		ORDER BY r.name, p.name
	`

	rows, err := db.Query(ctx, query, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query user roles: %w", err)
	}
	defer rows.Close()

	rolesMap := make(map[int]*models.Role)
	permissionNames := make(map[string]bool)

	for rows.Next() {
		var (
			roleID, permID                      int
			roleName, permName, resource, action string
			roleDesc, permDesc                  *string
			roleCreatedAt, roleUpdatedAt        interface{}
			permCreatedAt                       interface{}
		)

		err := rows.Scan(
			&roleID, &roleName, &roleDesc, &roleCreatedAt, &roleUpdatedAt,
			&permID, &permName, &permDesc, &resource, &action, &permCreatedAt,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Add role if not exists
		if _, exists := rolesMap[roleID]; !exists {
			rolesMap[roleID] = &models.Role{
				ID:          roleID,
				Name:        roleName,
				Description: roleDesc,
				Permissions: []models.Permission{},
			}
		}

		// Add permission to role
		perm := models.Permission{
			ID:          permID,
			Name:        permName,
			Description: permDesc,
			Resource:    resource,
			Action:      action,
		}
		rolesMap[roleID].Permissions = append(rolesMap[roleID].Permissions, perm)

		// Add permission name to set
		permissionNames[permName] = true
	}

	if err = rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("error iterating rows: %w", err)
	}

	// Convert maps to slices
	roles := make([]models.Role, 0, len(rolesMap))
	for _, role := range rolesMap {
		roles = append(roles, *role)
	}

	permissions := make([]string, 0, len(permissionNames))
	for perm := range permissionNames {
		permissions = append(permissions, perm)
	}

	return roles, permissions, nil
}

// HasPermission checks if a user has a specific permission
func HasPermission(permissions []string, required string) bool {
	for _, perm := range permissions {
		if perm == required {
			return true
		}
	}
	return false
}

// HasAnyPermission checks if a user has any of the specified permissions
func HasAnyPermission(permissions []string, required []string) bool {
	for _, req := range required {
		if HasPermission(permissions, req) {
			return true
		}
	}
	return false
}

// HasAllPermissions checks if a user has all of the specified permissions
func HasAllPermissions(permissions []string, required []string) bool {
	for _, req := range required {
		if !HasPermission(permissions, req) {
			return false
		}
	}
	return true
}

// HasRole checks if a user has a specific role
func HasRole(roles []models.Role, roleName string) bool {
	for _, role := range roles {
		if role.Name == roleName {
			return true
		}
	}
	return false
}

// IsAdmin checks if a user has the admin role
func IsAdmin(roles []models.Role) bool {
	return HasRole(roles, "admin")
}

// IsModerator checks if a user has the moderator role
func IsModerator(roles []models.Role) bool {
	return HasRole(roles, "moderator")
}

// GetUserRoles fetches roles for a user
func GetUserRoles(ctx context.Context, db *pgxpool.Pool, userID int) ([]models.Role, error) {
	query := `
		SELECT r.id, r.name, r.description, r.is_system, r.created_at, r.updated_at
		FROM roles r
		JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1
		ORDER BY r.name
	`

	rows, err := db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user roles: %w", err)
	}
	defer rows.Close()

	var roles []models.Role
	for rows.Next() {
		var role models.Role
		err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.IsSystem, &role.CreatedAt, &role.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}
		roles = append(roles, role)
	}

	return roles, nil
}

// GetUserPermissions fetches permission names for a user
func GetUserPermissions(ctx context.Context, db *pgxpool.Pool, userID int) ([]string, error) {
	query := `
		SELECT DISTINCT p.name
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		JOIN user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = $1
		ORDER BY p.name
	`

	rows, err := db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user permissions: %w", err)
	}
	defer rows.Close()

	var permissions []string
	for rows.Next() {
		var perm string
		if err := rows.Scan(&perm); err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}
		permissions = append(permissions, perm)
	}

	return permissions, nil
}
