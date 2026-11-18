package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"padduck/models"
)

const execLogCols = `id, organization_id, policy_id, policy_name, resource_type, resource_id, resource_name,
	workflow, action, decision, actions_taken::text, triggered_by, executed_at, rolled_back_at, rolled_back_by`

func scanExecLog(row interface{ Scan(dest ...any) error }) (*models.PolicyExecutionLog, error) {
	l := &models.PolicyExecutionLog{}
	var raw string
	if err := row.Scan(&l.ID, &l.OrganizationID, &l.PolicyID, &l.PolicyName, &l.ResourceType, &l.ResourceID,
		&l.ResourceName, &l.Workflow, &l.Action, &l.Decision, &raw, &l.TriggeredBy, &l.ExecutedAt,
		&l.RolledBackAt, &l.RolledBackBy); err != nil {
		return nil, err
	}
	l.ActionsTaken = []models.PolicyActionResult{}
	if raw != "" {
		if err := json.Unmarshal([]byte(raw), &l.ActionsTaken); err != nil {
			return nil, fmt.Errorf("decoding actions_taken: %w", err)
		}
	}
	return l, nil
}

// CreatePolicyExecutionLog stores one policy match.
func (r *Repository) CreatePolicyExecutionLog(ctx context.Context, l *models.PolicyExecutionLog) (*models.PolicyExecutionLog, error) {
	actions, err := json.Marshal(l.ActionsTaken)
	if err != nil {
		return nil, err
	}
	if l.ActionsTaken == nil {
		actions = []byte("[]")
	}
	return scanExecLog(r.db.QueryRow(ctx, `
		INSERT INTO policy_execution_logs
			(organization_id, policy_id, policy_name, resource_type, resource_id, resource_name,
			 workflow, action, decision, actions_taken, triggered_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING `+execLogCols,
		l.OrganizationID, l.PolicyID, l.PolicyName, l.ResourceType, l.ResourceID, l.ResourceName,
		l.Workflow, l.Action, l.Decision, actions, l.TriggeredBy))
}

// GetPolicyExecutionLog returns one log entry.
func (r *Repository) GetPolicyExecutionLog(ctx context.Context, id int64) (*models.PolicyExecutionLog, error) {
	return scanExecLog(r.db.QueryRow(ctx, `SELECT `+execLogCols+` FROM policy_execution_logs WHERE id=$1`, id))
}

// PolicyExecutionLogFilter narrows ListPolicyExecutionLogs; zero values are ignored.
type PolicyExecutionLogFilter struct {
	PolicyID     int64
	ResourceType string
	ResourceID   int64
	Limit        int
	Offset       int
}

// ListPolicyExecutionLogs returns matching entries newest first, and the total count.
func (r *Repository) ListPolicyExecutionLogs(ctx context.Context, f PolicyExecutionLogFilter) ([]*models.PolicyExecutionLog, int, error) {
	where, args := "WHERE true", []any{}
	if f.PolicyID > 0 {
		args = append(args, f.PolicyID)
		where += fmt.Sprintf(" AND policy_id=$%d", len(args))
	}
	if f.ResourceType != "" {
		args = append(args, f.ResourceType)
		where += fmt.Sprintf(" AND resource_type=$%d", len(args))
	}
	if f.ResourceID > 0 {
		args = append(args, f.ResourceID)
		where += fmt.Sprintf(" AND resource_id=$%d", len(args))
	}
	var total int
	if err := r.db.QueryRow(ctx, `SELECT count(*) FROM policy_execution_logs `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	limit := f.Limit
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	args = append(args, limit, f.Offset)
	rows, err := r.db.Query(ctx, fmt.Sprintf(`SELECT %s FROM policy_execution_logs %s
		ORDER BY executed_at DESC, id DESC LIMIT $%d OFFSET $%d`, execLogCols, where, len(args)-1, len(args)), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []*models.PolicyExecutionLog{}
	for rows.Next() {
		l, err := scanExecLog(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, l)
	}
	return out, total, rows.Err()
}

// CountPolicyExecutions returns the number of log entries per policy ID.
func (r *Repository) CountPolicyExecutions(ctx context.Context) (map[int64]int, error) {
	rows, err := r.db.Query(ctx, `SELECT policy_id, count(*) FROM policy_execution_logs WHERE policy_id IS NOT NULL GROUP BY policy_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[int64]int{}
	for rows.Next() {
		var id int64
		var n int
		if err := rows.Scan(&id, &n); err != nil {
			return nil, err
		}
		out[id] = n
	}
	return out, rows.Err()
}

// MarkPolicyExecutionRolledBack stores the per-action rollback outcomes and stamps the entry.
// It only succeeds for an entry that has not already been rolled back.
func (r *Repository) MarkPolicyExecutionRolledBack(ctx context.Context, id int64, actions []models.PolicyActionResult, by *int64, at time.Time) (bool, error) {
	raw, err := json.Marshal(actions)
	if err != nil {
		return false, err
	}
	tag, err := r.db.Exec(ctx, `
		UPDATE policy_execution_logs SET actions_taken=$2, rolled_back_at=$3, rolled_back_by=$4
		WHERE id=$1 AND rolled_back_at IS NULL`, id, raw, at, by)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() == 1, nil
}
