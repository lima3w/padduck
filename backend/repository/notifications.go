package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"ipam-next/models"
)

// --- Notification preferences ---

func (r *Repository) GetNotificationPreferences(ctx context.Context, userID int64) (*models.NotificationPreferences, error) {
	query := `SELECT id, user_id, login_success, login_failed, account_locked, password_changed,
	          mfa_changes, api_token_changes, role_changes, session_revoked, created_at, updated_at
	          FROM notification_preferences WHERE user_id = $1`
	p := &models.NotificationPreferences{}
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&p.ID, &p.UserID, &p.LoginSuccess, &p.LoginFailed, &p.AccountLocked, &p.PasswordChanged,
		&p.MFAChanges, &p.APITokenChanges, &p.RoleChanges, &p.SessionRevoked, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return &models.NotificationPreferences{
			UserID:          userID,
			LoginSuccess:    true,
			LoginFailed:     true,
			AccountLocked:   true,
			PasswordChanged: true,
			MFAChanges:      true,
			APITokenChanges: true,
			RoleChanges:     true,
			SessionRevoked:  true,
		}, nil
	}
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *Repository) UpsertNotificationPreferences(ctx context.Context, prefs *models.NotificationPreferences) (*models.NotificationPreferences, error) {
	query := `INSERT INTO notification_preferences
	          (user_id, login_success, login_failed, account_locked, password_changed,
	           mfa_changes, api_token_changes, role_changes, session_revoked)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	          ON CONFLICT (user_id) DO UPDATE SET
	              login_success    = EXCLUDED.login_success,
	              login_failed     = EXCLUDED.login_failed,
	              account_locked   = EXCLUDED.account_locked,
	              password_changed = EXCLUDED.password_changed,
	              mfa_changes      = EXCLUDED.mfa_changes,
	              api_token_changes = EXCLUDED.api_token_changes,
	              role_changes     = EXCLUDED.role_changes,
	              session_revoked  = EXCLUDED.session_revoked,
	              updated_at       = CURRENT_TIMESTAMP
	          RETURNING id, user_id, login_success, login_failed, account_locked, password_changed,
	                    mfa_changes, api_token_changes, role_changes, session_revoked, created_at, updated_at`
	p := &models.NotificationPreferences{}
	err := r.db.QueryRow(ctx, query,
		prefs.UserID, prefs.LoginSuccess, prefs.LoginFailed, prefs.AccountLocked, prefs.PasswordChanged,
		prefs.MFAChanges, prefs.APITokenChanges, prefs.RoleChanges, prefs.SessionRevoked,
	).Scan(
		&p.ID, &p.UserID, &p.LoginSuccess, &p.LoginFailed, &p.AccountLocked, &p.PasswordChanged,
		&p.MFAChanges, &p.APITokenChanges, &p.RoleChanges, &p.SessionRevoked, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// --- Notification queue ---

func (r *Repository) CreateNotificationQueueItem(ctx context.Context, userID int64, email, template, dataJSON string) (*models.NotificationQueue, error) {
	query := `INSERT INTO notification_queue (user_id, email, template, data)
	          VALUES ($1, $2, $3, $4::jsonb)
	          RETURNING id, user_id, email, template, data::text, status, retry_count,
	                    next_retry_at, sent_at, error_msg, created_at, updated_at`
	q := &models.NotificationQueue{}
	err := r.db.QueryRow(ctx, query, userID, email, template, dataJSON).Scan(
		&q.ID, &q.UserID, &q.Email, &q.Template, &q.Data, &q.Status, &q.RetryCount,
		&q.NextRetryAt, &q.SentAt, &q.ErrorMsg, &q.CreatedAt, &q.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return q, nil
}

func (r *Repository) GetPendingNotifications(ctx context.Context, limit int) ([]*models.NotificationQueue, error) {
	query := `SELECT id, user_id, email, template, data::text, status, retry_count,
	          next_retry_at, sent_at, error_msg, created_at, updated_at
	          FROM notification_queue
	          WHERE status IN ('pending', 'retrying')
	            AND (next_retry_at IS NULL OR next_retry_at <= NOW())
	          ORDER BY created_at ASC
	          LIMIT $1`
	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]*models.NotificationQueue, 0)
	for rows.Next() {
		q := &models.NotificationQueue{}
		if err := rows.Scan(
			&q.ID, &q.UserID, &q.Email, &q.Template, &q.Data, &q.Status, &q.RetryCount,
			&q.NextRetryAt, &q.SentAt, &q.ErrorMsg, &q.CreatedAt, &q.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, q)
	}
	return items, rows.Err()
}

func (r *Repository) MarkNotificationSent(ctx context.Context, id int64) error {
	query := `UPDATE notification_queue SET status = 'sent', sent_at = NOW(), updated_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *Repository) MarkNotificationFailed(ctx context.Context, id int64, errMsg string, retryCount int, nextRetryAt *time.Time) error {
	status := "failed"
	if nextRetryAt != nil {
		status = "retrying"
	}
	query := `UPDATE notification_queue
	          SET status = $2, error_msg = $3, retry_count = $4, next_retry_at = $5, updated_at = NOW()
	          WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id, status, errMsg, retryCount, nextRetryAt)
	return err
}

func (r *Repository) CountRecentNotificationsSent(ctx context.Context, userID int64, since time.Time) (int64, error) {
	query := `SELECT COUNT(*) FROM notification_queue WHERE user_id = $1 AND sent_at >= $2`
	var count int64
	err := r.db.QueryRow(ctx, query, userID, since).Scan(&count)
	return count, err
}

func (r *Repository) GetNotificationStats(ctx context.Context) (map[string]int64, error) {
	query := `SELECT status, COUNT(*) FROM notification_queue GROUP BY status`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := map[string]int64{
		"pending":  0,
		"sent":     0,
		"failed":   0,
		"retrying": 0,
	}
	for rows.Next() {
		var status string
		var count int64
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		stats[status] = count
	}
	return stats, rows.Err()
}

func (r *Repository) CleanupOldNotifications(ctx context.Context) error {
	query := `DELETE FROM notification_queue
	          WHERE (status = 'sent'   AND sent_at    < NOW() - INTERVAL '30 days')
	             OR (status = 'failed' AND updated_at < NOW() - INTERVAL '7 days')`
	_, err := r.db.Exec(ctx, query)
	return err
}
