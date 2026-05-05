package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateUser(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	testCases := []struct {
		name      string
		username  string
		email     string
		shouldErr bool
		errMsg    string
	}{
		{
			name:      "missing username",
			username:  "",
			email:     "user@example.com",
			shouldErr: true,
			errMsg:    "username is required",
		},
		{
			name:      "missing email",
			username:  "testuser",
			email:     "",
			shouldErr: true,
			errMsg:    "email is required",
		},
		{
			name:      "invalid email format",
			username:  "testuser",
			email:     "not-an-email",
			shouldErr: true,
			errMsg:    "invalid email format",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.CreateUser(ctx, tc.username, tc.email)
			if tc.shouldErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetUser(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	// Test validation - invalid ID
	_, err := svc.GetUser(ctx, -1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid user ID")
}
