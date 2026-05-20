package repository

import (
	"context"
	"fmt"

	"padduck/models"
)

// ---- RBAC ----

func (r *Repository) CreateRole(ctx context.Context, name, description string, isSystem bool) (*models.Role, error) {
	query := `INSERT INTO roles (name, description, is_system) VALUES ($1, $2, $3)
              RETURNING id, name, description, is_system, created_at, updated_at`
	row := r.db.QueryRow(ctx, query, name, description, isSystem)
	role := &models.Role{}
	err := row.Scan(&role.ID, &role.Name, &role.Description, &role.IsSystem, &role.CreatedAt, &role.UpdatedAt)
	return role, err
}

func (r *Repository) GetRoleByID(ctx context.Context, id int64) (*models.Role, error) {
	query := `SELECT id, name, description, is_system, created_at, updated_at FROM roles WHERE id = $1`
	row := r.db.QueryRow(ctx, query, id)
	role := &models.Role{}
	err := row.Scan(&role.ID, &role.Name, &role.Description, &role.IsSystem, &role.CreatedAt, &role.UpdatedAt)
	if err != nil {
		return nil, err
	}
	role.Permissions, err = r.GetRolePermissions(ctx, id)
	return role, err
}

func (r *Repository) ListRoles(ctx context.Context) ([]*models.Role, error) {
	query := `SELECT id, name, description, is_system, created_at, updated_at FROM roles ORDER BY is_system DESC, name`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var roles []*models.Role
	for rows.Next() {
		role := &models.Role{}
		if err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.IsSystem, &role.CreatedAt, &role.UpdatedAt); err != nil {
			return nil, err
		}
		role.Permissions, _ = r.GetRolePermissions(ctx, role.ID)
		roles = append(roles, role)
	}
	return roles, rows.Err()
}

func (r *Repository) UpdateRole(ctx context.Context, id int64, name, description string) (*models.Role, error) {
	query := `UPDATE roles SET name=$1, description=$2, updated_at=CURRENT_TIMESTAMP WHERE id=$3 AND is_system=FALSE
              RETURNING id, name, description, is_system, created_at, updated_at`
	row := r.db.QueryRow(ctx, query, name, description, id)
	role := &models.Role{}
	err := row.Scan(&role.ID, &role.Name, &role.Description, &role.IsSystem, &role.CreatedAt, &role.UpdatedAt)
	if err != nil {
		return nil, err
	}
	role.Permissions, _ = r.GetRolePermissions(ctx, id)
	return role, nil
}

func (r *Repository) DeleteRole(ctx context.Context, id int64) error {
	// Only delete non-system roles
	res, err := r.db.Exec(ctx, `DELETE FROM roles WHERE id=$1 AND is_system=FALSE`, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("role not found or is a system role")
	}
	return nil
}

func (r *Repository) GetRolePermissions(ctx context.Context, roleID int64) ([]*models.RolePermission, error) {
	query := `SELECT id, role_id, permission, resource_type, resource_id, created_at FROM role_permissions WHERE role_id=$1 ORDER BY permission`
	rows, err := r.db.Query(ctx, query, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var perms []*models.RolePermission
	for rows.Next() {
		p := &models.RolePermission{}
		if err := rows.Scan(&p.ID, &p.RoleID, &p.Permission, &p.ResourceType, &p.ResourceID, &p.CreatedAt); err != nil {
			return nil, err
		}
		perms = append(perms, p)
	}
	return perms, rows.Err()
}

func (r *Repository) AddPermissionToRole(ctx context.Context, roleID int64, permission string, resourceType *string, resourceID *int64) (*models.RolePermission, error) {
	query := `INSERT INTO role_permissions (role_id, permission, resource_type, resource_id) VALUES ($1, $2, $3, $4)
              RETURNING id, role_id, permission, resource_type, resource_id, created_at`
	row := r.db.QueryRow(ctx, query, roleID, permission, resourceType, resourceID)
	p := &models.RolePermission{}
	err := row.Scan(&p.ID, &p.RoleID, &p.Permission, &p.ResourceType, &p.ResourceID, &p.CreatedAt)
	return p, err
}

func (r *Repository) RemovePermissionFromRole(ctx context.Context, permissionID int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM role_permissions WHERE id=$1`, permissionID)
	return err
}

func (r *Repository) AssignRoleToUser(ctx context.Context, userID, roleID int64) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		userID, roleID)
	return err
}

// AssignRoleToUserWithLocation assigns a role to a user scoped to a specific location (or globally if locationID is nil).
func (r *Repository) AssignRoleToUserWithLocation(ctx context.Context, userID, roleID int64, locationID *int64) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO user_roles (user_id, role_id, location_id) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`,
		userID, roleID, locationID)
	return err
}

// GetUserRoleLocationIDs returns distinct location_id values from user_roles for a user (nil = global/unscoped).
func (r *Repository) GetUserRoleLocationIDs(ctx context.Context, userID int64) ([]int64, bool, error) {
	query := `SELECT DISTINCT location_id FROM user_roles WHERE user_id=$1`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	var scopedIDs []int64
	hasGlobal := false
	for rows.Next() {
		var locID *int64
		if err := rows.Scan(&locID); err != nil {
			return nil, false, err
		}
		if locID == nil {
			hasGlobal = true
		} else {
			scopedIDs = append(scopedIDs, *locID)
		}
	}
	return scopedIDs, hasGlobal, rows.Err()
}

func (r *Repository) RemoveRoleFromUser(ctx context.Context, userID, roleID int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM user_roles WHERE user_id=$1 AND role_id=$2`, userID, roleID)
	return err
}

func (r *Repository) GetUserRoles(ctx context.Context, userID int64) ([]*models.Role, error) {
	query := `SELECT r.id, r.name, r.description, r.is_system, r.created_at, r.updated_at
              FROM roles r JOIN user_roles ur ON r.id=ur.role_id WHERE ur.user_id=$1 ORDER BY r.name`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var roles []*models.Role
	for rows.Next() {
		role := &models.Role{}
		if err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.IsSystem, &role.CreatedAt, &role.UpdatedAt); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, rows.Err()
}

func (r *Repository) GetUserPermissions(ctx context.Context, userID int64) ([]*models.RolePermission, error) {
	query := `SELECT DISTINCT rp.id, rp.role_id, rp.permission, rp.resource_type, rp.resource_id, rp.created_at
              FROM role_permissions rp
              JOIN user_roles ur ON rp.role_id = ur.role_id
              WHERE ur.user_id = $1
              ORDER BY rp.permission`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var perms []*models.RolePermission
	for rows.Next() {
		p := &models.RolePermission{}
		if err := rows.Scan(&p.ID, &p.RoleID, &p.Permission, &p.ResourceType, &p.ResourceID, &p.CreatedAt); err != nil {
			return nil, err
		}
		perms = append(perms, p)
	}
	return perms, rows.Err()
}

func (r *Repository) CountUserRoles(ctx context.Context, userID int64) (int64, error) {
	var count int64
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM user_roles WHERE user_id=$1`, userID).Scan(&count)
	return count, err
}
