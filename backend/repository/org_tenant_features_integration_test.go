package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"padduck/models"
)

// TestTenantAwareFeatures_Integration covers the cross-cutting features of
// #10: API tokens, webhooks and their deliveries, scheduled reports and audit
// logs all stay inside the caller's organization.
func TestTenantAwareFeatures_Integration(t *testing.T) {
	r := testRepo(t)
	base := context.Background()
	orgA, orgB := newTestOrg(t, r, "tf-a"), newTestOrg(t, r, "tf-b")
	a, b := WithOrg(base, orgA), WithOrg(base, orgB)
	suffix := fmt.Sprint(time.Now().UnixNano())

	mkUser := func(name string, org int64) int64 {
		u, err := r.CreateUserWithPassword(base, name+suffix, name+suffix+"@example.com", "x", "admin")
		require.NoError(t, err)
		_, err = r.db.Exec(base, `UPDATE users SET organization_id = $1 WHERE id = $2`, org, u.ID)
		require.NoError(t, err)
		return u.ID
	}
	userA, userB := mkUser("tf-ua", orgA), mkUser("tf-ub", orgB)

	// --- API tokens carry their owner's organization.
	tokA, err := r.CreateAPITokenFull(a, userA, "hash-a-"+suffix, "tok-a", "read", nil)
	require.NoError(t, err)
	_, err = r.CreateAPIToken(b, userB, "hash-b-"+suffix, "tok-b")
	require.NoError(t, err)
	nA, err := r.CountAPITokensByOrg(base, orgA)
	require.NoError(t, err)
	nB, err := r.CountAPITokensByOrg(base, orgB)
	require.NoError(t, err)
	assert.EqualValues(t, 1, nA, "new tokens must count toward their org's quota")
	assert.EqualValues(t, 1, nB)
	tokens, err := r.ListAPITokenAnalytics(b, nil) // no explicit org: the request scope applies
	require.NoError(t, err)
	require.Len(t, tokens, 1)
	assert.NotEqual(t, tokA.ID, tokens[0].ID)

	// --- Webhooks and deliveries.
	hook, err := r.CreateWebhookEndpoint(a, &models.WebhookEndpoint{
		OrganizationID: &orgA, Name: "hook-a", URL: "https://example.com/a", Secret: "s",
		Events: []string{"x"}, ObjectTypes: []string{}, TagFilters: []string{}, IsActive: true,
	})
	require.NoError(t, err)
	del, err := r.CreateWebhookDelivery(a, hook.ID, "x", `{}`)
	require.NoError(t, err)
	list, err := r.ListWebhookEndpoints(b, nil)
	require.NoError(t, err)
	assert.Empty(t, list, "other org lists no webhooks")
	_, err = r.GetWebhookEndpoint(b, hook.ID)
	assert.Error(t, err)
	_, err = r.GetWebhookDelivery(b, del.ID)
	assert.Error(t, err)
	_, err = r.ReplayWebhookDelivery(b, del.ID)
	assert.Error(t, err, "other org cannot replay")
	ds, err := r.ListWebhookDeliveries(b, 50)
	require.NoError(t, err)
	assert.Empty(t, ds)
	_ = r.DeleteWebhookEndpoint(b, hook.ID)
	_, err = r.GetWebhookEndpoint(a, hook.ID)
	assert.NoError(t, err, "cross-org delete must not remove the webhook")
	var delOrg int64
	require.NoError(t, r.db.QueryRow(base, `SELECT organization_id FROM webhook_deliveries WHERE id = $1`, del.ID).Scan(&delOrg))
	assert.Equal(t, orgA, delOrg, "a delivery belongs to its webhook's org")

	// --- Scheduled reports.
	rep, err := r.CreateScheduledReport(a, &orgA, "rep-a", "inactive_ips", "0 6 * * *", []string{"a@example.com"}, map[string]any{}, "csv", userA)
	require.NoError(t, err)
	reps, err := r.ListScheduledReports(b, nil)
	require.NoError(t, err)
	assert.Empty(t, reps)
	_, err = r.GetScheduledReportByID(b, rep.ID)
	assert.Error(t, err)
	_ = r.DeleteScheduledReport(b, rep.ID)
	_, err = r.GetScheduledReportByID(a, rep.ID)
	assert.NoError(t, err)

	// --- Audit logs: an unfiltered query is limited to the request's org;
	// only an explicit all-orgs query crosses.
	for _, o := range []int64{orgA, orgB} {
		require.NoError(t, r.CreateAuditLog(base, &models.AuditLog{OrganizationID: &o, Username: "tf", Action: "tf_action_" + suffix, Status: "success"}))
	}
	scoped, err := r.ListAuditLogs(b, &models.AuditLogFilter{Action: "tf_action_" + suffix, Limit: 50})
	require.NoError(t, err)
	require.Len(t, scoped, 1)
	assert.Equal(t, orgB, *scoped[0].OrganizationID)
	total, err := r.CountAuditLogs(b, &models.AuditLogFilter{Action: "tf_action_" + suffix})
	require.NoError(t, err)
	assert.EqualValues(t, 1, total)
	all, err := r.ListAuditLogs(b, &models.AuditLogFilter{Action: "tf_action_" + suffix, Limit: 50, AllOrgs: true})
	require.NoError(t, err)
	assert.Len(t, all, 2, "platform all-orgs view sees both")
}

func TestSetPlatformAdmin_Integration(t *testing.T) {
	r := testRepo(t)
	ctx := context.Background()

	require.ErrorIs(t, r.SetPlatformAdmin(ctx, 999999, true), pgx.ErrNoRows)

	uid := createTestUser(t, r)
	require.NoError(t, r.SetPlatformAdmin(ctx, uid, true))
	u, err := r.GetUserByID(ctx, uid)
	require.NoError(t, err)
	assert.True(t, u.IsPlatformAdmin)
}
