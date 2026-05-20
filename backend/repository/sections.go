package repository

import (
	"context"

	"padduck/models"
)

// Section operations

func (r *Repository) CreateSection(ctx context.Context, name, description string, createdBy int64) (*models.Section, error) {
	query := `INSERT INTO sections (name, description, created_by) VALUES ($1, $2, $3) RETURNING id, name, description, created_by, created_at, updated_at`
	row := r.db.QueryRow(ctx, query, name, description, createdBy)

	section := &models.Section{}
	err := row.Scan(&section.ID, &section.Name, &section.Description, &section.CreatedBy, &section.CreatedAt, &section.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return section, nil
}

func (r *Repository) GetSectionByID(ctx context.Context, id int64) (*models.Section, error) {
	query := `SELECT id, name, description, created_by, created_at, updated_at FROM sections WHERE id = $1`
	row := r.db.QueryRow(ctx, query, id)

	section := &models.Section{}
	err := row.Scan(&section.ID, &section.Name, &section.Description, &section.CreatedBy, &section.CreatedAt, &section.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return section, nil
}

func (r *Repository) ListAllSections(ctx context.Context) ([]*models.Section, error) {
	query := `SELECT id, name, description, created_by, created_at, updated_at FROM sections ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sections := make([]*models.Section, 0)
	for rows.Next() {
		section := &models.Section{}
		err := rows.Scan(&section.ID, &section.Name, &section.Description, &section.CreatedBy, &section.CreatedAt, &section.UpdatedAt)
		if err != nil {
			return nil, err
		}
		sections = append(sections, section)
	}
	return sections, rows.Err()
}

func (r *Repository) UpdateSection(ctx context.Context, id int64, name, description string) (*models.Section, error) {
	query := `UPDATE sections SET name = $2, description = $3 WHERE id = $1 RETURNING id, name, description, created_by, created_at, updated_at`
	row := r.db.QueryRow(ctx, query, id, name, description)

	section := &models.Section{}
	err := row.Scan(&section.ID, &section.Name, &section.Description, &section.CreatedBy, &section.CreatedAt, &section.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return section, nil
}

func (r *Repository) DeleteSection(ctx context.Context, id int64) error {
	query := `DELETE FROM sections WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}
