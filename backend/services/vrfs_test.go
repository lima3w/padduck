package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateVRF_EmptyName(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	_, err := svc.CreateVRF(ctx, "", "rd:1", "some description")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "VRF name is required")
}

func TestGetVRF_InvalidID(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	tests := []struct {
		name string
		id   int64
	}{
		{"zero", 0},
		{"negative", -1},
		{"large negative", -999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.GetVRF(ctx, tt.id)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid VRF ID")
		})
	}
}

func TestUpdateVRF_InvalidID(t *testing.T) {
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
			_, err := svc.UpdateVRF(ctx, tt.id, "SomeName", "rd:1", "desc")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid VRF ID")
		})
	}
}

func TestUpdateVRF_EmptyName(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	// Valid id but empty name should return name-required error before hitting repo
	_, err := svc.UpdateVRF(ctx, 1, "", "rd:1", "desc")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "VRF name is required")
}

func TestDeleteVRF_InvalidID(t *testing.T) {
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
			err := svc.DeleteVRF(ctx, tt.id)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid VRF ID")
		})
	}
}
