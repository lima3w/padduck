package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// validateUsername — pure function, no repository required
// ---------------------------------------------------------------------------

func TestValidateUsername_Valid(t *testing.T) {
	cases := []string{
		"alice",
		"Bob123",
		"user_name",
		"user-name",
		"abc",
		"a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6",
	}
	for _, username := range cases {
		t.Run(username, func(t *testing.T) {
			err := validateUsername(username)
			assert.NoError(t, err, "expected valid username %q to pass", username)
		})
	}
}

func TestValidateUsername_Invalid_TooShort(t *testing.T) {
	err := validateUsername("ab")
	assert.ErrorIs(t, err, ErrInvalidUsername, "2-char username should fail")
}

func TestValidateUsername_Invalid_Empty(t *testing.T) {
	err := validateUsername("")
	assert.ErrorIs(t, err, ErrInvalidUsername, "empty username should fail")
}

func TestValidateUsername_Invalid_TooLong(t *testing.T) {
	// 33 characters — one over the limit
	err := validateUsername("a23456789012345678901234567890123")
	assert.ErrorIs(t, err, ErrInvalidUsername, "33-char username should fail")
}

func TestValidateUsername_Invalid_SpecialChars(t *testing.T) {
	cases := []string{
		"user name",  // space
		"user@name",  // @
		"user.name",  // dot
		"user!name",  // exclamation
	}
	for _, u := range cases {
		t.Run(u, func(t *testing.T) {
			err := validateUsername(u)
			assert.ErrorIs(t, err, ErrInvalidUsername, "username %q with special chars should fail", u)
		})
	}
}

// ---------------------------------------------------------------------------
// validateEmail — pure function, no repository required
// ---------------------------------------------------------------------------

func TestValidateEmail_Valid(t *testing.T) {
	cases := []string{
		"user@example.com",
		"User@Example.COM",
		"user+tag@example.org",
		"user.name@sub.domain.io",
	}
	for _, email := range cases {
		t.Run(email, func(t *testing.T) {
			err := validateEmail(email)
			assert.NoError(t, err, "expected valid email %q to pass", email)
		})
	}
}

func TestValidateEmail_Invalid_Empty(t *testing.T) {
	err := validateEmail("")
	assert.ErrorIs(t, err, ErrInvalidEmail, "empty email should fail")
}

func TestValidateEmail_Invalid_NoAtSign(t *testing.T) {
	err := validateEmail("notanemail")
	assert.ErrorIs(t, err, ErrInvalidEmail)
}

func TestValidateEmail_Invalid_NoDomain(t *testing.T) {
	err := validateEmail("user@")
	assert.ErrorIs(t, err, ErrInvalidEmail)
}

func TestValidateEmail_Invalid_NoTLD(t *testing.T) {
	err := validateEmail("user@example")
	assert.ErrorIs(t, err, ErrInvalidEmail)
}

// ---------------------------------------------------------------------------
// Exported error sentinels
// ---------------------------------------------------------------------------

func TestRegistrationErrorSentinels_NotNil(t *testing.T) {
	sentinels := []error{
		ErrRegistrationDisabled,
		ErrUsernameTaken,
		ErrEmailTaken,
		ErrInvalidUsername,
		ErrInvalidEmail,
		ErrPasswordTooShort,
		ErrEmailNotVerified,
		ErrPendingApproval,
		ErrAccountRejected,
		ErrAccountDisabled,
		ErrVerificationInvalid,
		ErrVerificationAlreadyUsed,
	}
	for _, e := range sentinels {
		assert.NotNil(t, e, "sentinel must not be nil")
		assert.NotEmpty(t, e.Error(), "sentinel must have a message")
	}
}

// ---------------------------------------------------------------------------
// RegistrationService.Register — validation paths with nil repository
//
// ConfigService.IsRegistrationEnabled returns true when repo is nil (safe
// default), so validation runs before the repository is touched.
// ---------------------------------------------------------------------------

func newRegistrationSvcNilRepo() *RegistrationService {
	configSvc := NewConfigService(nil) // nil repo → registration enabled by default
	return NewRegistrationService(nil, configSvc, nil)
}

func TestRegister_EmptyUsername_ReturnsErrInvalidUsername(t *testing.T) {
	svc := newRegistrationSvcNilRepo()
	_, err := svc.Register(context.Background(), RegisterRequest{
		Username: "",
		Email:    "valid@example.com",
		Password: "password123",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidUsername)
}

func TestRegister_EmptyEmail_ReturnsErrInvalidEmail(t *testing.T) {
	svc := newRegistrationSvcNilRepo()
	_, err := svc.Register(context.Background(), RegisterRequest{
		Username: "validuser",
		Email:    "",
		Password: "password123",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidEmail)
}

func TestRegister_InvalidEmail_ReturnsErrInvalidEmail(t *testing.T) {
	svc := newRegistrationSvcNilRepo()
	_, err := svc.Register(context.Background(), RegisterRequest{
		Username: "validuser",
		Email:    "not-an-email",
		Password: "password123",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidEmail)
}

func TestRegister_ShortPassword_ReturnsErrPasswordTooShort(t *testing.T) {
	svc := newRegistrationSvcNilRepo()
	_, err := svc.Register(context.Background(), RegisterRequest{
		Username: "validuser",
		Email:    "valid@example.com",
		Password: "short",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrPasswordTooShort)
}

func TestRegister_SevenCharPassword_ReturnsErrPasswordTooShort(t *testing.T) {
	svc := newRegistrationSvcNilRepo()
	_, err := svc.Register(context.Background(), RegisterRequest{
		Username: "validuser",
		Email:    "valid@example.com",
		Password: "seven77",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrPasswordTooShort)
}

func TestRegister_InvalidUsername_ReturnsErrInvalidUsername(t *testing.T) {
	svc := newRegistrationSvcNilRepo()
	_, err := svc.Register(context.Background(), RegisterRequest{
		Username: "ab", // too short
		Email:    "valid@example.com",
		Password: "password123",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidUsername)
}

// ---------------------------------------------------------------------------
// hashToken — pure helper, deterministic output
// ---------------------------------------------------------------------------

func TestHashToken_Deterministic(t *testing.T) {
	h1 := hashToken("test-token")
	h2 := hashToken("test-token")
	assert.Equal(t, h1, h2, "hashToken must be deterministic")
}

func TestHashToken_DifferentInputsDifferentOutputs(t *testing.T) {
	h1 := hashToken("token-a")
	h2 := hashToken("token-b")
	assert.NotEqual(t, h1, h2, "different tokens must produce different hashes")
}

func TestHashToken_NonEmpty(t *testing.T) {
	h := hashToken("anything")
	assert.NotEmpty(t, h)
}
