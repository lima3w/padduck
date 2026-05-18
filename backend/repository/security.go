package repository

import (
	"context"
	"time"

	"ipam-next/models"
)

// Login attempt operations

func (r *Repository) CreateLoginAttempt(ctx context.Context, username, ipAddress, userAgent string, success bool, failureReason string) error {
	query := `INSERT INTO login_attempts (username, ip_address, user_agent, success, failure_reason) VALUES ($1, $2::inet, $3, $4, $5)`
	_, err := r.db.Exec(ctx, query, username, nullableString(ipAddress), userAgent, success, nullableString(failureReason))
	return err
}

func (r *Repository) CountRecentFailedAttemptsByUsername(ctx context.Context, username string, since time.Time) (int, error) {
	query := `SELECT COUNT(*) FROM login_attempts WHERE username = $1 AND success = false AND created_at >= $2`
	var count int
	err := r.db.QueryRow(ctx, query, username, since).Scan(&count)
	return count, err
}

func (r *Repository) CountRecentFailedAttemptsByIP(ctx context.Context, username, ipAddress string, since time.Time) (int, error) {
	query := `SELECT COUNT(*) FROM login_attempts WHERE username = $1 AND ip_address = $2::inet AND success = false AND created_at >= $3`
	var count int
	err := r.db.QueryRow(ctx, query, username, ipAddress, since).Scan(&count)
	return count, err
}

// CountRecentFailedAttemptsByIPOnly counts failed login attempts from a specific IP address across
// all usernames within the given time window. Used for per-IP rate limiting.
func (r *Repository) CountRecentFailedAttemptsByIPOnly(ctx context.Context, ipAddress string, since time.Time) (int, error) {
	query := `SELECT COUNT(*) FROM login_attempts WHERE ip_address = $1::inet AND success = false AND created_at >= $2`
	var count int
	err := r.db.QueryRow(ctx, query, ipAddress, since).Scan(&count)
	return count, err
}

func (r *Repository) GetLoginHistory(ctx context.Context, username string, limit int) ([]*models.LoginAttempt, error) {
	query := `SELECT id, username, COALESCE(ip_address::text, ''), COALESCE(user_agent, ''), success, COALESCE(failure_reason, ''), created_at
	          FROM login_attempts WHERE username = $1 ORDER BY created_at DESC LIMIT $2`
	rows, err := r.db.Query(ctx, query, username, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	attempts := make([]*models.LoginAttempt, 0)
	for rows.Next() {
		a := &models.LoginAttempt{}
		if err := rows.Scan(&a.ID, &a.Username, &a.IPAddress, &a.UserAgent, &a.Success, &a.FailureReason, &a.CreatedAt); err != nil {
			return nil, err
		}
		attempts = append(attempts, a)
	}
	return attempts, rows.Err()
}

// Account lockout operations

func (r *Repository) CreateAccountLockout(ctx context.Context, userID int64, unlockAt time.Time, reason string, lockoutCount int) (*models.AccountLockout, error) {
	query := `INSERT INTO account_lockouts (user_id, unlock_at, reason, lockout_count)
	          VALUES ($1, $2, $3, $4)
	          RETURNING id, user_id, locked_at, unlock_at, unlock_token_hash, unlock_token_expires_at, unlock_token_used_at, reason, lockout_count, unlocked_at, unlocked_by, created_at`
	lo := &models.AccountLockout{}
	err := r.db.QueryRow(ctx, query, userID, unlockAt, reason, lockoutCount).Scan(
		&lo.ID, &lo.UserID, &lo.LockedAt, &lo.UnlockAt,
		&lo.UnlockTokenHash, &lo.UnlockTokenExpiresAt, &lo.UnlockTokenUsedAt,
		&lo.Reason, &lo.LockoutCount, &lo.UnlockedAt, &lo.UnlockedBy, &lo.CreatedAt,
	)
	return lo, err
}

func (r *Repository) GetActiveAccountLockout(ctx context.Context, userID int64) (*models.AccountLockout, error) {
	query := `SELECT id, user_id, locked_at, unlock_at, unlock_token_hash, unlock_token_expires_at, unlock_token_used_at, reason, lockout_count, unlocked_at, unlocked_by, created_at
	          FROM account_lockouts
	          WHERE user_id = $1 AND unlocked_at IS NULL AND unlock_at > NOW()
	          ORDER BY created_at DESC LIMIT 1`
	lo := &models.AccountLockout{}
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&lo.ID, &lo.UserID, &lo.LockedAt, &lo.UnlockAt,
		&lo.UnlockTokenHash, &lo.UnlockTokenExpiresAt, &lo.UnlockTokenUsedAt,
		&lo.Reason, &lo.LockoutCount, &lo.UnlockedAt, &lo.UnlockedBy, &lo.CreatedAt,
	)
	return lo, err
}

func (r *Repository) CountUserLockouts(ctx context.Context, userID int64) (int, error) {
	query := `SELECT COUNT(*) FROM account_lockouts WHERE user_id = $1`
	var count int
	err := r.db.QueryRow(ctx, query, userID).Scan(&count)
	return count, err
}

func (r *Repository) UnlockAccount(ctx context.Context, lockoutID int64, unlockedBy *int64) error {
	_, err := r.db.Exec(ctx, `UPDATE account_lockouts SET unlocked_at = NOW(), unlocked_by = $2 WHERE id = $1`, lockoutID, unlockedBy)
	return err
}

func (r *Repository) SetUnlockToken(ctx context.Context, lockoutID int64, tokenHash string, expiresAt time.Time) error {
	_, err := r.db.Exec(ctx, `UPDATE account_lockouts SET unlock_token_hash = $2, unlock_token_expires_at = $3 WHERE id = $1`, lockoutID, tokenHash, expiresAt)
	return err
}

func (r *Repository) GetLockoutByUnlockToken(ctx context.Context, tokenHash string) (*models.AccountLockout, error) {
	query := `SELECT id, user_id, locked_at, unlock_at, unlock_token_hash, unlock_token_expires_at, unlock_token_used_at, reason, lockout_count, unlocked_at, unlocked_by, created_at
	          FROM account_lockouts WHERE unlock_token_hash = $1`
	lo := &models.AccountLockout{}
	err := r.db.QueryRow(ctx, query, tokenHash).Scan(
		&lo.ID, &lo.UserID, &lo.LockedAt, &lo.UnlockAt,
		&lo.UnlockTokenHash, &lo.UnlockTokenExpiresAt, &lo.UnlockTokenUsedAt,
		&lo.Reason, &lo.LockoutCount, &lo.UnlockedAt, &lo.UnlockedBy, &lo.CreatedAt,
	)
	return lo, err
}

func (r *Repository) MarkUnlockTokenUsed(ctx context.Context, lockoutID int64) error {
	_, err := r.db.Exec(ctx, `UPDATE account_lockouts SET unlock_token_used_at = NOW() WHERE id = $1`, lockoutID)
	return err
}

// Security notification operations

func (r *Repository) CreateSecurityNotification(ctx context.Context, userID int64, notifType, ipAddress string) error {
	query := `INSERT INTO security_notifications (user_id, notification_type, ip_address) VALUES ($1, $2, $3::inet)`
	_, err := r.db.Exec(ctx, query, userID, notifType, nullableString(ipAddress))
	return err
}

func (r *Repository) CountRecentSecurityNotifications(ctx context.Context, userID int64, notifType string, since time.Time) (int, error) {
	query := `SELECT COUNT(*) FROM security_notifications WHERE user_id = $1 AND notification_type = $2 AND sent_at >= $3`
	var count int
	err := r.db.QueryRow(ctx, query, userID, notifType, since).Scan(&count)
	return count, err
}

func nullableString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
