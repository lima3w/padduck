package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"ipam-next/models"
	"ipam-next/utils"
)

const (
	TokenLength = 32
)

// GenerateAPIToken creates a new API token for a user.
// scope must be one of "read", "write", "admin" (defaults to "write" if empty).
// expiresInDays == 0 means read from config; if config is 0 or missing, no expiry.
func (s *Service) GenerateAPIToken(ctx context.Context, userID int64, tokenName, scope string, expiresInDays int) (string, error) {
	if userID <= 0 {
		return "", fmt.Errorf("invalid user ID")
	}
	if tokenName == "" {
		return "", fmt.Errorf("token name is required")
	}

	// Validate / default scope
	switch scope {
	case "read", "write", "admin":
		// valid
	default:
		scope = "write"
	}

	// Determine expiry
	var expiresAt *time.Time
	if expiresInDays == 0 {
		defaultDaysStr, _ := s.Config.Get("api_token_default_expiration_days")
		if n, err := strconv.Atoi(defaultDaysStr); err == nil && n > 0 {
			t := time.Now().Add(time.Duration(n) * 24 * time.Hour)
			expiresAt = &t
		}
		// else: no expiry
	} else if expiresInDays > 0 {
		t := time.Now().Add(time.Duration(expiresInDays) * 24 * time.Hour)
		expiresAt = &t
	}

	// Generate random token
	tokenBytes := make([]byte, TokenLength)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	token := hex.EncodeToString(tokenBytes)

	// Hash token for storage
	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])

	// Store token hash in database
	if _, err := s.repository.CreateAPITokenFull(ctx, userID, tokenHash, tokenName, scope, expiresAt); err != nil {
		return "", err
	}

	return token, nil
}

// ValidateAPIToken checks if a token is valid and returns the user and the token record.
// ip is recorded as the last-used IP address.
func (s *Service) ValidateAPIToken(ctx context.Context, token, ip string) (*models.User, *models.APIToken, error) {
	if token == "" {
		return nil, nil, fmt.Errorf("token is required")
	}

	// Hash the provided token
	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])

	// Look up token in database
	apiToken, err := s.repository.GetAPITokenByHash(ctx, tokenHash)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid token")
	}

	// Check if token has expired
	if apiToken.ExpiresAt != nil && apiToken.ExpiresAt.Before(time.Now()) {
		return nil, nil, fmt.Errorf("token has expired")
	}

	// Check rotation grace period
	if apiToken.RotationGraceExpiresAt != nil && apiToken.RotationGraceExpiresAt.Before(time.Now()) {
		return nil, nil, fmt.Errorf("token has been rotated and grace period has expired")
	}

	// Update last used timestamp and IP
	_ = s.repository.UpdateAPITokenLastUsed(ctx, apiToken.ID, ip)

	// Get user information
	user, err := s.repository.GetUserByID(ctx, apiToken.UserID)
	if err != nil {
		return nil, nil, fmt.Errorf("user not found")
	}

	return user, apiToken, nil
}

// RotateAPIToken marks the old token as rotated and creates a new one with the same name/scope.
// Returns the new raw token and when the old token's grace period expires.
func (s *Service) RotateAPIToken(ctx context.Context, tokenID, userID int64) (newToken string, graceExpiresAt time.Time, err error) {
	// Fetch old token
	oldToken, err := s.repository.GetAPITokenByID(ctx, tokenID)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("token not found")
	}
	if oldToken.UserID != userID {
		return "", time.Time{}, fmt.Errorf("token does not belong to this user")
	}

	// Read grace period from config (default 24h)
	gracePeriod := 24 * time.Hour
	if graceHoursStr, err2 := s.Config.Get("api_token_rotation_grace_period_hours"); err2 == nil {
		if n, err3 := strconv.Atoi(graceHoursStr); err3 == nil && n > 0 {
			gracePeriod = time.Duration(n) * time.Hour
		}
	}

	graceExpiresAt = time.Now().Add(gracePeriod)

	// Mark old token as rotated
	if err = s.repository.MarkAPITokenRotated(ctx, tokenID, graceExpiresAt); err != nil {
		return "", time.Time{}, fmt.Errorf("failed to rotate token: %w", err)
	}

	// Determine new token expiry: preserve remaining time if old token had expiry
	var expiresInDays int
	if oldToken.ExpiresAt != nil {
		remaining := time.Until(*oldToken.ExpiresAt)
		if remaining < 24*time.Hour {
			remaining = 24 * time.Hour
		}
		expiresInDays = int(remaining.Hours() / 24)
	}

	// Create new token with same name and scope
	newRawToken, err := s.GenerateAPIToken(ctx, userID, oldToken.Name, oldToken.Scope, expiresInDays)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to create replacement token: %w", err)
	}

	return newRawToken, graceExpiresAt, nil
}

// ExtendAPIToken extends the expiry of an existing token.
// days == 0 means read from config default.
func (s *Service) ExtendAPIToken(ctx context.Context, tokenID, userID int64, days int) (*models.APIToken, error) {
	// Fetch token and verify ownership
	existing, err := s.repository.GetAPITokenByID(ctx, tokenID)
	if err != nil {
		return nil, fmt.Errorf("token not found")
	}
	if existing.UserID != userID {
		return nil, fmt.Errorf("token does not belong to this user")
	}

	if days == 0 {
		defaultDaysStr, _ := s.Config.Get("api_token_default_expiration_days")
		if n, err2 := strconv.Atoi(defaultDaysStr); err2 == nil && n > 0 {
			days = n
		} else {
			days = 30
		}
	}

	newExpiresAt := time.Now().Add(time.Duration(days) * 24 * time.Hour)
	return s.repository.ExtendAPIToken(ctx, tokenID, userID, newExpiresAt)
}

// CleanupExpiredTokens removes tokens that have been expired for more than 30 days.
func (s *Service) CleanupExpiredTokens(ctx context.Context) error {
	return s.repository.DeleteExpiredAPITokens(ctx)
}

// RevokeAPIToken deletes a token
func (s *Service) RevokeAPIToken(ctx context.Context, tokenID int64) error {
	if tokenID <= 0 {
		return fmt.Errorf("invalid token ID")
	}

	return s.repository.DeleteAPIToken(ctx, tokenID)
}

// ListUserTokens returns all tokens for a user
func (s *Service) ListUserTokens(ctx context.Context, userID int64) ([]*models.APIToken, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("invalid user ID")
	}

	return s.repository.ListAPITokensByUser(ctx, userID)
}

// AuthResult is returned by AuthenticateUser; either the full user is set or MFAChallenge is set.
type AuthResult struct {
	User         *models.User
	MFARequired  bool
	MFAChallenge string // raw challenge token, only set when MFARequired is true
}

// AuthenticateUser verifies username and password.
// ipAddress and userAgent are recorded for audit and brute-force detection.
// When MFA is enabled, it returns an MFAChallenge instead of the full user.
func (s *Service) AuthenticateUser(ctx context.Context, username, password, ipAddress, userAgent string) (*AuthResult, error) {
	if username == "" || password == "" {
		return nil, fmt.Errorf("username and password required")
	}

	user, err := s.repository.GetUserByUsername(ctx, username)
	if err != nil {
		// Record attempt even for unknown usernames (best-effort)
		s.ProcessFailedLogin(ctx, 0, username, ipAddress, userAgent, "user not found")
		return nil, fmt.Errorf("user not found")
	}

	// Check lockout before verifying password (fail-fast, prevents enumeration)
	if locked, lockout, _ := s.IsAccountLocked(ctx, user.ID); locked {
		_ = s.repository.CreateLoginAttempt(ctx, username, ipAddress, userAgent, false, "account locked")
		return nil, fmt.Errorf("%w; locked until %s", ErrAccountLocked, lockout.UnlockAt.Format(time.RFC3339))
	}

	if user.PasswordHash == "" {
		s.ProcessFailedLogin(ctx, user.ID, username, ipAddress, userAgent, "no password set")
		return nil, fmt.Errorf("user has no password set")
	}

	if !utils.VerifyPassword(user.PasswordHash, password) {
		s.ProcessFailedLogin(ctx, user.ID, username, ipAddress, userAgent, "invalid password")
		return nil, fmt.Errorf("invalid password")
	}

	switch user.State {
	case "pending_email_verification":
		return nil, ErrEmailNotVerified
	case "pending_admin_approval":
		return nil, ErrPendingApproval
	case "rejected":
		return nil, ErrAccountRejected
	case "disabled":
		return nil, ErrAccountDisabled
	case "suspended":
		return nil, fmt.Errorf("account is suspended")
	}

	// If MFA is enabled, issue a challenge instead of returning the full user
	if s.MFA.IsMFAEnabled(ctx, user.ID) {
		challenge, err := s.MFA.CreateChallenge(ctx, user.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to create MFA challenge: %w", err)
		}
		return &AuthResult{MFARequired: true, MFAChallenge: challenge}, nil
	}

	s.RecordSuccessfulLogin(ctx, username, ipAddress, userAgent)
	return &AuthResult{User: user}, nil
}

// RevokeSessionToken revokes a session token by its hash
func (s *Service) RevokeSessionToken(ctx context.Context, userID int64, token string) error {
	if token == "" {
		return fmt.Errorf("token is required")
	}

	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])

	apiToken, err := s.repository.GetAPITokenByHash(ctx, tokenHash)
	if err != nil {
		return fmt.Errorf("token not found")
	}

	if apiToken.UserID != userID {
		return fmt.Errorf("token does not belong to this user")
	}

	return s.repository.DeleteAPIToken(ctx, apiToken.ID)
}

// CreatePasswordResetToken creates a password reset token for a user.
func (s *Service) CreatePasswordResetToken(ctx context.Context, email string) (token string, err error) {
	if email == "" {
		return "", fmt.Errorf("email is required")
	}

	user, err := s.repository.GetUserByEmail(ctx, email)
	if err != nil {
		return "", fmt.Errorf("user not found")
	}

	tokenBytes := make([]byte, TokenLength)
	_, err = rand.Read(tokenBytes)
	if err != nil {
		return "", err
	}
	token = hex.EncodeToString(tokenBytes)

	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])

	_, err = s.repository.CreatePasswordReset(ctx, user.ID, tokenHash)
	if err != nil {
		return "", err
	}

	return token, nil
}

// SendPasswordResetEmail creates a reset token for the given email and sends it.
// Returns a generic error (not revealing whether the email exists) to prevent enumeration.
func (s *Service) SendPasswordResetEmail(ctx context.Context, email string) error {
	token, err := s.CreatePasswordResetToken(ctx, email)
	if err != nil {
		return err
	}
	user, err := s.repository.GetUserByEmail(ctx, email)
	if err != nil {
		return err
	}
	return s.Email.SendPasswordResetEmail(user.Email, user.Username, token)
}

// SendPasswordResetEmailByID creates a reset token for the given user ID and sends it.
// Used by admins to trigger a reset on behalf of another user.
func (s *Service) SendPasswordResetEmailByID(ctx context.Context, userID int64) error {
	user, err := s.repository.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}
	token, err := s.CreatePasswordResetToken(ctx, user.Email)
	if err != nil {
		return err
	}
	return s.Email.SendPasswordResetEmail(user.Email, user.Username, token)
}

// ResetPasswordWithToken verifies a reset token and updates the password
// ResetPasswordWithToken verifies a reset token, updates the password, and returns the user ID.
func (s *Service) ResetPasswordWithToken(ctx context.Context, token, newPasswordHash string) (int64, error) {
	if token == "" || newPasswordHash == "" {
		return 0, fmt.Errorf("token and password hash required")
	}

	// Hash the provided token
	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])

	// Get the password reset record
	resetRecord, err := s.repository.GetPasswordResetByToken(ctx, tokenHash)
	if err != nil {
		return 0, fmt.Errorf("invalid reset token")
	}

	if resetRecord.ExpiresAt.Before(time.Now()) {
		return 0, fmt.Errorf("reset token has expired")
	}

	if resetRecord.UsedAt != nil {
		return 0, fmt.Errorf("reset token has already been used")
	}

	if err := s.repository.UpdateUserPassword(ctx, resetRecord.UserID, newPasswordHash); err != nil {
		return 0, err
	}

	if err := s.repository.MarkPasswordResetAsUsed(ctx, resetRecord.ID); err != nil {
		return 0, err
	}

	return resetRecord.UserID, nil
}

// InitAdminPassword sets the admin password when it is NULL (first boot).
// Returns true if the password was applied.
func (s *Service) InitAdminPassword(ctx context.Context, password string) (bool, error) {
	hash, err := utils.HashPassword(password)
	if err != nil {
		return false, fmt.Errorf("hashing admin password: %w", err)
	}
	return s.repository.InitAdminPassword(ctx, hash)
}

// ForceResetAdminPassword unconditionally sets the admin password.
func (s *Service) ForceResetAdminPassword(ctx context.Context, password string) error {
	hash, err := utils.HashPassword(password)
	if err != nil {
		return fmt.Errorf("hashing admin password: %w", err)
	}
	return s.repository.ForceSetAdminPassword(ctx, hash)
}
