package repository

import (
	"context"
	"fmt"

	"ipam-next/models"
)

// ---- Location management (v1.5.0) ----

const locationSelectCols = `id, parent_id, name, type, address, city, region, country, facility_code, time_zone, contact_name, contact_email, contact_phone, status, lat, lng, description, created_at, updated_at`

func scanLocation(row interface{ Scan(dest ...any) error }) (*models.Location, error) {
	l := &models.Location{}
	err := row.Scan(&l.ID, &l.ParentID, &l.Name, &l.Type, &l.Address, &l.City, &l.Region, &l.Country, &l.FacilityCode, &l.TimeZone, &l.ContactName, &l.ContactEmail, &l.ContactPhone, &l.Status, &l.Lat, &l.Lng, &l.Description, &l.CreatedAt, &l.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return l, nil
}

// LocationParams holds fields for creating or updating a location.
type LocationParams struct {
	ParentID     *int64   `json:"parent_id"`
	Name         string   `json:"name"`
	Type         string   `json:"type"`
	Address      *string  `json:"address"`
	City         *string  `json:"city"`
	Region       *string  `json:"region"`
	Country      *string  `json:"country"`
	FacilityCode *string  `json:"facility_code"`
	TimeZone     *string  `json:"time_zone"`
	ContactName  *string  `json:"contact_name"`
	ContactEmail *string  `json:"contact_email"`
	ContactPhone *string  `json:"contact_phone"`
	Status       string   `json:"status"`
	Lat          *float64 `json:"lat"`
	Lng          *float64 `json:"lng"`
	Description  *string  `json:"description"`
}

// CreateLocation inserts a new location record.
func (r *Repository) CreateLocation(ctx context.Context, p *LocationParams) (*models.Location, error) {
	query := `INSERT INTO locations (parent_id, name, type, address, city, region, country, facility_code, time_zone, contact_name, contact_email, contact_phone, status, lat, lng, description)
	          VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
	          RETURNING ` + locationSelectCols
	row := r.db.QueryRow(ctx, query, p.ParentID, p.Name, p.Type, p.Address, p.City, p.Region, p.Country, p.FacilityCode, p.TimeZone, p.ContactName, p.ContactEmail, p.ContactPhone, p.Status, p.Lat, p.Lng, p.Description)
	return scanLocation(row)
}

// GetLocationByID returns a single location.
func (r *Repository) GetLocationByID(ctx context.Context, id int64) (*models.Location, error) {
	query := `SELECT ` + locationSelectCols + ` FROM locations WHERE id=$1`
	row := r.db.QueryRow(ctx, query, id)
	l, err := scanLocation(row)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, fmt.Errorf("location not found")
		}
		return nil, err
	}
	return l, nil
}

// ListLocations returns all locations ordered by name.
func (r *Repository) ListLocations(ctx context.Context) ([]*models.Location, error) {
	query := `SELECT ` + locationSelectCols + ` FROM locations ORDER BY name`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	locs := make([]*models.Location, 0)
	for rows.Next() {
		l, err := scanLocation(rows)
		if err != nil {
			return nil, err
		}
		locs = append(locs, l)
	}
	return locs, rows.Err()
}

// ListLocationsPaginated returns a page of locations with a total count.
func (r *Repository) ListLocationsPaginated(ctx context.Context, limit, offset int) ([]*models.Location, int64, error) {
	var total int64
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM locations`).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `SELECT ` + locationSelectCols + ` FROM locations ORDER BY name LIMIT $1 OFFSET $2`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	locs := make([]*models.Location, 0)
	for rows.Next() {
		l, err := scanLocation(rows)
		if err != nil {
			return nil, 0, err
		}
		locs = append(locs, l)
	}
	return locs, total, rows.Err()
}

// UpdateLocation updates an existing location.
func (r *Repository) UpdateLocation(ctx context.Context, id int64, p *LocationParams) (*models.Location, error) {
	query := `UPDATE locations SET parent_id=$1, name=$2, type=$3, address=$4, city=$5, region=$6, country=$7, facility_code=$8, time_zone=$9, contact_name=$10, contact_email=$11, contact_phone=$12, status=$13, lat=$14, lng=$15, description=$16, updated_at=now()
	          WHERE id=$17
	          RETURNING ` + locationSelectCols
	row := r.db.QueryRow(ctx, query, p.ParentID, p.Name, p.Type, p.Address, p.City, p.Region, p.Country, p.FacilityCode, p.TimeZone, p.ContactName, p.ContactEmail, p.ContactPhone, p.Status, p.Lat, p.Lng, p.Description, id)
	l, err := scanLocation(row)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, fmt.Errorf("location not found")
		}
		return nil, err
	}
	return l, nil
}

// DeleteLocation deletes a location by ID.
func (r *Repository) DeleteLocation(ctx context.Context, id int64) error {
	ct, err := r.db.Exec(ctx, `DELETE FROM locations WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("location not found")
	}
	return nil
}

// GetLocationTree returns all locations in breadth-first order using a recursive CTE.
func (r *Repository) GetLocationTree(ctx context.Context) ([]*models.Location, error) {
	query := `
		WITH RECURSIVE loc_tree AS (
			SELECT id, parent_id, name, type, address, city, region, country, facility_code, time_zone, contact_name, contact_email, contact_phone, status, lat, lng, description, created_at, updated_at, 0 AS depth
			FROM locations WHERE parent_id IS NULL
			UNION ALL
			SELECT l.id, l.parent_id, l.name, l.type, l.address, l.city, l.region, l.country, l.facility_code, l.time_zone, l.contact_name, l.contact_email, l.contact_phone, l.status, l.lat, l.lng, l.description, l.created_at, l.updated_at, lt.depth + 1
			FROM locations l JOIN loc_tree lt ON l.parent_id = lt.id
		)
		SELECT ` + locationSelectCols + ` FROM loc_tree ORDER BY depth, name`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	locs := make([]*models.Location, 0)
	for rows.Next() {
		l, err := scanLocation(rows)
		if err != nil {
			return nil, err
		}
		locs = append(locs, l)
	}
	return locs, rows.Err()
}

// GetLocationAncestors returns the given location ID and all its ancestor IDs.
func (r *Repository) GetLocationAncestors(ctx context.Context, locationID int64) ([]int64, error) {
	query := `
		WITH RECURSIVE ancestors AS (
			SELECT id, parent_id FROM locations WHERE id=$1
			UNION ALL
			SELECT l.id, l.parent_id FROM locations l JOIN ancestors a ON l.id = a.parent_id
		)
		SELECT id FROM ancestors`
	rows, err := r.db.Query(ctx, query, locationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
