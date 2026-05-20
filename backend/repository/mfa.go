package repository

import (
	"context"
	"time"

	"padduck/models"
)

// MFA settings operations

func (r *Repository) GetMFASettings(ctx context.Context, userID int64) (*models.UserMFASettings, error) {
	query := `SELECT id, user_id, totp_enabled, backup_codes_generated_at, created_at, updated_at FROM user_mfa_settings WHERE user_id = $1`
	s := &models.UserMFASettings{}
	err := r.db.QueryRow(ctx, query, userID).Scan(&s.ID, &s.UserID, &s.TOTPEnabled, &s.BackupCodesGeneratedAt, &s.CreatedAt, &s.UpdatedAt)
	return s, err
}

func (r *Repository) UpsertMFASettings(ctx context.Context, userID int64, totpEnabled bool, backupCodesAt *time.Time) error {
	query := `INSERT INTO user_mfa_settings (user_id, totp_enabled, backup_codes_generated_at)
	          VALUES ($1, $2, $3)
	          ON CONFLICT (user_id) DO UPDATE SET totp_enabled = $2, backup_codes_generated_at = $3, updated_at = CURRENT_TIMESTAMP`
	_, err := r.db.Exec(ctx, query, userID, totpEnabled, backupCodesAt)
	return err
}

// TOTP secret operations

func (r *Repository) UpsertTOTPSecret(ctx context.Context, userID int64, encryptedSecret []byte) error {
	query := `INSERT INTO user_totp_secrets (user_id, encrypted_secret, verified)
	          VALUES ($1, $2, FALSE)
	          ON CONFLICT (user_id) DO UPDATE SET encrypted_secret = $2, verified = FALSE, updated_at = CURRENT_TIMESTAMP`
	_, err := r.db.Exec(ctx, query, userID, encryptedSecret)
	return err
}

func (r *Repository) GetTOTPSecret(ctx context.Context, userID int64) (*models.UserTOTPSecret, error) {
	query := `SELECT id, user_id, encrypted_secret, verified, created_at, updated_at FROM user_totp_secrets WHERE user_id = $1`
	s := &models.UserTOTPSecret{}
	err := r.db.QueryRow(ctx, query, userID).Scan(&s.ID, &s.UserID, &s.EncryptedSecret, &s.Verified, &s.CreatedAt, &s.UpdatedAt)
	return s, err
}

func (r *Repository) MarkTOTPVerified(ctx context.Context, userID int64) error {
	_, err := r.db.Exec(ctx, `UPDATE user_totp_secrets SET verified = TRUE, updated_at = CURRENT_TIMESTAMP WHERE user_id = $1`, userID)
	return err
}

func (r *Repository) DeleteTOTPSecret(ctx context.Context, userID int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM user_totp_secrets WHERE user_id = $1`, userID)
	return err
}

// Backup code operations

func (r *Repository) CreateBackupCodes(ctx context.Context, userID int64, hashes []string) error {
	// Delete existing codes first
	if _, err := r.db.Exec(ctx, `DELETE FROM user_backup_codes WHERE user_id = $1`, userID); err != nil {
		return err
	}
	for _, h := range hashes {
		if _, err := r.db.Exec(ctx, `INSERT INTO user_backup_codes (user_id, code_hash) VALUES ($1, $2)`, userID, h); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) ListBackupCodes(ctx context.Context, userID int64) ([]*models.UserBackupCode, error) {
	rows, err := r.db.Query(ctx, `SELECT id, user_id, code_hash, used, used_at, created_at FROM user_backup_codes WHERE user_id = $1 ORDER BY id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var codes []*models.UserBackupCode
	for rows.Next() {
		c := &models.UserBackupCode{}
		if err := rows.Scan(&c.ID, &c.UserID, &c.CodeHash, &c.Used, &c.UsedAt, &c.CreatedAt); err != nil {
			return nil, err
		}
		codes = append(codes, c)
	}
	return codes, rows.Err()
}

func (r *Repository) MarkBackupCodeUsed(ctx context.Context, codeID int64) error {
	_, err := r.db.Exec(ctx, `UPDATE user_backup_codes SET used = TRUE, used_at = CURRENT_TIMESTAMP WHERE id = $1`, codeID)
	return err
}

// MFA challenge operations

func (r *Repository) CreateMFAChallenge(ctx context.Context, userID int64, challengeHash string, expiresAt time.Time) (*models.MFAChallenge, error) {
	query := `INSERT INTO mfa_challenges (user_id, challenge_hash, expires_at) VALUES ($1, $2, $3) RETURNING id, user_id, challenge_hash, expires_at, completed_at, created_at`
	c := &models.MFAChallenge{}
	err := r.db.QueryRow(ctx, query, userID, challengeHash, expiresAt).Scan(&c.ID, &c.UserID, &c.ChallengeHash, &c.ExpiresAt, &c.CompletedAt, &c.CreatedAt)
	return c, err
}

func (r *Repository) GetMFAChallenge(ctx context.Context, challengeHash string) (*models.MFAChallenge, error) {
	query := `SELECT id, user_id, challenge_hash, expires_at, completed_at, created_at FROM mfa_challenges WHERE challenge_hash = $1`
	c := &models.MFAChallenge{}
	err := r.db.QueryRow(ctx, query, challengeHash).Scan(&c.ID, &c.UserID, &c.ChallengeHash, &c.ExpiresAt, &c.CompletedAt, &c.CreatedAt)
	return c, err
}

func (r *Repository) CompleteMFAChallenge(ctx context.Context, challengeID int64) error {
	_, err := r.db.Exec(ctx, `UPDATE mfa_challenges SET completed_at = CURRENT_TIMESTAMP WHERE id = $1`, challengeID)
	return err
}
