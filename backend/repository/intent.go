package repository

import (
	"context"
	"encoding/json"
	"time"

	"padduck/models"
)

const intentCols = `id, organization_id, resource_type, resource_id, operation,
		       desired_state::text, status, submitted_by, reviewed_by,
		       reviewer_note, submitted_at, reviewed_at, applied_at`

// CreateIntent inserts a new resource intent in the caller's organization.
func (r *Repository) CreateIntent(ctx context.Context, intent *models.ResourceIntent) (*models.ResourceIntent, error) {
	stateJSON, err := json.Marshal(intent.DesiredState)
	if err != nil {
		return nil, err
	}
	row := r.sQueryRow(ctx, `
		INSERT INTO resource_intents
		  (resource_type, resource_id, operation, desired_state, status, submitted_by, organization_id)
		VALUES ($1, $2, $3, $4::jsonb, $5, $6, @orgval)
		RETURNING id, organization_id, submitted_at`,
		intent.ResourceType, intent.ResourceID,
		intent.Operation, string(stateJSON), intent.Status, intent.SubmittedBy)
	return intent, row.Scan(&intent.ID, &intent.OrganizationID, &intent.SubmittedAt)
}

// GetIntent returns a single resource intent by ID within the caller's organization.
func (r *Repository) GetIntent(ctx context.Context, id int64) (*models.ResourceIntent, error) {
	row := r.sQueryRow(ctx, `SELECT `+intentCols+` FROM resource_intents WHERE @org(organization_id) AND id = $1`, id)
	return scanIntent(row)
}

// ListIntents returns intents in the caller's organization, filtered by optional
// status, resource type and resource ID, newest first.
func (r *Repository) ListIntents(ctx context.Context, status, resourceType string, resourceID *int64) ([]*models.ResourceIntent, error) {
	query := `SELECT ` + intentCols + ` FROM resource_intents WHERE @org(organization_id)`
	args := []any{}
	n := 1
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
	if resourceID != nil {
		query += ` AND resource_id = $` + itoa(n)
		args = append(args, *resourceID)
	}
	query += ` ORDER BY submitted_at DESC, id DESC LIMIT 500`

	rows, err := r.sQuery(ctx, query, args...)
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

// TransitionPendingIntent atomically moves a pending intent to a new status and
// returns it. It returns pgx.ErrNoRows when the intent does not exist in the
// caller's organization or is no longer pending, so two reviewers cannot both
// act on the same intent.
func (r *Repository) TransitionPendingIntent(ctx context.Context, id int64, status string, reviewedBy *int64, note *string, reviewedAt time.Time) (*models.ResourceIntent, error) {
	row := r.sQueryRow(ctx, `
		UPDATE resource_intents
		SET status = $2, reviewed_by = $3, reviewer_note = $4, reviewed_at = $5
		WHERE @org(organization_id) AND id = $1 AND status = 'pending'
		RETURNING `+intentCols,
		id, status, reviewedBy, note, reviewedAt)
	return scanIntent(row)
}

// UpdateIntentStatus records the outcome of applying an approved intent.
func (r *Repository) UpdateIntentStatus(ctx context.Context, id int64, status string, reviewedBy *int64, note *string, reviewedAt, appliedAt *time.Time) error {
	_, err := r.sExec(ctx, `
		UPDATE resource_intents
		SET status = $2, reviewed_by = $3, reviewer_note = $4,
		    reviewed_at = $5, applied_at = $6
		WHERE @org(organization_id) AND id = $1`,
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
