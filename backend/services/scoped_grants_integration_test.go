package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"padduck/repository"
)

func ptr[T any](v T) *T { return &v }

// TestScopedGrants_Integration covers delegated administration: a viewer made
// admin of one network can manage grants inside it, cannot reach outside it,
// cannot hand out what they do not hold, and nothing crosses organizations.
func TestScopedGrants_Integration(t *testing.T) {
	svc, repo, _ := testAuthService(t)
	id := svc.Ops.Identity
	base := context.Background()
	ctx := repository.WithOrg(base, 1)
	suffix := fmt.Sprint(time.Now().UnixNano())

	mk := func(name, role string) int64 {
		u, err := repo.CreateUserWithPassword(base, name+suffix, name+suffix+"@example.com", "x", role)
		require.NoError(t, err)
		return u.ID
	}
	admin, delegate, target := mk("sg-admin", "admin"), mk("sg-delegate", "viewer"), mk("sg-target", "viewer")

	n1, err := repo.CreateNetwork(ctx, "sg-n1-"+suffix, "", admin)
	require.NoError(t, err)
	n2, err := repo.CreateNetwork(ctx, "sg-n2-"+suffix, "", admin)
	require.NoError(t, err)
	s1, err := repo.CreateSubnet(ctx, n1.ID, "10.90.1.0", 24, "s1", nil, false, false)
	require.NoError(t, err)
	s2, err := repo.CreateSubnet(ctx, n2.ID, "10.90.2.0", 24, "s2", nil, false, false)
	require.NoError(t, err)
	loc, err := repo.CreateLocation(ctx, &repository.LocationParams{Name: "sg-loc-" + suffix, Type: "site", Status: "active"})
	require.NoError(t, err)
	s3, err := repo.CreateSubnetWithLocation(ctx, n2.ID, "10.90.3.0", 24, "s3", nil, false, false, &loc.ID)
	require.NoError(t, err)

	sub := func(sid int64) ResourceScope { return ResourceScope{Type: "subnet", ID: sid} }
	can := func(user int64, perm string, scopes ...ResourceScope) bool {
		return id.CheckPermission(ctx, user, perm, scopes...) == nil
	}
	grant := func(by, to int64, perm, st string, sid int64) error {
		req := CreateGrantRequest{UserID: to, Permission: perm}
		if st != "" {
			req.ScopeType, req.ScopeID = ptr(st), ptr(sid)
		}
		_, err := id.CreateGrant(ctx, by, req)
		return err
	}

	// A viewer holds nothing to start with.
	assert.False(t, can(delegate, PermV2SubnetWrite, sub(s1.ID)))

	// Admin makes the viewer an administrator of network n1.
	require.NoError(t, grant(admin, delegate, PermV2OrgWrite, "network", n1.ID))
	require.NoError(t, grant(admin, delegate, PermV2SubnetWrite, "section", n1.ID)) // legacy name

	// The grant covers the network's subnets, and only those.
	assert.True(t, can(delegate, PermV2SubnetWrite, sub(s1.ID)), "subnet in the granted network")
	assert.True(t, can(delegate, PermV2SubnetWrite, ResourceScope{Type: "section", ID: n1.ID}), "the network itself")
	assert.False(t, can(delegate, PermV2SubnetWrite, sub(s2.ID)), "subnet in another network must not match")
	assert.False(t, can(delegate, PermV2SubnetWrite, ResourceScope{Type: "section", ID: n2.ID}))
	assert.False(t, can(delegate, PermV2SubnetWrite), "no global access")
	assert.False(t, can(delegate, PermV2VLANWrite, sub(s1.ID)), "only the granted permission")

	// The delegate can delegate inside the scope...
	assert.NoError(t, grant(delegate, target, PermV2SubnetWrite, "subnet", s1.ID))
	assert.True(t, can(target, PermV2SubnetWrite, sub(s1.ID)))
	assert.False(t, can(target, PermV2SubnetWrite, sub(s2.ID)), "sub-delegation does not bleed")

	// ...but not outside it, not globally, and not what they do not hold.
	assert.ErrorIs(t, grant(delegate, target, PermV2SubnetWrite, "subnet", s2.ID), ErrGrantDenied)
	assert.ErrorIs(t, grant(delegate, target, PermV2SubnetWrite, "network", n2.ID), ErrGrantDenied)
	assert.ErrorIs(t, grant(delegate, target, PermV2SubnetWrite, "", 0), ErrGrantDenied)
	assert.ErrorIs(t, grant(delegate, target, PermV2VLANWrite, "network", n1.ID), ErrGrantDenied)
	// A plain user with the permission but no org-write cannot delegate it.
	assert.ErrorIs(t, grant(target, delegate, PermV2SubnetWrite, "subnet", s1.ID), ErrGrantDenied)

	// Validation: platform admin is never grantable, scopes must be complete and real.
	assert.Error(t, grant(admin, target, PermV2PlatformAdmin, "", 0))
	_, err = id.CreateGrant(ctx, admin, CreateGrantRequest{UserID: target, Permission: PermV2SubnetWrite, ScopeType: ptr("network")})
	assert.ErrorContains(t, err, "together")
	assert.ErrorContains(t, grant(admin, target, PermV2SubnetWrite, "bogus", 1), "unknown scope type")
	assert.ErrorContains(t, grant(admin, target, PermV2SubnetWrite, "network", 99999999), "not found")

	// A location grant covers subnets placed in it (and nothing else).
	require.NoError(t, grant(admin, target, PermV2IPAssign, "location", loc.ID))
	assert.True(t, can(target, PermV2IPAssign, sub(s3.ID)), "subnet in the granted location")
	assert.False(t, can(target, PermV2IPAssign, sub(s2.ID)), "other subnet of the same network")

	// Visibility and revocation follow the same boundary.
	require.NoError(t, grant(admin, target, PermV2VLANWrite, "network", n2.ID))
	all, err := id.ListUserGrants(ctx, admin, target)
	require.NoError(t, err)
	seen, err := id.ListUserGrants(ctx, delegate, target)
	require.NoError(t, err)
	assert.Greater(t, len(all), len(seen), "delegate sees only grants inside its scope")
	require.Len(t, seen, 1)
	assert.Equal(t, s1.ID, *seen[0].ScopeID)
	var outside, inside int64
	for _, g := range all {
		if g.Permission == PermV2VLANWrite {
			outside = g.ID
		}
		if g.ScopeType != nil && *g.ScopeType == "subnet" {
			inside = g.ID
		}
	}
	_, err = id.RevokeGrant(ctx, delegate, outside)
	assert.ErrorIs(t, err, ErrGrantDenied)
	_, err = id.RevokeGrant(ctx, delegate, inside)
	assert.NoError(t, err)
	assert.False(t, can(target, PermV2SubnetWrite, sub(s1.ID)), "revoked grant no longer applies")

	// Another organization cannot see, grant to or revoke anything here.
	orgB := repository.WithOrg(base, newScopedTestOrg(t, repo))
	_, err = id.CreateGrant(orgB, admin, CreateGrantRequest{UserID: target, Permission: PermV2IPAssign})
	assert.Error(t, err, "target is in another organization")
	_, err = id.RevokeGrant(orgB, admin, outside)
	assert.ErrorContains(t, err, "not found")
	got, err := id.ListUserGrants(orgB, admin, target)
	require.NoError(t, err)
	assert.Empty(t, got)
}

func newScopedTestOrg(t *testing.T, repo *repository.Repository) int64 {
	t.Helper()
	slug := fmt.Sprintf("sg-%d", time.Now().UnixNano())
	org, err := repo.CreateOrganization(context.Background(), "Org "+slug, slug)
	require.NoError(t, err)
	return org.ID
}
