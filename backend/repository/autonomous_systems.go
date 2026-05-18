package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"ipam-next/models"
)

func (r *Repository) CreateAutonomousSystem(ctx context.Context, asn int64, name, description, asType, rir string) (*models.AutonomousSystem, error) {
	query := `INSERT INTO autonomous_systems (asn, name, description, type, rir)
	          VALUES ($1, $2, $3, $4, $5)
	          RETURNING id, asn, name, description, type, rir, created_at, updated_at`
	a := &models.AutonomousSystem{}
	err := r.db.QueryRow(ctx, query, asn, name, description, asType, rir).Scan(
		&a.ID, &a.ASN, &a.Name, &a.Description, &a.Type, &a.RIR, &a.CreatedAt, &a.UpdatedAt,
	)
	return a, err
}

func (r *Repository) GetAutonomousSystemByID(ctx context.Context, id int64) (*models.AutonomousSystem, error) {
	query := `SELECT id, asn, name, description, type, rir, created_at, updated_at FROM autonomous_systems WHERE id = $1`
	a := &models.AutonomousSystem{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&a.ID, &a.ASN, &a.Name, &a.Description, &a.Type, &a.RIR, &a.CreatedAt, &a.UpdatedAt,
	)
	return a, err
}

func (r *Repository) ListAllAutonomousSystems(ctx context.Context) ([]*models.AutonomousSystem, error) {
	query := `SELECT id, asn, name, description, type, rir, created_at, updated_at FROM autonomous_systems ORDER BY asn ASC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]*models.AutonomousSystem, 0)
	for rows.Next() {
		a := &models.AutonomousSystem{}
		if err := rows.Scan(&a.ID, &a.ASN, &a.Name, &a.Description, &a.Type, &a.RIR, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, a)
	}
	return items, rows.Err()
}

// ListAutonomousSystemsPaginated returns a page of autonomous systems with a total count.
func (r *Repository) ListAutonomousSystemsPaginated(ctx context.Context, limit, offset int) ([]*models.AutonomousSystem, int64, error) {
	var total int64
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM autonomous_systems`).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `SELECT id, asn, name, description, type, rir, created_at, updated_at FROM autonomous_systems ORDER BY asn ASC LIMIT $1 OFFSET $2`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]*models.AutonomousSystem, 0)
	for rows.Next() {
		a := &models.AutonomousSystem{}
		if err := rows.Scan(&a.ID, &a.ASN, &a.Name, &a.Description, &a.Type, &a.RIR, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, a)
	}
	return items, total, rows.Err()
}

func (r *Repository) UpdateAutonomousSystem(ctx context.Context, id, asn int64, name, description, asType, rir string) (*models.AutonomousSystem, error) {
	query := `UPDATE autonomous_systems SET asn = $1, name = $2, description = $3, type = $4, rir = $5, updated_at = CURRENT_TIMESTAMP
	          WHERE id = $6
	          RETURNING id, asn, name, description, type, rir, created_at, updated_at`
	a := &models.AutonomousSystem{}
	err := r.db.QueryRow(ctx, query, asn, name, description, asType, rir, id).Scan(
		&a.ID, &a.ASN, &a.Name, &a.Description, &a.Type, &a.RIR, &a.CreatedAt, &a.UpdatedAt,
	)
	return a, err
}

func (r *Repository) DeleteAutonomousSystem(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM autonomous_systems WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
