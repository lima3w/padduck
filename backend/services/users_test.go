package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateUser_Validation(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	tests := []struct {
		name          string
		username      string
		email         string
		errorContains string
	}{
		{
			name:          "missing username",
			username:      "",
			email:         "user@example.com",
			errorContains: "username is required",
		},
		{
			name:          "missing email",
			username:      "testuser",
			email:         "",
			errorContains: "email is required",
		},
		{
			name:          "invalid email not-an-email",
			username:      "testuser",
			email:         "not-an-email",
			errorContains: "invalid email format",
		},
		{
			name:          "invalid email double-at",
			username:      "testuser",
			email:         "@@",
			errorContains: "invalid email format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.CreateUser(ctx, tt.username, tt.email)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errorContains)
		})
	}
}

func TestCreateUserWithPassword_RoleValidation(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	t.Run("invalid role superadmin", func(t *testing.T) {
		_, err := svc.CreateUserWithPassword(ctx, "alice", "alice@example.com", "hash", "superadmin")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid role")
	})

	t.Run("invalid role empty string", func(t *testing.T) {
		_, err := svc.CreateUserWithPassword(ctx, "alice", "alice@example.com", "hash", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid role")
	})

	t.Run("invalid role guest", func(t *testing.T) {
		_, err := svc.CreateUserWithPassword(ctx, "alice", "alice@example.com", "hash", "guest")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid role")
	})
}

func TestCreateUserWithPassword_UsernameEmailValidation(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	t.Run("missing username", func(t *testing.T) {
		_, err := svc.CreateUserWithPassword(ctx, "", "alice@example.com", "hash", "user")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "username is required")
	})

	t.Run("missing email", func(t *testing.T) {
		_, err := svc.CreateUserWithPassword(ctx, "alice", "", "hash", "user")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "email is required")
	})

	t.Run("invalid email format", func(t *testing.T) {
		_, err := svc.CreateUserWithPassword(ctx, "alice", "not-an-email", "hash", "user")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid email format")
	})
}

func TestUpdateUserRole_Validation(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	t.Run("userID zero", func(t *testing.T) {
		_, err := svc.UpdateUserRole(ctx, 0, "admin")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid user ID")
	})

	t.Run("userID negative", func(t *testing.T) {
		_, err := svc.UpdateUserRole(ctx, -1, "admin")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid user ID")
	})

	t.Run("invalid role superadmin", func(t *testing.T) {
		_, err := svc.UpdateUserRole(ctx, 1, "superadmin")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid role")
	})

	t.Run("invalid role empty string", func(t *testing.T) {
		_, err := svc.UpdateUserRole(ctx, 1, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid role")
	})

	t.Run("invalid role guest", func(t *testing.T) {
		_, err := svc.UpdateUserRole(ctx, 1, "guest")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid role")
	})
}

func TestGetUser_InvalidID(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	tests := []struct {
		name string
		id   int64
	}{
		{"zero ID", 0},
		{"negative ID", -1},
		{"large negative", -100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.GetUser(ctx, tt.id)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid user ID")
		})
	}
}

func TestGetUserByID_InvalidID(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	tests := []struct {
		name string
		id   int64
	}{
		{"zero ID", 0},
		{"negative ID", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.GetUserByID(ctx, tt.id)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid user ID")
		})
	}
}

func TestDeleteUser_InvalidID(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	tests := []struct {
		name string
		id   int64
	}{
		{"zero ID", 0},
		{"negative ID", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.DeleteUser(ctx, tt.id)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid user ID")
		})
	}
}
