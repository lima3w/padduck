package repository

import (
	"context"
	"fmt"

	"padduck/models"
)

// ---- Rack management (v1.5.0) ----

// RackParams holds fields for creating or updating a rack.
type RackParams struct {
	LocationID  *int64  `json:"location_id"`
	Name        string  `json:"name"`
	SizeU       int     `json:"size_u"`
	Description *string `json:"description"`
}

const rackSelectCols = `id, location_id, name, size_u, description, created_at, updated_at`

func scanRack(row interface{ Scan(dest ...any) error }) (*models.Rack, error) {
	r := &models.Rack{}
	err := row.Scan(&r.ID, &r.LocationID, &r.Name, &r.SizeU, &r.Description, &r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// CreateRack inserts a new rack record.
func (r *Repository) CreateRack(ctx context.Context, p *RackParams) (*models.Rack, error) {
	if p.SizeU <= 0 {
		p.SizeU = 42
	}
	query := `INSERT INTO racks (location_id, name, size_u, description)
	          VALUES ($1,$2,$3,$4)
	          RETURNING ` + rackSelectCols
	row := r.db.QueryRow(ctx, query, p.LocationID, p.Name, p.SizeU, p.Description)
	return scanRack(row)
}

// GetRackByID returns a single rack.
func (r *Repository) GetRackByID(ctx context.Context, id int64) (*models.Rack, error) {
	query := `SELECT ` + rackSelectCols + ` FROM racks WHERE id=$1`
	row := r.db.QueryRow(ctx, query, id)
	rack, err := scanRack(row)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, fmt.Errorf("rack not found")
		}
		return nil, err
	}
	return rack, nil
}

// ListRacks returns all racks, optionally filtered by location_id.
func (r *Repository) ListRacks(ctx context.Context, locationID *int64) ([]*models.Rack, error) {
	query := `SELECT ` + rackSelectCols + ` FROM racks`
	var args []interface{}
	if locationID != nil {
		query += ` WHERE location_id=$1`
		args = append(args, *locationID)
	}
	query += ` ORDER BY name`
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	racks := make([]*models.Rack, 0)
	for rows.Next() {
		rack, err := scanRack(rows)
		if err != nil {
			return nil, err
		}
		racks = append(racks, rack)
	}
	return racks, rows.Err()
}

// UpdateRack updates an existing rack.
func (r *Repository) UpdateRack(ctx context.Context, id int64, p *RackParams) (*models.Rack, error) {
	if p.SizeU <= 0 {
		p.SizeU = 42
	}
	query := `UPDATE racks SET location_id=$1, name=$2, size_u=$3, description=$4, updated_at=now()
	          WHERE id=$5
	          RETURNING ` + rackSelectCols
	row := r.db.QueryRow(ctx, query, p.LocationID, p.Name, p.SizeU, p.Description, id)
	rack, err := scanRack(row)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, fmt.Errorf("rack not found")
		}
		return nil, err
	}
	return rack, nil
}

// DeleteRack deletes a rack by ID.
func (r *Repository) DeleteRack(ctx context.Context, id int64) error {
	ct, err := r.db.Exec(ctx, `DELETE FROM racks WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("rack not found")
	}
	return nil
}

// ListRacksByLocation returns racks filtered by a specific location ID.
func (r *Repository) ListRacksByLocation(ctx context.Context, locationID int64) ([]*models.Rack, error) {
	return r.ListRacks(ctx, &locationID)
}
