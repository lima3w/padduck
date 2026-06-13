package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBulkDeleteIPAddresses_Integration(t *testing.T) {
	r := testRepo(t)
	ctx := context.Background()

	networkID := createTestNetwork(t, r)
	subnetID := createTestSubnet(t, r, networkID, "10.77.0.0", 24)

	ip1, err := r.CreateIPAddress(ctx, subnetID, "10.77.0.1", "bulk-host-1", "available", nil, nil, nil, nil)
	require.NoError(t, err)
	ip2, err := r.CreateIPAddress(ctx, subnetID, "10.77.0.2", "bulk-host-2", "available", nil, nil, nil, nil)
	require.NoError(t, err)
	ip3, err := r.CreateIPAddress(ctx, subnetID, "10.77.0.3", "bulk-host-3", "available", nil, nil, nil, nil)
	require.NoError(t, err)

	// Deleting two of three returns exactly those two IDs.
	deleted, err := r.BulkDeleteIPAddresses(ctx, []int64{ip1.ID, ip2.ID})
	require.NoError(t, err)
	assert.ElementsMatch(t, []int64{ip1.ID, ip2.ID}, deleted)

	// The third record is untouched.
	_, err = r.GetIPAddressByID(ctx, ip3.ID)
	require.NoError(t, err)

	// Deleted records are gone.
	_, err = r.GetIPAddressByID(ctx, ip1.ID)
	assert.Error(t, err, "ip1 must not be retrievable after bulk delete")

	// Nonexistent IDs are silently ignored — empty slice, no error.
	deleted, err = r.BulkDeleteIPAddresses(ctx, []int64{99999, 99998})
	require.NoError(t, err)
	assert.Empty(t, deleted)

	// Empty input returns empty slice with no error.
	deleted, err = r.BulkDeleteIPAddresses(ctx, []int64{})
	require.NoError(t, err)
	assert.Empty(t, deleted)
}
