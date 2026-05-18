package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"ipam-next/models"
)

var (
	ErrAccountLocked   = errors.New("account is temporarily locked due to too many failed login attempts")
	ErrInvalidUnlockToken = errors.New("unlock token is invalid or expired")
)

const (
	maxFailedAttempts   = 5
	bruteForceWindow    = 15 * time.Minute
	notifRateLimit      = 1 * time.Hour
	ipFailureThreshold  = 3 // send notification after this many failures from same IP
)

// lockoutDuration returns the lockout duration based on how many times the account has been locked.
// Durations escalate exponentially to deter persistent attackers.
func lockoutDuration(lockoutCount int) time.Duration {
	switch {
	case lockoutCount <= 1:
		return 5 * time.Minute
	case lockoutCount == 2:
		return 15 * time.Minute
	case lockoutCount == 3:
		return 1 * time.Hour
	case lockoutCount == 4:
		return 4 * time.Hour
	case lockoutCount == 5:
		return 24 * time.Hour
	default:
		return 7 * 24 * time.Hour // 7 days after 6+ lockouts
	}
}

// ProcessFailedLogin records a failed login attempt and triggers lockout/notifications if thresholds are met.
// userID may be 0 if the username doesn't exist.
func (s *Service) ProcessFailedLogin(ctx context.Context, userID int64, username, ipAddress, userAgent, failureReason string) {
	// Record the attempt (best-effort; don't surface errors to the caller)
	_ = s.repository.CreateLoginAttempt(ctx, username, ipAddress, userAgent, false, failureReason)

	if userID == 0 {
		// Unknown username — no lockout or notification possible
		return
	}

	// Check if brute force threshold reached
	since := time.Now().Add(-bruteForceWindow)
	count, err := s.repository.CountRecentFailedAttemptsByUsername(ctx, username, since)
	if err != nil {
		return
	}

	if count >= maxFailedAttempts {
		_ = s.lockAccount(ctx, userID, username, count)
	} else if count >= ipFailureThreshold {
		// Send a heads-up after 3 failures
		_ = s.sendFailedLoginAlert(ctx, userID, username, ipAddress, count)
	}
}

const (
	ipThrottleThreshold = 20             // max failed attempts per IP within the window
	ipThrottleWindow    = 15 * time.Minute
)

// IsIPThrottled returns true if the given IP address has accumulated too many
// failed login attempts across any usernames within the throttle window.
func (s *Service) IsIPThrottled(ctx context.Context, ipAddress string) (bool, error) {
	since := time.Now().Add(-ipThrottleWindow)
	count, err := s.repository.CountRecentFailedAttemptsByIPOnly(ctx, ipAddress, since)
	if err != nil {
		return false, err
	}
	return count >= ipThrottleThreshold, nil
}

// lockAccount creates a lockout record and sends a notification email.
func (s *Service) lockAccount(ctx context.Context, userID int64, username string, failCount int) error {
	lockoutCount, _ := s.repository.CountUserLockouts(ctx, userID)
	lockoutCount++ // this will be the new lockout
	duration := lockoutDuration(lockoutCount)
	unlockAt := time.Now().Add(duration)

	reason := fmt.Sprintf("%d failed login attempts within %s", failCount, bruteForceWindow)
	lockout, err := s.repository.CreateAccountLockout(ctx, userID, unlockAt, reason, lockoutCount)
	if err != nil {
		return err
	}

	// Generate unlock token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return err
	}
	rawToken := hex.EncodeToString(tokenBytes)
	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])
	tokenExpiry := time.Now().Add(24 * time.Hour)

	if err := s.repository.SetUnlockToken(ctx, lockout.ID, tokenHash, tokenExpiry); err != nil {
		return err
	}

	// Queue lockout notification (best-effort)
	appURL, _ := s.Config.GetCtx(ctx, "app_url")
	if appURL == "" {
		appURL = "http://localhost:3000"
	}
	unlockURL := fmt.Sprintf("%s/unlock-account?token=%s", appURL, rawToken)
	_ = s.Notification.Queue(ctx, userID, NotifAccountLocked, map[string]interface{}{
		"UnlockURL": unlockURL,
		"Duration":  duration.String(),
	})
	_ = s.repository.CreateSecurityNotification(ctx, userID, "account_locked", "")

	return nil
}

// sendFailedLoginAlert sends a notification email about multiple failed login attempts (rate-limited).
func (s *Service) sendFailedLoginAlert(ctx context.Context, userID int64, username, ipAddress string, count int) error {
	since := time.Now().Add(-notifRateLimit)
	recent, err := s.repository.CountRecentSecurityNotifications(ctx, userID, "failed_login_alert", since)
	if err != nil || recent > 0 {
		return err
	}

	if err := s.Notification.Queue(ctx, userID, NotifLoginFailed, map[string]interface{}{
		"IP":    ipAddress,
		"Count": count,
	}); err != nil {
		return err
	}

	return s.repository.CreateSecurityNotification(ctx, userID, "failed_login_alert", ipAddress)
}

// IsAccountLocked returns true and the lockout details if the account is currently locked.
func (s *Service) IsAccountLocked(ctx context.Context, userID int64) (bool, *models.AccountLockout, error) {
	lockout, err := s.repository.GetActiveAccountLockout(ctx, userID)
	if err != nil {
		// pgx returns pgx.ErrNoRows when no rows found; treat as not locked
		return false, nil, nil
	}
	return true, lockout, nil
}

// RecordSuccessfulLogin records a successful login attempt.
func (s *Service) RecordSuccessfulLogin(ctx context.Context, username, ipAddress, userAgent string) {
	_ = s.repository.CreateLoginAttempt(ctx, username, ipAddress, userAgent, true, "")
}

// GetLoginHistory returns the last N login attempts for a user.
func (s *Service) GetLoginHistory(ctx context.Context, userID int64, limit int) ([]*models.LoginAttempt, error) {
	user, err := s.repository.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.repository.GetLoginHistory(ctx, user.Username, limit)
}

// RequestUnlockEmail sends an unlock email for a locked account (identified by username).
func (s *Service) RequestUnlockEmail(ctx context.Context, username string) error {
	user, err := s.repository.GetUserByUsername(ctx, username)
	if err != nil {
		// Don't reveal whether username exists
		return nil
	}

	locked, lockout, err := s.IsAccountLocked(ctx, user.ID)
	if err != nil || !locked {
		return nil
	}

	// Rate-limit: don't send if a token already exists and hasn't expired
	if lockout.UnlockTokenHash != nil && lockout.UnlockTokenExpiresAt != nil && lockout.UnlockTokenExpiresAt.After(time.Now()) {
		return nil
	}

	// Generate new unlock token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return err
	}
	rawToken := hex.EncodeToString(tokenBytes)
	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])
	tokenExpiry := time.Now().Add(24 * time.Hour)

	if err := s.repository.SetUnlockToken(ctx, lockout.ID, tokenHash, tokenExpiry); err != nil {
		return err
	}

	appURL, _ := s.Config.GetCtx(ctx, "app_url")
	if appURL == "" {
		appURL = "http://localhost:3000"
	}
	unlockURL := fmt.Sprintf("%s/unlock-account?token=%s", appURL, rawToken)
	duration := time.Until(lockout.UnlockAt)
	_ = s.Email.SendAccountLockedEmail(user.Email, user.Username, unlockURL, duration)
	return nil
}

// UnlockAccountByToken verifies an unlock token and unlocks the account.
func (s *Service) UnlockAccountByToken(ctx context.Context, rawToken string) error {
	if rawToken == "" {
		return ErrInvalidUnlockToken
	}

	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])

	lockout, err := s.repository.GetLockoutByUnlockToken(ctx, tokenHash)
	if err != nil {
		return ErrInvalidUnlockToken
	}

	if lockout.UnlockTokenUsedAt != nil {
		return ErrInvalidUnlockToken
	}

	if lockout.UnlockTokenExpiresAt != nil && lockout.UnlockTokenExpiresAt.Before(time.Now()) {
		return ErrInvalidUnlockToken
	}

	if lockout.UnlockedAt != nil {
		// Already unlocked
		return nil
	}

	if err := s.repository.UnlockAccount(ctx, lockout.ID, nil); err != nil {
		return err
	}

	return s.repository.MarkUnlockTokenUsed(ctx, lockout.ID)
}

// UnlockAccountByAdmin unlocks an account on behalf of an admin.
func (s *Service) UnlockAccountByAdmin(ctx context.Context, userID, adminID int64) error {
	locked, lockout, err := s.IsAccountLocked(ctx, userID)
	if err != nil {
		return err
	}
	if !locked {
		return nil
	}
	return s.repository.UnlockAccount(ctx, lockout.ID, &adminID)
}
