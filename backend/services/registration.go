package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"ipam-next/models"
	"ipam-next/repository"
	"ipam-next/utils"
)

var (
	ErrRegistrationDisabled  = errors.New("registration is not open")
	ErrUsernameTaken         = errors.New("username already taken")
	ErrEmailTaken            = errors.New("email already in use")
	ErrInvalidUsername       = errors.New("username must be 3-32 characters, letters/numbers/underscore/hyphen only")
	ErrInvalidEmail          = errors.New("invalid email address")
	ErrPasswordTooShort      = errors.New("password must be at least 8 characters")
	ErrEmailNotVerified      = errors.New("email address not verified")
	ErrPendingApproval       = errors.New("account pending admin approval")
	ErrAccountRejected       = errors.New("account registration was rejected")
	ErrAccountDisabled       = errors.New("account has been disabled")
	ErrVerificationInvalid   = errors.New("verification token is invalid or expired")
	ErrVerificationAlreadyUsed = errors.New("verification token already used")
)

var (
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,32}$`)
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
)

type RegistrationService struct {
	repository *repository.Repository
	configSvc  *ConfigService
	emailSvc   *EmailService
}

func NewRegistrationService(repo *repository.Repository, configSvc *ConfigService, emailSvc *EmailService) *RegistrationService {
	return &RegistrationService{
		repository: repo,
		configSvc:  configSvc,
		emailSvc:   emailSvc,
	}
}

type RegisterRequest struct {
	Username string
	Email    string
	Password string
}

type RegisterResult struct {
	User  *models.User
	State string // active, pending_email_verification, pending_admin_approval
}

func (s *RegistrationService) Register(ctx context.Context, req RegisterRequest) (*RegisterResult, error) {
	if !s.configSvc.IsRegistrationEnabled() {
		return nil, ErrRegistrationDisabled
	}

	if err := validateUsername(req.Username); err != nil {
		return nil, err
	}
	if err := validateEmail(req.Email); err != nil {
		return nil, err
	}
	if len(req.Password) < 8 {
		return nil, ErrPasswordTooShort
	}

	// Check uniqueness
	if _, err := s.repository.GetUserByUsername(ctx, req.Username); err == nil {
		return nil, ErrUsernameTaken
	}
	if _, err := s.repository.GetUserByEmail(ctx, req.Email); err == nil {
		return nil, ErrEmailTaken
	}

	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("password hashing failed: %w", err)
	}

	requireVerification := s.configSvc.IsEmailVerificationRequired()
	requireApproval := s.configSvc.IsAdminApprovalRequired()

	state := "active"
	if requireVerification {
		state = "pending_email_verification"
	} else if requireApproval {
		state = "pending_admin_approval"
	}

	user, err := s.repository.CreateUserWithState(ctx, req.Username, req.Email, passwordHash, "user", state)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	if requireVerification {
		if err := s.sendVerificationEmail(ctx, user); err != nil {
			log.Printf("failed to send verification email to %s: %v", user.Email, err)
		}
	} else if requireApproval {
		if _, err := s.repository.CreateUserApproval(ctx, user.ID); err != nil {
			log.Printf("failed to create approval record for user %d: %v", user.ID, err)
		}
		s.notifyAdminsOfPendingApproval(ctx, user)
	}

	return &RegisterResult{User: user, State: state}, nil
}

func (s *RegistrationService) sendVerificationEmail(ctx context.Context, user *models.User) error {
	token, tokenHash, err := generateToken()
	if err != nil {
		return err
	}
	expiresAt := time.Now().Add(24 * time.Hour)

	if _, err := s.repository.CreateEmailVerification(ctx, user.ID, tokenHash, expiresAt); err != nil {
		return err
	}

	return s.emailSvc.SendVerificationEmail(user.Email, user.Username, token)
}

func (s *RegistrationService) VerifyEmail(ctx context.Context, rawToken string) error {
	tokenHash := hashToken(rawToken)

	ev, err := s.repository.GetEmailVerificationByToken(ctx, tokenHash)
	if err != nil {
		return ErrVerificationInvalid
	}
	if ev.UsedAt != nil {
		return ErrVerificationAlreadyUsed
	}
	if time.Now().After(ev.ExpiresAt) {
		return ErrVerificationInvalid
	}

	if err := s.repository.MarkEmailVerificationUsed(ctx, ev.ID); err != nil {
		return err
	}

	user, err := s.repository.GetUserByID(ctx, ev.UserID)
	if err != nil {
		return err
	}

	requireApproval := s.configSvc.IsAdminApprovalRequired()
	newState := "active"
	if requireApproval {
		newState = "pending_admin_approval"
	}

	if err := s.repository.UpdateUserState(ctx, user.ID, newState); err != nil {
		return err
	}

	if requireApproval {
		if _, err := s.repository.CreateUserApproval(ctx, user.ID); err != nil {
			log.Printf("failed to create approval record for user %d: %v", user.ID, err)
		}
		s.notifyAdminsOfPendingApproval(ctx, user)
	}

	return nil
}

func (s *RegistrationService) ResendVerification(ctx context.Context, email string) error {
	user, err := s.repository.GetUserByEmail(ctx, email)
	if err != nil {
		return nil // don't reveal whether email exists
	}
	if user.State != "pending_email_verification" {
		return nil
	}

	if err := s.repository.DeleteEmailVerificationsByUser(ctx, user.ID); err != nil {
		return err
	}
	return s.sendVerificationEmail(ctx, user)
}

func (s *RegistrationService) ApproveUser(ctx context.Context, approvalID, reviewerID int64) error {
	approval, err := s.getApprovalByID(ctx, approvalID)
	if err != nil {
		return err
	}

	if err := s.repository.UpdateUserApproval(ctx, approvalID, "approved", reviewerID, nil); err != nil {
		return err
	}
	if err := s.repository.UpdateUserState(ctx, approval.UserID, "active"); err != nil {
		return err
	}

	user, err := s.repository.GetUserByID(ctx, approval.UserID)
	if err == nil {
		if err := s.emailSvc.SendApprovedEmail(user.Email, user.Username); err != nil {
			log.Printf("failed to send approval email to %s: %v", user.Email, err)
		}
	}

	return nil
}

func (s *RegistrationService) RejectUser(ctx context.Context, approvalID, reviewerID int64, reason string) error {
	approval, err := s.getApprovalByID(ctx, approvalID)
	if err != nil {
		return err
	}

	var rejectionReason *string
	if reason != "" {
		rejectionReason = &reason
	}

	if err := s.repository.UpdateUserApproval(ctx, approvalID, "rejected", reviewerID, rejectionReason); err != nil {
		return err
	}
	if err := s.repository.UpdateUserState(ctx, approval.UserID, "rejected"); err != nil {
		return err
	}

	user, err := s.repository.GetUserByID(ctx, approval.UserID)
	if err == nil {
		if err := s.emailSvc.SendRejectedEmail(user.Email, user.Username, reason); err != nil {
			log.Printf("failed to send rejection email to %s: %v", user.Email, err)
		}
	}

	return nil
}

func (s *RegistrationService) ListPendingApprovals(ctx context.Context) ([]*models.UserApproval, error) {
	return s.repository.ListPendingApprovals(ctx)
}

func (s *RegistrationService) getApprovalByID(ctx context.Context, approvalID int64) (*models.UserApproval, error) {
	return s.repository.GetUserApprovalByID(ctx, approvalID)
}

func (s *RegistrationService) notifyAdminsOfPendingApproval(ctx context.Context, user *models.User) {
	admins, err := s.repository.ListAllUsers(ctx)
	if err != nil {
		return
	}
	for _, admin := range admins {
		if admin.Role == "admin" && admin.Email != "" {
			if err := s.emailSvc.SendApprovalRequestEmail(admin.Email, user.Username, user.Email); err != nil {
				log.Printf("failed to notify admin %s: %v", admin.Email, err)
			}
		}
	}
}

func generateToken() (rawToken, tokenHash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", fmt.Errorf("rand.Read: %w", err)
	}
	rawToken = hex.EncodeToString(b)
	tokenHash = hashToken(rawToken)
	return rawToken, tokenHash, nil
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func validateUsername(username string) error {
	if !usernameRegex.MatchString(username) {
		return ErrInvalidUsername
	}
	return nil
}

func validateEmail(email string) error {
	if !emailRegex.MatchString(strings.ToLower(email)) {
		return ErrInvalidEmail
	}
	return nil
}
