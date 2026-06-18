package services

import (
	"context"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
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
			_, err := svc.Ops.IPAM.CreateIPAddress(ctx, tt.subnetID, tt.address, "", tt.status, nil, nil, nil, nil)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errorContains)
		})
	}
}

func TestNormalizeCreateIPAddressError_DuplicateAddress(t *testing.T) {
	err := normalizeCreateIPAddressError(&pgconn.PgError{
		Code:           "23505",
		ConstraintName: "ip_addresses_subnet_id_address_key",
	}, "192.168.0.10")

	assert.EqualError(t, err, "IP address 192.168.0.10 already exists in this subnet")
}

func TestNormalizeCreateIPAddressError_WrappedDuplicateAddress(t *testing.T) {
	err := normalizeCreateIPAddressError(fmt.Errorf("insert failed: %w", &pgconn.PgError{
		Code:           "23505",
		ConstraintName: "ip_addresses_subnet_id_address_key",
	}), "192.168.0.10")

	assert.EqualError(t, err, "IP address 192.168.0.10 already exists in this subnet")
}

func TestNormalizeCreateIPAddressError_PreservesOtherErrors(t *testing.T) {
	original := fmt.Errorf("database unavailable")
	err := normalizeCreateIPAddressError(original, "192.168.0.10")

	assert.Same(t, original, err)
}

func TestAssignIPAddress_Validation(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	t.Run("id zero returns invalid IP address ID", func(t *testing.T) {
		_, err := svc.Ops.IPAM.AssignIPAddress(ctx, 0, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid IP address ID")
	})

	t.Run("id negative returns invalid IP address ID", func(t *testing.T) {
		_, err := svc.Ops.IPAM.AssignIPAddress(ctx, -1, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid IP address ID")
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
			_, err := svc.Ops.IPAM.ReleaseIPAddress(ctx, tt.id)
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
			err := svc.Ops.IPAM.DeleteIPAddress(ctx, tt.id)
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
			_, err := svc.Ops.IPAM.GetIPAddress(ctx, tt.id)
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
			_, err := svc.Ops.IPAM.FindNextAvailableIP(ctx, tt.subnetID)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid subnet ID")
		})
	}
}

func TestAllocateIPAddress_Validation(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	t.Run("subnetID zero", func(t *testing.T) {
		_, err := svc.Ops.IPAM.AllocateIPAddress(ctx, 0, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid subnet ID")
	})

	t.Run("subnetID negative", func(t *testing.T) {
		_, err := svc.Ops.IPAM.AllocateIPAddress(ctx, -1, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid subnet ID")
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
			_, err := svc.Ops.IPAM.GetSubnetUtilization(ctx, tt.subnetID)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid subnet ID")
		})
	}
}
