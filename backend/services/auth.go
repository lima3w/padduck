package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"ipam-next/models"
	"ipam-next/utils"
)

const (
	TokenLength = 32
)

// GenerateAPIToken creates a new API token for a user
func (s *Service) GenerateAPIToken(ctx context.Context, userID int64, tokenName string) (token string, err error) {
	if userID <= 0 {
		return "", fmt.Errorf("invalid user ID")
	}
	if tokenName == "" {
		return "", fmt.Errorf("token name is required")
	}

	// Generate random token
	tokenBytes := make([]byte, TokenLength)
	_, err = rand.Read(tokenBytes)
	if err != nil {
		return "", err
	}
	token = hex.EncodeToString(tokenBytes)

	// Hash token for storage
	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])

	// Store token hash in database
	_, err = s.repository.CreateAPIToken(ctx, userID, tokenHash, tokenName)
	if err != nil {
		return "", err
	}

	return token, nil
}

// ValidateAPIToken checks if a token is valid and returns the user
func (s *Service) ValidateAPIToken(ctx context.Context, token string) (*models.User, error) {
	if token == "" {
		return nil, fmt.Errorf("token is required")
	}

	// Hash the provided token
	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])

	// Look up token in database
	apiToken, err := s.repository.GetAPITokenByHash(ctx, tokenHash)
	if err != nil {
		return nil, fmt.Errorf("invalid token")
	}

	// Check if token has expired
	if apiToken.ExpiresAt != nil && apiToken.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("token has expired")
	}

	// Update last used timestamp
	_ = s.repository.UpdateAPITokenLastUsed(ctx, apiToken.ID)

	// Get user information
	user, err := s.repository.GetUserByID(ctx, apiToken.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
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

// CreatePasswordResetToken creates a password reset token for a user
func (s *Service) CreatePasswordResetToken(ctx context.Context, email string) (token string, err error) {
	if email == "" {
		return "", fmt.Errorf("email is required")
	}

	user, err := s.repository.GetUserByEmail(ctx, email)
	if err != nil {
		return "", fmt.Errorf("user not found")
	}

	// Generate random token
	tokenBytes := make([]byte, TokenLength)
	_, err = rand.Read(tokenBytes)
	if err != nil {
		return "", err
	}
	token = hex.EncodeToString(tokenBytes)

	// Hash token for storage
	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])

	// Store token in database with 1 hour expiration
	_, err = s.repository.CreatePasswordReset(ctx, user.ID, tokenHash)
	if err != nil {
		return "", err
	}

	return token, nil
}

// ResetPasswordWithToken verifies a reset token and updates the password
func (s *Service) ResetPasswordWithToken(ctx context.Context, token, newPasswordHash string) error {
	if token == "" || newPasswordHash == "" {
		return fmt.Errorf("token and password hash required")
	}

	// Hash the provided token
	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])

	// Get the password reset record
	resetRecord, err := s.repository.GetPasswordResetByToken(ctx, tokenHash)
	if err != nil {
		return fmt.Errorf("invalid reset token")
	}

	// Check if token has expired
	if resetRecord.ExpiresAt.Before(time.Now()) {
		return fmt.Errorf("reset token has expired")
	}

	// Check if token has already been used
	if resetRecord.UsedAt != nil {
		return fmt.Errorf("reset token has already been used")
	}

	// Update user's password
	if err := s.repository.UpdateUserPassword(ctx, resetRecord.UserID, newPasswordHash); err != nil {
		return err
	}

	// Mark token as used
	if err := s.repository.MarkPasswordResetAsUsed(ctx, resetRecord.ID); err != nil {
		return err
	}

	return nil
}

// InitAdminPassword sets the admin password on first boot when it is NULL.
// Returns true and the password used if it was set (or generated), false if already set.
func (s *Service) InitAdminPassword(ctx context.Context, password string) (bool, error) {
	hash, err := utils.HashPassword(password)
	if err != nil {
		return false, fmt.Errorf("hashing admin password: %w", err)
	}
	return s.repository.InitAdminPassword(ctx, hash)
}
