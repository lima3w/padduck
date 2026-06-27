package repository

import (
	"context"
	"encoding/json"
	"time"

	"padduck/models"
)

// CreateIntent inserts a new resource intent and returns it with its assigned ID.
func (r *Repository) CreateIntent(ctx context.Context, intent *models.ResourceIntent) (*models.ResourceIntent, error) {
	stateJSON, err := json.Marshal(intent.DesiredState)
	if err != nil {
		return nil, err
	}
	row := r.db.QueryRow(ctx, `
		INSERT INTO resource_intents
		  (organization_id, resource_type, resource_id, operation, desired_state, status, submitted_by)
		VALUES ($1, $2, $3, $4, $5::jsonb, $6, $7)
		RETURNING id, submitted_at`,
		intent.OrganizationID, intent.ResourceType, intent.ResourceID,
		intent.Operation, string(stateJSON), intent.Status, intent.SubmittedBy)
	return intent, row.Scan(&intent.ID, &intent.SubmittedAt)
}

// GetIntent returns a single resource intent by ID.
func (r *Repository) GetIntent(ctx context.Context, id int64) (*models.ResourceIntent, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, organization_id, resource_type, resource_id, operation,
		       desired_state::text, status, submitted_by, reviewed_by,
		       reviewer_note, submitted_at, reviewed_at, applied_at
		FROM resource_intents WHERE id = $1`, id)
	return scanIntent(row)
}

// ListIntents returns intents filtered by optional org, status, and resource_type.
func (r *Repository) ListIntents(ctx context.Context, orgID *int64, status, resourceType string) ([]*models.ResourceIntent, error) {
	query := `
		SELECT id, organization_id, resource_type, resource_id, operation,
		       desired_state::text, status, submitted_by, reviewed_by,
		       reviewer_note, submitted_at, reviewed_at, applied_at
		FROM resource_intents WHERE 1=1`
	args := []any{}
	n := 1
	if orgID != nil {
		query += ` AND organization_id = $` + itoa(n)
		args = append(args, *orgID)
		n++
	}
	if status != "" {
		query += ` AND status = $` + itoa(n)
		args = append(args, status)
		n++
	}
	if resourceType != "" {
		query += ` AND resource_type = $` + itoa(n)
		args = append(args, resourceType)
		n++
	}
	query += ` ORDER BY submitted_at DESC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*models.ResourceIntent
	for rows.Next() {
		item, err := scanIntent(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

// UpdateIntentStatus transitions an intent to a new status and records review/apply metadata.
func (r *Repository) UpdateIntentStatus(ctx context.Context, id int64, status string, reviewedBy *int64, note *string, reviewedAt, appliedAt *time.Time) error {
	_, err := r.db.Exec(ctx, `
		UPDATE resource_intents
		SET status = $2, reviewed_by = $3, reviewer_note = $4,
		    reviewed_at = $5, applied_at = $6
		WHERE id = $1`,
		id, status, reviewedBy, note, reviewedAt, appliedAt)
	return err
}

type intentScanner interface {
	Scan(dest ...any) error
}

func scanIntent(row intentScanner) (*models.ResourceIntent, error) {
	item := &models.ResourceIntent{}
	var rawState string
	err := row.Scan(
		&item.ID, &item.OrganizationID, &item.ResourceType, &item.ResourceID,
		&item.Operation, &rawState, &item.Status,
		&item.SubmittedBy, &item.ReviewedBy, &item.ReviewerNote,
		&item.SubmittedAt, &item.ReviewedAt, &item.AppliedAt,
	)
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal([]byte(rawState), &item.DesiredState)
	return item, nil
}
