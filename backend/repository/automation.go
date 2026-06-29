package repository

import (
	"context"
	"encoding/json"
	"strings"

	"padduck/models"
)

func (r *Repository) ListAutomationPolicies(ctx context.Context) ([]*models.AutomationPolicy, error) {
	query := `SELECT id, name, workflow, action, effect, enabled, conditions::text, message, created_at, updated_at
	          FROM automation_policies
	          ORDER BY workflow ASC, action ASC, name ASC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	policies := make([]*models.AutomationPolicy, 0)
	for rows.Next() {
		p := &models.AutomationPolicy{}
		var conditions string
		if err := rows.Scan(&p.ID, &p.Name, &p.Workflow, &p.Action, &p.Effect, &p.Enabled, &conditions, &p.Message, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		p.Conditions = decodePolicyConditions(conditions)
		policies = append(policies, p)
	}
	return policies, rows.Err()
}

func (r *Repository) ListEnabledAutomationPolicies(ctx context.Context, workflow, action string) ([]*models.AutomationPolicy, error) {
	query := `SELECT id, name, workflow, action, effect, enabled, conditions::text, message, created_at, updated_at
	          FROM automation_policies
	          WHERE enabled = TRUE
	            AND (workflow = '*' OR workflow = $1)
	            AND (action = '*' OR action = $2)
	          ORDER BY id ASC`
	rows, err := r.db.Query(ctx, query, workflow, action)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	policies := make([]*models.AutomationPolicy, 0)
	for rows.Next() {
		p := &models.AutomationPolicy{}
		var conditions string
		if err := rows.Scan(&p.ID, &p.Name, &p.Workflow, &p.Action, &p.Effect, &p.Enabled, &conditions, &p.Message, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		p.Conditions = decodePolicyConditions(conditions)
		policies = append(policies, p)
	}
	return policies, rows.Err()
}

func (r *Repository) CreateAutomationPolicy(ctx context.Context, policy *models.AutomationPolicy) (*models.AutomationPolicy, error) {
	query := `INSERT INTO automation_policies (name, workflow, action, effect, enabled, conditions, message)
	          VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7)
	          RETURNING id, name, workflow, action, effect, enabled, conditions::text, message, created_at, updated_at`
	out := &models.AutomationPolicy{}
	var conditions string
	err := r.db.QueryRow(ctx, query, policy.Name, policy.Workflow, policy.Action, policy.Effect, policy.Enabled, encodePolicyConditions(policy.Conditions), policy.Message).
		Scan(&out.ID, &out.Name, &out.Workflow, &out.Action, &out.Effect, &out.Enabled, &conditions, &out.Message, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return nil, err
	}
	out.Conditions = decodePolicyConditions(conditions)
	return out, nil
}

func (r *Repository) UpdateAutomationPolicy(ctx context.Context, policy *models.AutomationPolicy) (*models.AutomationPolicy, error) {
	query := `UPDATE automation_policies
	          SET name = $2, workflow = $3, action = $4, effect = $5, enabled = $6,
	              conditions = $7::jsonb, message = $8, updated_at = NOW()
	          WHERE id = $1
	          RETURNING id, name, workflow, action, effect, enabled, conditions::text, message, created_at, updated_at`
	out := &models.AutomationPolicy{}
	var conditions string
	err := r.db.QueryRow(ctx, query, policy.ID, policy.Name, policy.Workflow, policy.Action, policy.Effect, policy.Enabled, encodePolicyConditions(policy.Conditions), policy.Message).
		Scan(&out.ID, &out.Name, &out.Workflow, &out.Action, &out.Effect, &out.Enabled, &conditions, &out.Message, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return nil, err
	}
	out.Conditions = decodePolicyConditions(conditions)
	return out, nil
}

func (r *Repository) DeleteAutomationPolicy(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM automation_policies WHERE id = $1`, id)
	return err
}

// decodePolicyConditions parses the stored JSONB into []PolicyCondition.
// It handles both the legacy map format {"field": "value"} and the new
// array format [{"field":"f","operator":"eq","value":"v"}].
func decodePolicyConditions(raw string) []models.PolicyCondition {
	if raw == "" {
		return nil
	}
	raw = strings.TrimSpace(raw)

	// New array format.
	if strings.HasPrefix(raw, "[") {
		var out []models.PolicyCondition
		_ = json.Unmarshal([]byte(raw), &out)
		return out
	}

	// Legacy object format — migrate on read.
	var legacy map[string]string
	if err := json.Unmarshal([]byte(raw), &legacy); err != nil {
		return nil
	}
	out := make([]models.PolicyCondition, 0, len(legacy))
	for field, value := range legacy {
		op := "eq"
		if strings.HasSuffix(value, "*") || value == "*" {
			op = "glob"
		}
		out = append(out, models.PolicyCondition{Field: field, Operator: op, Value: value})
	}
	return out
}

func encodePolicyConditions(conds []models.PolicyCondition) string {
	if len(conds) == 0 {
		return "[]"
	}
	b, err := json.Marshal(conds)
	if err != nil {
		return "[]"
	}
	return string(b)
}
