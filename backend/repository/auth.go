package repository

import (
	"context"
	"time"

	"ipam-next/models"
)

// API Token operations

func (r *Repository) CreateAPIToken(ctx context.Context, userID int64, tokenHash, name string) (*models.APIToken, error) {
	query := `INSERT INTO api_tokens (user_id, token_hash, name) VALUES ($1, $2, $3)
	          RETURNING id, user_id, token_hash, name, scope, usage_count, last_used_at, last_used_ip, expires_at, rotation_grace_expires_at, created_at, updated_at`
	row := r.db.QueryRow(ctx, query, userID, tokenHash, name)

	token := &models.APIToken{}
	err := row.Scan(&token.ID, &token.UserID, &token.TokenHash, &token.Name, &token.Scope,
		&token.UsageCount, &token.LastUsedAt, &token.LastUsedIP,
		&token.ExpiresAt, &token.RotationGraceExpiresAt, &token.CreatedAt, &token.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (r *Repository) CreateAPITokenFull(ctx context.Context, userID int64, tokenHash, name, scope string, expiresAt *time.Time) (*models.APIToken, error) {
	query := `INSERT INTO api_tokens (user_id, token_hash, name, scope, expires_at)
	          VALUES ($1, $2, $3, $4, $5)
	          RETURNING id, user_id, token_hash, name, scope, usage_count, last_used_at, last_used_ip, expires_at, rotation_grace_expires_at, created_at, updated_at`
	row := r.db.QueryRow(ctx, query, userID, tokenHash, name, scope, expiresAt)
	token := &models.APIToken{}
	err := row.Scan(&token.ID, &token.UserID, &token.TokenHash, &token.Name, &token.Scope,
		&token.UsageCount, &token.LastUsedAt, &token.LastUsedIP,
		&token.ExpiresAt, &token.RotationGraceExpiresAt, &token.CreatedAt, &token.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (r *Repository) GetAPITokenByHash(ctx context.Context, tokenHash string) (*models.APIToken, error) {
	query := `SELECT id, user_id, token_hash, name, scope, usage_count, last_used_at, last_used_ip, expires_at, rotation_grace_expires_at, created_at, updated_at FROM api_tokens WHERE token_hash = $1`
	row := r.db.QueryRow(ctx, query, tokenHash)

	token := &models.APIToken{}
	err := row.Scan(&token.ID, &token.UserID, &token.TokenHash, &token.Name, &token.Scope,
		&token.UsageCount, &token.LastUsedAt, &token.LastUsedIP,
		&token.ExpiresAt, &token.RotationGraceExpiresAt, &token.CreatedAt, &token.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (r *Repository) ListAPITokensByUser(ctx context.Context, userID int64) ([]*models.APIToken, error) {
	query := `SELECT id, user_id, token_hash, name, scope, usage_count, last_used_at, last_used_ip, expires_at, rotation_grace_expires_at, created_at, updated_at FROM api_tokens WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tokens := make([]*models.APIToken, 0)
	for rows.Next() {
		token := &models.APIToken{}
		err := rows.Scan(&token.ID, &token.UserID, &token.TokenHash, &token.Name, &token.Scope,
			&token.UsageCount, &token.LastUsedAt, &token.LastUsedIP,
			&token.ExpiresAt, &token.RotationGraceExpiresAt, &token.CreatedAt, &token.UpdatedAt)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}
	return tokens, rows.Err()
}

func (r *Repository) ListAPITokenAnalytics(ctx context.Context) ([]*models.APIToken, error) {
	query := `SELECT t.id, t.user_id, COALESCE(u.username, ''), t.token_hash, t.name, t.scope,
	                 t.usage_count, t.last_used_at, t.last_used_ip, t.expires_at,
	                 t.rotation_grace_expires_at, t.created_at, t.updated_at
	          FROM api_tokens t
	          LEFT JOIN users u ON u.id = t.user_id
	          ORDER BY t.usage_count DESC, t.last_used_at DESC NULLS LAST, t.created_at DESC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tokens := make([]*models.APIToken, 0)
	for rows.Next() {
		token := &models.APIToken{}
		err := rows.Scan(&token.ID, &token.UserID, &token.Username, &token.TokenHash, &token.Name, &token.Scope,
			&token.UsageCount, &token.LastUsedAt, &token.LastUsedIP,
			&token.ExpiresAt, &token.RotationGraceExpiresAt, &token.CreatedAt, &token.UpdatedAt)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}
	return tokens, rows.Err()
}

func (r *Repository) UpdateAPITokenLastUsed(ctx context.Context, tokenID int64, ip string) error {
	query := `UPDATE api_tokens SET last_used_at = CURRENT_TIMESTAMP, last_used_ip = $2, usage_count = usage_count + 1 WHERE id = $1`
	_, err := r.db.Exec(ctx, query, tokenID, nullableString(ip))
	return err
}

func (r *Repository) DeleteAPIToken(ctx context.Context, tokenID int64) error {
	query := `DELETE FROM api_tokens WHERE id = $1`
	_, err := r.db.Exec(ctx, query, tokenID)
	return err
}

func (r *Repository) MarkAPITokenRotated(ctx context.Context, tokenID int64, graceExpiresAt time.Time) error {
	query := `UPDATE api_tokens SET rotation_grace_expires_at = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.db.Exec(ctx, query, tokenID, graceExpiresAt)
	return err
}

func (r *Repository) ExtendAPIToken(ctx context.Context, tokenID, userID int64, newExpiresAt time.Time) (*models.APIToken, error) {
	query := `UPDATE api_tokens SET expires_at = $3, updated_at = CURRENT_TIMESTAMP
	          WHERE id = $1 AND user_id = $2
	          RETURNING id, user_id, token_hash, name, scope, usage_count, last_used_at, last_used_ip, expires_at, rotation_grace_expires_at, created_at, updated_at`
	row := r.db.QueryRow(ctx, query, tokenID, userID, newExpiresAt)
	token := &models.APIToken{}
	err := row.Scan(&token.ID, &token.UserID, &token.TokenHash, &token.Name, &token.Scope,
		&token.UsageCount, &token.LastUsedAt, &token.LastUsedIP,
		&token.ExpiresAt, &token.RotationGraceExpiresAt, &token.CreatedAt, &token.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (r *Repository) GetAPITokenByID(ctx context.Context, tokenID int64) (*models.APIToken, error) {
	query := `SELECT id, user_id, token_hash, name, scope, usage_count, last_used_at, last_used_ip, expires_at, rotation_grace_expires_at, created_at, updated_at FROM api_tokens WHERE id = $1`
	row := r.db.QueryRow(ctx, query, tokenID)
	token := &models.APIToken{}
	err := row.Scan(&token.ID, &token.UserID, &token.TokenHash, &token.Name, &token.Scope,
		&token.UsageCount, &token.LastUsedAt, &token.LastUsedIP,
		&token.ExpiresAt, &token.RotationGraceExpiresAt, &token.CreatedAt, &token.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (r *Repository) DeleteExpiredAPITokens(ctx context.Context) error {
	// Delete tokens expired more than 30 days ago with no grace period active
	query := `DELETE FROM api_tokens WHERE expires_at IS NOT NULL AND expires_at < NOW() - INTERVAL '30 days' AND (rotation_grace_expires_at IS NULL OR rotation_grace_expires_at < NOW() - INTERVAL '30 days')`
	_, err := r.db.Exec(ctx, query)
	return err
}

// Session operations

func (r *Repository) CreateSession(ctx context.Context, userID int64, tokenHash, deviceName, ipAddress, userAgent string, absoluteExpiresAt time.Time) (*models.Session, error) {
	query := `INSERT INTO sessions (user_id, token_hash, device_name, ip_address, user_agent, absolute_expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, user_id, token_hash, device_name, ip_address, user_agent, last_used_at, absolute_expires_at, is_impersonation, impersonated_by, created_at, updated_at`
	row := r.db.QueryRow(ctx, query, userID, tokenHash, deviceName, ipAddress, userAgent, absoluteExpiresAt)

	s := &models.Session{}
	err := row.Scan(&s.ID, &s.UserID, &s.TokenHash, &s.DeviceName, &s.IPAddress, &s.UserAgent, &s.LastUsedAt, &s.AbsoluteExpiresAt, &s.IsImpersonation, &s.ImpersonatedBy, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *Repository) GetSessionByHash(ctx context.Context, tokenHash string) (*models.Session, error) {
	query := `SELECT id, user_id, token_hash, device_name, ip_address, user_agent, last_used_at, absolute_expires_at, is_impersonation, impersonated_by, created_at, updated_at FROM sessions WHERE token_hash = $1`
	row := r.db.QueryRow(ctx, query, tokenHash)

	s := &models.Session{}
	err := row.Scan(&s.ID, &s.UserID, &s.TokenHash, &s.DeviceName, &s.IPAddress, &s.UserAgent, &s.LastUsedAt, &s.AbsoluteExpiresAt, &s.IsImpersonation, &s.ImpersonatedBy, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *Repository) ListSessionsByUser(ctx context.Context, userID int64) ([]*models.Session, error) {
	query := `SELECT id, user_id, token_hash, device_name, ip_address, user_agent, last_used_at, absolute_expires_at, is_impersonation, impersonated_by, created_at, updated_at FROM sessions WHERE user_id = $1 ORDER BY last_used_at DESC`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessions := make([]*models.Session, 0)
	for rows.Next() {
		s := &models.Session{}
		err := rows.Scan(&s.ID, &s.UserID, &s.TokenHash, &s.DeviceName, &s.IPAddress, &s.UserAgent, &s.LastUsedAt, &s.AbsoluteExpiresAt, &s.IsImpersonation, &s.ImpersonatedBy, &s.CreatedAt, &s.UpdatedAt)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}

func (r *Repository) UpdateSessionLastUsed(ctx context.Context, sessionID int64) error {
	query := `UPDATE sessions SET last_used_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.db.Exec(ctx, query, sessionID)
	return err
}

func (r *Repository) DeleteSession(ctx context.Context, sessionID int64) error {
	query := `DELETE FROM sessions WHERE id = $1`
	_, err := r.db.Exec(ctx, query, sessionID)
	return err
}

func (r *Repository) DeleteSessionByHash(ctx context.Context, tokenHash string) error {
	query := `DELETE FROM sessions WHERE token_hash = $1`
	_, err := r.db.Exec(ctx, query, tokenHash)
	return err
}

func (r *Repository) DeleteAllUserSessions(ctx context.Context, userID int64) error {
	query := `DELETE FROM sessions WHERE user_id = $1`
	_, err := r.db.Exec(ctx, query, userID)
	return err
}

func (r *Repository) DeleteExpiredSessions(ctx context.Context) error {
	query := `DELETE FROM sessions WHERE absolute_expires_at < CURRENT_TIMESTAMP`
	_, err := r.db.Exec(ctx, query)
	return err
}

// ListAllActiveSessions returns all non-expired sessions across all users,
// joined with username from the users table for admin risk review.
func (r *Repository) ListAllActiveSessions(ctx context.Context) ([]*models.Session, error) {
	query := `SELECT s.id, s.user_id, COALESCE(u.username, ''), s.token_hash, s.device_name,
	                 s.ip_address, s.user_agent, s.last_used_at, s.absolute_expires_at,
	                 s.is_impersonation, s.impersonated_by, s.created_at, s.updated_at
	          FROM sessions s
	          LEFT JOIN users u ON u.id = s.user_id
	          WHERE s.absolute_expires_at > CURRENT_TIMESTAMP
	          ORDER BY s.last_used_at DESC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessions := make([]*models.Session, 0)
	for rows.Next() {
		s := &models.Session{}
		err := rows.Scan(&s.ID, &s.UserID, &s.Username, &s.TokenHash, &s.DeviceName,
			&s.IPAddress, &s.UserAgent, &s.LastUsedAt, &s.AbsoluteExpiresAt,
			&s.IsImpersonation, &s.ImpersonatedBy, &s.CreatedAt, &s.UpdatedAt)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}

// CreateImpersonationSession creates a session flagged as impersonation
func (r *Repository) CreateImpersonationSession(ctx context.Context, targetUserID, adminID int64, tokenHash, deviceName, ipAddress, userAgent string, absoluteExpiresAt time.Time) (*models.Session, error) {
	query := `INSERT INTO sessions (user_id, token_hash, device_name, ip_address, user_agent, absolute_expires_at, is_impersonation, impersonated_by)
		VALUES ($1, $2, $3, $4, $5, $6, TRUE, $7)
		RETURNING id, user_id, token_hash, device_name, ip_address, user_agent, last_used_at, absolute_expires_at, is_impersonation, impersonated_by, created_at, updated_at`
	row := r.db.QueryRow(ctx, query, targetUserID, tokenHash, deviceName, ipAddress, userAgent, absoluteExpiresAt, adminID)

	s := &models.Session{}
	err := row.Scan(&s.ID, &s.UserID, &s.TokenHash, &s.DeviceName, &s.IPAddress, &s.UserAgent, &s.LastUsedAt, &s.AbsoluteExpiresAt, &s.IsImpersonation, &s.ImpersonatedBy, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return s, nil
}
