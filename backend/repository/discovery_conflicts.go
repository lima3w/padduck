package repository

import (
	"context"
	"fmt"
	"time"

	"ipam-next/models"
)

// ListDiscoveryConflicts returns all discovery conflicts, optionally filtered by status.
// Pass "" to return all statuses.
func (r *Repository) ListDiscoveryConflicts(ctx context.Context, status string) ([]*models.DiscoveryConflict, error) {
	query := `
		SELECT id, device_id, field_name, current_value, discovered_value,
		       confidence_score, source, status, reviewed_by, reviewed_at, created_at
		FROM discovery_conflicts`
	var args []interface{}
	if status != "" {
		query += ` WHERE status = $1`
		args = append(args, status)
	}
	query += ` ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conflicts []*models.DiscoveryConflict
	for rows.Next() {
		c := &models.DiscoveryConflict{}
		if err := rows.Scan(
			&c.ID, &c.DeviceID, &c.FieldName, &c.CurrentValue, &c.DiscoveredValue,
			&c.ConfidenceScore, &c.Source, &c.Status, &c.ReviewedBy, &c.ReviewedAt, &c.CreatedAt,
		); err != nil {
			return nil, err
		}
		conflicts = append(conflicts, c)
	}
	return conflicts, rows.Err()
}

// GetDiscoveryConflict retrieves a single discovery conflict by ID.
func (r *Repository) GetDiscoveryConflict(ctx context.Context, id int64) (*models.DiscoveryConflict, error) {
	c := &models.DiscoveryConflict{}
	err := r.db.QueryRow(ctx, `
		SELECT id, device_id, field_name, current_value, discovered_value,
		       confidence_score, source, status, reviewed_by, reviewed_at, created_at
		FROM discovery_conflicts WHERE id = $1`, id).Scan(
		&c.ID, &c.DeviceID, &c.FieldName, &c.CurrentValue, &c.DiscoveredValue,
		&c.ConfidenceScore, &c.Source, &c.Status, &c.ReviewedBy, &c.ReviewedAt, &c.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// CreateDiscoveryConflict inserts a new pending discovery conflict.
func (r *Repository) CreateDiscoveryConflict(ctx context.Context, deviceID int64, fieldName, discoveredValue string, currentValue *string, confidenceScore float64, source string) (*models.DiscoveryConflict, error) {
	c := &models.DiscoveryConflict{}
	err := r.db.QueryRow(ctx, `
		INSERT INTO discovery_conflicts
		    (device_id, field_name, current_value, discovered_value, confidence_score, source)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, device_id, field_name, current_value, discovered_value,
		          confidence_score, source, status, reviewed_by, reviewed_at, created_at`,
		deviceID, fieldName, currentValue, discoveredValue, confidenceScore, source,
	).Scan(
		&c.ID, &c.DeviceID, &c.FieldName, &c.CurrentValue, &c.DiscoveredValue,
		&c.ConfidenceScore, &c.Source, &c.Status, &c.ReviewedBy, &c.ReviewedAt, &c.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// ResolveDiscoveryConflict marks a conflict as accepted or rejected.
func (r *Repository) ResolveDiscoveryConflict(ctx context.Context, id int64, action string, reviewedBy string) (*models.DiscoveryConflict, error) {
	now := time.Now()
	c := &models.DiscoveryConflict{}
	err := r.db.QueryRow(ctx, `
		UPDATE discovery_conflicts
		SET status = $1, reviewed_by = $2, reviewed_at = $3
		WHERE id = $4
		RETURNING id, device_id, field_name, current_value, discovered_value,
		          confidence_score, source, status, reviewed_by, reviewed_at, created_at`,
		action, reviewedBy, now, id,
	).Scan(
		&c.ID, &c.DeviceID, &c.FieldName, &c.CurrentValue, &c.DiscoveredValue,
		&c.ConfidenceScore, &c.Source, &c.Status, &c.ReviewedBy, &c.ReviewedAt, &c.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("resolve discovery conflict: %w", err)
	}
	return c, nil
}
