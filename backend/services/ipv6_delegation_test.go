package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"ipam-next/models"
)

// Tests for pure validation logic in ipv6_delegation.go that fires before
// any repository call, so a nil-repository Service is safe to use here.

var nilRepoSvc = &Service{}

// ---------------------------------------------------------------------------
// ListDelegations — validates subnetID > 0
// ---------------------------------------------------------------------------

func TestListDelegations_InvalidSubnetID_ReturnsError(t *testing.T) {
	_, err := nilRepoSvc.ListDelegations(context.Background(), 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid subnet ID")
}

func TestListDelegations_NegativeSubnetID_ReturnsError(t *testing.T) {
	_, err := nilRepoSvc.ListDelegations(context.Background(), -5)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid subnet ID")
}

// ---------------------------------------------------------------------------
// CreateDelegation — validates ParentSubnetID and DelegatedPrefix
// ---------------------------------------------------------------------------

func TestCreateDelegation_InvalidParentSubnetID_ReturnsError(t *testing.T) {
	d := &models.IPv6Delegation{ParentSubnetID: 0, DelegatedPrefix: "2001:db8::/48"}
	_, err := nilRepoSvc.CreateDelegation(context.Background(), d)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid parent subnet ID")
}

func TestCreateDelegation_MissingPrefix_ReturnsError(t *testing.T) {
	d := &models.IPv6Delegation{ParentSubnetID: 1, DelegatedPrefix: ""}
	_, err := nilRepoSvc.CreateDelegation(context.Background(), d)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delegated prefix is required")
}

// ---------------------------------------------------------------------------
// UpdateDelegation — validates id and DelegatedPrefix
// ---------------------------------------------------------------------------

func TestUpdateDelegation_InvalidID_ReturnsError(t *testing.T) {
	d := &models.IPv6Delegation{DelegatedPrefix: "2001:db8::/48"}
	_, err := nilRepoSvc.UpdateDelegation(context.Background(), 0, d)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid delegation ID")
}

func TestUpdateDelegation_MissingPrefix_ReturnsError(t *testing.T) {
	d := &models.IPv6Delegation{DelegatedPrefix: ""}
	_, err := nilRepoSvc.UpdateDelegation(context.Background(), 1, d)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delegated prefix is required")
}

// ---------------------------------------------------------------------------
// DeleteDelegation — validates id > 0
// ---------------------------------------------------------------------------

func TestDeleteDelegation_InvalidID_ReturnsError(t *testing.T) {
	err := nilRepoSvc.DeleteDelegation(context.Background(), 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid delegation ID")
}

func TestDeleteDelegation_NegativeID_ReturnsError(t *testing.T) {
	err := nilRepoSvc.DeleteDelegation(context.Background(), -1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid delegation ID")
}

// ---------------------------------------------------------------------------
// isExpiredNow — pure helper
// ---------------------------------------------------------------------------

func TestIsExpiredNow_NilExpiresAt_ReturnsFalse(t *testing.T) {
	assert.False(t, isExpiredNow(nil))
}

func TestIsExpiredNow_FutureTime_ReturnsFalse(t *testing.T) {
	future := time.Now().Add(1 * time.Hour)
	assert.False(t, isExpiredNow(&future))
}

func TestIsExpiredNow_PastTime_ReturnsTrue(t *testing.T) {
	past := time.Now().Add(-1 * time.Hour)
	assert.True(t, isExpiredNow(&past))
}
