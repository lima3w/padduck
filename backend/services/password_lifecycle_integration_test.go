package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"padduck/internal/testdb"
	"padduck/repository"
	"padduck/utils"
)

// testAuthService returns a full Service backed by a scratch database with a
// password-holding fixture user.
func testAuthService(t *testing.T) (*Service, *repository.Repository, int64) {
	t.Helper()
	pool := testdb.Connect(t, "services")
	testdb.Truncate(t, pool,
		"password_resets", "api_tokens", "sessions",
		"login_attempts", "account_lockouts", "security_notifications", "users")
	repo := repository.NewRepository(pool)
	svc := NewService(repo, testMFAKey)

	hash, err := utils.HashPassword("original-password")
	require.NoError(t, err)
	u, err := repo.CreateUserWithPassword(context.Background(), "pw-user", "pw@example.com", hash, "user")
	require.NoError(t, err)
	return svc, repo, u.ID
}

func userPasswordVerifies(t *testing.T, repo *repository.Repository, userID int64, password string) bool {
	t.Helper()
	u, err := repo.GetUserByID(context.Background(), userID)
	require.NoError(t, err)
	return utils.VerifyPassword(u.PasswordHash, password)
}

func TestChangePassword_Integration(t *testing.T) {
	svc, repo, userID := testAuthService(t)
	ctx := context.Background()

	// Wrong current password: rejected, password unchanged.
	err := svc.Ops.Identity.ChangePassword(ctx, userID, "wrong", "new-password-123", "")
	assert.ErrorContains(t, err, "current password is incorrect")
	assert.True(t, userPasswordVerifies(t, repo, userID, "original-password"))

	// Correct current password: old stops working, new works.
	require.NoError(t, svc.Ops.Identity.ChangePassword(ctx, userID, "original-password", "new-password-123", ""))
	assert.False(t, userPasswordVerifies(t, repo, userID, "original-password"))
	assert.True(t, userPasswordVerifies(t, repo, userID, "new-password-123"))
}

func TestChangePassword_RevokesOtherSessions_Integration(t *testing.T) {
	svc, _, userID := testAuthService(t)
	ctx := context.Background()

	current, err := svc.Ops.Identity.CreateWebSession(ctx, userID, "10.0.0.1", "browser-a")
	require.NoError(t, err)
	other, err := svc.Ops.Identity.CreateWebSession(ctx, userID, "10.0.0.2", "browser-b")
	require.NoError(t, err)

	// Keeping the current session: it survives, the other is revoked.
	require.NoError(t, svc.Ops.Identity.ChangePassword(ctx, userID, "original-password", "new-password-123", current))
	_, _, err = svc.Ops.Identity.ValidateSession(ctx, current)
	assert.NoError(t, err, "session making the change must stay valid")
	_, _, err = svc.Ops.Identity.ValidateSession(ctx, other)
	assert.Error(t, err, "other sessions must be revoked on password change")

	// Empty keep token revokes everything.
	require.NoError(t, svc.Ops.Identity.ChangePassword(ctx, userID, "new-password-123", "third-password-123", ""))
	_, _, err = svc.Ops.Identity.ValidateSession(ctx, current)
	assert.Error(t, err)
}

func TestPasswordResetFlow_Integration(t *testing.T) {
	svc, repo, userID := testAuthService(t)
	ctx := context.Background()

	// Unknown email.
	_, err := svc.Ops.Identity.CreatePasswordResetToken(ctx, "nobody@example.com")
	assert.Error(t, err)

	token, err := svc.Ops.Identity.CreatePasswordResetToken(ctx, "pw@example.com")
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// The raw token is not stored: only its hash.
	var stored string
	require.NoError(t, repo.GetPool().QueryRow(ctx, "SELECT token_hash FROM password_resets").Scan(&stored))
	assert.NotEqual(t, token, stored)

	// A session exists before the reset; recovery must kill it.
	session, err := svc.Ops.Identity.CreateWebSession(ctx, userID, "10.0.0.9", "intruder")
	require.NoError(t, err)

	// Wrong token.
	newHash, err := utils.HashPassword("reset-password-123")
	require.NoError(t, err)
	_, err = svc.Ops.Identity.ResetPasswordWithToken(ctx, "bogus-token", newHash)
	assert.ErrorContains(t, err, "invalid reset token")

	// Valid token resets the password and reports the right user.
	gotUser, err := svc.Ops.Identity.ResetPasswordWithToken(ctx, token, newHash)
	require.NoError(t, err)
	assert.Equal(t, userID, gotUser)
	assert.True(t, userPasswordVerifies(t, repo, userID, "reset-password-123"))

	// All sessions are revoked by the reset.
	_, _, err = svc.Ops.Identity.ValidateSession(ctx, session)
	assert.Error(t, err, "sessions must not survive a password reset")

	// Replaying the consumed token fails.
	_, err = svc.Ops.Identity.ResetPasswordWithToken(ctx, token, newHash)
	assert.ErrorContains(t, err, "already been used")
}

func TestPasswordResetToken_Expiry_Integration(t *testing.T) {
	svc, repo, _ := testAuthService(t)
	ctx := context.Background()

	token, err := svc.Ops.Identity.CreatePasswordResetToken(ctx, "pw@example.com")
	require.NoError(t, err)

	// Force the token past its expiry window.
	_, err = repo.GetPool().Exec(ctx, "UPDATE password_resets SET expires_at = CURRENT_TIMESTAMP - INTERVAL '1 minute'")
	require.NoError(t, err)

	newHash, err := utils.HashPassword("reset-password-123")
	require.NoError(t, err)
	_, err = svc.Ops.Identity.ResetPasswordWithToken(ctx, token, newHash)
	assert.ErrorContains(t, err, "expired")
}

func TestInitAndForceResetAdminPassword_Integration(t *testing.T) {
	svc, repo, _ := testAuthService(t)
	ctx := context.Background()

	// Fixture admin with no password yet (first boot state).
	admin, err := repo.CreateUser(ctx, "admin", "admin@example.com")
	require.NoError(t, err)

	// First boot: password applied.
	applied, err := svc.Ops.Identity.InitAdminPassword(ctx, "first-boot-password")
	require.NoError(t, err)
	assert.True(t, applied)
	assert.True(t, userPasswordVerifies(t, repo, admin.ID, "first-boot-password"))

	// Second boot: hash already set, init must not overwrite.
	applied, err = svc.Ops.Identity.InitAdminPassword(ctx, "second-boot-password")
	require.NoError(t, err)
	assert.False(t, applied)
	assert.True(t, userPasswordVerifies(t, repo, admin.ID, "first-boot-password"))

	// Force reset overrides unconditionally.
	require.NoError(t, svc.Ops.Identity.ForceResetAdminPassword(ctx, "forced-password"))
	assert.True(t, userPasswordVerifies(t, repo, admin.ID, "forced-password"))
}

func TestRotateAPIToken_Integration(t *testing.T) {
	svc, repo, userID := testAuthService(t)
	ctx := context.Background()

	raw, err := svc.Ops.Identity.GenerateAPIToken(ctx, userID, "automation", "write", 30)
	require.NoError(t, err)
	tokens, err := svc.Ops.Identity.ListUserTokens(ctx, userID)
	require.NoError(t, err)
	require.Len(t, tokens, 1)
	oldID := tokens[0].ID

	// Wrong owner.
	_, _, err = svc.Ops.Identity.RotateAPIToken(ctx, oldID, userID+999)
	assert.ErrorContains(t, err, "does not belong")

	newRaw, graceExpiresAt, err := svc.Ops.Identity.RotateAPIToken(ctx, oldID, userID)
	require.NoError(t, err)
	assert.NotEqual(t, raw, newRaw)
	assert.WithinDuration(t, time.Now().Add(24*time.Hour), graceExpiresAt, time.Minute,
		"default grace period is 24h")

	// Old token is marked rotated; the replacement carries name and scope.
	oldTok, err := repo.GetAPITokenByID(ctx, oldID)
	require.NoError(t, err)
	assert.NotNil(t, oldTok.RotationGraceExpiresAt)

	tokens, err = svc.Ops.Identity.ListUserTokens(ctx, userID)
	require.NoError(t, err)
	require.Len(t, tokens, 2)
	var replacement *struct {
		Name, Scope string
	}
	for _, tok := range tokens {
		if tok.ID != oldID {
			replacement = &struct{ Name, Scope string }{tok.Name, tok.Scope}
		}
	}
	require.NotNil(t, replacement)
	assert.Equal(t, "automation", replacement.Name)
	assert.Equal(t, "write", replacement.Scope)
}

func TestExtendAPIToken_Integration(t *testing.T) {
	svc, _, userID := testAuthService(t)
	ctx := context.Background()

	_, err := svc.Ops.Identity.GenerateAPIToken(ctx, userID, "extend-me", "read", 1)
	require.NoError(t, err)
	tokens, err := svc.Ops.Identity.ListUserTokens(ctx, userID)
	require.NoError(t, err)
	require.Len(t, tokens, 1)

	// Wrong owner.
	_, err = svc.Ops.Identity.ExtendAPIToken(ctx, tokens[0].ID, userID+999, 10)
	assert.ErrorContains(t, err, "does not belong")

	extended, err := svc.Ops.Identity.ExtendAPIToken(ctx, tokens[0].ID, userID, 10)
	require.NoError(t, err)
	require.NotNil(t, extended.ExpiresAt)
	assert.WithinDuration(t, time.Now().Add(10*24*time.Hour), *extended.ExpiresAt, time.Minute)
}

func TestRevokeSessionToken_Integration(t *testing.T) {
	svc, _, userID := testAuthService(t)
	ctx := context.Background()

	raw, err := svc.Ops.Identity.GenerateAPIToken(ctx, userID, "revoke-me", "read", 0)
	require.NoError(t, err)

	// Empty and unknown tokens.
	assert.ErrorContains(t, svc.Ops.Identity.RevokeSessionToken(ctx, userID, ""), "token is required")
	assert.ErrorContains(t, svc.Ops.Identity.RevokeSessionToken(ctx, userID, "unknown"), "token not found")

	// Wrong owner cannot revoke.
	assert.ErrorContains(t, svc.Ops.Identity.RevokeSessionToken(ctx, userID+999, raw), "does not belong")

	require.NoError(t, svc.Ops.Identity.RevokeSessionToken(ctx, userID, raw))
	tokens, err := svc.Ops.Identity.ListUserTokens(ctx, userID)
	require.NoError(t, err)
	assert.Empty(t, tokens)
}
