package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"padduck/internal/export"
	"padduck/models"
	"padduck/repository"
)

// ─────────────────────────────────────────────────────────────────────────────
// Helpers / stub repo
// ─────────────────────────────────────────────────────────────────────────────

// stubReportsRepo is a minimal in-memory implementation of reportsRepo for unit tests.
type stubReportsRepo struct {
	subnets          []*models.Subnet
	thresholdSubnets []*models.Subnet
	ips              map[int64][]*models.IPAddress
	history          map[int64][]*models.SubnetUtilizationPoint
	cooldowns        map[int64]*models.AlertCooldown
	scheduledReports map[int64]*models.ScheduledReport
	nextID           int64
	snapshots        []snapshotCall
	inactiveIPs      []*models.InactiveIPReport
	sections         []*models.Network
	trends           []*models.SubnetUtilizationTrend
	duplicates       *models.DuplicatesReport

	utilizationTrendCalls int
	inactiveIPCalls       int
	duplicateCalls        int
	listScheduledCalls    int
}

type snapshotCall struct {
	subnetID int64
	used     int
	total    int
	pct      float64
}

func newStubRepo() *stubReportsRepo {
	return &stubReportsRepo{
		ips:              make(map[int64][]*models.IPAddress),
		history:          make(map[int64][]*models.SubnetUtilizationPoint),
		cooldowns:        make(map[int64]*models.AlertCooldown),
		scheduledReports: make(map[int64]*models.ScheduledReport),
	}
}

func (r *stubReportsRepo) ListAllSubnets(_ context.Context) ([]*models.Subnet, error) {
	return r.subnets, nil
}
func (r *stubReportsRepo) ListSubnetsBySection(_ context.Context, networkID int64) ([]*models.Subnet, error) {
	return r.subnets, nil
}
func (r *stubReportsRepo) ListIPAddressesBySubnet(_ context.Context, subnetID int64) ([]*models.IPAddress, error) {
	return r.ips[subnetID], nil
}
func (r *stubReportsRepo) BulkSubnetUtilization(_ context.Context) ([]repository.SubnetUtil, error) {
	var out []repository.SubnetUtil
	for _, s := range r.subnets {
		used := 0
		for _, ip := range r.ips[s.ID] {
			if ip.Status == "assigned" || ip.Status == "reserved" {
				used++
			}
		}
		out = append(out, repository.SubnetUtil{SubnetID: s.ID, PrefixLength: s.PrefixLength, Used: used})
	}
	return out, nil
}
func (r *stubReportsRepo) RecordUtilizationSnapshot(_ context.Context, subnetID int64, used, total int, pct float64) error {
	r.snapshots = append(r.snapshots, snapshotCall{subnetID: subnetID, used: used, total: total, pct: pct})
	return nil
}
func (r *stubReportsRepo) GetUtilizationHistory(_ context.Context, subnetID int64, days int) ([]*models.SubnetUtilizationPoint, error) {
	return r.history[subnetID], nil
}
func (r *stubReportsRepo) GetUtilizationTrends(_ context.Context) ([]*models.SubnetUtilizationTrend, error) {
	r.utilizationTrendCalls++
	return r.trends, nil
}
func (r *stubReportsRepo) GetLatestUtilizationForSubnet(_ context.Context, subnetID int64) (*models.SubnetUtilizationPoint, error) {
	pts := r.history[subnetID]
	if len(pts) == 0 {
		return nil, nil
	}
	return pts[len(pts)-1], nil
}
func (r *stubReportsRepo) GetSubnetsByUtilizationThreshold(_ context.Context, _ float64) ([]*models.SubnetUtilizationTrend, error) {
	return nil, nil
}
func (r *stubReportsRepo) ListSubnetsWithThresholds(_ context.Context) ([]*models.Subnet, error) {
	return r.thresholdSubnets, nil
}
func (r *stubReportsRepo) GetAlertCooldown(_ context.Context, subnetID int64) (*models.AlertCooldown, error) {
	return r.cooldowns[subnetID], nil
}
func (r *stubReportsRepo) SetAlertCooldown(_ context.Context, subnetID int64, pct float64) error {
	r.cooldowns[subnetID] = &models.AlertCooldown{SubnetID: subnetID, AlertedPct: pct, AlertedAt: time.Now()}
	return nil
}
func (r *stubReportsRepo) ClearAlertCooldown(_ context.Context, subnetID int64) error {
	delete(r.cooldowns, subnetID)
	return nil
}
func (r *stubReportsRepo) CreateScheduledReport(_ context.Context, name, reportType, scheduleCron string, recipientEmails []string, filters map[string]any, format string, createdBy int64) (*models.ScheduledReport, error) {
	r.nextID++
	rpt := &models.ScheduledReport{
		ID: r.nextID, Name: name, ReportType: reportType,
		ScheduleCron: scheduleCron, RecipientEmails: recipientEmails,
		Filters: filters, Format: format, CreatedBy: createdBy,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	r.scheduledReports[rpt.ID] = rpt
	return rpt, nil
}
func (r *stubReportsRepo) GetScheduledReportByID(_ context.Context, id int64) (*models.ScheduledReport, error) {
	rpt, ok := r.scheduledReports[id]
	if !ok {
		return nil, context.Canceled // any error acts as not-found
	}
	return rpt, nil
}
func (r *stubReportsRepo) ListScheduledReports(_ context.Context) ([]*models.ScheduledReport, error) {
	r.listScheduledCalls++
	var out []*models.ScheduledReport
	for _, rpt := range r.scheduledReports {
		out = append(out, rpt)
	}
	return out, nil
}
func (r *stubReportsRepo) UpdateScheduledReport(_ context.Context, id int64, name, reportType, scheduleCron string, recipientEmails []string, filters map[string]any, format string) (*models.ScheduledReport, error) {
	rpt, ok := r.scheduledReports[id]
	if !ok {
		return nil, context.Canceled
	}
	rpt.Name = name
	rpt.ReportType = reportType
	rpt.ScheduleCron = scheduleCron
	rpt.RecipientEmails = recipientEmails
	rpt.Filters = filters
	rpt.Format = format
	return rpt, nil
}
func (r *stubReportsRepo) UpdateScheduledReportRunTime(_ context.Context, id int64, t time.Time) error {
	if rpt, ok := r.scheduledReports[id]; ok {
		rpt.LastRunAt = &t
	}
	return nil
}
func (r *stubReportsRepo) DeleteScheduledReport(_ context.Context, id int64) error {
	delete(r.scheduledReports, id)
	return nil
}
func (r *stubReportsRepo) GetInactiveIPs(_ context.Context, days int, networkID *int64) ([]*models.InactiveIPReport, error) {
	r.inactiveIPCalls++
	return r.inactiveIPs, nil
}
func (r *stubReportsRepo) BulkReleaseIPs(_ context.Context, ipIDs []int64) (int64, error) {
	return int64(len(ipIDs)), nil
}
func (r *stubReportsRepo) ListAllNetworks(_ context.Context) ([]*models.Network, error) {
	return r.sections, nil
}
func (r *stubReportsRepo) GetSubnetByID(_ context.Context, id int64) (*models.Subnet, error) {
	for _, s := range r.subnets {
		if s.ID == id {
			return s, nil
		}
	}
	return nil, context.Canceled
}
func (r *stubReportsRepo) GetSubnetGaps(_ context.Context) ([]*repository.SubnetGapRow, error) {
	return nil, nil
}
func (r *stubReportsRepo) GetVLANAssignment(_ context.Context) ([]*repository.VLANAssignmentRow, error) {
	return nil, nil
}
func (r *stubReportsRepo) GetIPAge(_ context.Context) ([]*repository.IPAgeRow, error) {
	return nil, nil
}
func (r *stubReportsRepo) GetDNSAudit(_ context.Context) ([]*repository.DNSAuditRow, error) {
	return nil, nil
}
func (r *stubReportsRepo) GetInactiveDevices(_ context.Context, days int) ([]*models.InactiveDeviceReport, error) {
	return nil, nil
}
func (r *stubReportsRepo) GetOverdueScanJobs(_ context.Context, days int) ([]*models.FailedScanJobReport, error) {
	return nil, nil
}
func (r *stubReportsRepo) GetDuplicates(_ context.Context) (*models.DuplicatesReport, error) {
	r.duplicateCalls++
	if r.duplicates != nil {
		return r.duplicates, nil
	}
	return &models.DuplicatesReport{
		DuplicateHostnames: []models.DuplicateHostname{},
		ConflictingIPs:     []models.ConflictingIP{},
	}, nil
}

func newTestReportsService(repo *stubReportsRepo) *ReportsService {
	cfg := &ConfigService{}
	email := &EmailService{configSvc: cfg}
	return NewReportsService(repo, cfg, email, nil)
}

// ─────────────────────────────────────────────────────────────────────────────
// TakeUtilizationSnapshots
// ─────────────────────────────────────────────────────────────────────────────

func TestTakeUtilizationSnapshots_RecordsSnapshot(t *testing.T) {
	repo := newStubRepo()
	subnet := &models.Subnet{ID: 1, NetworkAddress: "10.0.0.0", PrefixLength: 24}
	repo.subnets = []*models.Subnet{subnet}
	repo.ips[1] = []*models.IPAddress{
		{ID: 1, Status: "assigned"},
		{ID: 2, Status: "available"},
		{ID: 3, Status: "reserved"},
	}

	svc := newTestReportsService(repo)
	svc.TakeUtilizationSnapshots(context.Background())

	require.Len(t, repo.snapshots, 1)
	snap := repo.snapshots[0]
	assert.Equal(t, int64(1), snap.subnetID)
	assert.Equal(t, 2, snap.used) // assigned + reserved
	assert.Equal(t, 256, snap.total)
	assert.InDelta(t, 0.78, snap.pct, 0.01)
}

func TestTakeUtilizationSnapshots_EmptySubnet(t *testing.T) {
	repo := newStubRepo()
	repo.subnets = []*models.Subnet{
		{ID: 1, NetworkAddress: "192.168.0.0", PrefixLength: 24},
	}
	// No IPs

	svc := newTestReportsService(repo)
	svc.TakeUtilizationSnapshots(context.Background())

	require.Len(t, repo.snapshots, 1)
	assert.Equal(t, 0, repo.snapshots[0].used)
	assert.Equal(t, 256, repo.snapshots[0].total)
	assert.Equal(t, 0.0, repo.snapshots[0].pct)
}

// ─────────────────────────────────────────────────────────────────────────────
// CheckThresholdAlerts — cooldown logic
// ─────────────────────────────────────────────────────────────────────────────

func TestCheckThresholdAlerts_AboveThreshold_SetsCooldown(t *testing.T) {
	repo := newStubRepo()
	threshold := 80
	subnet := &models.Subnet{
		ID:                1,
		NetworkAddress:    "10.0.0.0",
		PrefixLength:      24,
		AlertThresholdPct: &threshold,
	}
	repo.thresholdSubnets = []*models.Subnet{subnet}
	pct := 85.0
	repo.history[1] = []*models.SubnetUtilizationPoint{
		{UtilizationPct: pct, UsedCount: 217, TotalCount: 256, RecordedAt: time.Now()},
	}

	svc := newTestReportsService(repo)
	svc.CheckThresholdAlerts(context.Background())

	// Cooldown should have been set
	assert.NotNil(t, repo.cooldowns[1])
	assert.InDelta(t, pct, repo.cooldowns[1].AlertedPct, 0.01)
}

func TestCheckThresholdAlerts_AlreadyCooledDown_NoSecondAlert(t *testing.T) {
	repo := newStubRepo()
	threshold := 80
	subnet := &models.Subnet{
		ID:                1,
		NetworkAddress:    "10.0.0.0",
		PrefixLength:      24,
		AlertThresholdPct: &threshold,
	}
	repo.thresholdSubnets = []*models.Subnet{subnet}
	repo.history[1] = []*models.SubnetUtilizationPoint{
		{UtilizationPct: 85.0, UsedCount: 217, TotalCount: 256, RecordedAt: time.Now()},
	}
	// Cooldown already set
	repo.cooldowns[1] = &models.AlertCooldown{SubnetID: 1, AlertedPct: 85.0, AlertedAt: time.Now()}

	svc := newTestReportsService(repo)
	initialCooldownTime := repo.cooldowns[1].AlertedAt
	svc.CheckThresholdAlerts(context.Background())

	// Cooldown should NOT be updated (same timestamp)
	assert.Equal(t, initialCooldownTime, repo.cooldowns[1].AlertedAt)
}

func TestCheckThresholdAlerts_BelowThresholdMinus5_ClearsCooldown(t *testing.T) {
	repo := newStubRepo()
	threshold := 80
	subnet := &models.Subnet{
		ID:                1,
		NetworkAddress:    "10.0.0.0",
		PrefixLength:      24,
		AlertThresholdPct: &threshold,
	}
	repo.thresholdSubnets = []*models.Subnet{subnet}
	// Utilization is now 70%, which is < 80 - 5 = 75
	repo.history[1] = []*models.SubnetUtilizationPoint{
		{UtilizationPct: 70.0, UsedCount: 179, TotalCount: 256, RecordedAt: time.Now()},
	}
	// Existing cooldown
	repo.cooldowns[1] = &models.AlertCooldown{SubnetID: 1, AlertedPct: 85.0, AlertedAt: time.Now()}

	svc := newTestReportsService(repo)
	svc.CheckThresholdAlerts(context.Background())

	// Cooldown should be cleared
	assert.Nil(t, repo.cooldowns[1])
}

func TestCheckThresholdAlerts_NoHistory_NoAlert(t *testing.T) {
	repo := newStubRepo()
	threshold := 80
	subnet := &models.Subnet{
		ID:                1,
		NetworkAddress:    "10.0.0.0",
		PrefixLength:      24,
		AlertThresholdPct: &threshold,
	}
	repo.thresholdSubnets = []*models.Subnet{subnet}
	// No history

	svc := newTestReportsService(repo)
	svc.CheckThresholdAlerts(context.Background())

	// No cooldown set
	assert.Nil(t, repo.cooldowns[1])
}

// ─────────────────────────────────────────────────────────────────────────────
// totalAddressesFromPrefix
// ─────────────────────────────────────────────────────────────────────────────

func TestTotalAddressesFromPrefix(t *testing.T) {
	cases := []struct {
		prefix   int
		expected int
	}{
		{24, 256},
		{25, 128},
		{32, 1},
		{0, 1 << 32},
		{16, 65536},
		{64, 1 << 32}, // IPv6 capped at 32 bits
		{128, 1},      // /128 → 1 address
	}
	for _, tc := range cases {
		assert.Equal(t, tc.expected, totalAddressesFromPrefix(tc.prefix),
			"prefix /%d", tc.prefix)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// BulkReleaseIPs
// ─────────────────────────────────────────────────────────────────────────────

func TestBulkReleaseIPs_ReturnsCount(t *testing.T) {
	repo := newStubRepo()
	svc := newTestReportsService(repo)

	count, err := svc.BulkReleaseIPs(context.Background(), []int64{1, 2, 3}, 0)
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)
}

// ─────────────────────────────────────────────────────────────────────────────
// Scheduled report CRUD round-trip
// ─────────────────────────────────────────────────────────────────────────────

func TestScheduledReport_CreateAndGet(t *testing.T) {
	repo := newStubRepo()
	svc := newTestReportsService(repo)
	ctx := context.Background()

	rpt, err := svc.CreateScheduledReport(ctx,
		"Test Report", "utilisation_summary", "0 9 * * 1",
		[]string{"admin@example.com"}, map[string]any{}, "csv", 1,
	)
	require.NoError(t, err)
	assert.Equal(t, "Test Report", rpt.Name)
	assert.Equal(t, "csv", rpt.Format)

	got, err := svc.GetScheduledReport(ctx, rpt.ID)
	require.NoError(t, err)
	assert.Equal(t, rpt.ID, got.ID)
}

func TestScheduledReport_Delete(t *testing.T) {
	repo := newStubRepo()
	svc := newTestReportsService(repo)
	ctx := context.Background()

	rpt, err := svc.CreateScheduledReport(ctx,
		"Temp", "inactive_ips", "0 0 * * *",
		[]string{}, map[string]any{}, "csv", 1,
	)
	require.NoError(t, err)

	err = svc.DeleteScheduledReport(ctx, rpt.ID)
	require.NoError(t, err)

	_, err = svc.GetScheduledReport(ctx, rpt.ID)
	assert.Error(t, err)
}

// ─────────────────────────────────────────────────────────────────────────────
// Performance budgets
// ─────────────────────────────────────────────────────────────────────────────

func TestPerformanceBudget_UtilizationTrendWorkflowUsesCachedRead(t *testing.T) {
	repo := newStubRepo()
	repo.trends = []*models.SubnetUtilizationTrend{
		{SubnetID: 1, CIDR: "10.0.0.0/24", CurrentPct: 50},
	}
	svc := newTestReportsService(repo)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		trends, err := svc.GetUtilizationTrends(ctx)
		require.NoError(t, err)
		require.Len(t, trends, 1)
	}

	assert.LessOrEqual(t, repo.utilizationTrendCalls, 1, "utilization trends should stay within the repository-call budget")
}

func TestPerformanceBudget_InactiveIPWorkflowUsesParameterizedCachedRead(t *testing.T) {
	deviceID := int64(5)
	repo := newStubRepo()
	repo.inactiveIPs = []*models.InactiveIPReport{
		{IPID: 1, IPAddress: "10.0.0.10", DeviceID: &deviceID, DaysInactive: 90},
	}
	svc := newTestReportsService(repo)
	ctx := context.Background()
	networkID := int64(42)

	for i := 0; i < 3; i++ {
		ips, err := svc.GetInactiveIPs(ctx, 90, &networkID)
		require.NoError(t, err)
		require.Len(t, ips, 1)
	}
	if _, err := svc.GetInactiveIPs(ctx, 30, &networkID); err != nil {
		t.Fatal(err)
	}

	assert.LessOrEqual(t, repo.inactiveIPCalls, 2, "inactive IP reads should cache identical day/section parameters")
}

func TestPerformanceBudget_DuplicateReportWorkflowUsesCachedRead(t *testing.T) {
	repo := newStubRepo()
	repo.duplicates = &models.DuplicatesReport{
		DuplicateHostnames: []models.DuplicateHostname{{Hostname: "router-1", Count: 2, DeviceIDs: []int64{1, 2}}},
		ConflictingIPs:     []models.ConflictingIP{{IPAddress: "10.0.0.5", Count: 2, Hostnames: []string{"a", "b"}}},
	}
	svc := newTestReportsService(repo)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		report, err := svc.GetDuplicates(ctx)
		require.NoError(t, err)
		require.Len(t, report.DuplicateHostnames, 1)
	}

	assert.LessOrEqual(t, repo.duplicateCalls, 1, "duplicate report should stay within the repository-call budget")
}

func TestPerformanceBudget_ScheduledReportListInvalidatesOnMutation(t *testing.T) {
	repo := newStubRepo()
	svc := newTestReportsService(repo)
	ctx := context.Background()

	_, err := svc.CreateScheduledReport(ctx, "Weekly", "utilisation_summary", "0 9 * * 1", []string{"admin@example.com"}, map[string]any{}, "csv", 1)
	require.NoError(t, err)

	for i := 0; i < 3; i++ {
		reports, err := svc.ListScheduledReports(ctx)
		require.NoError(t, err)
		require.Len(t, reports, 1)
	}
	require.Equal(t, 1, repo.listScheduledCalls)

	_, err = svc.CreateScheduledReport(ctx, "Daily", "inactive_ips", "0 7 * * *", []string{"admin@example.com"}, map[string]any{"days": 90}, "csv", 1)
	require.NoError(t, err)
	reports, err := svc.ListScheduledReports(ctx)
	require.NoError(t, err)
	require.Len(t, reports, 2)

	assert.Equal(t, 2, repo.listScheduledCalls, "scheduled report list cache should be invalidated by writes")
}

// ─────────────────────────────────────────────────────────────────────────────
// GenerateCSV / GeneratePDF round-trip (proxy through export package)
// ─────────────────────────────────────────────────────────────────────────────

func TestGenerateCSV_NonEmpty(t *testing.T) {
	headers := []string{"a", "b"}
	rows := []map[string]string{{"a": "1", "b": "2"}}
	data, err := export.GenerateCSV(headers, rows)
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestGeneratePDF_NonEmpty(t *testing.T) {
	data, err := export.GeneratePDF("Test", []string{"col1"}, [][]string{{"val1"}})
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}
