package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateAPIToken_Validation(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	tests := []struct {
		name          string
		userID        int64
		tokenName     string
		errorContains string
	}{
		{
			name:          "userID zero",
			userID:        0,
			tokenName:     "my-token",
			errorContains: "invalid user ID",
		},
		{
			name:          "userID negative",
			userID:        -1,
			tokenName:     "my-token",
			errorContains: "invalid user ID",
		},
		{
			name:          "empty token name",
			userID:        1,
			tokenName:     "",
			errorContains: "token name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.GenerateAPIToken(ctx, tt.userID, tt.tokenName)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errorContains)
		})
	}
}

func TestValidateAPIToken_Validation(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	t.Run("empty token", func(t *testing.T) {
		_, err := svc.ValidateAPIToken(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token is required")
	})
}

func TestRevokeAPIToken_Validation(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	tests := []struct {
		name    string
		tokenID int64
	}{
		{"tokenID zero", 0},
		{"tokenID negative", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.RevokeAPIToken(ctx, tt.tokenID)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid token ID")
		})
	}
}

func TestListUserTokens_Validation(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	tests := []struct {
		name   string
		userID int64
	}{
		{"userID zero", 0},
		{"userID negative", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.ListUserTokens(ctx, tt.userID)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid user ID")
		})
	}
}

func TestCreatePasswordResetToken_Validation(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	t.Run("empty email", func(t *testing.T) {
		_, err := svc.CreatePasswordResetToken(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "email is required")
	})
}

func TestResetPasswordWithToken_Validation(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	tests := []struct {
		name            string
		token           string
		newPasswordHash string
		errorContains   string
	}{
		{
			name:            "empty token",
			token:           "",
			newPasswordHash: "somehash",
			errorContains:   "token and password hash required",
		},
		{
			name:            "empty password hash",
			token:           "sometoken",
			newPasswordHash: "",
			errorContains:   "token and password hash required",
		},
		{
			name:            "both empty",
			token:           "",
			newPasswordHash: "",
			errorContains:   "token and password hash required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.ResetPasswordWithToken(ctx, tt.token, tt.newPasswordHash)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errorContains)
		})
	}
}

func TestAuthenticateUser_Validation(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	tests := []struct {
		name          string
		username      string
		password      string
		errorContains string
	}{
		{
			name:          "empty username",
			username:      "",
			password:      "secret",
			errorContains: "username and password required",
		},
		{
			name:          "empty password",
			username:      "alice",
			password:      "",
			errorContains: "username and password required",
		},
		{
			name:          "both empty",
			username:      "",
			password:      "",
			errorContains: "username and password required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.AuthenticateUser(ctx, tt.username, tt.password, "127.0.0.1", "test-agent")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errorContains)
		})
	}
}
