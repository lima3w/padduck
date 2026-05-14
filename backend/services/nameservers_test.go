package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"ipam-next/repository"
)

// ---------------------------------------------------------------------------
// CreateNameserver — input validation (no DB needed)
// ---------------------------------------------------------------------------

func TestCreateNameserver_EmptyName_ReturnsError(t *testing.T) {
	svc := &Service{}
	_, err := svc.CreateNameserver(context.Background(), &repository.NameserverParams{
		Name: "", Server1: "8.8.8.8",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")
}

func TestCreateNameserver_EmptyServer1_ReturnsError(t *testing.T) {
	svc := &Service{}
	_, err := svc.CreateNameserver(context.Background(), &repository.NameserverParams{
		Name: "Google DNS", Server1: "",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "server1 is required")
}

// ---------------------------------------------------------------------------
// GetNameserver — input validation
// ---------------------------------------------------------------------------

func TestGetNameserver_InvalidID_ReturnsError(t *testing.T) {
	svc := &Service{}

	_, err := svc.GetNameserver(context.Background(), 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid nameserver ID")

	_, err = svc.GetNameserver(context.Background(), -3)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid nameserver ID")
}

// ---------------------------------------------------------------------------
// UpdateNameserver — input validation
// ---------------------------------------------------------------------------

func TestUpdateNameserver_InvalidID_ReturnsError(t *testing.T) {
	svc := &Service{}
	_, err := svc.UpdateNameserver(context.Background(), 0, &repository.NameserverParams{Name: "X", Server1: "1.1.1.1"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid nameserver ID")
}

func TestUpdateNameserver_EmptyName_ReturnsError(t *testing.T) {
	svc := &Service{}
	_, err := svc.UpdateNameserver(context.Background(), 1, &repository.NameserverParams{Name: "", Server1: "1.1.1.1"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")
}

func TestUpdateNameserver_EmptyServer1_ReturnsError(t *testing.T) {
	svc := &Service{}
	_, err := svc.UpdateNameserver(context.Background(), 1, &repository.NameserverParams{Name: "Cloudflare", Server1: ""})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "server1 is required")
}

// ---------------------------------------------------------------------------
// DeleteNameserver — input validation
// ---------------------------------------------------------------------------

func TestDeleteNameserver_InvalidID_ReturnsError(t *testing.T) {
	svc := &Service{}
	err := svc.DeleteNameserver(context.Background(), 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid nameserver ID")
}
