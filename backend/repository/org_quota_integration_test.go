package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"padduck/models"
)

// TestOrgQuota_Integration checks subnet and IP limits are per organization,
// that unlimited and system contexts are unaffected, and that the audit prune
// leaves an org with a longer retention override alone.
func TestOrgQuota_Integration(t *testing.T) {
	r := testRepo(t)
	base := context.Background()
	orgA, orgB := newTestOrg(t, r, "quota-a"), newTestOrg(t, r, "quota-b")
	a, b := WithOrg(base, orgA), WithOrg(base, orgB)
	uid := createTestUser(t, r)

	one, two := 1, 2
	_, err := r.UpsertOrganizationSettings(base, &models.OrganizationSettings{
		OrganizationID: orgA, MaxSubnets: &one, MaxIPAddresses: &two,
	})
	require.NoError(t, err)

	netA, err := r.CreateNetwork(a, "q-net-a", "", uid)
	require.NoError(t, err)
	subA, err := r.CreateSubnet(a, netA.ID, "10.70.0.0", 24, "q1", nil, false, false)
	require.NoError(t, err)

	var qe *QuotaExceededError
	_, err = r.CreateSubnet(a, netA.ID, "10.70.1.0", 24, "q2", nil, false, false)
	require.True(t, errors.As(err, &qe), "second subnet should hit the limit: %v", err)
	assert.Equal(t, "subnets", qe.Resource)
	assert.Equal(t, 1, qe.Limit)

	_, err = r.CreateIPAddress(a, subA.ID, "10.70.0.1", "", "assigned", nil, nil, nil, nil)
	require.NoError(t, err)
	_, err = r.CreateIPAddress(a, subA.ID, "10.70.0.2", "", "assigned", nil, nil, nil, nil)
	require.NoError(t, err)
	_, err = r.CreateIPAddress(a, subA.ID, "10.70.0.3", "", "assigned", nil, nil, nil, nil)
	require.True(t, errors.As(err, &qe), "third IP should hit the limit: %v", err)
	assert.Equal(t, "ip_addresses", qe.Resource)

	// Org B has no settings row, so it is unlimited, and A's usage is not B's.
	netB, err := r.CreateNetwork(b, "q-net-b", "", uid)
	require.NoError(t, err)
	for i, cidr := range []string{"10.71.0.0", "10.71.1.0", "10.71.2.0"} {
		_, err = r.CreateSubnet(b, netB.ID, cidr, 24, "b", nil, false, false)
		require.NoError(t, err, "org B subnet %d", i)
	}

	// A system context (no org) is never limited.
	_, err = r.CreateSubnet(base, netA.ID, "10.72.0.0", 24, "sys", nil, false, false)
	assert.False(t, errors.As(err, &qe))

	// Raising the limit unblocks creation.
	_, err = r.UpsertOrganizationSettings(base, &models.OrganizationSettings{OrganizationID: orgA, MaxSubnets: &two})
	require.NoError(t, err)
	_, err = r.CreateSubnet(a, netA.ID, "10.70.1.0", 24, "q2", nil, false, false)
	assert.NoError(t, err)

	// Audit retention: the global prune skips an org with its own override.
	_, err = r.db.Exec(base, `INSERT INTO audit_logs (organization_id, action, resource_type, created_at)
		VALUES ($1, 'quota_old', 'x', NOW() - INTERVAL '40 days'), ($2, 'quota_old', 'x', NOW() - INTERVAL '40 days')`, orgA, orgB)
	require.NoError(t, err)
	days := 90
	_, err = r.UpsertOrganizationSettings(base, &models.OrganizationSettings{OrganizationID: orgA, AuditRetentionDays: &days})
	require.NoError(t, err)
	_, err = r.PruneAuditLogs(base, 30)
	require.NoError(t, err)
	var nA, nB int
	require.NoError(t, r.db.QueryRow(base, `SELECT COUNT(*) FROM audit_logs WHERE action='quota_old' AND organization_id=$1`, orgA).Scan(&nA))
	require.NoError(t, r.db.QueryRow(base, `SELECT COUNT(*) FROM audit_logs WHERE action='quota_old' AND organization_id=$1`, orgB).Scan(&nB))
	assert.Equal(t, 1, nA, "org with a 90 day override keeps 40 day old logs")
	assert.Equal(t, 0, nB, "org on the global 30 day window loses them")
}
