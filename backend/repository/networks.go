package repository

import (
	"context"

	"padduck/models"
)

// Network operations

func (r *Repository) CreateNetwork(ctx context.Context, name, description string, createdBy int64) (*models.Network, error) {
	query := `INSERT INTO networks (name, description, created_by) VALUES ($1, $2, $3) RETURNING id, name, description, created_by, created_at, updated_at`
	row := r.db.QueryRow(ctx, query, name, description, createdBy)

	section := &models.Network{}
	err := row.Scan(&section.ID, &section.Name, &section.Description, &section.CreatedBy, &section.CreatedAt, &section.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return section, nil
}

func (r *Repository) GetNetworkByID(ctx context.Context, id int64) (*models.Network, error) {
	query := `SELECT id, name, description, created_by, created_at, updated_at FROM networks WHERE id = $1`
	row := r.db.QueryRow(ctx, query, id)

	section := &models.Network{}
	err := row.Scan(&section.ID, &section.Name, &section.Description, &section.CreatedBy, &section.CreatedAt, &section.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return section, nil
}

func (r *Repository) ListAllNetworks(ctx context.Context) ([]*models.Network, error) {
	query := `SELECT id, name, description, created_by, created_at, updated_at FROM networks ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sections := make([]*models.Network, 0)
	for rows.Next() {
		section := &models.Network{}
		err := rows.Scan(&section.ID, &section.Name, &section.Description, &section.CreatedBy, &section.CreatedAt, &section.UpdatedAt)
		if err != nil {
			return nil, err
		}
		sections = append(sections, section)
	}
	return sections, rows.Err()
}

func (r *Repository) UpdateNetwork(ctx context.Context, id int64, name, description string) (*models.Network, error) {
	query := `UPDATE networks SET name = $2, description = $3 WHERE id = $1 RETURNING id, name, description, created_by, created_at, updated_at`
	row := r.db.QueryRow(ctx, query, id, name, description)

	section := &models.Network{}
	err := row.Scan(&section.ID, &section.Name, &section.Description, &section.CreatedBy, &section.CreatedAt, &section.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return section, nil
}

func (r *Repository) DeleteNetwork(ctx context.Context, id int64) error {
	query := `DELETE FROM networks WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}
