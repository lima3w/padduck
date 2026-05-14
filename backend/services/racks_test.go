package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"ipam-next/repository"
)

// ---------------------------------------------------------------------------
// CreateRack — input validation (no DB needed)
// ---------------------------------------------------------------------------

func TestCreateRack_EmptyName_ReturnsError(t *testing.T) {
	svc := &Service{}
	_, err := svc.CreateRack(context.Background(), &repository.RackParams{Name: "", SizeU: 42})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")
}

func TestCreateRack_ZeroSizeU_DefaultsTo42(t *testing.T) {
	svc := &Service{}
	req := &repository.RackParams{Name: "Rack A", SizeU: 0}
	// Validation passes; repo is nil so call will panic — recover it.
	func() {
		defer func() { recover() }()
		_, _ = svc.CreateRack(context.Background(), req)
	}()
	assert.Equal(t, 42, req.SizeU)
}

func TestCreateRack_NegativeSizeU_DefaultsTo42(t *testing.T) {
	svc := &Service{}
	req := &repository.RackParams{Name: "Rack B", SizeU: -1}
	func() {
		defer func() { recover() }()
		_, _ = svc.CreateRack(context.Background(), req)
	}()
	assert.Equal(t, 42, req.SizeU)
}

// ---------------------------------------------------------------------------
// GetRack — input validation
// ---------------------------------------------------------------------------

func TestGetRack_InvalidID_ReturnsError(t *testing.T) {
	svc := &Service{}
	_, err := svc.GetRack(context.Background(), 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid rack ID")

	_, err = svc.GetRack(context.Background(), -1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid rack ID")
}

// ---------------------------------------------------------------------------
// UpdateRack — input validation
// ---------------------------------------------------------------------------

func TestUpdateRack_InvalidID_ReturnsError(t *testing.T) {
	svc := &Service{}
	_, err := svc.UpdateRack(context.Background(), 0, &repository.RackParams{Name: "X"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid rack ID")
}

func TestUpdateRack_EmptyName_ReturnsError(t *testing.T) {
	svc := &Service{}
	_, err := svc.UpdateRack(context.Background(), 1, &repository.RackParams{Name: ""})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")
}

func TestUpdateRack_ZeroSizeU_DefaultsTo42(t *testing.T) {
	svc := &Service{}
	req := &repository.RackParams{Name: "Rack A", SizeU: 0}
	func() {
		defer func() { recover() }()
		_, _ = svc.UpdateRack(context.Background(), 1, req)
	}()
	assert.Equal(t, 42, req.SizeU)
}

// ---------------------------------------------------------------------------
// DeleteRack — input validation
// ---------------------------------------------------------------------------

func TestDeleteRack_InvalidID_ReturnsError(t *testing.T) {
	svc := &Service{}
	err := svc.DeleteRack(context.Background(), 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid rack ID")
}

// ---------------------------------------------------------------------------
// ListDevicesInRack — input validation
// ---------------------------------------------------------------------------

func TestListDevicesInRack_InvalidID_ReturnsError(t *testing.T) {
	svc := &Service{}
	_, err := svc.ListDevicesInRack(context.Background(), 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid rack ID")

	_, err = svc.ListDevicesInRack(context.Background(), -3)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid rack ID")
}
