package services

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	qrcode "github.com/skip2/go-qrcode"
	"golang.org/x/crypto/bcrypt"
	"ipam-next/models"
	"ipam-next/repository"
)

var (
	ErrMFAAlreadyEnabled   = errors.New("MFA is already enabled")
	ErrMFANotEnabled       = errors.New("MFA is not enabled")
	ErrMFANotSetup         = errors.New("MFA setup not started")
	ErrInvalidTOTPCode     = errors.New("invalid MFA code")
	ErrInvalidChallenge    = errors.New("invalid or expired MFA challenge")
	ErrChallengeExpired    = errors.New("MFA challenge expired")
	ErrChallengeCompleted  = errors.New("MFA challenge already completed")
)

const (
	totpIssuer    = "IPAM Next"
	backupCodeLen = 10
	challengeTTL  = 5 * time.Minute
)

type MFAService struct {
	repository    *repository.Repository
	encryptionKey []byte
}

func NewMFAService(repo *repository.Repository, encryptionKeyHex string) (*MFAService, error) {
	key, err := hex.DecodeString(encryptionKeyHex)
	if err != nil || len(key) != 32 {
		return nil, fmt.Errorf("MFA_ENCRYPTION_KEY must be a 64-character hex string (32 bytes)")
	}
	return &MFAService{repository: repo, encryptionKey: key}, nil
}

// SetupTOTP generates a new TOTP secret for the user and returns QR code data.
// The secret is stored unverified until ConfirmTOTP is called.
func (s *MFAService) SetupTOTP(ctx context.Context, userID int64, username, email string) (secret, qrDataURL string, err error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      totpIssuer,
		AccountName: email,
		Algorithm:   otp.AlgorithmSHA1,
		Digits:      otp.DigitsSix,
		Period:      30,
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to generate TOTP key: %w", err)
	}

	encrypted, err := s.encrypt([]byte(key.Secret()))
	if err != nil {
		return "", "", fmt.Errorf("failed to encrypt secret: %w", err)
	}

	if err := s.repository.UpsertTOTPSecret(ctx, userID, encrypted); err != nil {
		return "", "", fmt.Errorf("failed to store TOTP secret: %w", err)
	}

	png, err := qrcode.Encode(key.URL(), qrcode.Medium, 256)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate QR code: %w", err)
	}

	qrDataURL = "data:image/png;base64," + base64.StdEncoding.EncodeToString(png)
	return key.Secret(), qrDataURL, nil
}

// ConfirmTOTP verifies the first code entered after setup, enabling TOTP for the user.
// Also generates and returns backup codes.
func (s *MFAService) ConfirmTOTP(ctx context.Context, userID int64, code string) ([]string, error) {
	secret, err := s.getTOTPSecret(ctx, userID)
	if err != nil {
		return nil, ErrMFANotSetup
	}

	if !s.validateTOTP(code, secret) {
		return nil, ErrInvalidTOTPCode
	}

	if err := s.repository.MarkTOTPVerified(ctx, userID); err != nil {
		return nil, err
	}

	now := time.Now()
	if err := s.repository.UpsertMFASettings(ctx, userID, true, &now); err != nil {
		return nil, err
	}

	return s.regenerateBackupCodes(ctx, userID, &now)
}

// DisableTOTP disables MFA after verifying the current TOTP code.
func (s *MFAService) DisableTOTP(ctx context.Context, userID int64, code string) error {
	settings, err := s.repository.GetMFASettings(ctx, userID)
	if err != nil || !settings.TOTPEnabled {
		return ErrMFANotEnabled
	}

	if err := s.VerifyTOTPOrBackupCode(ctx, userID, code); err != nil {
		return err
	}

	if err := s.repository.DeleteTOTPSecret(ctx, userID); err != nil {
		return err
	}
	return s.repository.UpsertMFASettings(ctx, userID, false, nil)
}

// RegenerateBackupCodes creates a fresh set of backup codes (requires TOTP verification).
func (s *MFAService) RegenerateBackupCodes(ctx context.Context, userID int64, code string) ([]string, error) {
	if err := s.VerifyTOTPOrBackupCode(ctx, userID, code); err != nil {
		return nil, err
	}
	now := time.Now()
	return s.regenerateBackupCodes(ctx, userID, &now)
}

func (s *MFAService) regenerateBackupCodes(ctx context.Context, userID int64, ts *time.Time) ([]string, error) {
	codes := make([]string, backupCodeLen)
	hashes := make([]string, backupCodeLen)
	for i := range codes {
		raw := make([]byte, 6)
		rand.Read(raw)
		code := strings.ToUpper(hex.EncodeToString(raw))
		code = code[:6] + "-" + code[6:]
		codes[i] = code
		h, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		hashes[i] = string(h)
	}
	if err := s.repository.CreateBackupCodes(ctx, userID, hashes); err != nil {
		return nil, err
	}
	if ts != nil {
		s.repository.UpsertMFASettings(ctx, userID, true, ts)
	}
	return codes, nil
}

// VerifyTOTPOrBackupCode validates a TOTP code or a backup code.
func (s *MFAService) VerifyTOTPOrBackupCode(ctx context.Context, userID int64, code string) error {
	// Try TOTP first
	secret, err := s.getTOTPSecret(ctx, userID)
	if err == nil && s.validateTOTP(code, secret) {
		return nil
	}

	// Try backup code
	return s.verifyBackupCode(ctx, userID, code)
}

func (s *MFAService) verifyBackupCode(ctx context.Context, userID int64, code string) error {
	codes, err := s.repository.ListBackupCodes(ctx, userID)
	if err != nil {
		return ErrInvalidTOTPCode
	}
	normalised := strings.ToUpper(strings.ReplaceAll(code, " ", ""))
	for _, c := range codes {
		if c.Used {
			continue
		}
		if bcrypt.CompareHashAndPassword([]byte(c.CodeHash), []byte(normalised)) == nil {
			_ = s.repository.MarkBackupCodeUsed(ctx, c.ID)
			return nil
		}
	}
	return ErrInvalidTOTPCode
}

// GetMFAStatus returns whether TOTP is enabled and how many backup codes remain.
func (s *MFAService) GetMFAStatus(ctx context.Context, userID int64) (enabled bool, backupRemaining int) {
	settings, err := s.repository.GetMFASettings(ctx, userID)
	if err != nil || !settings.TOTPEnabled {
		return false, 0
	}
	codes, err := s.repository.ListBackupCodes(ctx, userID)
	if err != nil {
		return true, 0
	}
	for _, c := range codes {
		if !c.Used {
			backupRemaining++
		}
	}
	return true, backupRemaining
}

// IsMFAEnabled returns true if the user has verified TOTP enabled.
func (s *MFAService) IsMFAEnabled(ctx context.Context, userID int64) bool {
	settings, err := s.repository.GetMFASettings(ctx, userID)
	return err == nil && settings.TOTPEnabled
}

// CreateChallenge issues a short-lived MFA challenge after password auth succeeds.
func (s *MFAService) CreateChallenge(ctx context.Context, userID int64) (string, error) {
	b := make([]byte, 32)
	rand.Read(b)
	raw := hex.EncodeToString(b)
	hash := sha256.Sum256([]byte(raw))
	hashHex := hex.EncodeToString(hash[:])

	_, err := s.repository.CreateMFAChallenge(ctx, userID, hashHex, time.Now().Add(challengeTTL))
	if err != nil {
		return "", err
	}
	return raw, nil
}

// CompleteChallenge verifies a challenge token + MFA code and returns the user ID.
func (s *MFAService) CompleteChallenge(ctx context.Context, rawChallenge, code string) (int64, error) {
	hash := sha256.Sum256([]byte(rawChallenge))
	hashHex := hex.EncodeToString(hash[:])

	ch, err := s.repository.GetMFAChallenge(ctx, hashHex)
	if err != nil {
		return 0, ErrInvalidChallenge
	}
	if ch.CompletedAt != nil {
		return 0, ErrChallengeCompleted
	}
	if time.Now().After(ch.ExpiresAt) {
		return 0, ErrChallengeExpired
	}

	if err := s.VerifyTOTPOrBackupCode(ctx, ch.UserID, code); err != nil {
		return 0, err
	}

	_ = s.repository.CompleteMFAChallenge(ctx, ch.ID)
	return ch.UserID, nil
}

func (s *MFAService) getTOTPSecret(ctx context.Context, userID int64) (string, error) {
	rec, err := s.repository.GetTOTPSecret(ctx, userID)
	if err != nil {
		return "", err
	}
	raw, err := s.decrypt(rec.EncryptedSecret)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func (s *MFAService) validateTOTP(code, secret string) bool {
	valid, err := totp.ValidateCustom(code, secret, time.Now(), totp.ValidateOpts{
		Period:    30,
		Skew:      1,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})
	return err == nil && valid
}

// ValidateTOTPCode validates a TOTP code for a user (used by per-action MFA middleware).
func (s *MFAService) ValidateTOTPCode(ctx context.Context, userID int64, code string) bool {
	secret, err := s.getTOTPSecret(ctx, userID)
	if err != nil {
		return false
	}
	return s.validateTOTP(code, secret)
}

func (s *MFAService) encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func (s *MFAService) decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(ciphertext) < gcm.NonceSize() {
		return nil, fmt.Errorf("ciphertext too short")
	}
	nonce, ct := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]
	return gcm.Open(nil, nonce, ct, nil)
}

// BackupCodesRemaining returns how many unused backup codes a user has.
func (s *MFAService) BackupCodesRemaining(ctx context.Context, userID int64) int {
	codes, err := s.repository.ListBackupCodes(ctx, userID)
	if err != nil {
		return 0
	}
	n := 0
	for _, c := range codes {
		if !c.Used {
			n++
		}
	}
	return n
}

// unused import guard
var _ = models.UserMFASettings{}
