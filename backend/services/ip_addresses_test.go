package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateIPAddress_Validation(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	tests := []struct {
		name          string
		subnetID      int64
		address       string
		status        string
		errorContains string
	}{
		{
			name:          "subnetID zero",
			subnetID:      0,
			address:       "192.168.0.10",
			status:        "available",
			errorContains: "invalid subnet ID",
		},
		{
			name:          "subnetID negative",
			subnetID:      -1,
			address:       "192.168.0.10",
			status:        "available",
			errorContains: "invalid subnet ID",
		},
		{
			name:          "invalid IP address",
			subnetID:      1,
			address:       "999.x.x.x",
			status:        "available",
			errorContains: "invalid IP address",
		},
		{
			name:          "invalid status unknown",
			subnetID:      1,
			address:       "192.168.0.10",
			status:        "unknown",
			errorContains: "invalid IP status",
		},
		{
			name:          "empty status",
			subnetID:      1,
			address:       "192.168.0.10",
			status:        "",
			errorContains: "invalid IP status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.CreateIPAddress(ctx, tt.subnetID, tt.address, "", tt.status, nil, nil, nil)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errorContains)
		})
	}
}

func TestAssignIPAddress_Validation(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	t.Run("id zero returns invalid IP address ID", func(t *testing.T) {
		_, err := svc.AssignIPAddress(ctx, 0, "somedevice")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid IP address ID")
	})

	t.Run("id negative returns invalid IP address ID", func(t *testing.T) {
		_, err := svc.AssignIPAddress(ctx, -1, "somedevice")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid IP address ID")
	})

	t.Run("valid id with empty assignedTo", func(t *testing.T) {
		// id=1 passes the id guard; assignedTo="" triggers the next guard
		_, err := svc.AssignIPAddress(ctx, 1, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "assigned_to cannot be empty")
	})
}

func TestReleaseIPAddress_InvalidID(t *testing.T) {
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
			_, err := svc.ReleaseIPAddress(ctx, tt.id)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid IP address ID")
		})
	}
}

func TestDeleteIPAddress_InvalidID(t *testing.T) {
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
			err := svc.DeleteIPAddress(ctx, tt.id)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid IP address ID")
		})
	}
}

func TestGetIPAddress_InvalidID(t *testing.T) {
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
			_, err := svc.GetIPAddress(ctx, tt.id)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid IP address ID")
		})
	}
}

func TestFindNextAvailableIP_InvalidSubnetID(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	tests := []struct {
		name     string
		subnetID int64
	}{
		{"zero subnetID", 0},
		{"negative subnetID", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.FindNextAvailableIP(ctx, tt.subnetID)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid subnet ID")
		})
	}
}

func TestAllocateIPAddress_Validation(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	t.Run("subnetID zero", func(t *testing.T) {
		_, err := svc.AllocateIPAddress(ctx, 0, "device1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid subnet ID")
	})

	t.Run("subnetID negative", func(t *testing.T) {
		_, err := svc.AllocateIPAddress(ctx, -1, "device1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid subnet ID")
	})

	t.Run("valid subnetID with empty assignedTo", func(t *testing.T) {
		_, err := svc.AllocateIPAddress(ctx, 1, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "assigned_to cannot be empty")
	})
}

func TestGetSubnetUtilization_InvalidSubnetID(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	tests := []struct {
		name     string
		subnetID int64
	}{
		{"zero subnetID", 0},
		{"negative subnetID", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.GetSubnetUtilization(ctx, tt.subnetID)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid subnet ID")
		})
	}
}
