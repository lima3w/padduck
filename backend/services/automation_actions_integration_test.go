package services

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"padduck/internal/testdb"
	"padduck/models"
	"padduck/repository"
)

type recordingScanner struct {
	mu   sync.Mutex
	jobs []*models.ScanJob
}

func (r *recordingScanner) RunJob(_ context.Context, j *models.ScanJob) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.jobs = append(r.jobs, j)
	return nil
}

func (r *recordingScanner) got() []*models.ScanJob {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]*models.ScanJob(nil), r.jobs...)
}

func eventually(t *testing.T, what string, f func() bool) {
	t.Helper()
	require.Eventually(t, f, 10*time.Second, 50*time.Millisecond, what)
}

// TestPolicyActions_Integration covers #19: each built-in action does what the
// policy says, malformed actions are rejected on save, failures are recorded
// without undoing anything, and changes applied through an intent fire actions
// just like direct writes.
func TestPolicyActions_Integration(t *testing.T) {
	pool := testdb.Connect(t, "services")
	repo := repository.NewRepository(pool)
	svc := NewService(repo, testMFAKey)
	auto := svc.Ops.Automation
	scanner := &recordingScanner{}
	auto.SetScanner(scanner)
	base := context.Background()
	ctx := repository.WithOrg(base, 1)
	suffix := fmt.Sprint(time.Now().UnixNano())

	admin, err := repo.CreateUserWithPassword(base, "act-admin"+suffix, "act-admin"+suffix+"@example.com", "x", "admin")
	require.NoError(t, err)
	net, err := repo.CreateNetwork(ctx, "act-net-"+suffix, "", admin.ID)
	require.NoError(t, err)
	sub, err := repo.CreateSubnet(ctx, net.ID, "10.66.0.0", 24, "act", nil, false, false)
	require.NoError(t, err)

	// --- Validation on save.
	bad := map[string]models.PolicyAction{
		"unknown type":       {Type: "explode"},
		"notify no target":   {Type: "notify"},
		"notify bad user":    {Type: "notify", Params: map[string]string{"user_id": "abc"}},
		"webhook no id":      {Type: "webhook"},
		"webhook bad id":     {Type: "webhook", Params: map[string]string{"webhook_id": "x"}},
		"annotation empty":   {Type: "audit_annotation", Params: map[string]string{"message": " "}},
		"scan bad profile":   {Type: "scan", Params: map[string]string{"profile_id": "x"}},
		"tag no id":          {Type: "tag"},
		"tag non-numeric id": {Type: "tag", Params: map[string]string{"tag_id": "x"}},
	}
	for name, act := range bad {
		_, err := auto.SavePolicy(ctx, &models.AutomationPolicy{Name: "bad " + name + suffix, Effect: "allow", Enabled: true, Actions: []models.PolicyAction{act}})
		assert.Error(t, err, name)
	}

	// --- Webhook goes to the named endpoint only.
	mkHook := func(name string, active bool) *models.WebhookEndpoint {
		h, err := repo.CreateWebhookEndpoint(ctx, &models.WebhookEndpoint{
			OrganizationID: ptrInt64(1), Name: name + suffix, URL: "https://example.com/" + name, Secret: "s",
			Events: []string{}, ObjectTypes: []string{}, TagFilters: []string{}, IsActive: active,
		})
		require.NoError(t, err)
		return h
	}
	target, other, off := mkHook("target", true), mkHook("other", true), mkHook("off", false)
	profile, err := repo.CreateScanProfile(ctx, "prof"+suffix, "ping+snmp", nil, 7, nil, true, nil, "v2c")
	require.NoError(t, err)

	_, err = auto.SavePolicy(ctx, &models.AutomationPolicy{
		Name: "actions" + suffix, Workflow: "ip_address", Action: "create", Effect: "allow", Enabled: true,
		Conditions: []models.PolicyCondition{{Field: "hostname", Operator: "eq", Value: "act-host"}},
		Actions: []models.PolicyAction{
			{Type: "webhook", Params: map[string]string{"webhook_id": itoa(target.ID)}},
			{Type: "webhook", Params: map[string]string{"webhook_id": itoa(off.ID)}},
			{Type: "webhook", Params: map[string]string{"webhook_id": "999999"}},
			{Type: "audit_annotation", Params: map[string]string{"message": "needs review"}},
			{Type: "notify", Params: map[string]string{"role": "admin"}},
		},
	})
	require.NoError(t, err)
	_, err = auto.SavePolicy(ctx, &models.AutomationPolicy{
		Name: "scan" + suffix, Workflow: "subnet", Action: "create", Effect: "allow", Enabled: true,
		Actions: []models.PolicyAction{{Type: "scan", Params: map[string]string{"profile_id": itoa(profile.ID)}}},
	})
	require.NoError(t, err)

	ip, err := repo.CreateIPAddress(ctx, sub.ID, "10.66.0.9", "act-host", "assigned", nil, nil, nil, nil)
	require.NoError(t, err)
	d, err := auto.Evaluate(ctx, AutomationRequest{Workflow: "ip_address", Action: "create",
		Values: map[string]string{"hostname": "act-host"}, TriggeredBy: "api", OrgID: ptrInt64(1)})
	require.NoError(t, err)
	auto.RecordApplied(ctx, d, "ip_address", ip.ID, ip.Address)

	count := func(q string, args ...any) int {
		var n int
		require.NoError(t, pool.QueryRow(base, q, args...).Scan(&n))
		return n
	}
	assert.Equal(t, 1, count(`SELECT count(*) FROM webhook_deliveries WHERE endpoint_id=$1`, target.ID))
	assert.Equal(t, 0, count(`SELECT count(*) FROM webhook_deliveries WHERE endpoint_id=$1`, other.ID), "only the named endpoint receives it")
	assert.Equal(t, 0, count(`SELECT count(*) FROM webhook_deliveries WHERE endpoint_id=$1`, off.ID), "inactive endpoint is not delivered to")
	assert.Equal(t, 1, count(`SELECT count(*) FROM audit_logs WHERE action='policy_annotation' AND resource_id=$1 AND new_values->>'message'='needs review'`, ip.ID))

	logs, _, err := auto.ListExecutions(ctx, repository.PolicyExecutionLogFilter{ResourceType: "ip_address", ResourceID: ip.ID})
	require.NoError(t, err)
	require.Len(t, logs, 1)
	got := logs[0].ActionsTaken
	require.Len(t, got, 5)
	assert.Equal(t, "success", got[0].Status)
	assert.Equal(t, "failed", got[1].Status)
	assert.Contains(t, got[1].Error, "inactive")
	assert.Equal(t, "failed", got[2].Status, "a missing endpoint is recorded, not fatal")
	assert.Equal(t, "success", got[3].Status)
	assert.Equal(t, "success", got[4].Status)
	still, err := repo.GetIPAddressByID(ctx, ip.ID)
	require.NoError(t, err)
	assert.Equal(t, ip.ID, still.ID, "action failures never undo the primary write")

	// --- Scan: the job takes the profile's settings and is actually started.
	sd, err := auto.Evaluate(ctx, AutomationRequest{Workflow: "subnet", Action: "create", Values: map[string]string{}, TriggeredBy: "api", OrgID: ptrInt64(1)})
	require.NoError(t, err)
	auto.RecordApplied(ctx, sd, "subnet", sub.ID, "10.66.0.0/24")
	eventually(t, "scan job started", func() bool { return len(scanner.got()) == 1 })
	job := scanner.got()[0]
	assert.Equal(t, "ping+snmp", job.ScanType)
	assert.Equal(t, 7, job.PingConcurrency)
	assert.Equal(t, []int64{sub.ID}, job.SubnetIDs)

	// --- Intents fire actions too.
	require.NoError(t, svc.Config.SetCtx(ctx, "intent_auto_approve", "true"))
	_, err = auto.SavePolicy(ctx, &models.AutomationPolicy{
		Name: "intent-note" + suffix, Workflow: "ip_address", Action: "create", Effect: "allow", Enabled: true,
		Conditions: []models.PolicyCondition{{Field: "hostname", Operator: "eq", Value: "via-intent"}},
		Actions:    []models.PolicyAction{{Type: "audit_annotation", Params: map[string]string{"message": "from intent"}}},
	})
	require.NoError(t, err)
	in, err := svc.Ops.Intent.SubmitIntent(ctx, "ip_address", nil, "create",
		map[string]any{"subnet_id": sub.ID, "address": "10.66.0.20", "hostname": "via-intent", "status": "assigned"}, &admin.ID)
	require.NoError(t, err)
	require.Equal(t, "applied", in.Status)
	eventually(t, "intent-applied change ran its policy actions", func() bool {
		return count(`SELECT count(*) FROM audit_logs WHERE action='policy_annotation' AND new_values->>'message'='from intent' AND resource_name='10.66.0.20'`) == 1
	})
	hist, _, err := auto.ListExecutions(ctx, repository.PolicyExecutionLogFilter{ResourceType: "ip_address"})
	require.NoError(t, err)
	var seen bool
	for _, h := range hist {
		if h.PolicyName == "intent-note"+suffix && h.ResourceName == "10.66.0.20" {
			seen = true
		}
	}
	assert.True(t, seen, "the intent's match is in the execution log")
}

func ptrInt64(v int64) *int64 { return &v }
