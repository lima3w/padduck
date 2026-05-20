package repository

import (
	"context"
	"time"

	"padduck/models"
)

// User operations

func (r *Repository) CreateUser(ctx context.Context, username, email string) (*models.User, error) {
	query := `INSERT INTO users (username, email, role) VALUES ($1, $2, 'user') RETURNING id, username, email, password_hash, role, state, last_login_at, suspended_at, suspended_by, suspension_reason, privacy_accepted_at, privacy_accepted_version, deletion_requested_at, anonymized_at, created_at, updated_at`
	row := r.db.QueryRow(ctx, query, username, email)

	user := &models.User{}
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.State, &user.LastLoginAt, &user.SuspendedAt, &user.SuspendedBy, &user.SuspensionReason, &user.PrivacyAcceptedAt, &user.PrivacyAcceptedVersion, &user.DeletionRequestedAt, &user.AnonymizedAt, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *Repository) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	query := `SELECT id, username, email, password_hash, role, state, last_login_at, suspended_at, suspended_by, suspension_reason, privacy_accepted_at, privacy_accepted_version, deletion_requested_at, anonymized_at, created_at, updated_at FROM users WHERE id = $1`
	row := r.db.QueryRow(ctx, query, id)

	user := &models.User{}
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.State, &user.LastLoginAt, &user.SuspendedAt, &user.SuspendedBy, &user.SuspensionReason, &user.PrivacyAcceptedAt, &user.PrivacyAcceptedVersion, &user.DeletionRequestedAt, &user.AnonymizedAt, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *Repository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `SELECT id, username, email, password_hash, role, state, last_login_at, suspended_at, suspended_by, suspension_reason, privacy_accepted_at, privacy_accepted_version, deletion_requested_at, anonymized_at, created_at, updated_at FROM users WHERE username = $1`
	row := r.db.QueryRow(ctx, query, username)

	user := &models.User{}
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.State, &user.LastLoginAt, &user.SuspendedAt, &user.SuspendedBy, &user.SuspensionReason, &user.PrivacyAcceptedAt, &user.PrivacyAcceptedVersion, &user.DeletionRequestedAt, &user.AnonymizedAt, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *Repository) ListAllUsers(ctx context.Context) ([]*models.User, error) {
	query := `SELECT id, username, email, password_hash, role, state, last_login_at, suspended_at, suspended_by, suspension_reason, privacy_accepted_at, privacy_accepted_version, deletion_requested_at, anonymized_at, created_at, updated_at FROM users ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]*models.User, 0)
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.State, &user.LastLoginAt, &user.SuspendedAt, &user.SuspendedBy, &user.SuspensionReason, &user.PrivacyAcceptedAt, &user.PrivacyAcceptedVersion, &user.DeletionRequestedAt, &user.AnonymizedAt, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

// ListUsersPaginated returns a page of users with a total count.
func (r *Repository) ListUsersPaginated(ctx context.Context, limit, offset int) ([]*models.User, int64, error) {
	var total int64
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `SELECT id, username, email, password_hash, role, state, last_login_at, suspended_at, suspended_by, suspension_reason, privacy_accepted_at, privacy_accepted_version, deletion_requested_at, anonymized_at, created_at, updated_at FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	users := make([]*models.User, 0)
	for rows.Next() {
		user := &models.User{}
		if err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.State, &user.LastLoginAt, &user.SuspendedAt, &user.SuspendedBy, &user.SuspensionReason, &user.PrivacyAcceptedAt, &user.PrivacyAcceptedVersion, &user.DeletionRequestedAt, &user.AnonymizedAt, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, 0, err
		}
		users = append(users, user)
	}
	return users, total, rows.Err()
}

func (r *Repository) CreateUserWithPassword(ctx context.Context, username, email, passwordHash, role string) (*models.User, error) {
	query := `INSERT INTO users (username, email, password_hash, role) VALUES ($1, $2, $3, $4) RETURNING id, username, email, password_hash, role, state, last_login_at, suspended_at, suspended_by, suspension_reason, privacy_accepted_at, privacy_accepted_version, deletion_requested_at, anonymized_at, created_at, updated_at`
	row := r.db.QueryRow(ctx, query, username, email, passwordHash, role)

	user := &models.User{}
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.State, &user.LastLoginAt, &user.SuspendedAt, &user.SuspendedBy, &user.SuspensionReason, &user.PrivacyAcceptedAt, &user.PrivacyAcceptedVersion, &user.DeletionRequestedAt, &user.AnonymizedAt, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *Repository) CreateUserWithState(ctx context.Context, username, email, passwordHash, role, state string) (*models.User, error) {
	query := `INSERT INTO users (username, email, password_hash, role, state) VALUES ($1, $2, $3, $4, $5) RETURNING id, username, email, password_hash, role, state, last_login_at, suspended_at, suspended_by, suspension_reason, privacy_accepted_at, privacy_accepted_version, deletion_requested_at, anonymized_at, created_at, updated_at`
	row := r.db.QueryRow(ctx, query, username, email, passwordHash, role, state)

	user := &models.User{}
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.State, &user.LastLoginAt, &user.SuspendedAt, &user.SuspendedBy, &user.SuspensionReason, &user.PrivacyAcceptedAt, &user.PrivacyAcceptedVersion, &user.DeletionRequestedAt, &user.AnonymizedAt, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *Repository) UpdateUserState(ctx context.Context, userID int64, state string) error {
	query := `UPDATE users SET state = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.db.Exec(ctx, query, userID, state)
	return err
}

func (r *Repository) UpdateUserEmail(ctx context.Context, userID int64, email string) error {
	query := `UPDATE users SET email = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.db.Exec(ctx, query, userID, email)
	return err
}

func (r *Repository) UpdateUserRole(ctx context.Context, userID int64, role string) (*models.User, error) {
	query := `UPDATE users SET role = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $1 RETURNING id, username, email, password_hash, role, state, last_login_at, suspended_at, suspended_by, suspension_reason, privacy_accepted_at, privacy_accepted_version, deletion_requested_at, anonymized_at, created_at, updated_at`
	row := r.db.QueryRow(ctx, query, userID, role)

	user := &models.User{}
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.State, &user.LastLoginAt, &user.SuspendedAt, &user.SuspendedBy, &user.SuspensionReason, &user.PrivacyAcceptedAt, &user.PrivacyAcceptedVersion, &user.DeletionRequestedAt, &user.AnonymizedAt, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *Repository) DeleteUser(ctx context.Context, userID int64) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.Exec(ctx, query, userID)
	return err
}

func (r *Repository) UpdateLastLogin(ctx context.Context, userID int64) error {
	query := `UPDATE users SET last_login_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.db.Exec(ctx, query, userID)
	return err
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `SELECT id, username, email, password_hash, role, state, last_login_at, suspended_at, suspended_by, suspension_reason, privacy_accepted_at, privacy_accepted_version, deletion_requested_at, anonymized_at, created_at, updated_at FROM users WHERE email = $1`
	row := r.db.QueryRow(ctx, query, email)

	user := &models.User{}
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.State, &user.LastLoginAt, &user.SuspendedAt, &user.SuspendedBy, &user.SuspensionReason, &user.PrivacyAcceptedAt, &user.PrivacyAcceptedVersion, &user.DeletionRequestedAt, &user.AnonymizedAt, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *Repository) CreatePasswordReset(ctx context.Context, userID int64, tokenHash string) (*models.PasswordReset, error) {
	query := `INSERT INTO password_resets (user_id, token_hash, expires_at) VALUES ($1, $2, CURRENT_TIMESTAMP + INTERVAL '1 hour') RETURNING id, user_id, token_hash, expires_at, used_at, created_at, updated_at`
	row := r.db.QueryRow(ctx, query, userID, tokenHash)

	reset := &models.PasswordReset{}
	err := row.Scan(&reset.ID, &reset.UserID, &reset.TokenHash, &reset.ExpiresAt, &reset.UsedAt, &reset.CreatedAt, &reset.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return reset, nil
}

func (r *Repository) GetPasswordResetByToken(ctx context.Context, tokenHash string) (*models.PasswordReset, error) {
	query := `SELECT id, user_id, token_hash, expires_at, used_at, created_at, updated_at FROM password_resets WHERE token_hash = $1`
	row := r.db.QueryRow(ctx, query, tokenHash)

	reset := &models.PasswordReset{}
	err := row.Scan(&reset.ID, &reset.UserID, &reset.TokenHash, &reset.ExpiresAt, &reset.UsedAt, &reset.CreatedAt, &reset.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return reset, nil
}

func (r *Repository) MarkPasswordResetAsUsed(ctx context.Context, resetID int64) error {
	query := `UPDATE password_resets SET used_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.db.Exec(ctx, query, resetID)
	return err
}

func (r *Repository) UpdateUserPassword(ctx context.Context, userID int64, passwordHash string) error {
	query := `UPDATE users SET password_hash = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.db.Exec(ctx, query, userID, passwordHash)
	return err
}

// InitAdminPassword sets the admin password only when it is currently NULL (i.e. first boot).
// Returns true if the password was set, false if it was already set.
func (r *Repository) InitAdminPassword(ctx context.Context, passwordHash string) (bool, error) {
	query := `UPDATE users SET password_hash = $1, updated_at = CURRENT_TIMESTAMP WHERE username = 'admin' AND password_hash IS NULL`
	result, err := r.db.Exec(ctx, query, passwordHash)
	if err != nil {
		return false, err
	}
	return result.RowsAffected() > 0, nil
}

// ForceSetAdminPassword unconditionally updates the admin user's password hash.
func (r *Repository) ForceSetAdminPassword(ctx context.Context, passwordHash string) error {
	query := `UPDATE users SET password_hash = $1, updated_at = CURRENT_TIMESTAMP WHERE username = 'admin'`
	_, err := r.db.Exec(ctx, query, passwordHash)
	return err
}

// Email verification operations

func (r *Repository) CreateEmailVerification(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) (*models.EmailVerification, error) {
	query := `INSERT INTO email_verifications (user_id, token_hash, expires_at) VALUES ($1, $2, $3)
	          ON CONFLICT (token_hash) DO NOTHING
	          RETURNING id, user_id, token_hash, expires_at, used_at, created_at, updated_at`
	ev := &models.EmailVerification{}
	err := r.db.QueryRow(ctx, query, userID, tokenHash, expiresAt).Scan(
		&ev.ID, &ev.UserID, &ev.TokenHash, &ev.ExpiresAt, &ev.UsedAt, &ev.CreatedAt, &ev.UpdatedAt,
	)
	return ev, err
}

func (r *Repository) GetEmailVerificationByToken(ctx context.Context, tokenHash string) (*models.EmailVerification, error) {
	query := `SELECT id, user_id, token_hash, expires_at, used_at, created_at, updated_at FROM email_verifications WHERE token_hash = $1`
	ev := &models.EmailVerification{}
	err := r.db.QueryRow(ctx, query, tokenHash).Scan(
		&ev.ID, &ev.UserID, &ev.TokenHash, &ev.ExpiresAt, &ev.UsedAt, &ev.CreatedAt, &ev.UpdatedAt,
	)
	return ev, err
}

func (r *Repository) MarkEmailVerificationUsed(ctx context.Context, verificationID int64) error {
	query := `UPDATE email_verifications SET used_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.db.Exec(ctx, query, verificationID)
	return err
}

func (r *Repository) DeleteEmailVerificationsByUser(ctx context.Context, userID int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM email_verifications WHERE user_id = $1`, userID)
	return err
}

// User approval operations

func (r *Repository) CreateUserApproval(ctx context.Context, userID int64) (*models.UserApproval, error) {
	query := `INSERT INTO user_approvals (user_id) VALUES ($1) RETURNING id, user_id, status, reviewed_by, reviewed_at, rejection_reason, created_at, updated_at`
	ua := &models.UserApproval{}
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&ua.ID, &ua.UserID, &ua.Status, &ua.ReviewedBy, &ua.ReviewedAt, &ua.RejectionReason, &ua.CreatedAt, &ua.UpdatedAt,
	)
	return ua, err
}

func (r *Repository) GetUserApprovalByUserID(ctx context.Context, userID int64) (*models.UserApproval, error) {
	query := `SELECT id, user_id, status, reviewed_by, reviewed_at, rejection_reason, created_at, updated_at FROM user_approvals WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1`
	ua := &models.UserApproval{}
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&ua.ID, &ua.UserID, &ua.Status, &ua.ReviewedBy, &ua.ReviewedAt, &ua.RejectionReason, &ua.CreatedAt, &ua.UpdatedAt,
	)
	return ua, err
}

func (r *Repository) ListPendingApprovals(ctx context.Context) ([]*models.UserApproval, error) {
	query := `SELECT id, user_id, status, reviewed_by, reviewed_at, rejection_reason, created_at, updated_at FROM user_approvals WHERE status = 'pending' ORDER BY created_at ASC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	approvals := make([]*models.UserApproval, 0)
	for rows.Next() {
		ua := &models.UserApproval{}
		if err := rows.Scan(&ua.ID, &ua.UserID, &ua.Status, &ua.ReviewedBy, &ua.ReviewedAt, &ua.RejectionReason, &ua.CreatedAt, &ua.UpdatedAt); err != nil {
			return nil, err
		}
		approvals = append(approvals, ua)
	}
	return approvals, rows.Err()
}

func (r *Repository) UpdateUserApproval(ctx context.Context, approvalID int64, status string, reviewedBy int64, rejectionReason *string) error {
	query := `UPDATE user_approvals SET status = $2, reviewed_by = $3, reviewed_at = CURRENT_TIMESTAMP, rejection_reason = $4, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.db.Exec(ctx, query, approvalID, status, reviewedBy, rejectionReason)
	return err
}

func (r *Repository) GetUserApprovalByID(ctx context.Context, approvalID int64) (*models.UserApproval, error) {
	query := `SELECT id, user_id, status, reviewed_by, reviewed_at, rejection_reason, created_at, updated_at FROM user_approvals WHERE id = $1`
	ua := &models.UserApproval{}
	err := r.db.QueryRow(ctx, query, approvalID).Scan(
		&ua.ID, &ua.UserID, &ua.Status, &ua.ReviewedBy, &ua.ReviewedAt, &ua.RejectionReason, &ua.CreatedAt, &ua.UpdatedAt,
	)
	return ua, err
}

// SuspendUser sets a user's state to suspended with reason and admin tracking
func (r *Repository) SuspendUser(ctx context.Context, userID, adminID int64, reason string) error {
	query := `UPDATE users SET state = 'suspended', suspended_at = CURRENT_TIMESTAMP, suspended_by = $2, suspension_reason = $3, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.db.Exec(ctx, query, userID, adminID, reason)
	return err
}

// UnsuspendUser restores a user to active state
func (r *Repository) UnsuspendUser(ctx context.Context, userID int64) error {
	query := `UPDATE users SET state = 'active', suspended_at = NULL, suspended_by = NULL, suspension_reason = NULL, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.db.Exec(ctx, query, userID)
	return err
}

// BulkUpdateUserState updates the state of multiple users
func (r *Repository) BulkUpdateUserState(ctx context.Context, userIDs []int64, state string) (int64, error) {
	query := `UPDATE users SET state = $1, updated_at = CURRENT_TIMESTAMP WHERE id = ANY($2)`
	result, err := r.db.Exec(ctx, query, state, userIDs)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

// BulkDeleteUsers deletes multiple users
func (r *Repository) BulkDeleteUsers(ctx context.Context, userIDs []int64) (int64, error) {
	query := `DELETE FROM users WHERE id = ANY($1)`
	result, err := r.db.Exec(ctx, query, userIDs)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

// CountAdminsExcluding returns the number of active admin users whose IDs are not in the exclusion list.
func (r *Repository) CountAdminsExcluding(ctx context.Context, excludeIDs []int64) (int64, error) {
	query := `SELECT COUNT(*) FROM users WHERE role = 'admin' AND id != ALL($1)`
	var count int64
	err := r.db.QueryRow(ctx, query, excludeIDs).Scan(&count)
	return count, err
}

// UpdatePrivacyConsent records user acceptance of the privacy policy
func (r *Repository) UpdatePrivacyConsent(ctx context.Context, userID int64, version string) error {
	query := `UPDATE users SET privacy_accepted_at = CURRENT_TIMESTAMP, privacy_accepted_version = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.db.Exec(ctx, query, userID, version)
	return err
}

// RequestDeletion marks a user as having requested account deletion
func (r *Repository) RequestDeletion(ctx context.Context, userID int64) error {
	query := `UPDATE users SET deletion_requested_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.db.Exec(ctx, query, userID)
	return err
}

// AnonymizeUser replaces PII with anonymized values (GDPR right to erasure)
func (r *Repository) AnonymizeUser(ctx context.Context, userID int64) error {
	query := `UPDATE users SET
		username = 'deleted_' || id::text,
		email = 'deleted_' || id::text || '@deleted.invalid',
		password_hash = '',
		state = 'disabled',
		anonymized_at = CURRENT_TIMESTAMP,
		updated_at = CURRENT_TIMESTAMP
	WHERE id = $1`
	_, err := r.db.Exec(ctx, query, userID)
	return err
}

// GetUserAllData returns all data associated with a user for GDPR export
func (r *Repository) GetUserAllData(ctx context.Context, userID int64) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	// Get user record
	user, err := r.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	data["user"] = user

	// Get sessions
	sessions, err := r.ListSessionsByUser(ctx, userID)
	if err == nil {
		data["sessions"] = sessions
	}

	// Get API tokens
	tokens, err := r.ListAPITokensByUser(ctx, userID)
	if err == nil {
		data["api_tokens"] = tokens
	}

	// Get audit logs
	logs, err := r.ListAuditLogs(ctx, &models.AuditLogFilter{UserID: &userID, Limit: 1000})
	if err == nil {
		data["audit_logs"] = logs
	}

	return data, nil
}
