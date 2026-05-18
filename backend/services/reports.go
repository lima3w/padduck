package services

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"ipam-next/internal/export"
	"ipam-next/models"
	"ipam-next/repository"
)

// reportsRepo lists the repository methods used by ReportsService.
type reportsRepo interface {
	// utilisation
	ListAllSubnets(ctx context.Context) ([]*models.Subnet, error)
	ListSubnetsBySection(ctx context.Context, sectionID int64) ([]*models.Subnet, error)
	ListIPAddressesBySubnet(ctx context.Context, subnetID int64) ([]*models.IPAddress, error)
	RecordUtilisationSnapshot(ctx context.Context, subnetID int64, used, total int, pct float64) error
	GetUtilisationHistory(ctx context.Context, subnetID int64, days int) ([]*models.SubnetUtilisationPoint, error)
	GetUtilisationTrends(ctx context.Context) ([]*models.SubnetUtilisationTrend, error)
	GetLatestUtilisationForSubnet(ctx context.Context, subnetID int64) (*models.SubnetUtilisationPoint, error)
	GetSubnetsByUtilisationThreshold(ctx context.Context, thresholdPct float64) ([]*models.SubnetUtilisationTrend, error)
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
	GetInactiveIPs(ctx context.Context, days int, sectionID *int64) ([]*models.InactiveIPReport, error)
	BulkReleaseIPs(ctx context.Context, ipIDs []int64) (int64, error)
	// sections and subnets for export
	ListAllSections(ctx context.Context) ([]*models.Section, error)
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
}

// NewReportsService creates a new ReportsService.
func NewReportsService(repo reportsRepo, config *ConfigService, email *EmailService, audit *AuditService) *ReportsService {
	return &ReportsService{
		repo:   repo,
		config: config,
		email:  email,
		audit:  audit,
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Utilisation snapshots (#220)
// ─────────────────────────────────────────────────────────────────────────────

// TakeUtilisationSnapshots iterates all subnets and records a utilisation snapshot for each.
func (rs *ReportsService) TakeUtilisationSnapshots(ctx context.Context) {
	subnets, err := rs.repo.ListAllSubnets(ctx)
	if err != nil {
		log.Printf("[reports] list subnets for snapshot: %v", err)
		return
	}

	for _, subnet := range subnets {
		ips, err := rs.repo.ListIPAddressesBySubnet(ctx, subnet.ID)
		if err != nil {
			log.Printf("[reports] list IPs for subnet %d: %v", subnet.ID, err)
			continue
		}

		// Count used IPs (assigned or reserved)
		used := 0
		for _, ip := range ips {
			if ip.Status == "assigned" || ip.Status == "reserved" {
				used++
			}
		}

		// Total addresses in the CIDR
		total := totalAddressesFromPrefix(subnet.PrefixLength)

		var pct float64
		if total > 0 {
			pct = float64(used) / float64(total) * 100
		}

		if err := rs.repo.RecordUtilisationSnapshot(ctx, subnet.ID, used, total, pct); err != nil {
			log.Printf("[reports] record snapshot for subnet %d: %v", subnet.ID, err)
		}
	}

	// After recording snapshots, check threshold alerts.
	rs.CheckThresholdAlerts(ctx)
}

// GetUtilisationHistory returns utilisation history for a subnet.
func (rs *ReportsService) GetUtilisationHistory(ctx context.Context, subnetID int64, days int) ([]*models.SubnetUtilisationPoint, error) {
	return rs.repo.GetUtilisationHistory(ctx, subnetID, days)
}

// GetUtilisationTrends returns trend data for all subnets.
func (rs *ReportsService) GetUtilisationTrends(ctx context.Context) ([]*models.SubnetUtilisationTrend, error) {
	return rs.repo.GetUtilisationTrends(ctx)
}

// StartUtilisationSnapshotJob launches a background goroutine that periodically takes snapshots.
func (rs *ReportsService) StartUtilisationSnapshotJob(ctx context.Context) {
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
				rs.TakeUtilisationSnapshots(ctx)
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
		log.Printf("[reports] list subnets with thresholds: %v", err)
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

		latest, err := rs.repo.GetLatestUtilisationForSubnet(ctx, subnet.ID)
		if err != nil || latest == nil {
			continue
		}

		currentPct := latest.UtilisationPct
		thresholdF := float64(threshold)

		cidr := fmt.Sprintf("%s/%d", subnet.NetworkAddress, subnet.PrefixLength)

		if currentPct >= thresholdF {
			// Check cooldown
			cooldown, err := rs.repo.GetAlertCooldown(ctx, subnet.ID)
			if err != nil {
				log.Printf("[reports] get cooldown for subnet %d: %v", subnet.ID, err)
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
					log.Printf("[reports] send alert email for subnet %d: %v", subnet.ID, err)
				}
			}

			// Set cooldown
			if err := rs.repo.SetAlertCooldown(ctx, subnet.ID, currentPct); err != nil {
				log.Printf("[reports] set cooldown for subnet %d: %v", subnet.ID, err)
			}
		} else if currentPct < thresholdF-5 {
			// Clear cooldown to allow future re-alerting
			if err := rs.repo.ClearAlertCooldown(ctx, subnet.ID); err != nil {
				log.Printf("[reports] clear cooldown for subnet %d: %v", subnet.ID, err)
			}
		}
	}
}

// GetSubnetsNearCapacity returns subnets whose latest utilisation exceeds the threshold.
func (rs *ReportsService) GetSubnetsNearCapacity(ctx context.Context, thresholdPct float64) ([]*models.SubnetUtilisationTrend, error) {
	return rs.repo.GetSubnetsByUtilisationThreshold(ctx, thresholdPct)
}

// ─────────────────────────────────────────────────────────────────────────────
// Scheduled reports (#222)
// ─────────────────────────────────────────────────────────────────────────────

// CreateScheduledReport creates a new scheduled report.
func (rs *ReportsService) CreateScheduledReport(ctx context.Context, name, reportType, scheduleCron string, recipientEmails []string, filters map[string]any, format string, createdBy int64) (*models.ScheduledReport, error) {
	return rs.repo.CreateScheduledReport(ctx, name, reportType, scheduleCron, recipientEmails, filters, format, createdBy)
}

// GetScheduledReport retrieves a scheduled report by ID.
func (rs *ReportsService) GetScheduledReport(ctx context.Context, id int64) (*models.ScheduledReport, error) {
	return rs.repo.GetScheduledReportByID(ctx, id)
}

// ListScheduledReports returns all scheduled reports.
func (rs *ReportsService) ListScheduledReports(ctx context.Context) ([]*models.ScheduledReport, error) {
	return rs.repo.ListScheduledReports(ctx)
}

// UpdateScheduledReport updates a scheduled report.
func (rs *ReportsService) UpdateScheduledReport(ctx context.Context, id int64, name, reportType, scheduleCron string, recipientEmails []string, filters map[string]any, format string) (*models.ScheduledReport, error) {
	return rs.repo.UpdateScheduledReport(ctx, id, name, reportType, scheduleCron, recipientEmails, filters, format)
}

// DeleteScheduledReport removes a scheduled report.
func (rs *ReportsService) DeleteScheduledReport(ctx context.Context, id int64) error {
	return rs.repo.DeleteScheduledReport(ctx, id)
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
			log.Printf("[reports] send scheduled report to %s: %v", recipient, err)
		}
	}

	return rs.repo.UpdateScheduledReportRunTime(ctx, report.ID, time.Now())
}

// generateReportData builds the report payload for a given report definition.
func (rs *ReportsService) generateReportData(ctx context.Context, report *models.ScheduledReport) ([]byte, string, string, error) {
	switch report.ReportType {
	case "utilisation_summary":
		trends, err := rs.repo.GetUtilisationTrends(ctx)
		if err != nil {
			return nil, "", "", err
		}
		headers := []string{"subnet_id", "cidr", "description", "current_pct", "week_ago_pct", "delta_pct"}
		rows := make([]map[string]string, len(trends))
		for i, t := range trends {
			rows[i] = map[string]string{
				"subnet_id":   strconv.FormatInt(t.SubnetID, 10),
				"cidr":        t.CIDR,
				"description": t.Description,
				"current_pct": fmt.Sprintf("%.2f", t.CurrentPct),
				"week_ago_pct": fmt.Sprintf("%.2f", t.WeekAgoPct),
				"delta_pct":   fmt.Sprintf("%.2f", t.DeltaPct),
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
		ips, err := rs.repo.GetInactiveIPs(ctx, days, nil)
		if err != nil {
			return nil, "", "", err
		}
		headers := []string{"ip_id", "ip_address", "hostname", "subnet_cidr", "section_name", "assigned_to", "days_inactive"}
		rows := make([]map[string]string, len(ips))
		for i, ip := range ips {
			assignedTo := ""
			if ip.AssignedTo != nil {
				assignedTo = *ip.AssignedTo
			}
			rows[i] = map[string]string{
				"ip_id":        strconv.FormatInt(ip.IPID, 10),
				"ip_address":   ip.IPAddress,
				"hostname":     ip.Hostname,
				"subnet_cidr":  ip.SubnetCIDR,
				"section_name": ip.SectionName,
				"assigned_to":  assignedTo,
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
		headers := []string{"ip_id", "address", "status", "assigned_to", "days_old", "days_since_seen"}
		rows := make([]map[string]string, len(ips))
		for i, ip := range ips {
			daysSinceSeen := "never"
			if ip.DaysSinceSeen >= 0 {
				daysSinceSeen = strconv.Itoa(ip.DaysSinceSeen)
			}
			rows[i] = map[string]string{
				"ip_id":           strconv.FormatInt(ip.IPID, 10),
				"address":         ip.Address,
				"status":          ip.Status,
				"assigned_to":     ip.AssignedTo,
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
		ips, err := rs.repo.GetInactiveIPs(ctx, days, nil)
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
				"section_name":  ip.SectionName,
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
				"job_id":        strconv.FormatInt(j.JobID, 10),
				"job_name":      j.JobName,
				"schedule_cron": j.ScheduleCron,
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
				reports, err := rs.repo.ListScheduledReports(ctx)
				if err != nil {
					log.Printf("[reports] list scheduled reports: %v", err)
					continue
				}
				for _, rpt := range reports {
					if reportIsDue(rpt, now) {
						rptCopy := rpt
						go func() {
							if err := rs.RunScheduledReport(ctx, rptCopy); err != nil {
								log.Printf("[reports] run scheduled report %d: %v", rptCopy.ID, err)
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
func (rs *ReportsService) GetInactiveIPs(ctx context.Context, days int, sectionID *int64) ([]*models.InactiveIPReport, error) {
	return rs.repo.GetInactiveIPs(ctx, days, sectionID)
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

	if rs.audit != nil {
		rs.audit.Log(ctx, AuditEntry{
			UserID: &operatorUserID,
			Action: "ip.bulk_release",
			ResourceType: "ip_address",
			ResourceName: fmt.Sprintf("%d IPs", count),
			Status: "success",
		})
	}

	return count, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Duplicate detection (#425)
// ─────────────────────────────────────────────────────────────────────────────

// GetDuplicates returns a report of duplicate device hostnames and conflicting IP assignments.
func (rs *ReportsService) GetDuplicates(ctx context.Context) (*models.DuplicatesReport, error) {
	return rs.repo.GetDuplicates(ctx)
}

// ─────────────────────────────────────────────────────────────────────────────
// Export helpers (#223)
// ─────────────────────────────────────────────────────────────────────────────

// ExportSubnets builds a CSV/PDF of all subnets with utilisation data.
func (rs *ReportsService) ExportSubnets(ctx context.Context, format string) ([]byte, string, string, error) {
	trends, err := rs.repo.GetUtilisationTrends(ctx)
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

	headers := []string{"ip_address", "hostname", "status", "assigned_to"}
	rows := make([]map[string]string, len(ips))
	for i, ip := range ips {
		assignedTo := ""
		if ip.AssignedTo != nil {
			assignedTo = *ip.AssignedTo
		}
		rows[i] = map[string]string{
			"ip_address":  ip.Address,
			"hostname":    ip.Hostname,
			"status":      ip.Status,
			"assigned_to": assignedTo,
		}
	}
	return buildReport(format, "IP Addresses", headers, rows)
}

// ExportInactiveIPs builds a CSV/PDF of inactive IPs.
func (rs *ReportsService) ExportInactiveIPs(ctx context.Context, days int, format string) ([]byte, string, string, error) {
	ips, err := rs.repo.GetInactiveIPs(ctx, days, nil)
	if err != nil {
		return nil, "", "", err
	}

	headers := []string{"ip_address", "hostname", "subnet_cidr", "section_name", "days_inactive"}
	rows := make([]map[string]string, len(ips))
	for i, ip := range ips {
		rows[i] = map[string]string{
			"ip_address":   ip.IPAddress,
			"hostname":     ip.Hostname,
			"subnet_cidr":  ip.SubnetCIDR,
			"section_name": ip.SectionName,
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

// cidrFromSubnet constructs the CIDR string for a subnet.
func cidrFromSubnet(s *models.Subnet) string {
	return fmt.Sprintf("%s/%d", s.NetworkAddress, s.PrefixLength)
}

// parseIPNet is a helper to parse a CIDR string.
func parseIPNet(cidr string) (*net.IPNet, error) {
	_, ipNet, err := net.ParseCIDR(cidr)
	return ipNet, err
}

// subnetSectionName returns the section name by extracting it from a join. Placeholder.
func subnetSectionName(s *models.Subnet) string {
	_ = s
	return ""
}

// Ensure unused imports are used.
var _ = strings.TrimSpace
var _ = cidrFromSubnet
var _ = parseIPNet
var _ = subnetSectionName
