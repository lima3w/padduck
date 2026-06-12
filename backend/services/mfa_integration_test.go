package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"padduck/internal/testdb"
	"padduck/repository"
)

const testMFAKey = "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"

// testMFAService returns an MFA service backed by a scratch database with a
// fixture user, truncating MFA tables for isolation.
func testMFAService(t *testing.T) (*MFAService, int64) {
	t.Helper()
	pool := testdb.Connect(t, "services")
	testdb.Truncate(t, pool,
		"mfa_challenges", "user_backup_codes", "user_totp_secrets", "user_mfa_settings", "users")

	// Production-cost bcrypt for 10 codes per enrollment dominates test time
	// (especially under -race); single-use and matching semantics are
	// identical at MinCost.
	prevCost := backupCodeBcryptCost
	backupCodeBcryptCost = bcrypt.MinCost
	t.Cleanup(func() { backupCodeBcryptCost = prevCost })
	repo := repository.NewRepository(pool)
	svc, err := NewMFAService(repo, testMFAKey)
	require.NoError(t, err)
	u, err := repo.CreateUser(context.Background(), "mfa-user", "mfa@example.com")
	require.NoError(t, err)
	return svc, u.ID
}

// totpCode computes the current valid code for a secret, like an
// authenticator app would.
func totpCode(t *testing.T, secret string) string {
	t.Helper()
	code, err := totp.GenerateCode(secret, time.Now())
	require.NoError(t, err)
	return code
}

// setupAndConfirm walks a user through the full enrollment flow and returns
// the TOTP secret and backup codes.
func setupAndConfirm(t *testing.T, svc *MFAService, userID int64) (string, []string) {
	t.Helper()
	ctx := context.Background()
	secret, _, err := svc.SetupTOTP(ctx, userID, "mfa-user", "mfa@example.com")
	require.NoError(t, err)
	codes, err := svc.ConfirmTOTP(ctx, userID, totpCode(t, secret))
	require.NoError(t, err)
	return secret, codes
}

func TestNewMFAService_KeyValidation(t *testing.T) {
	for _, bad := range []string{"", "not-hex", "abcd", strings.Repeat("0", 63), strings.Repeat("0", 66)} {
		_, err := NewMFAService(nil, bad)
		assert.Error(t, err, "key %q must be rejected", bad)
	}
	_, err := NewMFAService(nil, testMFAKey)
	assert.NoError(t, err)
}

func TestMFAEncryptDecrypt(t *testing.T) {
	svc, err := NewMFAService(nil, testMFAKey)
	require.NoError(t, err)

	plaintext := []byte("JBSWY3DPEHPK3PXP")
	ct, err := svc.encrypt(plaintext)
	require.NoError(t, err)
	assert.NotEqual(t, plaintext, ct)

	got, err := svc.decrypt(ct)
	require.NoError(t, err)
	assert.Equal(t, plaintext, got)

	// Same plaintext encrypts to different ciphertexts (random nonce).
	ct2, err := svc.encrypt(plaintext)
	require.NoError(t, err)
	assert.NotEqual(t, ct, ct2)

	// Wrong key must fail, not return garbage.
	other, err := NewMFAService(nil, strings.Repeat("ff", 32))
	require.NoError(t, err)
	_, err = other.decrypt(ct)
	assert.Error(t, err)

	// Corrupted ciphertext must fail.
	corrupted := append([]byte{}, ct...)
	corrupted[len(corrupted)-1] ^= 0xff
	_, err = svc.decrypt(corrupted)
	assert.Error(t, err)

	// Truncated below the nonce size must fail without panicking.
	_, err = svc.decrypt(ct[:5])
	assert.Error(t, err)
}

func TestSetupTOTP_Integration(t *testing.T) {
	svc, userID := testMFAService(t)
	ctx := context.Background()

	secret, qr, err := svc.SetupTOTP(ctx, userID, "mfa-user", "mfa@example.com")
	require.NoError(t, err)
	assert.NotEmpty(t, secret)
	assert.True(t, strings.HasPrefix(qr, "data:image/png;base64,"), "QR must be a PNG data URL")

	// The stored secret round-trips through encryption.
	stored, err := svc.getTOTPSecret(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, secret, stored)

	// Setup is repeatable until confirmed (replaces the pending secret)...
	secret2, _, err := svc.SetupTOTP(ctx, userID, "mfa-user", "mfa@example.com")
	require.NoError(t, err)
	assert.NotEqual(t, secret, secret2)

	// ...but not once MFA is enabled.
	_, err = svc.ConfirmTOTP(ctx, userID, totpCode(t, secret2))
	require.NoError(t, err)
	_, _, err = svc.SetupTOTP(ctx, userID, "mfa-user", "mfa@example.com")
	assert.ErrorIs(t, err, ErrMFAAlreadyEnabled)
}

func TestConfirmTOTP_Integration(t *testing.T) {
	svc, userID := testMFAService(t)
	ctx := context.Background()

	// Confirm without setup.
	_, err := svc.ConfirmTOTP(ctx, userID, "123456")
	assert.ErrorIs(t, err, ErrMFANotSetup)

	secret, _, err := svc.SetupTOTP(ctx, userID, "mfa-user", "mfa@example.com")
	require.NoError(t, err)

	// Wrong code: MFA must stay disabled.
	_, err = svc.ConfirmTOTP(ctx, userID, "000000")
	assert.ErrorIs(t, err, ErrInvalidTOTPCode)
	assert.False(t, svc.IsMFAEnabled(ctx, userID))

	// Correct code: enabled, 10 backup codes in XXXXXX-XXXXXX hex format.
	codes, err := svc.ConfirmTOTP(ctx, userID, totpCode(t, secret))
	require.NoError(t, err)
	require.Len(t, codes, 10)
	format := regexp.MustCompile(`^[0-9A-F]{6}-[0-9A-F]{6}$`)
	for _, c := range codes {
		assert.Regexp(t, format, c)
	}
	assert.True(t, svc.IsMFAEnabled(ctx, userID))

	enabled, remaining := svc.GetMFAStatus(ctx, userID)
	assert.True(t, enabled)
	assert.Equal(t, 10, remaining)

	assert.True(t, svc.ValidateTOTPCode(ctx, userID, totpCode(t, secret)))
	assert.False(t, svc.ValidateTOTPCode(ctx, userID, "000000"))
}

func TestDisableTOTP_Integration(t *testing.T) {
	svc, userID := testMFAService(t)
	ctx := context.Background()

	// Disable before enabling.
	assert.ErrorIs(t, svc.DisableTOTP(ctx, userID, "123456"), ErrMFANotEnabled)

	secret, _ := setupAndConfirm(t, svc, userID)

	// Wrong code keeps MFA on.
	assert.ErrorIs(t, svc.DisableTOTP(ctx, userID, "000000"), ErrInvalidTOTPCode)
	assert.True(t, svc.IsMFAEnabled(ctx, userID))

	require.NoError(t, svc.DisableTOTP(ctx, userID, totpCode(t, secret)))
	assert.False(t, svc.IsMFAEnabled(ctx, userID))
	enabled, remaining := svc.GetMFAStatus(ctx, userID)
	assert.False(t, enabled)
	assert.Zero(t, remaining)
}

func TestBackupCodes_SingleUse_Integration(t *testing.T) {
	svc, userID := testMFAService(t)
	ctx := context.Background()

	_, codes := setupAndConfirm(t, svc, userID)

	// A backup code authenticates once...
	require.NoError(t, svc.VerifyTOTPOrBackupCode(ctx, userID, codes[0]))
	assert.Equal(t, 9, svc.BackupCodesRemaining(ctx, userID))

	// ...and is rejected on replay.
	assert.ErrorIs(t, svc.VerifyTOTPOrBackupCode(ctx, userID, codes[0]), ErrInvalidTOTPCode)
	assert.Equal(t, 9, svc.BackupCodesRemaining(ctx, userID))

	// Input is normalized: lowercase with spaces still matches.
	sloppy := strings.ToLower(codes[1][:3] + " " + codes[1][3:])
	require.NoError(t, svc.VerifyTOTPOrBackupCode(ctx, userID, sloppy))
	assert.Equal(t, 8, svc.BackupCodesRemaining(ctx, userID))

	// A made-up code is rejected.
	assert.ErrorIs(t, svc.VerifyTOTPOrBackupCode(ctx, userID, "AAAAAA-AAAAAA"), ErrInvalidTOTPCode)
}

func TestRegenerateBackupCodes_Integration(t *testing.T) {
	svc, userID := testMFAService(t)
	ctx := context.Background()

	secret, oldCodes := setupAndConfirm(t, svc, userID)

	// Requires a valid code.
	_, err := svc.RegenerateBackupCodes(ctx, userID, "000000")
	assert.ErrorIs(t, err, ErrInvalidTOTPCode)

	newCodes, err := svc.RegenerateBackupCodes(ctx, userID, totpCode(t, secret))
	require.NoError(t, err)
	require.Len(t, newCodes, 10)
	assert.Equal(t, 10, svc.BackupCodesRemaining(ctx, userID))

	// Old codes are invalidated, new ones work.
	assert.ErrorIs(t, svc.VerifyTOTPOrBackupCode(ctx, userID, oldCodes[0]), ErrInvalidTOTPCode)
	require.NoError(t, svc.VerifyTOTPOrBackupCode(ctx, userID, newCodes[0]))
}

func TestChallengeFlow_Integration(t *testing.T) {
	svc, userID := testMFAService(t)
	ctx := context.Background()

	secret, _ := setupAndConfirm(t, svc, userID)

	raw, err := svc.CreateChallenge(ctx, userID)
	require.NoError(t, err)
	require.NotEmpty(t, raw)

	// Unknown challenge token.
	_, err = svc.CompleteChallenge(ctx, "bogus-token", totpCode(t, secret))
	assert.ErrorIs(t, err, ErrInvalidChallenge)

	// Valid token, wrong code: challenge stays open.
	_, err = svc.CompleteChallenge(ctx, raw, "000000")
	assert.ErrorIs(t, err, ErrInvalidTOTPCode)

	// Valid token + valid code returns the user.
	gotUser, err := svc.CompleteChallenge(ctx, raw, totpCode(t, secret))
	require.NoError(t, err)
	assert.Equal(t, userID, gotUser)

	// Replaying a completed challenge fails even with a valid code.
	_, err = svc.CompleteChallenge(ctx, raw, totpCode(t, secret))
	assert.ErrorIs(t, err, ErrChallengeCompleted)
}

func TestChallengeExpiry_Integration(t *testing.T) {
	svc, userID := testMFAService(t)
	ctx := context.Background()

	secret, _ := setupAndConfirm(t, svc, userID)

	// Insert an already-expired challenge directly, mirroring
	// CreateChallenge's token hashing.
	raw := "expired-challenge-raw-token"
	hash := sha256.Sum256([]byte(raw))
	_, err := svc.repository.CreateMFAChallenge(ctx, userID, hex.EncodeToString(hash[:]), time.Now().Add(-time.Minute))
	require.NoError(t, err)

	_, err = svc.CompleteChallenge(ctx, raw, totpCode(t, secret))
	assert.ErrorIs(t, err, ErrChallengeExpired)
}
