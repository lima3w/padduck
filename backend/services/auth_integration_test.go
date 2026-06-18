package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthenticateUser_Integration(t *testing.T) {
	svc, _, userID := testAuthService(t)
	ctx := context.Background()

	// Success returns the full user with no MFA challenge.
	result, err := svc.AuthenticateUser(ctx, "pw-user", "original-password", "10.0.0.1", "ua")
	require.NoError(t, err)
	require.NotNil(t, result.User)
	assert.Equal(t, userID, result.User.ID)
	assert.False(t, result.MFARequired)

	// Wrong password and unknown user both fail.
	_, err = svc.AuthenticateUser(ctx, "pw-user", "wrong", "10.0.0.1", "ua")
	assert.ErrorContains(t, err, "invalid password")
	_, err = svc.AuthenticateUser(ctx, "ghost", "whatever", "10.0.0.1", "ua")
	assert.ErrorContains(t, err, "user not found")

	// Empty credentials are rejected up front.
	_, err = svc.AuthenticateUser(ctx, "", "", "10.0.0.1", "ua")
	assert.ErrorContains(t, err, "username and password required")
}

func TestAuthenticateUser_AccountStates_Integration(t *testing.T) {
	svc, repo, userID := testAuthService(t)
	ctx := context.Background()

	for state, wantErr := range map[string]error{
		"disabled":               ErrAccountDisabled,
		"pending_admin_approval": ErrPendingApproval,
		"rejected":               ErrAccountRejected,
	} {
		require.NoError(t, repo.UpdateUserState(ctx, userID, state))

		// The state error is only reachable with the correct password...
		_, err := svc.AuthenticateUser(ctx, "pw-user", "original-password", "10.0.0.1", "ua")
		assert.ErrorIs(t, err, wantErr, "state %s with correct password", state)

		// ...a wrong password gets the generic failure (no existence leak).
		_, err = svc.AuthenticateUser(ctx, "pw-user", "wrong", "10.0.0.1", "ua")
		assert.ErrorContains(t, err, "invalid password", "state %s with wrong password", state)
	}
}

func TestAuthenticateUser_Lockout_Integration(t *testing.T) {
	svc, _, userID := testAuthService(t)
	ctx := context.Background()

	// Five failed attempts inside the brute-force window trigger a lockout.
	for i := 0; i < 5; i++ {
		_, err := svc.AuthenticateUser(ctx, "pw-user", "wrong", "10.0.0.9", "ua")
		assert.ErrorContains(t, err, "invalid password")
	}

	locked, lockout, err := svc.IsAccountLocked(ctx, userID)
	require.NoError(t, err)
	require.True(t, locked, "5 failures within the window must lock the account")
	assert.True(t, lockout.UnlockAt.After(time.Now()), "first lockout lasts 5 minutes")

	// Even the correct password is rejected with the lockout error...
	_, err = svc.AuthenticateUser(ctx, "pw-user", "original-password", "10.0.0.9", "ua")
	assert.ErrorIs(t, err, ErrAccountLocked)

	// ...while a wrong password keeps getting the generic failure, so the
	// distinct lockout response never confirms account existence pre-auth.
	_, err = svc.AuthenticateUser(ctx, "pw-user", "wrong", "10.0.0.9", "ua")
	assert.ErrorContains(t, err, "invalid password")
	assert.NotContains(t, err.Error(), "locked")
}

func TestAuthenticateUser_MFAChallenge_Integration(t *testing.T) {
	svc, _, userID := testAuthService(t)
	ctx := context.Background()

	// Enroll the user in MFA (MinCost backup codes; see testMFAService).
	prevCost := backupCodeBcryptCost
	backupCodeBcryptCost = 1 // bcrypt.MinCost
	t.Cleanup(func() { backupCodeBcryptCost = prevCost })
	secret, _ := setupAndConfirm(t, svc.Auth.MFA, userID)

	result, err := svc.AuthenticateUser(ctx, "pw-user", "original-password", "10.0.0.1", "ua")
	require.NoError(t, err)
	assert.True(t, result.MFARequired, "MFA-enabled users get a challenge, not a session")
	assert.Nil(t, result.User)
	require.NotEmpty(t, result.MFAChallenge)

	// The returned challenge completes with a valid TOTP code.
	gotUser, err := svc.Auth.MFA.CompleteChallenge(ctx, result.MFAChallenge, totpCode(t, secret))
	require.NoError(t, err)
	assert.Equal(t, userID, gotUser)
}

func TestSessionManagement_Integration(t *testing.T) {
	svc, _, userID := testAuthService(t)
	ctx := context.Background()

	tokenA, err := svc.CreateWebSession(ctx, userID, "10.0.0.1", "Mozilla/5.0 (X11; Linux x86_64) Firefox/120.0")
	require.NoError(t, err)
	_, err = svc.CreateWebSession(ctx, userID, "10.0.0.2", "curl/8.0")
	require.NoError(t, err)

	sessions, err := svc.ListUserSessions(ctx, userID)
	require.NoError(t, err)
	require.Len(t, sessions, 2)

	// RevokeSessionByID enforces ownership.
	assert.Error(t, svc.RevokeSessionByID(ctx, userID+999, sessions[0].ID))
	require.NoError(t, svc.RevokeSessionByID(ctx, userID, sessions[0].ID))
	sessions, err = svc.ListUserSessions(ctx, userID)
	require.NoError(t, err)
	assert.Len(t, sessions, 1)

	// RevokeSession by raw token enforces ownership too.
	assert.Error(t, svc.RevokeSession(ctx, userID+999, tokenA))
	assert.Error(t, svc.RevokeSession(ctx, userID, "unknown-token"))

	// RevokeAllSessions clears the rest.
	require.NoError(t, svc.RevokeAllSessions(ctx, userID))
	sessions, err = svc.ListUserSessions(ctx, userID)
	require.NoError(t, err)
	assert.Empty(t, sessions)
	_, _, err = svc.ValidateSession(ctx, tokenA)
	assert.Error(t, err)

	require.NoError(t, svc.UpdateLastLogin(ctx, userID))

	throttled, err := svc.IsIPThrottled(ctx, "10.0.0.1")
	require.NoError(t, err)
	assert.False(t, throttled)
}

func TestValidateAPIToken_Integration(t *testing.T) {
	svc, repo, userID := testAuthService(t)
	ctx := context.Background()

	raw, err := svc.GenerateAPIToken(ctx, userID, "validate-me", "read", 30)
	require.NoError(t, err)

	user, tok, err := svc.ValidateAPIToken(ctx, raw, "10.0.0.5")
	require.NoError(t, err)
	assert.Equal(t, userID, user.ID)
	assert.Equal(t, "read", tok.Scope)

	_, _, err = svc.ValidateAPIToken(ctx, "unknown", "10.0.0.5")
	assert.ErrorContains(t, err, "invalid token")

	// Expired token.
	expiredRaw := "expired-token-raw"
	hash := sha256.Sum256([]byte(expiredRaw))
	past := time.Now().UTC().Add(-time.Hour)
	_, err = repo.CreateAPITokenFull(ctx, userID, hex.EncodeToString(hash[:]), "old", "read", &past)
	require.NoError(t, err)
	_, _, err = svc.ValidateAPIToken(ctx, expiredRaw, "10.0.0.5")
	assert.ErrorContains(t, err, "expired")

	// Rotated token past its grace period.
	tokens, err := svc.ListUserTokens(ctx, userID)
	require.NoError(t, err)
	require.NotEmpty(t, tokens)
	require.NoError(t, repo.MarkAPITokenRotated(ctx, tok.ID, time.Now().UTC().Add(-time.Minute)))
	_, _, err = svc.ValidateAPIToken(ctx, raw, "10.0.0.5")
	assert.ErrorContains(t, err, "grace period has expired")

	require.NoError(t, svc.CleanupExpiredTokens(ctx))
}
