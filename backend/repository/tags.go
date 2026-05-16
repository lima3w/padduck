package repository

import (
	"context"
	"fmt"

	"ipam-next/models"
)

// IP Tag operations

func scanTag(row interface {
	Scan(dest ...any) error
}) (*models.IPTag, error) {
	tag := &models.IPTag{}
	return tag, row.Scan(&tag.ID, &tag.Name, &tag.Colour, &tag.Description, &tag.IsSystem, &tag.CreatedAt)
}

func (r *Repository) CreateIPTag(ctx context.Context, name, colour string, description *string) (*models.IPTag, error) {
	query := `INSERT INTO ip_tags (name, colour, description) VALUES ($1, $2, $3)
	          RETURNING id, name, colour, description, is_system, created_at`
	row := r.db.QueryRow(ctx, query, name, colour, description)
	return scanTag(row)
}

func (r *Repository) GetIPTagByID(ctx context.Context, id int64) (*models.IPTag, error) {
	query := `SELECT id, name, colour, description, is_system, created_at FROM ip_tags WHERE id = $1`
	row := r.db.QueryRow(ctx, query, id)
	return scanTag(row)
}

func (r *Repository) ListIPTags(ctx context.Context) ([]*models.IPTag, error) {
	query := `SELECT id, name, colour, description, is_system, created_at FROM ip_tags ORDER BY name`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tags := make([]*models.IPTag, 0)
	for rows.Next() {
		tag, err := scanTag(rows)
		if err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, rows.Err()
}

func (r *Repository) UpdateIPTag(ctx context.Context, id int64, name, colour string, description *string) (*models.IPTag, error) {
	query := `UPDATE ip_tags SET name = $2, colour = $3, description = $4 WHERE id = $1
	          RETURNING id, name, colour, description, is_system, created_at`
	row := r.db.QueryRow(ctx, query, id, name, colour, description)
	return scanTag(row)
}

func (r *Repository) DeleteIPTag(ctx context.Context, id int64) error {
	// Prevent deleting system tags
	var isSystem bool
	err := r.db.QueryRow(ctx, `SELECT is_system FROM ip_tags WHERE id = $1`, id).Scan(&isSystem)
	if err != nil {
		return fmt.Errorf("tag not found")
	}
	if isSystem {
		return fmt.Errorf("cannot delete system tag")
	}
	// Prevent deleting tags in use
	var count int64
	err = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM ip_addresses WHERE tag_id = $1`, id).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("tag is in use by %d IP address(es)", count)
	}
	_, err = r.db.Exec(ctx, `DELETE FROM ip_tags WHERE id = $1`, id)
	return err
}
