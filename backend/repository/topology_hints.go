package repository

import (
	"context"

	"padduck/models"
)

// ListTopologyHints returns all topology hints, optionally filtered by status,
// ordered by confidence_score DESC, created_at DESC.
func (r *Repository) ListTopologyHints(ctx context.Context, status string) ([]*models.TopologyHint, error) {
	query := `
		SELECT id, source_type, source_id, target_type, target_id, hint_type,
		       confidence_score, evidence, status, created_at, updated_at
		FROM topology_hints`
	args := []any{}
	if status != "" {
		query += ` WHERE status = $1`
		args = append(args, status)
	}
	query += ` ORDER BY confidence_score DESC, created_at DESC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hints []*models.TopologyHint
	for rows.Next() {
		h := &models.TopologyHint{}
		if err := rows.Scan(
			&h.ID, &h.SourceType, &h.SourceID, &h.TargetType, &h.TargetID, &h.HintType,
			&h.ConfidenceScore, &h.Evidence, &h.Status, &h.CreatedAt, &h.UpdatedAt,
		); err != nil {
			return nil, err
		}
		hints = append(hints, h)
	}
	return hints, rows.Err()
}

// GetTopologyHint returns a single topology hint by ID.
func (r *Repository) GetTopologyHint(ctx context.Context, id int64) (*models.TopologyHint, error) {
	h := &models.TopologyHint{}
	err := r.db.QueryRow(ctx, `
		SELECT id, source_type, source_id, target_type, target_id, hint_type,
		       confidence_score, evidence, status, created_at, updated_at
		FROM topology_hints WHERE id = $1`, id).Scan(
		&h.ID, &h.SourceType, &h.SourceID, &h.TargetType, &h.TargetID, &h.HintType,
		&h.ConfidenceScore, &h.Evidence, &h.Status, &h.CreatedAt, &h.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return h, nil
}

// UpsertTopologyHint creates or updates a topology hint identified by the unique
// combination of source_type, source_id, target_type, target_id, hint_type.
func (r *Repository) UpsertTopologyHint(ctx context.Context, sourceType string, sourceID int64, targetType string, targetID int64, hintType string, confidence float64, evidence *string) (*models.TopologyHint, error) {
	h := &models.TopologyHint{}
	err := r.db.QueryRow(ctx, `
		INSERT INTO topology_hints
			(source_type, source_id, target_type, target_id, hint_type, confidence_score, evidence)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT ON CONSTRAINT uq_topology_hints DO UPDATE SET
			confidence_score = EXCLUDED.confidence_score,
			evidence         = EXCLUDED.evidence,
			updated_at       = now()
		RETURNING id, source_type, source_id, target_type, target_id, hint_type,
		          confidence_score, evidence, status, created_at, updated_at`,
		sourceType, sourceID, targetType, targetID, hintType, confidence, evidence,
	).Scan(
		&h.ID, &h.SourceType, &h.SourceID, &h.TargetType, &h.TargetID, &h.HintType,
		&h.ConfidenceScore, &h.Evidence, &h.Status, &h.CreatedAt, &h.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return h, nil
}

// UpdateTopologyHintStatus updates the status of a topology hint.
func (r *Repository) UpdateTopologyHintStatus(ctx context.Context, id int64, status string) (*models.TopologyHint, error) {
	h := &models.TopologyHint{}
	err := r.db.QueryRow(ctx, `
		UPDATE topology_hints SET status = $2, updated_at = now()
		WHERE id = $1
		RETURNING id, source_type, source_id, target_type, target_id, hint_type,
		          confidence_score, evidence, status, created_at, updated_at`,
		id, status,
	).Scan(
		&h.ID, &h.SourceType, &h.SourceID, &h.TargetType, &h.TargetID, &h.HintType,
		&h.ConfidenceScore, &h.Evidence, &h.Status, &h.CreatedAt, &h.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return h, nil
}
