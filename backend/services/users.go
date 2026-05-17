package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/mail"
	"time"

	"ipam-next/models"
	"ipam-next/utils"
)

// CreateUser creates a new user with validation
func (s *Service) CreateUser(ctx context.Context, username, email string) (*models.User, error) {
	if username == "" {
		return nil, fmt.Errorf("username is required")
	}

	if email == "" {
		return nil, fmt.Errorf("email is required")
	}

	if _, err := mail.ParseAddress(email); err != nil {
		return nil, fmt.Errorf("invalid email format: %w", err)
	}

	return s.repository.CreateUser(ctx, username, email)
}

// GetUser retrieves a user by ID
func (s *Service) GetUser(ctx context.Context, id int64) (*models.User, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid user ID")
	}

	return s.repository.GetUserByID(ctx, id)
}

// GetUserByID retrieves a user by ID
func (s *Service) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid user ID")
	}

	return s.repository.GetUserByID(ctx, id)
}

// ListAllUsers returns all users
func (s *Service) ListAllUsers(ctx context.Context) ([]*models.User, error) {
	return s.repository.ListAllUsers(ctx)
}

// CreateUserWithPassword creates a new user with password hash
func (s *Service) CreateUserWithPassword(ctx context.Context, username, email, passwordHash, role string) (*models.User, error) {
	if username == "" {
		return nil, fmt.Errorf("username is required")
	}

	if email == "" {
		return nil, fmt.Errorf("email is required")
	}

	if _, err := mail.ParseAddress(email); err != nil {
		return nil, fmt.Errorf("invalid email format: %w", err)
	}

	if role != "admin" && role != "user" && role != "viewer" {
		return nil, fmt.Errorf("invalid role")
	}

	return s.repository.CreateUserWithPassword(ctx, username, email, passwordHash, role)
}

// UpdateUserRole updates a user's role
func (s *Service) UpdateUserRole(ctx context.Context, userID int64, role string) (*models.User, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("invalid user ID")
	}

	if role != "admin" && role != "user" && role != "viewer" {
		return nil, fmt.Errorf("invalid role")
	}

	return s.repository.UpdateUserRole(ctx, userID, role)
}

// DeleteUser deletes a user
func (s *Service) DeleteUser(ctx context.Context, userID int64) error {
	if userID <= 0 {
		return fmt.Errorf("invalid user ID")
	}

	return s.repository.DeleteUser(ctx, userID)
}

// UpdateUserEmail updates a user's email address (admin operation)
func (s *Service) UpdateUserEmail(ctx context.Context, userID int64, email string) error {
	if userID <= 0 {
		return fmt.Errorf("invalid user ID")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return fmt.Errorf("invalid email format")
	}
	return s.repository.UpdateUserEmail(ctx, userID, email)
}

// SuspendUser suspends a user account
func (s *Service) SuspendUser(ctx context.Context, userID, adminID int64, reason string) error {
	if userID <= 0 {
		return fmt.Errorf("invalid user ID")
	}
	user, err := s.repository.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}
	if user.Role == "admin" {
		return fmt.Errorf("cannot suspend admin users")
	}
	return s.repository.SuspendUser(ctx, userID, adminID, reason)
}

// UnsuspendUser restores a suspended user
func (s *Service) UnsuspendUser(ctx context.Context, userID int64) error {
	if userID <= 0 {
		return fmt.Errorf("invalid user ID")
	}
	return s.repository.UnsuspendUser(ctx, userID)
}

// BulkSuspendUsers suspends multiple users
func (s *Service) BulkSuspendUsers(ctx context.Context, userIDs []int64, adminID int64, reason string) (int64, error) {
	// Filter out admin users
	filtered := make([]int64, 0, len(userIDs))
	for _, id := range userIDs {
		u, err := s.repository.GetUserByID(ctx, id)
		if err == nil && u.Role != "admin" {
			filtered = append(filtered, id)
		}
	}
	if len(filtered) == 0 {
		return 0, nil
	}
	return s.repository.BulkUpdateUserState(ctx, filtered, "suspended")
}

// BulkActivateUsers activates multiple users
func (s *Service) BulkActivateUsers(ctx context.Context, userIDs []int64) (int64, error) {
	return s.repository.BulkUpdateUserState(ctx, userIDs, "active")
}

// BulkDeleteUsers deletes multiple users
func (s *Service) BulkDeleteUsers(ctx context.Context, userIDs []int64) (int64, error) {
	// Ensure at least one admin would remain after deletion
	remainingAdmins, err := s.repository.CountAdminsExcluding(ctx, userIDs)
	if err != nil {
		return 0, fmt.Errorf("failed to check admin count: %w", err)
	}
	if remainingAdmins == 0 {
		return 0, fmt.Errorf("cannot delete all admins")
	}
	return s.repository.BulkDeleteUsers(ctx, userIDs)
}

// BulkImportUsers creates users from a slice of import records
func (s *Service) BulkImportUsers(ctx context.Context, records []BulkUserImportRecord, defaultPassword string) ([]BulkImportResult, error) {
	results := make([]BulkImportResult, 0, len(records))
	for _, rec := range records {
		role := rec.Role
		if role == "" {
			role = "user"
		}
		if role != "admin" && role != "user" && role != "viewer" {
			results = append(results, BulkImportResult{Username: rec.Username, Error: "invalid role"})
			continue
		}
		if rec.Username == "" || rec.Email == "" {
			results = append(results, BulkImportResult{Username: rec.Username, Error: "username and email required"})
			continue
		}
		hash, err := utils.HashPassword(defaultPassword)
		if err != nil {
			results = append(results, BulkImportResult{Username: rec.Username, Error: "password hash error"})
			continue
		}
		user, err := s.repository.CreateUserWithPassword(ctx, rec.Username, rec.Email, hash, role)
		if err != nil {
			results = append(results, BulkImportResult{Username: rec.Username, Error: err.Error()})
			continue
		}
		results = append(results, BulkImportResult{Username: rec.Username, UserID: user.ID})
	}
	return results, nil
}

// BulkUserImportRecord is a single row from CSV import
type BulkUserImportRecord struct {
	Username string
	Email    string
	Role     string
}

// BulkImportResult is the outcome of importing a single user
type BulkImportResult struct {
	Username string `json:"username"`
	UserID   int64  `json:"user_id,omitempty"`
	Error    string `json:"error,omitempty"`
}

// AcceptPrivacyPolicy records user's consent to the current privacy policy
func (s *Service) AcceptPrivacyPolicy(ctx context.Context, userID int64) error {
	version, err := s.Config.GetCtx(ctx, "privacy_policy_version")
	if err != nil || version == "" {
		version = "1.0"
	}
	return s.repository.UpdatePrivacyConsent(ctx, userID, version)
}

// GetPrivacyPolicyVersion returns the current privacy policy version
func (s *Service) GetPrivacyPolicyVersion(ctx context.Context) string {
	v, err := s.Config.GetCtx(ctx, "privacy_policy_version")
	if err != nil || v == "" {
		return "1.0"
	}
	return v
}

// RequestAccountDeletion marks a user as requesting deletion
func (s *Service) RequestAccountDeletion(ctx context.Context, userID int64) error {
	return s.repository.RequestDeletion(ctx, userID)
}

// GDPRDeleteUser anonymizes a user's data for GDPR compliance
func (s *Service) GDPRDeleteUser(ctx context.Context, userID int64) error {
	// Delete sessions and API tokens first
	_ = s.repository.DeleteAllUserSessions(ctx, userID)
	// Anonymize user record
	return s.repository.AnonymizeUser(ctx, userID)
}

// ExportUserData returns all data for a user (GDPR data export)
func (s *Service) ExportUserData(ctx context.Context, userID int64) (map[string]interface{}, error) {
	return s.repository.GetUserAllData(ctx, userID)
}

// StartImpersonation creates an impersonation session for an admin
func (s *Service) StartImpersonation(ctx context.Context, targetUserID, adminID int64, ipAddress, userAgent string) (string, error) {
	target, err := s.repository.GetUserByID(ctx, targetUserID)
	if err != nil {
		return "", fmt.Errorf("target user not found")
	}
	if target.Role == "admin" {
		return "", fmt.Errorf("cannot impersonate admin users")
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	rawToken := hex.EncodeToString(tokenBytes)
	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])

	expiry := time.Now().Add(1 * time.Hour)
	_, err = s.repository.CreateImpersonationSession(ctx, targetUserID, adminID, tokenHash,
		"impersonation", ipAddress, userAgent, expiry)
	if err != nil {
		return "", err
	}
	return rawToken, nil
}
