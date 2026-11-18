package repository

import (
	"context"

	"padduck/models"
)

// VRF operations

func (r *Repository) CreateVRF(ctx context.Context, name, rd, description string) (*models.VRF, error) {
	query := `INSERT INTO vrfs (name, route_distinguisher, description, organization_id)
	          VALUES ($1, $2, $3, @orgval)
	          RETURNING id, name, route_distinguisher, description, created_at, updated_at`
	vrf := &models.VRF{}
	err := r.sQueryRow(ctx, query, name, rd, description).Scan(
		&vrf.ID, &vrf.Name, &vrf.RouteDistinguisher, &vrf.Description, &vrf.CreatedAt, &vrf.UpdatedAt,
	)
	return vrf, err
}

func (r *Repository) GetVRFByID(ctx context.Context, id int64) (*models.VRF, error) {
	query := `SELECT id, name, route_distinguisher, description, created_at, updated_at FROM vrfs WHERE @org(organization_id) AND id = $1`
	vrf := &models.VRF{}
	err := r.sQueryRow(ctx, query, id).Scan(
		&vrf.ID, &vrf.Name, &vrf.RouteDistinguisher, &vrf.Description, &vrf.CreatedAt, &vrf.UpdatedAt,
	)
	return vrf, err
}

func (r *Repository) ListAllVRFs(ctx context.Context) ([]*models.VRF, error) {
	query := `SELECT id, name, route_distinguisher, description, created_at, updated_at FROM vrfs WHERE @org(organization_id) ORDER BY name ASC`
	rows, err := r.sQuery(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	vrfs := make([]*models.VRF, 0)
	for rows.Next() {
		vrf := &models.VRF{}
		err := rows.Scan(&vrf.ID, &vrf.Name, &vrf.RouteDistinguisher, &vrf.Description, &vrf.CreatedAt, &vrf.UpdatedAt)
		if err != nil {
			return nil, err
		}
		vrfs = append(vrfs, vrf)
	}
	return vrfs, rows.Err()
}

// ListVRFsPaginated returns a page of VRFs with a total count.
func (r *Repository) ListVRFsPaginated(ctx context.Context, limit, offset int) ([]*models.VRF, int64, error) {
	var total int64
	if err := r.sQueryRow(ctx, `SELECT COUNT(*) FROM vrfs WHERE @org(organization_id) `).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `SELECT id, name, route_distinguisher, description, created_at, updated_at FROM vrfs WHERE @org(organization_id) ORDER BY name ASC LIMIT $1 OFFSET $2`
	rows, err := r.sQuery(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	vrfs := make([]*models.VRF, 0)
	for rows.Next() {
		vrf := &models.VRF{}
		if err := rows.Scan(&vrf.ID, &vrf.Name, &vrf.RouteDistinguisher, &vrf.Description, &vrf.CreatedAt, &vrf.UpdatedAt); err != nil {
			return nil, 0, err
		}
		vrfs = append(vrfs, vrf)
	}
	return vrfs, total, rows.Err()
}

func (r *Repository) UpdateVRF(ctx context.Context, id int64, name, rd, description string) (*models.VRF, error) {
	query := `UPDATE vrfs SET name = $1, route_distinguisher = $2, description = $3, updated_at = CURRENT_TIMESTAMP
	          WHERE @org(organization_id) AND id = $4
	          RETURNING id, name, route_distinguisher, description, created_at, updated_at`
	vrf := &models.VRF{}
	err := r.sQueryRow(ctx, query, name, rd, description, id).Scan(
		&vrf.ID, &vrf.Name, &vrf.RouteDistinguisher, &vrf.Description, &vrf.CreatedAt, &vrf.UpdatedAt,
	)
	return vrf, err
}

func (r *Repository) DeleteVRF(ctx context.Context, id int64) error {
	_, err := r.sExec(ctx, `DELETE FROM vrfs WHERE @org(organization_id) AND id = $1`, id)
	return err
}
