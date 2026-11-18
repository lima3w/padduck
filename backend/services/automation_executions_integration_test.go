package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"padduck/internal/testdb"
	"padduck/models"
	"padduck/repository"
)

func TestAutomationExecutionLog_Integration(t *testing.T) {
	pool := testdb.Connect(t, "services")
	testdb.Truncate(t, pool, "policy_execution_logs", "automation_policies", "scan_jobs", "ip_addresses", "ip_tags", "subnets", "networks", "users")
	repo := repository.NewRepository(pool)
	svc := NewService(repo, testMFAKey)
	auto := svc.Ops.Automation
	ctx := context.Background()

	u, err := repo.CreateUser(ctx, "auto-user", "auto@example.com")
	require.NoError(t, err)
	n, err := repo.CreateNetwork(ctx, "net", "", u.ID)
	require.NoError(t, err)
	sub, err := repo.CreateSubnet(ctx, n.ID, "10.9.0.0", 24, "lab", nil, false, false)
	require.NoError(t, err)
	oldTag, err := repo.CreateIPTag(ctx, "old", "#111111", nil)
	require.NoError(t, err)
	newTag, err := repo.CreateIPTag(ctx, "managed", "#222222", nil)
	require.NoError(t, err)
	ip, err := repo.CreateIPAddress(ctx, sub.ID, "10.9.0.5", "host", "assigned", &oldTag.ID, nil, nil, nil)
	require.NoError(t, err)

	tagPolicy, err := auto.SavePolicy(ctx, &models.AutomationPolicy{
		Name: "tag lab hosts", Workflow: "ip_address", Action: "create", Effect: "allow", Enabled: true,
		Conditions: []models.PolicyCondition{{Field: "hostname", Operator: "starts_with", Value: "host"}},
		Actions: []models.PolicyAction{
			{Type: "tag", Params: map[string]string{"tag_id": itoa(newTag.ID)}},
			{Type: "audit_annotation", Params: map[string]string{"message": "tagged"}},
			{Type: "tag", Params: map[string]string{"tag_id": "not-a-number"}},
		},
	})
	require.NoError(t, err)
	_, err = auto.SavePolicy(ctx, &models.AutomationPolicy{
		Name: "no guests", Workflow: "ip_address", Action: "create", Effect: "deny", Enabled: true, Message: "guests not allowed",
		Conditions: []models.PolicyCondition{{Field: "hostname", Operator: "eq", Value: "guest"}},
	})
	require.NoError(t, err)

	// A denied request is logged with no resource.
	d, err := auto.Evaluate(ctx, AutomationRequest{Workflow: "ip_address", Action: "create",
		Values: map[string]string{"hostname": "guest"}, TriggeredBy: "api"})
	require.NoError(t, err)
	assert.False(t, d.Allowed)

	// A pure check leaves no trace.
	_, err = auto.Evaluate(ctx, AutomationRequest{Workflow: "ip_address", Action: "create", Values: map[string]string{"hostname": "guest"}})
	require.NoError(t, err)
	_, err = auto.Evaluate(ctx, AutomationRequest{Workflow: "ip_address", Action: "create", Values: map[string]string{"hostname": "guest"}, DryRun: true, TriggeredBy: "api"})
	require.NoError(t, err)

	// An allowed request runs the matching policy's actions and logs each outcome.
	d, err = auto.Evaluate(ctx, AutomationRequest{Workflow: "ip_address", Action: "create",
		Values: map[string]string{"hostname": "host-1"}, TriggeredBy: "api"})
	require.NoError(t, err)
	require.True(t, d.Allowed)
	require.Len(t, d.Matched, 1)
	auto.RecordApplied(ctx, d, "ip_address", ip.ID, ip.Address)

	logs, total, err := auto.ListExecutions(ctx, repository.PolicyExecutionLogFilter{})
	require.NoError(t, err)
	require.Equal(t, 2, total, "one deny and one allow; the check and the dry run left nothing")
	allow, deny := logs[0], logs[1]
	assert.Equal(t, "deny", deny.Decision)
	assert.Nil(t, deny.ResourceID)
	assert.Empty(t, deny.ActionsTaken)
	assert.Equal(t, "allow", allow.Decision)
	assert.Equal(t, tagPolicy.ID, *allow.PolicyID)
	assert.Equal(t, "api", allow.TriggeredBy)
	assert.Equal(t, ip.Address, allow.ResourceName)
	require.Len(t, allow.ActionsTaken, 3)
	assert.Equal(t, "success", allow.ActionsTaken[0].Status)
	assert.Equal(t, "success", allow.ActionsTaken[1].Status)
	assert.Equal(t, "failed", allow.ActionsTaken[2].Status, "a failing action is recorded, not fatal")
	assert.Contains(t, allow.ActionsTaken[2].Error, "invalid tag_id")
	got, err := repo.GetIPAddressByID(ctx, ip.ID)
	require.NoError(t, err)
	assert.Equal(t, newTag.ID, *got.TagID, "the allowed request's primary effect and the tag action both applied")

	// Filters and per-policy counts.
	byRes, total, err := auto.ListExecutions(ctx, repository.PolicyExecutionLogFilter{ResourceType: "ip_address", ResourceID: ip.ID})
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Equal(t, allow.ID, byRes[0].ID)
	byPolicy, _, err := auto.ListExecutions(ctx, repository.PolicyExecutionLogFilter{PolicyID: tagPolicy.ID})
	require.NoError(t, err)
	assert.Len(t, byPolicy, 1)
	counts, err := auto.ExecutionCounts(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, counts[tagPolicy.ID])

	// Rolling back a deny has nothing to reverse.
	_, err = auto.RollbackExecution(ctx, deny.ID, &u.ID)
	assert.ErrorIs(t, err, ErrNothingToRollBack)

	// Roll back the tag action: the previous tag comes back, the failed action is untouched.
	rolled, err := auto.RollbackExecution(ctx, allow.ID, &u.ID)
	require.NoError(t, err)
	require.NotNil(t, rolled.RolledBackAt)
	assert.Equal(t, u.ID, *rolled.RolledBackBy)
	assert.Equal(t, "rolled_back", rolled.ActionsTaken[0].Rollback.Status)
	assert.Equal(t, "not_reversible", rolled.ActionsTaken[1].Rollback.Status)
	assert.Nil(t, rolled.ActionsTaken[2].Rollback)
	got, err = repo.GetIPAddressByID(ctx, ip.ID)
	require.NoError(t, err)
	assert.Equal(t, oldTag.ID, *got.TagID)
	_, err = auto.RollbackExecution(ctx, allow.ID, &u.ID)
	assert.ErrorIs(t, err, ErrExecutionAlreadyRolledBack)

	// A tag changed by someone else after the policy ran is not clobbered.
	d, err = auto.Evaluate(ctx, AutomationRequest{Workflow: "ip_address", Action: "create",
		Values: map[string]string{"hostname": "host-2"}, TriggeredBy: "api"})
	require.NoError(t, err)
	auto.RecordApplied(ctx, d, "ip_address", ip.ID, ip.Address)
	_, err = repo.UpdateIPAddressTag(ctx, ip.ID, nil)
	require.NoError(t, err)
	logs, _, err = auto.ListExecutions(ctx, repository.PolicyExecutionLogFilter{ResourceID: ip.ID})
	require.NoError(t, err)
	rolled, err = auto.RollbackExecution(ctx, logs[0].ID, nil)
	require.NoError(t, err)
	assert.Equal(t, "skipped", rolled.ActionsTaken[0].Rollback.Status)
	got, err = repo.GetIPAddressByID(ctx, ip.ID)
	require.NoError(t, err)
	assert.Nil(t, got.TagID, "left as the user set it")

	// Scan action: the created job is removed on rollback while it has not run.
	_, err = auto.SavePolicy(ctx, &models.AutomationPolicy{
		Name: "scan new subnets", Workflow: "subnet", Action: "create", Effect: "allow", Enabled: true,
		Actions: []models.PolicyAction{{Type: "scan"}},
	})
	require.NoError(t, err)
	d, err = auto.Evaluate(ctx, AutomationRequest{Workflow: "subnet", Action: "create", Values: map[string]string{}, TriggeredBy: "api"})
	require.NoError(t, err)
	auto.RecordApplied(ctx, d, "subnet", sub.ID, "10.9.0.0/24")
	scanLogs, _, err := auto.ListExecutions(ctx, repository.PolicyExecutionLogFilter{ResourceType: "subnet"})
	require.NoError(t, err)
	require.Len(t, scanLogs, 1)
	require.Equal(t, "success", scanLogs[0].ActionsTaken[0].Status)
	jobID := scanLogs[0].ActionsTaken[0].Undo["scan_job_id"]
	var jobs int
	require.NoError(t, pool.QueryRow(ctx, `SELECT count(*) FROM scan_jobs WHERE id=$1`, jobID).Scan(&jobs))
	assert.Equal(t, 1, jobs)
	rolled, err = auto.RollbackExecution(ctx, scanLogs[0].ID, nil)
	require.NoError(t, err)
	assert.Equal(t, "rolled_back", rolled.ActionsTaken[0].Rollback.Status)
	require.NoError(t, pool.QueryRow(ctx, `SELECT count(*) FROM scan_jobs WHERE id=$1`, jobID).Scan(&jobs))
	assert.Equal(t, 0, jobs)

	// Deleting a policy keeps its history under the saved name.
	require.NoError(t, auto.DeletePolicy(ctx, tagPolicy.ID))
	kept, err := auto.GetExecution(ctx, allow.ID)
	require.NoError(t, err)
	assert.Nil(t, kept.PolicyID)
	assert.Equal(t, "tag lab hosts", kept.PolicyName)
}
