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

// TestIntentFlow_Integration covers the desired-state intent model: auto-approve
// applies synchronously, manual review applies only on approval, a rejected or
// already-reviewed intent changes nothing, the diff shows what a pending update
// would change, and nothing crosses organizations.
func TestIntentFlow_Integration(t *testing.T) {
	svc, repo, _ := testAuthService(t)
	in := svc.Ops.Intent
	base := context.Background()
	ctx := repository.WithOrg(base, 1)
	suffix := fmt.Sprint(time.Now().UnixNano())

	admin, err := repo.CreateUserWithPassword(base, "in-admin"+suffix, "in-admin"+suffix+"@example.com", "x", "admin")
	require.NoError(t, err)
	net, err := repo.CreateNetwork(ctx, "in-net-"+suffix, "", admin.ID)
	require.NoError(t, err)

	subnetState := func(addr, desc string) map[string]any {
		return map[string]any{"network_id": net.ID, "network_address": addr, "prefix_length": 24, "description": desc}
	}
	hasSubnet := func(addr string) bool {
		list, err := repo.ListSubnetsBySection(ctx, net.ID)
		require.NoError(t, err)
		for _, s := range list {
			if s.NetworkAddress == addr {
				return true
			}
		}
		return false
	}

	// Auto-approve (the default) applies the change synchronously.
	got, err := in.SubmitIntent(ctx, "subnet", nil, "create", subnetState("10.77.1.0", "auto"), &admin.ID)
	require.NoError(t, err)
	assert.Equal(t, "applied", got.Status)
	assert.True(t, hasSubnet("10.77.1.0"))

	// Manual review: nothing is applied until someone approves.
	require.NoError(t, svc.Config.SetCtx(ctx, "intent_auto_approve", "false"))
	pending, err := in.SubmitIntent(ctx, "subnet", nil, "create", subnetState("10.77.2.0", "manual"), &admin.ID)
	require.NoError(t, err)
	assert.Equal(t, "pending", pending.Status)
	assert.False(t, hasSubnet("10.77.2.0"))

	approved, err := in.ApproveIntent(ctx, pending.ID, &admin.ID, "ok")
	require.NoError(t, err)
	assert.Equal(t, "applied", approved.Status)
	assert.True(t, hasSubnet("10.77.2.0"))
	require.NotNil(t, approved.ReviewedBy)

	// A reviewed intent cannot be approved or rejected again.
	_, err = in.ApproveIntent(ctx, pending.ID, &admin.ID, "")
	assert.ErrorIs(t, err, ErrIntentNotPending)
	_, err = in.RejectIntent(ctx, pending.ID, &admin.ID, "")
	assert.ErrorIs(t, err, ErrIntentNotPending)

	// Rejection leaves the data untouched.
	rej, err := in.SubmitIntent(ctx, "subnet", nil, "create", subnetState("10.77.3.0", "rejected"), &admin.ID)
	require.NoError(t, err)
	out, err := in.RejectIntent(ctx, rej.ID, &admin.ID, "not wanted")
	require.NoError(t, err)
	assert.Equal(t, "rejected", out.Status)
	assert.False(t, hasSubnet("10.77.3.0"))

	// A failing apply is recorded as failed and cannot be retried.
	bad, err := in.SubmitIntent(ctx, "subnet", nil, "create", subnetState("not-an-address", "bad"), &admin.ID)
	require.NoError(t, err)
	_, err = in.ApproveIntent(ctx, bad.ID, &admin.ID, "")
	require.Error(t, err)
	failed, err := in.GetIntent(ctx, bad.ID)
	require.NoError(t, err)
	assert.Equal(t, "failed", failed.Status)

	// Validation.
	_, err = in.SubmitIntent(ctx, "widget", nil, "create", nil, &admin.ID)
	assert.Error(t, err)
	_, err = in.SubmitIntent(ctx, "subnet", nil, "update", nil, &admin.ID)
	assert.Error(t, err, "update needs a resource_id")

	// The diff lists only the fields a pending update would change.
	var existing int64
	list, err := repo.ListSubnetsBySection(ctx, net.ID)
	require.NoError(t, err)
	for _, s := range list {
		if s.NetworkAddress == "10.77.1.0" {
			existing = s.ID
		}
	}
	require.NotZero(t, existing)
	upd, err := in.SubmitIntent(ctx, "subnet", &existing, "update", map[string]any{"description": "renamed", "prefix_length": 24}, &admin.ID)
	require.NoError(t, err)
	assert.Equal(t, "pending", upd.Status)
	diff, err := in.GetCurrentVsDesired(ctx, "subnet", existing)
	require.NoError(t, err)
	require.NotNil(t, diff.Pending)
	require.Len(t, diff.Changes, 1)
	assert.Equal(t, "description", diff.Changes[0].Field)
	assert.Equal(t, "auto", diff.Changes[0].Current)
	assert.Equal(t, "renamed", diff.Changes[0].Desired)

	hist, err := in.ListIntents(ctx, "", "subnet", &existing)
	require.NoError(t, err)
	assert.Len(t, hist, 1)

	// Another organization sees none of it and cannot act on it.
	slug := "in-org-" + suffix
	org, err := repo.CreateOrganization(base, "Org "+slug, slug)
	require.NoError(t, err)
	other := repository.WithOrg(base, org.ID)
	seen, err := in.ListIntents(other, "", "", nil)
	require.NoError(t, err)
	assert.Empty(t, seen)
	_, err = in.GetIntent(other, upd.ID)
	assert.Error(t, err)
	_, err = in.ApproveIntent(other, upd.ID, &admin.ID, "")
	assert.ErrorIs(t, err, ErrIntentNotPending)
	still, err := in.GetIntent(ctx, upd.ID)
	require.NoError(t, err)
	assert.Equal(t, "pending", still.Status)
}
