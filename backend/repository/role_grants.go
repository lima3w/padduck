package repository

import (
	"context"

	"padduck/models"
)

func (r *Repository) CreateRoleGrant(ctx context.Context, orgID, userID int64, permission string, scopeType *string, scopeID *int64, grantedBy *int64) (*models.RoleGrant, error) {
	g := &models.RoleGrant{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO role_grants (organization_id, user_id, permission, scope_type, scope_id, granted_by)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, organization_id, user_id, permission, scope_type, scope_id, granted_by, granted_at`,
		orgID, userID, permission, scopeType, scopeID, grantedBy,
	).Scan(&g.ID, &g.OrganizationID, &g.UserID, &g.Permission, &g.ScopeType, &g.ScopeID, &g.GrantedBy, &g.GrantedAt)
	return g, err
}

func (r *Repository) GetRoleGrant(ctx context.Context, id int64) (*models.RoleGrant, error) {
	g := &models.RoleGrant{}
	err := r.sQueryRow(ctx,
		`SELECT id, organization_id, user_id, permission, scope_type, scope_id, granted_by, granted_at FROM role_grants WHERE id = $1 AND @org(organization_id)`,
		id,
	).Scan(&g.ID, &g.OrganizationID, &g.UserID, &g.Permission, &g.ScopeType, &g.ScopeID, &g.GrantedBy, &g.GrantedAt)
	return g, err
}

func (r *Repository) DeleteRoleGrant(ctx context.Context, id int64) error {
	_, err := r.sExec(ctx, `DELETE FROM role_grants WHERE id = $1 AND @org(organization_id)`, id)
	return err
}

func (r *Repository) ListUserGrants(ctx context.Context, userID int64) ([]*models.RoleGrant, error) {
	rows, err := r.sQuery(ctx,
		`SELECT id, organization_id, user_id, permission, scope_type, scope_id, granted_by, granted_at
		 FROM role_grants WHERE user_id = $1 AND @org(organization_id) ORDER BY granted_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	grants := make([]*models.RoleGrant, 0)
	for rows.Next() {
		g := &models.RoleGrant{}
		if err := rows.Scan(&g.ID, &g.OrganizationID, &g.UserID, &g.Permission, &g.ScopeType, &g.ScopeID, &g.GrantedBy, &g.GrantedAt); err != nil {
			return nil, err
		}
		grants = append(grants, g)
	}
	return grants, rows.Err()
}

// ListUserPermissionGrants returns a user's grants of one permission. It is
// used by permission checks; the user already belongs to exactly one org.
func (r *Repository) ListUserPermissionGrants(ctx context.Context, userID int64, permission string) ([]*models.RoleGrant, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, organization_id, user_id, permission, scope_type, scope_id, granted_by, granted_at
		 FROM role_grants WHERE user_id = $1 AND permission = $2`, userID, permission)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	grants := make([]*models.RoleGrant, 0)
	for rows.Next() {
		g := &models.RoleGrant{}
		if err := rows.Scan(&g.ID, &g.OrganizationID, &g.UserID, &g.Permission, &g.ScopeType, &g.ScopeID, &g.GrantedBy, &g.GrantedAt); err != nil {
			return nil, err
		}
		grants = append(grants, g)
	}
	return grants, rows.Err()
}
