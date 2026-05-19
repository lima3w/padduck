package repository

import (
	"context"

	"ipam-next/models"
)

// GetActiveBreakGlassSession returns the current active session if one exists (not expired, not ended).
func (r *Repository) GetActiveBreakGlassSession(ctx context.Context) (*models.BreakGlassSession, error) {
	query := `
		SELECT id, initiated_by_user_id, justification, expires_at, ended_at, ended_by_user_id, created_at,
		       (ended_at IS NULL AND expires_at > now()) AS is_active
		FROM break_glass_sessions
		WHERE ended_at IS NULL AND expires_at > now()
		ORDER BY created_at DESC LIMIT 1`
	row := r.db.QueryRow(ctx, query)
	s := &models.BreakGlassSession{}
	err := row.Scan(
		&s.ID, &s.InitiatedByUserID, &s.Justification,
		&s.ExpiresAt, &s.EndedAt, &s.EndedByUserID, &s.CreatedAt, &s.IsActive,
	)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// CreateBreakGlassSession inserts a new session with a 1-hour expiry.
func (r *Repository) CreateBreakGlassSession(ctx context.Context, userID int64, justification string) (*models.BreakGlassSession, error) {
	query := `
		INSERT INTO break_glass_sessions (initiated_by_user_id, justification, expires_at)
		VALUES ($1, $2, now() + interval '1 hour')
		RETURNING id, initiated_by_user_id, justification, expires_at, ended_at, ended_by_user_id, created_at,
		          (ended_at IS NULL AND expires_at > now()) AS is_active`
	row := r.db.QueryRow(ctx, query, userID, justification)
	s := &models.BreakGlassSession{}
	err := row.Scan(
		&s.ID, &s.InitiatedByUserID, &s.Justification,
		&s.ExpiresAt, &s.EndedAt, &s.EndedByUserID, &s.CreatedAt, &s.IsActive,
	)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// EndBreakGlassSession marks the active session as ended.
func (r *Repository) EndBreakGlassSession(ctx context.Context, sessionID, endedByUserID int64) (*models.BreakGlassSession, error) {
	query := `
		UPDATE break_glass_sessions
		SET ended_at = now(), ended_by_user_id = $2
		WHERE id = $1
		RETURNING id, initiated_by_user_id, justification, expires_at, ended_at, ended_by_user_id, created_at,
		          (ended_at IS NULL AND expires_at > now()) AS is_active`
	row := r.db.QueryRow(ctx, query, sessionID, endedByUserID)
	s := &models.BreakGlassSession{}
	err := row.Scan(
		&s.ID, &s.InitiatedByUserID, &s.Justification,
		&s.ExpiresAt, &s.EndedAt, &s.EndedByUserID, &s.CreatedAt, &s.IsActive,
	)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// ListBreakGlassSessions returns all sessions (newest first) for audit review.
func (r *Repository) ListBreakGlassSessions(ctx context.Context) ([]*models.BreakGlassSession, error) {
	query := `
		SELECT id, initiated_by_user_id, justification, expires_at, ended_at, ended_by_user_id, created_at,
		       (ended_at IS NULL AND expires_at > now()) AS is_active
		FROM break_glass_sessions
		ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*models.BreakGlassSession
	for rows.Next() {
		s := &models.BreakGlassSession{}
		if err := rows.Scan(
			&s.ID, &s.InitiatedByUserID, &s.Justification,
			&s.ExpiresAt, &s.EndedAt, &s.EndedByUserID, &s.CreatedAt, &s.IsActive,
		); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}
