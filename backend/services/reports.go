package services

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"padduck/internal/export"
	"padduck/models"
	"padduck/repository"
)

// reportsRepo lists the repository methods used by ReportsService.
type reportsRepo interface {
	// utilization
	ListAllSubnets(ctx context.Context) ([]*models.Subnet, error)
	ListSubnetsBySection(ctx context.Context, networkID int64) ([]*models.Subnet, error)
	ListIPAddressesBySubnet(ctx context.Context, subnetID int64) ([]*models.IPAddress, error)
	BulkSubnetUtilization(ctx context.Context) ([]repository.SubnetUtil, error)
	RecordUtilizationSnapshot(ctx context.Context, subnetID int64, used, total int, pct float64) error
	GetUtilizationHistory(ctx context.Context, subnetID int64, days int) ([]*models.SubnetUtilizationPoint, error)
	GetUtilizationTrends(ctx context.Context) ([]*models.SubnetUtilizationTrend, error)
	GetLatestUtilizationForSubnet(ctx context.Context, subnetID int64) (*models.SubnetUtilizationPoint, error)
	GetSubnetsByUtilizationThreshold(ctx context.Context, thresholdPct float64) ([]*models.SubnetUtilizationTrend, error)
	// alert cooldowns
	ListSubnetsWithThresholds(ctx context.Context) ([]*models.Subnet, error)
	GetAlertCooldown(ctx context.Context, subnetID int64) (*models.AlertCooldown, error)
	SetAlertCooldown(ctx context.Context, subnetID int64, pct float64) error
	ClearAlertCooldown(ctx context.Context, subnetID int64) error
	// scheduled reports
	CreateScheduledReport(ctx context.Context, name, reportType, scheduleCron string, recipientEmails []string, filters map[string]any, format string, createdBy int64) (*models.ScheduledReport, error)
	GetScheduledReportByID(ctx context.Context, id int64) (*models.ScheduledReport, error)
	ListScheduledReports(ctx context.Context) ([]*models.ScheduledReport, error)
	UpdateScheduledReport(ctx context.Context, id int64, name, reportType, scheduleCron string, recipientEmails []string, filters map[string]any, format string) (*models.ScheduledReport, error)
	UpdateScheduledReportRunTime(ctx context.Context, id int64, t time.Time) error
	DeleteScheduledReport(ctx context.Context, id int64) error
	// inactive IPs
	GetInactiveIPs(ctx context.Context, days int, networkID *int64) ([]*models.InactiveIPReport, error)
	BulkReleaseIPs(ctx context.Context, ipIDs []int64) (int64, error)
	// sections and subnets for export
	ListAllNetworks(ctx context.Context) ([]*models.Network, error)
	GetSubnetByID(ctx context.Context, id int64) (*models.Subnet, error)
	// expanded report types
	GetSubnetGaps(ctx context.Context) ([]*repository.SubnetGapRow, error)
	GetVLANAssignment(ctx context.Context) ([]*repository.VLANAssignmentRow, error)
	GetIPAge(ctx context.Context) ([]*repository.IPAgeRow, error)
	GetDNSAudit(ctx context.Context) ([]*repository.DNSAuditRow, error)
	// remediation report types
	GetInactiveDevices(ctx context.Context, days int) ([]*models.InactiveDeviceReport, error)
	GetOverdueScanJobs(ctx context.Context, days int) ([]*models.FailedScanJobReport, error)
	// duplicate detection (#425)
	GetDuplicates(ctx context.Context) (*models.DuplicatesReport, error)
}

// ReportsService provides reporting and analytics functionality.
type ReportsService struct {
	repo   reportsRepo
	config *ConfigService
	email  *EmailService
	audit  *AuditService
	cache  *ttlCache[any]
}

// NewReportsService creates a new ReportsService.
func NewReportsService(repo reportsRepo, config *ConfigService, email *EmailService, audit *AuditService) *ReportsService {
	return &ReportsService{
		repo:   repo,
		config: config,
		email:  email,
		audit:  audit,
		cache:  newTTLCache[any](30 * time.Second),
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Utilization snapshots (#220)
// ─────────────────────────────────────────────────────────────────────────────

// TakeUtilizationSnapshots records a utilization snapshot for every subnet in two queries.
func (rs *ReportsService) TakeUtilizationSnapshots(ctx context.Context) {
	utils, err := rs.repo.BulkSubnetUtilization(ctx)
	if err != nil {
		slog.Error("reports: bulk subnet utilization failed", "error", err)
		return
	}

	for _, u := range utils {
		total := totalAddressesFromPrefix(u.PrefixLength)
		var pct float64
		if total > 0 {
			pct = float64(u.Used) / float64(total) * 100
		}
		if err := rs.repo.RecordUtilizationSnapshot(ctx, u.SubnetID, u.Used, total, pct); err != nil {
			slog.Error("reports: record snapshot failed", "subnet_id", u.SubnetID, "error", err)
		}
	}

	// After recording snapshots, check threshold alerts.
	rs.CheckThresholdAlerts(ctx)
}

// GetUtilizationHistory returns utilization history for a subnet.
func (rs *ReportsService) GetUtilizationHistory(ctx context.Context, subnetID int64, days int) ([]*models.SubnetUtilizationPoint, error) {
	return rs.repo.GetUtilizationHistory(ctx, subnetID, days)
}

// GetUtilizationTrends returns trend data for all subnets.
func (rs *ReportsService) GetUtilizationTrends(ctx context.Context) ([]*models.SubnetUtilizationTrend, error) {
	if value, ok := rs.reportCache().get("utilization_trends"); ok {
		return cloneUtilizationTrends(value.([]*models.SubnetUtilizationTrend)), nil
	}
	trends, err := rs.repo.GetUtilizationTrends(ctx)
	if err != nil {
		return nil, err
	}
	rs.reportCache().set("utilization_trends", cloneUtilizationTrends(trends))
	return cloneUtilizationTrends(trends), nil
}

// StartUtilizationSnapshotJob launches a background goroutine that periodically takes snapshots.
func (rs *ReportsService) StartUtilizationSnapshotJob(ctx context.Context) {
	go func() {
		intervalHours := 1
		if rs.config != nil && rs.config.repository != nil {
			if val, err := rs.config.Get("utilisation_snapshot_interval_hours"); err == nil && val != "" {
				if h, err := strconv.Atoi(val); err == nil && h > 0 {
					intervalHours = h
				}
			}
		}

		ticker := time.NewTicker(time.Duration(intervalHours) * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				rs.TakeUtilizationSnapshots(ctx)
			}
		}
	}()
}

// ─────────────────────────────────────────────────────────────────────────────
// Threshold alerts (#221)
// ─────────────────────────────────────────────────────────────────────────────

// CheckThresholdAlerts checks all subnets with alert thresholds and sends alerts as needed.
func (rs *ReportsService) CheckThresholdAlerts(ctx context.Context) {
	subnets, err := rs.repo.ListSubnetsWithThresholds(ctx)
	if err != nil {
		slog.Error("reports: list subnets with thresholds failed", "error", err)
		return
	}

	// Also check global default threshold (ignore config errors in test environments)
	var globalThreshold *int
	if rs.config != nil && rs.config.repository != nil {
		globalThresholdStr, _ := rs.config.Get("default_alert_threshold_pct")
		if globalThresholdStr != "" {
			if v, err := strconv.Atoi(globalThresholdStr); err == nil {
				globalThreshold = &v
			}
		}
	}

	// Collect all subnets: those with per-subnet threshold + those relying on global
	toCheck := make([]*models.Subnet, 0, len(subnets))
	toCheck = append(toCheck, subnets...)

	// If global threshold is set, also check subnets without per-subnet threshold
	if globalThreshold != nil {
		all, err := rs.repo.ListAllSubnets(ctx)
		if err == nil {
			for _, s := range all {
				if s.AlertThresholdPct == nil {
					// Apply global threshold by setting it transiently
					gv := *globalThreshold
					s.AlertThresholdPct = &gv
					toCheck = append(toCheck, s)
				}
			}
		}
	}

	for _, subnet := range toCheck {
		threshold := *subnet.AlertThresholdPct

		latest, err := rs.repo.GetLatestUtilizationForSubnet(ctx, subnet.ID)
		if err != nil || latest == nil {
			continue
		}

		currentPct := latest.UtilizationPct
		thresholdF := float64(threshold)

		cidr := fmt.Sprintf("%s/%d", subnet.NetworkAddress, subnet.PrefixLength)

		if currentPct >= thresholdF {
			// Check cooldown
			cooldown, err := rs.repo.GetAlertCooldown(ctx, subnet.ID)
			if err != nil {
				slog.Error("reports: get cooldown failed", "subnet_id", subnet.ID, "error", err)
				continue
			}
			if cooldown != nil {
				// Already alerted, skip
				continue
			}

			// Send alert
			alertEmail := ""
			if subnet.AlertEmailOverride != nil && *subnet.AlertEmailOverride != "" {
				alertEmail = *subnet.AlertEmailOverride
			} else if rs.config != nil && rs.config.repository != nil {
				alertEmail, _ = rs.config.Get("alert_email")
			}

			if alertEmail != "" {
				subject := fmt.Sprintf("Subnet Capacity Alert: %s", cidr)
				body := fmt.Sprintf(
					"Subnet capacity alert\n\nCIDR: %s\nDescription: %s\nUsed: %d / %d (%.2f%%)\nThreshold: %d%%\n\nPlease review and take action.",
					cidr, subnet.Description,
					latest.UsedCount, latest.TotalCount, currentPct,
					threshold,
				)
				if err := rs.email.Send(alertEmail, subject, body); err != nil {
					slog.Error("reports: send alert email failed", "subnet_id", subnet.ID, "error", err)
				}
			}

			// Set cooldown
			if err := rs.repo.SetAlertCooldown(ctx, subnet.ID, currentPct); err != nil {
				slog.Error("reports: set cooldown failed", "subnet_id", subnet.ID, "error", err)
			}
		} else if currentPct < thresholdF-5 {
			// Clear cooldown to allow future re-alerting
			if err := rs.repo.ClearAlertCooldown(ctx, subnet.ID); err != nil {
				slog.Error("reports: clear cooldown failed", "subnet_id", subnet.ID, "error", err)
			}
		}
	}
}

// GetSubnetsNearCapacity returns subnets whose latest utilization exceeds the threshold.
func (rs *ReportsService) GetSubnetsNearCapacity(ctx context.Context, thresholdPct float64) ([]*models.SubnetUtilizationTrend, error) {
	key := fmt.Sprintf("subnets_near_capacity:%.2f", thresholdPct)
	if value, ok := rs.reportCache().get(key); ok {
		return cloneUtilizationTrends(value.([]*models.SubnetUtilizationTrend)), nil
	}
	trends, err := rs.repo.GetSubnetsByUtilizationThreshold(ctx, thresholdPct)
	if err != nil {
		return nil, err
	}
	rs.reportCache().set(key, cloneUtilizationTrends(trends))
	return cloneUtilizationTrends(trends), nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Scheduled reports (#222)
// ─────────────────────────────────────────────────────────────────────────────

// CreateScheduledReport creates a new scheduled report.
func (rs *ReportsService) CreateScheduledReport(ctx context.Context, name, reportType, scheduleCron string, recipientEmails []string, filters map[string]any, format string, createdBy int64) (*models.ScheduledReport, error) {
	report, err := rs.repo.CreateScheduledReport(ctx, name, reportType, scheduleCron, recipientEmails, filters, format, createdBy)
	if err == nil {
		rs.clearReportCache()
	}
	return report, err
}

// GetScheduledReport retrieves a scheduled report by ID.
func (rs *ReportsService) GetScheduledReport(ctx context.Context, id int64) (*models.ScheduledReport, error) {
	return rs.repo.GetScheduledReportByID(ctx, id)
}

// ListScheduledReports returns all scheduled reports.
func (rs *ReportsService) ListScheduledReports(ctx context.Context) ([]*models.ScheduledReport, error) {
	if value, ok := rs.reportCache().get("scheduled_reports"); ok {
		return cloneScheduledReports(value.([]*models.ScheduledReport)), nil
	}
	reports, err := rs.repo.ListScheduledReports(ctx)
	if err != nil {
		return nil, err
	}
	rs.reportCache().set("scheduled_reports", cloneScheduledReports(reports))
	return cloneScheduledReports(reports), nil
}

// UpdateScheduledReport updates a scheduled report.
func (rs *ReportsService) UpdateScheduledReport(ctx context.Context, id int64, name, reportType, scheduleCron string, recipientEmails []string, filters map[string]any, format string) (*models.ScheduledReport, error) {
	report, err := rs.repo.UpdateScheduledReport(ctx, id, name, reportType, scheduleCron, recipientEmails, filters, format)
	if err == nil {
		rs.clearReportCache()
	}
	return report, err
}

// DeleteScheduledReport removes a scheduled report.
func (rs *ReportsService) DeleteScheduledReport(ctx context.Context, id int64) error {
	err := rs.repo.DeleteScheduledReport(ctx, id)
	if err == nil {
		rs.clearReportCache()
	}
	return err
}

// RunScheduledReport generates and emails a report.
func (rs *ReportsService) RunScheduledReport(ctx context.Context, report *models.ScheduledReport) error {
	data, filename, contentType, err := rs.generateReportData(ctx, report)
	if err != nil {
		return fmt.Errorf("generating report: %w", err)
	}

	for _, recipient := range report.RecipientEmails {
		subject := fmt.Sprintf("Scheduled Report: %s", report.Name)
		body := fmt.Sprintf(
			"Please find attached the scheduled report: %s\nType: %s\nFormat: %s\nFile: %s\nContent-Type: %s\n\n---\n%s",
			report.Name, report.ReportType, report.Format, filename, contentType, string(data),
		)
		if err := rs.email.Send(recipient, subject, body); err != nil {
			slog.Error("reports: send scheduled report failed", "recipient", recipient, "error", err)
		}
	}

	err = rs.repo.UpdateScheduledReportRunTime(ctx, report.ID, time.Now().UTC())
	if err == nil {
		rs.clearReportCache()
	}
	return err
}

// generateReportData builds the report payload for a given report definition.
func (rs *ReportsService) generateReportData(ctx context.Context, report *models.ScheduledReport) ([]byte, string, string, error) {
	switch report.ReportType {
	case "utilisation_summary":
		trends, err := rs.GetUtilizationTrends(ctx)
		if err != nil {
			return nil, "", "", err
		}
		headers := []string{"subnet_id", "cidr", "description", "current_pct", "week_ago_pct", "delta_pct"}
		rows := make([]map[string]string, len(trends))
		for i, t := range trends {
			rows[i] = map[string]string{
				"subnet_id":    strconv.FormatInt(t.SubnetID, 10),
				"cidr":         t.CIDR,
				"description":  t.Description,
				"current_pct":  fmt.Sprintf("%.2f", t.CurrentPct),
				"week_ago_pct": fmt.Sprintf("%.2f", t.WeekAgoPct),
				"delta_pct":    fmt.Sprintf("%.2f", t.DeltaPct),
			}
		}
		return buildReport(report.Format, "Utilisation Summary", headers, rows)

	case "inactive_ips":
		days := 90
		if v, ok := report.Filters["days"]; ok {
			if d, ok := v.(float64); ok {
				days = int(d)
			}
		}
		ips, err := rs.GetInactiveIPs(ctx, days, nil)
		if err != nil {
			return nil, "", "", err
		}
		headers := []string{"ip_id", "ip_address", "hostname", "subnet_cidr", "section_name", "device_id", "days_inactive"}
		rows := make([]map[string]string, len(ips))
		for i, ip := range ips {
			deviceID := ""
			if ip.DeviceID != nil {
				deviceID = strconv.FormatInt(*ip.DeviceID, 10)
			}
			rows[i] = map[string]string{
				"ip_id":         strconv.FormatInt(ip.IPID, 10),
				"ip_address":    ip.IPAddress,
				"hostname":      ip.Hostname,
				"subnet_cidr":   ip.SubnetCIDR,
				"section_name":  ip.NetworkName,
				"device_id":     deviceID,
				"days_inactive": strconv.Itoa(ip.DaysInactive),
			}
		}
		return buildReport(report.Format, "Inactive IPs", headers, rows)

	case "subnet_gaps":
		gaps, err := rs.repo.GetSubnetGaps(ctx)
		if err != nil {
			return nil, "", "", err
		}
		headers := []string{"subnet_id", "cidr", "description", "total_ips", "used_ips", "free_ips", "used_pct"}
		rows := make([]map[string]string, len(gaps))
		for i, g := range gaps {
			rows[i] = map[string]string{
				"subnet_id":   strconv.FormatInt(g.SubnetID, 10),
				"cidr":        g.CIDR,
				"description": g.Description,
				"total_ips":   strconv.Itoa(g.TotalIPs),
				"used_ips":    strconv.Itoa(g.UsedIPs),
				"free_ips":    strconv.Itoa(g.FreeIPs),
				"used_pct":    fmt.Sprintf("%.2f", g.UsedPct),
			}
		}
		return buildReport(report.Format, "Subnet Gaps", headers, rows)

	case "vlan_assignment":
		vlans, err := rs.repo.GetVLANAssignment(ctx)
		if err != nil {
			return nil, "", "", err
		}
		headers := []string{"vlan_id", "vlan_name", "vlan_tag", "subnet_count", "subnet_cidrs"}
		rows := make([]map[string]string, len(vlans))
		for i, v := range vlans {
			rows[i] = map[string]string{
				"vlan_id":      strconv.FormatInt(v.VLANID, 10),
				"vlan_name":    v.VLANName,
				"vlan_tag":     strconv.Itoa(v.VLANTag),
				"subnet_count": strconv.Itoa(v.SubnetCount),
				"subnet_cidrs": v.SubnetCIDRs,
			}
		}
		return buildReport(report.Format, "VLAN Assignment", headers, rows)

	case "ip_age":
		ips, err := rs.repo.GetIPAge(ctx)
		if err != nil {
			return nil, "", "", err
		}
		headers := []string{"ip_id", "address", "status", "device_id", "days_old", "days_since_seen"}
		rows := make([]map[string]string, len(ips))
		for i, ip := range ips {
			daysSinceSeen := "never"
			if ip.DaysSinceSeen >= 0 {
				daysSinceSeen = strconv.Itoa(ip.DaysSinceSeen)
			}
			deviceID := ""
			if ip.DeviceID != nil {
				deviceID = strconv.FormatInt(*ip.DeviceID, 10)
			}
			rows[i] = map[string]string{
				"ip_id":           strconv.FormatInt(ip.IPID, 10),
				"address":         ip.Address,
				"status":          ip.Status,
				"device_id":       deviceID,
				"days_old":        strconv.Itoa(ip.DaysOld),
				"days_since_seen": daysSinceSeen,
			}
		}
		return buildReport(report.Format, "IP Age", headers, rows)

	case "dns_audit":
		entries, err := rs.repo.GetDNSAudit(ctx)
		if err != nil {
			return nil, "", "", err
		}
		headers := []string{"ip_id", "address", "dns_name", "ptr_record", "dns_last_checked"}
		rows := make([]map[string]string, len(entries))
		for i, e := range entries {
			rows[i] = map[string]string{
				"ip_id":            strconv.FormatInt(e.IPID, 10),
				"address":          e.Address,
				"dns_name":         e.DNSName,
				"ptr_record":       e.PTRRecord,
				"dns_last_checked": e.DNSLastChecked,
			}
		}
		return buildReport(report.Format, "DNS Audit", headers, rows)

	case "stale_leases":
		days := 30
		if v, ok := report.Filters["days"]; ok {
			if d, ok := v.(float64); ok {
				days = int(d)
			}
		}
		ips, err := rs.GetInactiveIPs(ctx, days, nil)
		if err != nil {
			return nil, "", "", err
		}
		headers := []string{"ip_id", "ip_address", "hostname", "subnet_cidr", "section_name", "days_inactive"}
		rows := make([]map[string]string, len(ips))
		for i, ip := range ips {
			rows[i] = map[string]string{
				"ip_id":         strconv.FormatInt(ip.IPID, 10),
				"ip_address":    ip.IPAddress,
				"hostname":      ip.Hostname,
				"subnet_cidr":   ip.SubnetCIDR,
				"section_name":  ip.NetworkName,
				"days_inactive": strconv.Itoa(ip.DaysInactive),
			}
		}
		return buildReport(report.Format, "Stale Leases", headers, rows)

	case "inactive_devices":
		days := 30
		if v, ok := report.Filters["days"]; ok {
			if d, ok := v.(float64); ok {
				days = int(d)
			}
		}
		devices, err := rs.repo.GetInactiveDevices(ctx, days)
		if err != nil {
			return nil, "", "", err
		}
		headers := []string{"device_id", "hostname", "vendor", "model", "days_inactive"}
		rows := make([]map[string]string, len(devices))
		for i, d := range devices {
			rows[i] = map[string]string{
				"device_id":     strconv.FormatInt(d.DeviceID, 10),
				"hostname":      d.Hostname,
				"vendor":        d.Vendor,
				"model":         d.Model,
				"days_inactive": strconv.Itoa(d.DaysInactive),
			}
		}
		return buildReport(report.Format, "Inactive Devices", headers, rows)

	case "failed_scans":
		days := 7
		if v, ok := report.Filters["days"]; ok {
			if d, ok := v.(float64); ok {
				days = int(d)
			}
		}
		jobs, err := rs.repo.GetOverdueScanJobs(ctx, days)
		if err != nil {
			return nil, "", "", err
		}
		headers := []string{"job_id", "job_name", "schedule_cron", "days_since_run"}
		rows := make([]map[string]string, len(jobs))
		for i, j := range jobs {
			rows[i] = map[string]string{
				"job_id":         strconv.FormatInt(j.JobID, 10),
				"job_name":       j.JobName,
				"schedule_cron":  j.ScheduleCron,
				"days_since_run": strconv.Itoa(j.DaysSinceRun),
			}
		}
		return buildReport(report.Format, "Failed Scans", headers, rows)

	default:
		return nil, "", "", fmt.Errorf("unsupported report type %q", report.ReportType)
	}
}

// buildReport produces either a CSV or PDF byte slice.
func buildReport(format, title string, headers []string, rows []map[string]string) ([]byte, string, string, error) {
	switch format {
	case "pdf":
		pdfRows := make([][]string, len(rows))
		for i, row := range rows {
			r := make([]string, len(headers))
			for j, h := range headers {
				r[j] = row[h]
			}
			pdfRows[i] = r
		}
		data, err := export.GeneratePDF(title, headers, pdfRows)
		return data, "report.pdf", "application/pdf", err
	default:
		data, err := export.GenerateCSV(headers, rows)
		return data, "report.csv", "text/csv", err
	}
}

func (rs *ReportsService) reportCache() *ttlCache[any] {
	if rs.cache == nil {
		rs.cache = newTTLCache[any](30 * time.Second)
	}
	return rs.cache
}

func (rs *ReportsService) clearReportCache() {
	if rs.cache != nil {
		rs.cache.clear()
	}
}

func sectionCacheKey(networkID *int64) string {
	if networkID == nil {
		return "all"
	}
	return strconv.FormatInt(*networkID, 10)
}

func cloneUtilizationTrends(trends []*models.SubnetUtilizationTrend) []*models.SubnetUtilizationTrend {
	out := make([]*models.SubnetUtilizationTrend, 0, len(trends))
	for _, trend := range trends {
		if trend == nil {
			out = append(out, nil)
			continue
		}
		clone := *trend
		out = append(out, &clone)
	}
	return out
}

func cloneInactiveIPs(ips []*models.InactiveIPReport) []*models.InactiveIPReport {
	out := make([]*models.InactiveIPReport, 0, len(ips))
	for _, ip := range ips {
		if ip == nil {
			out = append(out, nil)
			continue
		}
		clone := *ip
		if ip.LastSeen != nil {
			lastSeen := *ip.LastSeen
			clone.LastSeen = &lastSeen
		}
		out = append(out, &clone)
	}
	return out
}

func cloneScheduledReports(reports []*models.ScheduledReport) []*models.ScheduledReport {
	out := make([]*models.ScheduledReport, 0, len(reports))
	for _, report := range reports {
		if report == nil {
			out = append(out, nil)
			continue
		}
		out = append(out, cloneScheduledReport(report))
	}
	return out
}

func cloneScheduledReport(report *models.ScheduledReport) *models.ScheduledReport {
	if report == nil {
		return nil
	}
	clone := *report
	clone.RecipientEmails = append([]string(nil), report.RecipientEmails...)
	if report.Filters != nil {
		clone.Filters = make(map[string]any, len(report.Filters))
		for key, value := range report.Filters {
			clone.Filters[key] = value
		}
	}
	if report.LastRunAt != nil {
		lastRunAt := *report.LastRunAt
		clone.LastRunAt = &lastRunAt
	}
	return &clone
}

func cloneDuplicatesReport(report *models.DuplicatesReport) *models.DuplicatesReport {
	if report == nil {
		return nil
	}
	clone := &models.DuplicatesReport{
		DuplicateHostnames: append([]models.DuplicateHostname(nil), report.DuplicateHostnames...),
		ConflictingIPs:     append([]models.ConflictingIP(nil), report.ConflictingIPs...),
	}
	for i := range clone.DuplicateHostnames {
		clone.DuplicateHostnames[i].DeviceIDs = append([]int64(nil), clone.DuplicateHostnames[i].DeviceIDs...)
	}
	for i := range clone.ConflictingIPs {
		clone.ConflictingIPs[i].Hostnames = append([]string(nil), clone.ConflictingIPs[i].Hostnames...)
	}
	return clone
}

// StartScheduledReportJob launches a background goroutine that checks and runs due reports.
func (rs *ReportsService) StartScheduledReportJob(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case now := <-ticker.C:
				reports, err := rs.ListScheduledReports(ctx)
				if err != nil {
					slog.Error("reports: list scheduled reports failed", "error", err)
					continue
				}
				for _, rpt := range reports {
					if reportIsDue(rpt, now) {
						rptCopy := rpt
						go func() {
							if err := rs.RunScheduledReport(ctx, rptCopy); err != nil {
								slog.Error("reports: run scheduled report failed", "report_id", rptCopy.ID, "error", err)
							}
						}()
					}
				}
			}
		}
	}()
}

// reportIsDue returns true if the report's cron schedule matches the given time.
func reportIsDue(rpt *models.ScheduledReport, now time.Time) bool {
	return matchesCron(rpt.ScheduleCron, now)
}

// ─────────────────────────────────────────────────────────────────────────────
// Inactive IP reclamation (#224)
// ─────────────────────────────────────────────────────────────────────────────

// GetInactiveIPs returns IPs that have been inactive for the given number of days.
func (rs *ReportsService) GetInactiveIPs(ctx context.Context, days int, networkID *int64) ([]*models.InactiveIPReport, error) {
	key := fmt.Sprintf("inactive_ips:%d:%s", days, sectionCacheKey(networkID))
	if value, ok := rs.reportCache().get(key); ok {
		return cloneInactiveIPs(value.([]*models.InactiveIPReport)), nil
	}
	ips, err := rs.repo.GetInactiveIPs(ctx, days, networkID)
	if err != nil {
		return nil, err
	}
	rs.reportCache().set(key, cloneInactiveIPs(ips))
	return cloneInactiveIPs(ips), nil
}

// GetDNSAudit returns all IPs with DNS name tracking data.
func (rs *ReportsService) GetDNSAudit(ctx context.Context) ([]*repository.DNSAuditRow, error) {
	return rs.repo.GetDNSAudit(ctx)
}

// BulkReleaseIPs releases a set of IPs back to 'available' and logs the action.
func (rs *ReportsService) BulkReleaseIPs(ctx context.Context, ipIDs []int64, operatorUserID int64) (int64, error) {
	count, err := rs.repo.BulkReleaseIPs(ctx, ipIDs)
	if err != nil {
		return 0, err
	}
	rs.clearReportCache()

	if rs.audit != nil {
		rs.audit.Log(ctx, AuditEntry{
			UserID:       &operatorUserID,
			Action:       "ip.bulk_release",
			ResourceType: "ip_address",
			ResourceName: fmt.Sprintf("%d IPs", count),
			Status:       "success",
		})
	}

	return count, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Duplicate detection (#425)
// ─────────────────────────────────────────────────────────────────────────────

// GetDuplicates returns a report of duplicate device hostnames and conflicting IP assignments.
func (rs *ReportsService) GetDuplicates(ctx context.Context) (*models.DuplicatesReport, error) {
	if value, ok := rs.reportCache().get("duplicates"); ok {
		return cloneDuplicatesReport(value.(*models.DuplicatesReport)), nil
	}
	report, err := rs.repo.GetDuplicates(ctx)
	if err != nil {
		return nil, err
	}
	rs.reportCache().set("duplicates", cloneDuplicatesReport(report))
	return cloneDuplicatesReport(report), nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Export helpers (#223)
// ─────────────────────────────────────────────────────────────────────────────

// ExportSubnets builds a CSV/PDF of all subnets with utilization data.
func (rs *ReportsService) ExportSubnets(ctx context.Context, format string) ([]byte, string, string, error) {
	trends, err := rs.GetUtilizationTrends(ctx)
	if err != nil {
		return nil, "", "", err
	}

	headers := []string{"cidr", "description", "current_pct"}
	rows := make([]map[string]string, len(trends))
	for i, t := range trends {
		rows[i] = map[string]string{
			"cidr":        t.CIDR,
			"description": t.Description,
			"current_pct": fmt.Sprintf("%.2f", t.CurrentPct),
		}
	}
	return buildReport(format, "Subnet Report", headers, rows)
}

// ExportIPs builds a CSV/PDF of IPs for a specific subnet.
func (rs *ReportsService) ExportIPs(ctx context.Context, subnetID int64, format string) ([]byte, string, string, error) {
	ips, err := rs.repo.ListIPAddressesBySubnet(ctx, subnetID)
	if err != nil {
		return nil, "", "", err
	}

	headers := []string{"ip_address", "hostname", "status", "device_id"}
	rows := make([]map[string]string, len(ips))
	for i, ip := range ips {
		deviceID := ""
		if ip.DeviceID != nil {
			deviceID = strconv.FormatInt(*ip.DeviceID, 10)
		}
		rows[i] = map[string]string{
			"ip_address": ip.Address,
			"hostname":   ip.Hostname,
			"status":     ip.Status,
			"device_id":  deviceID,
		}
	}
	return buildReport(format, "IP Addresses", headers, rows)
}

// ExportInactiveIPs builds a CSV/PDF of inactive IPs.
func (rs *ReportsService) ExportInactiveIPs(ctx context.Context, days int, format string) ([]byte, string, string, error) {
	ips, err := rs.GetInactiveIPs(ctx, days, nil)
	if err != nil {
		return nil, "", "", err
	}

	headers := []string{"ip_address", "hostname", "subnet_cidr", "section_name", "days_inactive"}
	rows := make([]map[string]string, len(ips))
	for i, ip := range ips {
		rows[i] = map[string]string{
			"ip_address":    ip.IPAddress,
			"hostname":      ip.Hostname,
			"subnet_cidr":   ip.SubnetCIDR,
			"section_name":  ip.NetworkName,
			"days_inactive": strconv.Itoa(ip.DaysInactive),
		}
	}
	return buildReport(format, "Inactive IPs", headers, rows)
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

// totalAddressesFromPrefix returns the total number of host addresses in a prefix.
// For IPv4 /24 → 256, /32 → 1. For IPv6 we cap at a reasonable max.
func totalAddressesFromPrefix(prefix int) int {
	if prefix < 0 {
		return 0
	}
	if prefix <= 32 {
		// IPv4
		return 1 << uint(32-prefix)
	}
	// IPv6 — cap at 2^32 for sanity
	bits := 128 - prefix
	if bits > 32 {
		bits = 32
	}
	return 1 << uint(bits)
}

