package repository

import (
	"context"

	"padduck/models"
)

// ListDeviceTypes returns all device types ordered by name.
func (r *Repository) ListDeviceTypes(ctx context.Context) ([]*models.DeviceType, error) {
	query := `SELECT id, name, COALESCE(icon, ''), description, created_at, updated_at FROM device_types ORDER BY name`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	types := make([]*models.DeviceType, 0)
	for rows.Next() {
		dt := &models.DeviceType{}
		if err := rows.Scan(&dt.ID, &dt.Name, &dt.Icon, &dt.Description, &dt.CreatedAt, &dt.UpdatedAt); err != nil {
			return nil, err
		}
		types = append(types, dt)
	}
	return types, rows.Err()
}
