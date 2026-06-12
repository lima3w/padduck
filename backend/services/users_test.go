package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"image"
	"image/jpeg"
	"image/png"
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

func TestUpdateUserEmail_Validation(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	t.Run("invalid user ID", func(t *testing.T) {
		err := svc.UpdateUserEmail(ctx, 0, "valid@example.com")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid user ID")
	})

	t.Run("invalid email", func(t *testing.T) {
		err := svc.UpdateUserEmail(ctx, 1, "not-an-email")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid email")
	})
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
			results, err := svc.BulkImportUsers(ctx, tt.records)
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

func encodeTestPNG(t *testing.T, width, height int) string {
	t.Helper()
	var buf bytes.Buffer
	require.NoError(t, png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, width, height))))
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

func TestValidateAvatarImage(t *testing.T) {
	pngPayload := encodeTestPNG(t, 64, 64)

	var jpegBuf bytes.Buffer
	require.NoError(t, jpeg.Encode(&jpegBuf, image.NewRGBA(image.Rect(0, 0, 64, 64)), nil))
	jpegPayload := base64.StdEncoding.EncodeToString(jpegBuf.Bytes())

	scriptPayload := base64.StdEncoding.EncodeToString([]byte("#!/bin/sh\necho pwned\n"))

	t.Run("valid png", func(t *testing.T) {
		out, err := validateAvatarImage("data:image/png;base64," + pngPayload)
		require.NoError(t, err)
		assert.Equal(t, "data:image/png;base64,"+pngPayload, out)
	})

	t.Run("declared type is rewritten to real format", func(t *testing.T) {
		// JPEG bytes declared as PNG: must store the real type
		out, err := validateAvatarImage("data:image/png;base64," + jpegPayload)
		require.NoError(t, err)
		assert.Equal(t, "data:image/jpeg;base64,"+jpegPayload, out)
	})

	t.Run("script disguised as png", func(t *testing.T) {
		_, err := validateAvatarImage("data:image/png;base64," + scriptPayload)
		assert.ErrorContains(t, err, "not a recognized image")
	})

	t.Run("svg rejected", func(t *testing.T) {
		svg := base64.StdEncoding.EncodeToString([]byte(`<svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script></svg>`))
		_, err := validateAvatarImage("data:image/svg+xml;base64," + svg)
		assert.ErrorContains(t, err, "not a recognized image")
	})

	t.Run("invalid base64", func(t *testing.T) {
		_, err := validateAvatarImage("data:image/png;base64,!!!not-base64!!!")
		assert.ErrorContains(t, err, "not valid base64")
	})

	t.Run("not a data url", func(t *testing.T) {
		_, err := validateAvatarImage("https://example.com/avatar.png")
		assert.ErrorContains(t, err, "data URL")
	})

	t.Run("oversized dimensions", func(t *testing.T) {
		_, err := validateAvatarImage("data:image/png;base64," + encodeTestPNG(t, 1, maxAvatarDimension+1))
		assert.ErrorContains(t, err, "dimensions")
	})
}

func TestUpdateUserAvatar_RejectsNonImage(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	data := "data:image/png;base64," + base64.StdEncoding.EncodeToString([]byte("not an image"))
	err := svc.UpdateUserAvatar(context.Background(), 1, "custom", &data)
	assert.ErrorContains(t, err, "not a recognized image")
}
