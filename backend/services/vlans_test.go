package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"padduck/models"
)

func TestCreateVLAN_InvalidVLANID(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	tests := []struct {
		name   string
		vlanID int
	}{
		// 0 is now valid (native/default VLAN)
		{"too high 4095", 4095},
		{"negative", -1},
		{"large negative", -100},
		{"way too high", 9999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Ops.IPAM.CreateVLAN(ctx, nil, nil, nil, tt.vlanID, "SomeVLAN", "desc")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "VLAN ID must be between 0 and 4094")
		})
	}
}

func TestCreateVLAN_EmptyName(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	tests := []struct {
		name   string
		vlanID int
	}{
		{"vlanID=1 empty name", 1},
		{"vlanID=4094 empty name", 4094},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Ops.IPAM.CreateVLAN(ctx, nil, nil, nil, tt.vlanID, "", "desc")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "VLAN name is required")
		})
	}
}

func TestGetVLAN_InvalidID(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	tests := []struct {
		name string
		id   int64
	}{
		{"zero", 0},
		{"negative", -1},
		{"large negative", -100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Ops.IPAM.GetVLAN(ctx, tt.id)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid VLAN ID")
		})
	}
}

func TestUpdateVLAN_InvalidID(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	tests := []struct {
		name string
		id   int64
	}{
		{"zero", 0},
		{"negative", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Ops.IPAM.UpdateVLAN(ctx, tt.id, nil, nil, "SomeName", "desc")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid VLAN ID")
		})
	}
}

func TestUpdateVLAN_EmptyName(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	// Valid id but empty name should return name-required error before hitting repo
	_, err := svc.Ops.IPAM.UpdateVLAN(ctx, 1, nil, nil, "", "desc")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "VLAN name is required")
}

func TestDeleteVLAN_InvalidID(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	tests := []struct {
		name string
		id   int64
	}{
		{"zero", 0},
		{"negative", -1},
		{"large negative", -50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.Ops.IPAM.DeleteVLAN(ctx, tt.id)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid VLAN ID")
		})
	}
}

func TestListVLANsByVRF_InvalidVRFID(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	tests := []struct {
		name  string
		vrfID int64
	}{
		{"zero", 0},
		{"negative", -1},
		{"large negative", -999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Ops.IPAM.ListVLANsByVRF(ctx, tt.vrfID)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid VRF ID")
		})
	}
}

// VLANUsageReport model structure test

func TestVLANUsageEntry_Fields(t *testing.T) {
	entry := &models.VLANUsageEntry{
		VLANID:         1,
		VLANName:       "Management",
		VLANTag:        10,
		SubnetCount:    3,
		IPCount:        42,
		TotalIPs:       256,
		UtilizationPct: 16.41,
	}
	assert.Equal(t, int64(1), entry.VLANID)
	assert.Equal(t, "Management", entry.VLANName)
	assert.Equal(t, 10, entry.VLANTag)
	assert.Equal(t, int64(3), entry.SubnetCount)
	assert.Equal(t, int64(42), entry.IPCount)
	assert.Equal(t, int64(256), entry.TotalIPs)
	assert.InDelta(t, 16.41, entry.UtilizationPct, 0.01)
}

func TestVLANUsageReport_Fields(t *testing.T) {
	report := &models.VLANUsageReport{
		Entries:     []*models.VLANUsageEntry{},
		GeneratedAt: "2026-05-14T00:00:00Z",
	}
	assert.NotNil(t, report.Entries)
	assert.Equal(t, "2026-05-14T00:00:00Z", report.GeneratedAt)
}

// GetVLANSubnets service unit tests

func TestGetVLANSubnets_InvalidID(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	tests := []struct {
		name   string
		vlanID int64
	}{
		{"zero", 0},
		{"negative", -1},
		{"large negative", -100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Ops.IPAM.GetVLANSubnets(ctx, tt.vlanID)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid VLAN ID")
		})
	}
}

func TestAssignSubnetToVLAN_InvalidIDs(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	tests := []struct {
		name     string
		vlanID   int64
		subnetID int64
		want     string
	}{
		{"bad vlan", 0, 1, "invalid VLAN ID"},
		{"bad subnet", 1, 0, "invalid subnet ID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Ops.IPAM.AssignSubnetToVLAN(ctx, tt.vlanID, tt.subnetID)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.want)
		})
	}
}

func TestRemoveSubnetFromVLAN_InvalidIDs(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	tests := []struct {
		name     string
		vlanID   int64
		subnetID int64
		want     string
	}{
		{"bad vlan", 0, 1, "invalid VLAN ID"},
		{"bad subnet", 1, 0, "invalid subnet ID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Ops.IPAM.RemoveSubnetFromVLAN(ctx, tt.vlanID, tt.subnetID)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.want)
		})
	}
}

// VLANDomain service unit tests

func TestCreateVLANDomain_EmptyName(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	_, err := svc.Ops.IPAM.CreateVLANDomain(ctx, "", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "VLAN domain name is required")
}

func TestGetVLANDomain_InvalidID(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	tests := []struct {
		name string
		id   int64
	}{
		{"zero", 0},
		{"negative", -1},
		{"large negative", -100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Ops.IPAM.GetVLANDomain(ctx, tt.id)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid VLAN domain ID")
		})
	}
}

func TestUpdateVLANDomain_InvalidID(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	tests := []struct {
		name string
		id   int64
	}{
		{"zero", 0},
		{"negative", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Ops.IPAM.UpdateVLANDomain(ctx, tt.id, "Foo", nil)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid VLAN domain ID")
		})
	}
}

func TestUpdateVLANDomain_EmptyName(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	_, err := svc.Ops.IPAM.UpdateVLANDomain(ctx, 1, "", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "VLAN domain name is required")
}

// VLANGroup service unit tests

func TestCreateVLANGroup_EmptyName(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	_, err := svc.Ops.IPAM.CreateVLANGroup(ctx, "", nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "VLAN group name is required")
}

func TestGetVLANGroup_InvalidID(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	tests := []struct {
		name string
		id   int64
	}{
		{"zero", 0},
		{"negative", -1},
		{"large negative", -100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Ops.IPAM.GetVLANGroup(ctx, tt.id)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid VLAN group ID")
		})
	}
}

func TestUpdateVLANGroup_InvalidID(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	tests := []struct {
		name string
		id   int64
	}{
		{"zero", 0},
		{"negative", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Ops.IPAM.UpdateVLANGroup(ctx, tt.id, "Foo", nil, nil)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid VLAN group ID")
		})
	}
}

func TestUpdateVLANGroup_EmptyName(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	_, err := svc.Ops.IPAM.UpdateVLANGroup(ctx, 1, "", nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "VLAN group name is required")
}

func TestDeleteVLANGroup_InvalidID(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	tests := []struct {
		name string
		id   int64
	}{
		{"zero", 0},
		{"negative", -1},
		{"large negative", -50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.Ops.IPAM.DeleteVLANGroup(ctx, tt.id)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid VLAN group ID")
		})
	}
}

func TestDeleteVLANDomain_InvalidID(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	tests := []struct {
		name string
		id   int64
	}{
		{"zero", 0},
		{"negative", -1},
		{"large negative", -50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.Ops.IPAM.DeleteVLANDomain(ctx, tt.id)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid VLAN domain ID")
		})
	}
}
