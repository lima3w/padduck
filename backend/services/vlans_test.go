package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateVLAN_InvalidVLANID(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	tests := []struct {
		name   string
		vlanID int
	}{
		{"zero", 0},
		{"too high 4095", 4095},
		{"negative", -1},
		{"large negative", -100},
		{"way too high", 9999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.CreateVLAN(ctx, nil, tt.vlanID, "SomeVLAN", "desc")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "VLAN ID must be between 1 and 4094")
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
			_, err := svc.CreateVLAN(ctx, nil, tt.vlanID, "", "desc")
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
			_, err := svc.GetVLAN(ctx, tt.id)
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
			_, err := svc.UpdateVLAN(ctx, tt.id, "SomeName", "desc")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid VLAN ID")
		})
	}
}

func TestUpdateVLAN_EmptyName(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	// Valid id but empty name should return name-required error before hitting repo
	_, err := svc.UpdateVLAN(ctx, 1, "", "desc")
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
			err := svc.DeleteVLAN(ctx, tt.id)
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
			_, err := svc.ListVLANsByVRF(ctx, tt.vrfID)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid VRF ID")
		})
	}
}
