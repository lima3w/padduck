package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"padduck/internal/testdb"
	"padduck/models"
	"padduck/repository"
)

// testDiscoveryService returns the discovery service wired to a scratch
// database, plus a network/subnet fixture (TEST-NET-1, guaranteed dead, so
// real ping scans terminate fast without touching live hosts).
func testDiscoveryService(t *testing.T) (*DiscoveryService, *repository.Repository, int64) {
	t.Helper()
	pool := testdb.Connect(t, "services")
	testdb.Truncate(t, pool,
		"scan_results", "scan_runs", "scan_jobs", "scan_agents", "scan_profiles",
		"ip_addresses", "subnets", "networks", "users")
	repo := repository.NewRepository(pool)
	svc := NewService(repo, testMFAKey)

	ctx := context.Background()
	u, err := repo.CreateUser(ctx, "scan-user", "scan@example.com")
	require.NoError(t, err)
	n, err := repo.CreateNetwork(ctx, "scan-net", "fixture", u.ID)
	require.NoError(t, err)
	s, err := repo.CreateSubnet(ctx, n.ID, "192.0.2.0", 30, "fixture", nil, false, false)
	require.NoError(t, err)
	return svc.Discovery, repo, s.ID
}

func TestMatchesCron(t *testing.T) {
	at := time.Date(2026, 6, 12, 14, 30, 0, 0, time.UTC) // Friday 14:30
	for _, tc := range []struct {
		cron string
		want bool
	}{
		{"* * * * *", true},
		{"30 14 * * *", true},
		{"30 14 12 6 *", true},
		{"30 14 * * 5", true},  // Friday == 5
		{"31 14 * * *", false}, // wrong minute
		{"30 15 * * *", false}, // wrong hour
		{"30 14 13 * *", false},
		{"30 14 * 7 *", false},
		{"30 14 * * 0", false},
		{"* * * *", false},      // 4 fields
		{"bogus * * * *", false}, // non-numeric
	} {
		assert.Equal(t, tc.want, matchesCron(tc.cron, at), "cron %q", tc.cron)
	}
}

func TestShouldRunJob(t *testing.T) {
	now := time.Date(2026, 6, 12, 14, 30, 0, 0, time.UTC)
	cron := "30 14 * * *"
	future := now.Add(time.Hour)

	job := &models.ScanJob{}
	assert.False(t, shouldRunJob(job, now), "no schedule means never auto-run")

	empty := ""
	job.ScheduleCron = &empty
	assert.False(t, shouldRunJob(job, now))

	job.ScheduleCron = &cron
	assert.True(t, shouldRunJob(job, now))

	// NextRunAt in the future suppresses the run even when the cron matches.
	job.NextRunAt = &future
	assert.False(t, shouldRunJob(job, now))
}

func TestParsePorts(t *testing.T) {
	assert.Equal(t, []string{"22", "443", "8080"}, parsePorts("22, 443 ,8080"))
	assert.Equal(t, []string{"80"}, parsePorts("80"))
	assert.Empty(t, parsePorts(""))
	assert.Empty(t, parsePorts(" , ,"))
}

func TestScanJobLifecycle_Integration(t *testing.T) {
	d, _, subnetID := testDiscoveryService(t)
	ctx := context.Background()

	cron := "0 2 * * *"
	job, err := d.CreateJob(ctx, "nightly", []int64{subnetID}, &cron, 1, true)
	require.NoError(t, err)
	require.NotZero(t, job.ID)

	got, err := d.GetJob(ctx, job.ID)
	require.NoError(t, err)
	assert.Equal(t, "nightly", got.Name)
	assert.Equal(t, []int64{subnetID}, got.SubnetIDs)

	jobs, err := d.ListJobs(ctx)
	require.NoError(t, err)
	assert.Len(t, jobs, 1)

	// Full update changes the scan settings and clamps concurrency.
	updated, err := d.UpdateJobFull(ctx, job.ID, "nightly-v2", []int64{subnetID}, nil, true, 500, true, "ping", nil, false, false, false)
	require.NoError(t, err)
	assert.Equal(t, "nightly-v2", updated.Name)
	assert.Equal(t, 100, updated.PingConcurrency, "concurrency above 100 is clamped")

	require.NoError(t, d.DeleteJob(ctx, job.ID))
	_, err = d.GetJob(ctx, job.ID)
	assert.Error(t, err)
}

func TestRunJob_StateAndResults_Integration(t *testing.T) {
	d, _, subnetID := testDiscoveryService(t)
	ctx := context.Background()

	job, err := d.CreateJob(ctx, "run-me", []int64{subnetID}, nil, 1, false)
	require.NoError(t, err)
	full, err := d.GetJob(ctx, job.ID)
	require.NoError(t, err)

	// 192.0.2.0/30 has two host IPs, both in TEST-NET (never alive).
	require.NoError(t, d.RunJob(ctx, full))

	runs, err := d.ListScanRuns(ctx, job.ID)
	require.NoError(t, err)
	require.Len(t, runs, 1, "RunJob must create exactly one scan run")
	assert.NotNil(t, runs[0].FinishedAt, "scan run must be finished")

	results, err := d.ListResults(ctx, job.ID, 10)
	require.NoError(t, err)
	require.Len(t, results, 2, "one result per host IP in the /30")
	for _, r := range results {
		assert.False(t, r.IsAlive, "TEST-NET IPs must scan dead")
	}

	updatedJob, err := d.GetJob(ctx, job.ID)
	require.NoError(t, err)
	assert.NotNil(t, updatedJob.LastRunAt, "job run time recorded")
}

func TestRunJob_DuplicateGuard_Integration(t *testing.T) {
	d, _, subnetID := testDiscoveryService(t)
	ctx := context.Background()

	job, err := d.CreateJob(ctx, "dup-guard", []int64{subnetID}, nil, 1, false)
	require.NoError(t, err)
	full, err := d.GetJob(ctx, job.ID)
	require.NoError(t, err)

	// Simulate the job already being in flight: RunJob must skip without
	// creating a scan run.
	d.inFlight.Store(job.ID, struct{}{})
	defer d.inFlight.Delete(job.ID)
	assert.True(t, d.IsRunning(job.ID))

	require.NoError(t, d.RunJob(ctx, full))
	runs, err := d.ListScanRuns(ctx, job.ID)
	require.NoError(t, err)
	assert.Empty(t, runs, "a duplicate run must not create a scan run")
}

func TestScanProfileLifecycle_Integration(t *testing.T) {
	d, _, subnetID := testDiscoveryService(t)
	ctx := context.Background()

	_, err := d.CreateScanProfile(ctx, "", "ping", nil, 10, nil, false, nil, "")
	assert.ErrorContains(t, err, "name is required")
	_, err = d.CreateScanProfile(ctx, "bad", "warpdrive", nil, 10, nil, false, nil, "")
	assert.ErrorContains(t, err, "invalid scan_type")

	profile, err := d.CreateScanProfile(ctx, "fast-ping", "", nil, 0, nil, false, nil, "")
	require.NoError(t, err)
	assert.Equal(t, "ping", profile.ScanType, "scan type defaults to ping")
	assert.Equal(t, 20, profile.PingConcurrency, "concurrency defaults to 20")
	assert.Equal(t, "v2c", profile.SNMPVersion)

	got, err := d.GetScanProfileByID(ctx, profile.ID)
	require.NoError(t, err)
	assert.Equal(t, "fast-ping", got.Name)

	profiles, err := d.ListScanProfiles(ctx)
	require.NoError(t, err)
	assert.Len(t, profiles, 1)

	_, err = d.UpdateScanProfile(ctx, profile.ID, "fast-ping-v2", "ping", nil, 50, nil, true, nil, "v2c")
	require.NoError(t, err)

	// Profiles attach to subnets (and detach).
	require.NoError(t, d.SetSubnetScanProfile(ctx, subnetID, &profile.ID))
	require.NoError(t, d.SetSubnetScanProfile(ctx, subnetID, nil))

	require.NoError(t, d.DeleteScanProfile(ctx, profile.ID))
	_, err = d.GetScanProfileByID(ctx, profile.ID)
	assert.Error(t, err)
}

func TestScanAgentLifecycle_Integration(t *testing.T) {
	d, _, _ := testDiscoveryService(t)
	ctx := context.Background()

	agent, rawToken, err := d.CreateAgent(ctx, "lab-agent", 30)
	require.NoError(t, err)
	require.NotEmpty(t, rawToken)

	// The raw token authenticates; garbage does not.
	authed, err := d.AuthenticateAgent(ctx, rawToken)
	require.NoError(t, err)
	assert.Equal(t, agent.ID, authed.ID)
	_, err = d.AuthenticateAgent(ctx, "bogus-token")
	assert.Error(t, err)

	// Rotation invalidates the old token and issues a working replacement.
	_, newToken, err := d.RotateToken(ctx, agent.ID, 30)
	require.NoError(t, err)
	assert.NotEqual(t, rawToken, newToken)
	_, err = d.AuthenticateAgent(ctx, rawToken)
	assert.Error(t, err, "old token must stop working after rotation")
	_, err = d.AuthenticateAgent(ctx, newToken)
	assert.NoError(t, err)

	version := "1.2.3"
	require.NoError(t, d.HeartbeatAgent(ctx, agent.ID, &version, []string{"icmp"}, "healthy", nil))

	agents, err := d.ListAgents(ctx)
	require.NoError(t, err)
	require.Len(t, agents, 1)

	require.NoError(t, d.DeleteAgent(ctx, agent.ID))
	_, err = d.AuthenticateAgent(ctx, newToken)
	assert.Error(t, err, "deleted agent must not authenticate")
}

func TestAcceptAgentResults_Integration(t *testing.T) {
	d, repo, subnetID := testDiscoveryService(t)
	ctx := context.Background()

	agent, _, err := d.CreateAgent(ctx, "results-agent", 0)
	require.NoError(t, err)
	otherAgent, _, err := d.CreateAgent(ctx, "other-agent", 0)
	require.NoError(t, err)

	job, err := d.CreateJob(ctx, "agent-job", []int64{subnetID}, nil, 1, true)
	require.NoError(t, err)
	_, err = d.UpdateJobFull(ctx, job.ID, "agent-job", []int64{subnetID}, nil, true, 10, false, "ping", &agent.ID, true, false, false)
	require.NoError(t, err)

	results := []AgentScanResult{
		{SubnetID: subnetID, IPAddress: "192.0.2.1", IsAlive: true, ResponseTimeMs: 12},
		{SubnetID: subnetID, IPAddress: "192.0.2.2", IsAlive: false},
	}

	// A job not assigned to the submitting agent is rejected.
	err = d.AcceptAgentResults(ctx, otherAgent.ID, job.ID, results)
	assert.ErrorContains(t, err, "not assigned to this agent")

	require.NoError(t, d.AcceptAgentResults(ctx, agent.ID, job.ID, results))

	stored, err := d.ListResults(ctx, job.ID, 10)
	require.NoError(t, err)
	assert.Len(t, stored, 2)

	// auto_add_ips: the alive, unknown address was created as an IP record.
	ips, err := repo.ListIPAddressesBySubnet(ctx, subnetID)
	require.NoError(t, err)
	require.Len(t, ips, 1)
	assert.Equal(t, "192.0.2.1", ips[0].Address)
	assert.Equal(t, "assigned", ips[0].Status)
}
