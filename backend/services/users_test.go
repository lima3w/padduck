package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestSuspendUser_InvalidID(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	err := svc.SuspendUser(ctx, 0, 1, "test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid user ID")

	err = svc.SuspendUser(ctx, -1, 1, "test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid user ID")
}

func TestUnsuspendUser_InvalidID(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	err := svc.UnsuspendUser(ctx, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid user ID")
}


func TestBulkImportUsersValidation(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	tests := []struct {
		name    string
		records []BulkUserImportRecord
		wantErr string
	}{
		{
			name:    "empty username gets error in result",
			records: []BulkUserImportRecord{{Username: "", Email: "a@b.com"}},
			wantErr: "username and email required",
		},
		{
			name:    "empty email gets error in result",
			records: []BulkUserImportRecord{{Username: "user1", Email: ""}},
			wantErr: "username and email required",
		},
		{
			name:    "invalid role gets error in result",
			records: []BulkUserImportRecord{{Username: "user1", Email: "a@b.com", Role: "superadmin"}},
			wantErr: "invalid role",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := svc.BulkImportUsers(ctx, tt.records, "pass")
			assert.NoError(t, err) // function-level error is nil; errors are per-record
			require.Len(t, results, 1)
			assert.Contains(t, results[0].Error, tt.wantErr)
		})
	}
}

func TestValidateCron(t *testing.T) {
	assert.NoError(t, validateCron("* * * * *"))
	assert.NoError(t, validateCron("0 12 * * 1"))
	assert.Error(t, validateCron(""))
	assert.Error(t, validateCron("* * *"))
	assert.Error(t, validateCron("* * * * * *"))
}

func TestEnumerateCIDR(t *testing.T) {
	ips, err := enumerateCIDR("192.168.1.0", 24)
	assert.NoError(t, err)
	assert.Equal(t, 254, len(ips))
	assert.Equal(t, "192.168.1.1", ips[0])
	assert.Equal(t, "192.168.1.254", ips[len(ips)-1])
}

func TestEnumerateCIDR_Slash30(t *testing.T) {
	ips, err := enumerateCIDR("10.0.0.0", 30)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(ips))
}

func TestEnumerateCIDR_Invalid(t *testing.T) {
	_, err := enumerateCIDR("not-an-ip", 24)
	assert.Error(t, err)
}
