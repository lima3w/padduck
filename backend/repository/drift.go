package repository

import (
	"context"
	"encoding/json"
	"strconv"

	"padduck/models"
)

// UpsertDriftItem inserts or updates the single open drift item for a resource.
func (r *Repository) UpsertDriftItem(ctx context.Context, item *models.DriftItem) error {
	diffsJSON, err := json.Marshal(item.FieldDiffs)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, `
		INSERT INTO drift_items (organization_id, resource_type, resource_id, observed_state_id, field_diffs)
		VALUES ($1, $2, $3, $4, $5::jsonb)
		ON CONFLICT (resource_type, resource_id) WHERE status = 'open'
		DO UPDATE SET
		  observed_state_id = EXCLUDED.observed_state_id,
		  field_diffs       = EXCLUDED.field_diffs,
		  updated_at        = NOW()`,
		item.OrganizationID, item.ResourceType, item.ResourceID,
		item.ObservedStateID, string(diffsJSON))
	return err
}

// GetDriftItem returns a single drift item by ID.
func (r *Repository) GetDriftItem(ctx context.Context, id int64) (*models.DriftItem, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, organization_id, resource_type, resource_id, observed_state_id,
		       field_diffs::text, status, resolved_by, resolved_at, created_at, updated_at
		FROM drift_items WHERE id = $1`, id)
	return scanDriftItem(row)
}

// ListDriftItems returns drift items, optionally filtered by status and org.
func (r *Repository) ListDriftItems(ctx context.Context, orgID *int64, status string) ([]*models.DriftItem, error) {
	query := `
		SELECT id, organization_id, resource_type, resource_id, observed_state_id,
		       field_diffs::text, status, resolved_by, resolved_at, created_at, updated_at
		FROM drift_items WHERE 1=1`
	args := []any{}
	n := 1
	if status != "" {
		query += ` AND status = $` + itoa(n)
		args = append(args, status)
		n++
	}
	if orgID != nil {
		query += ` AND organization_id = $` + itoa(n)
		args = append(args, *orgID)
		n++
	}
	query += ` ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*models.DriftItem
	for rows.Next() {
		item, err := scanDriftItem(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

// ResolveDriftItem sets the status (accepted/dismissed/escalated) on a drift item.
func (r *Repository) ResolveDriftItem(ctx context.Context, id int64, status string, resolvedBy *int64) error {
	_, err := r.db.Exec(ctx, `
		UPDATE drift_items
		SET status = $2, resolved_by = $3, resolved_at = NOW(), updated_at = NOW()
		WHERE id = $1`, id, status, resolvedBy)
	return err
}

type driftScanner interface {
	Scan(dest ...any) error
}

func scanDriftItem(row driftScanner) (*models.DriftItem, error) {
	item := &models.DriftItem{}
	var rawDiffs string
	err := row.Scan(&item.ID, &item.OrganizationID, &item.ResourceType, &item.ResourceID,
		&item.ObservedStateID, &rawDiffs, &item.Status,
		&item.ResolvedBy, &item.ResolvedAt, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal([]byte(rawDiffs), &item.FieldDiffs)
	return item, nil
}

func itoa(n int) string { return strconv.Itoa(n) }
