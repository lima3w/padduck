package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"padduck/models"
)

// ─────────────────────────────────────────────────────────────────────────────
// Utilisation history (#220)
// ─────────────────────────────────────────────────────────────────────────────

// RecordUtilisationSnapshot inserts a utilisation data point for a subnet.
func (r *Repository) RecordUtilisationSnapshot(ctx context.Context, subnetID int64, used, total int, pct float64) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO subnet_utilisation_history (subnet_id, used_count, total_count, utilisation_pct)
		 VALUES ($1, $2, $3, $4)`,
		subnetID, used, total, pct,
	)
	return err
}

// GetUtilisationHistory returns ordered utilisation data points for a subnet over the last N days.
func (r *Repository) GetUtilisationHistory(ctx context.Context, subnetID int64, days int) ([]*models.SubnetUtilisationPoint, error) {
	rows, err := r.db.Query(ctx,
		`SELECT recorded_at, used_count, total_count, utilisation_pct
		 FROM subnet_utilisation_history
		 WHERE subnet_id = $1 AND recorded_at >= now() - ($2 || ' days')::interval
		 ORDER BY recorded_at ASC`,
		subnetID, days,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []*models.SubnetUtilisationPoint
	for rows.Next() {
		p := &models.SubnetUtilisationPoint{}
		if err := rows.Scan(&p.RecordedAt, &p.UsedCount, &p.TotalCount, &p.UtilisationPct); err != nil {
			return nil, err
		}
		points = append(points, p)
	}
	return points, rows.Err()
}

// GetUtilisationTrends returns current utilisation and the delta vs 7 days ago for all subnets.
func (r *Repository) GetUtilisationTrends(ctx context.Context) ([]*models.SubnetUtilisationTrend, error) {
	rows, err := r.db.Query(ctx, `
		WITH latest AS (
			SELECT DISTINCT ON (subnet_id)
				subnet_id, utilisation_pct AS current_pct
			FROM subnet_utilisation_history
			ORDER BY subnet_id, recorded_at DESC
		),
		week_ago AS (
			SELECT DISTINCT ON (subnet_id)
				subnet_id, utilisation_pct AS week_ago_pct
			FROM subnet_utilisation_history
			WHERE recorded_at BETWEEN now() - interval '8 days' AND now() - interval '6 days'
			ORDER BY subnet_id, recorded_at DESC
		)
		SELECT s.id,
		       host(s.network_address) || '/' || s.prefix_length AS cidr,
		       s.description,
		       COALESCE(l.current_pct, 0) AS current_pct,
		       COALESCE(w.week_ago_pct, 0) AS week_ago_pct,
		       COALESCE(l.current_pct, 0) - COALESCE(w.week_ago_pct, 0) AS delta_pct
		FROM subnets s
		LEFT JOIN latest l ON l.subnet_id = s.id
		LEFT JOIN week_ago w ON w.subnet_id = s.id
		ORDER BY delta_pct DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trends []*models.SubnetUtilisationTrend
	for rows.Next() {
		t := &models.SubnetUtilisationTrend{}
		if err := rows.Scan(&t.SubnetID, &t.CIDR, &t.Description, &t.CurrentPct, &t.WeekAgoPct, &t.DeltaPct); err != nil {
			return nil, err
		}
		trends = append(trends, t)
	}
	return trends, rows.Err()
}

// ─────────────────────────────────────────────────────────────────────────────
// Alert cooldowns (#221)
// ─────────────────────────────────────────────────────────────────────────────

// GetAlertCooldown returns the cooldown record for a subnet, or nil if none exists.
func (r *Repository) GetAlertCooldown(ctx context.Context, subnetID int64) (*models.AlertCooldown, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, subnet_id, alerted_at, alerted_pct FROM alert_cooldowns WHERE subnet_id = $1`,
		subnetID,
	)
	cd := &models.AlertCooldown{}
	err := row.Scan(&cd.ID, &cd.SubnetID, &cd.AlertedAt, &cd.AlertedPct)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return cd, nil
}

// SetAlertCooldown upserts a cooldown record for a subnet.
func (r *Repository) SetAlertCooldown(ctx context.Context, subnetID int64, pct float64) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO alert_cooldowns (subnet_id, alerted_at, alerted_pct)
		 VALUES ($1, now(), $2)
		 ON CONFLICT (subnet_id) DO UPDATE SET alerted_at = now(), alerted_pct = EXCLUDED.alerted_pct`,
		subnetID, pct,
	)
	return err
}

// ClearAlertCooldown removes the cooldown record for a subnet.
func (r *Repository) ClearAlertCooldown(ctx context.Context, subnetID int64) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM alert_cooldowns WHERE subnet_id = $1`,
		subnetID,
	)
	return err
}

// ListSubnetsWithThresholds returns subnets that have alert_threshold_pct set.
func (r *Repository) ListSubnetsWithThresholds(ctx context.Context) ([]*models.Subnet, error) {
	query := `SELECT ` + subnetSelectCols + ` ` + subnetFromJoin + `
	          WHERE s.alert_threshold_pct IS NOT NULL
	          ORDER BY s.network_address`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subnets []*models.Subnet
	for rows.Next() {
		subnet, err := scanSubnet(rows)
		if err != nil {
			return nil, err
		}
		subnets = append(subnets, subnet)
	}
	return subnets, rows.Err()
}

// ListAllSubnets returns every subnet across all sections.
func (r *Repository) ListAllSubnets(ctx context.Context) ([]*models.Subnet, error) {
	query := `SELECT ` + subnetSelectCols + ` ` + subnetFromJoin + ` ORDER BY s.network_address`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subnets []*models.Subnet
	for rows.Next() {
		subnet, err := scanSubnet(rows)
		if err != nil {
			return nil, err
		}
		subnets = append(subnets, subnet)
	}
	return subnets, rows.Err()
}

// GetLatestUtilisationForSubnet returns the most recent utilisation record for a subnet.
func (r *Repository) GetLatestUtilisationForSubnet(ctx context.Context, subnetID int64) (*models.SubnetUtilisationPoint, error) {
	row := r.db.QueryRow(ctx,
		`SELECT recorded_at, used_count, total_count, utilisation_pct
		 FROM subnet_utilisation_history
		 WHERE subnet_id = $1
		 ORDER BY recorded_at DESC
		 LIMIT 1`,
		subnetID,
	)
	p := &models.SubnetUtilisationPoint{}
	err := row.Scan(&p.RecordedAt, &p.UsedCount, &p.TotalCount, &p.UtilisationPct)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return p, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Scheduled reports (#222)
// ─────────────────────────────────────────────────────────────────────────────

// CreateScheduledReport inserts a new scheduled report.
func (r *Repository) CreateScheduledReport(ctx context.Context, name, reportType, scheduleCron string, recipientEmails []string, filters map[string]any, format string, createdBy int64) (*models.ScheduledReport, error) {
	emailsJSON, err := json.Marshal(recipientEmails)
	if err != nil {
		return nil, fmt.Errorf("marshalling recipient emails: %w", err)
	}
	filtersJSON, err := json.Marshal(filters)
	if err != nil {
		return nil, fmt.Errorf("marshalling filters: %w", err)
	}

	row := r.db.QueryRow(ctx,
		`INSERT INTO scheduled_reports (name, report_type, schedule_cron, recipient_emails, filters, format, created_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, name, report_type, schedule_cron, recipient_emails, filters, format, last_run_at, created_by, created_at, updated_at`,
		name, reportType, scheduleCron, emailsJSON, filtersJSON, format, createdBy,
	)
	return scanScheduledReport(row)
}

// GetScheduledReportByID returns a scheduled report by primary key.
func (r *Repository) GetScheduledReportByID(ctx context.Context, id int64) (*models.ScheduledReport, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, name, report_type, schedule_cron, recipient_emails, filters, format, last_run_at, created_by, created_at, updated_at
		 FROM scheduled_reports WHERE id = $1`,
		id,
	)
	return scanScheduledReport(row)
}

// ListScheduledReports returns all scheduled reports ordered by name.
func (r *Repository) ListScheduledReports(ctx context.Context) ([]*models.ScheduledReport, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, report_type, schedule_cron, recipient_emails, filters, format, last_run_at, created_by, created_at, updated_at
		 FROM scheduled_reports ORDER BY name`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []*models.ScheduledReport
	for rows.Next() {
		rpt, err := scanScheduledReport(rows)
		if err != nil {
			return nil, err
		}
		reports = append(reports, rpt)
	}
	return reports, rows.Err()
}

// UpdateScheduledReport updates the mutable fields of a scheduled report.
func (r *Repository) UpdateScheduledReport(ctx context.Context, id int64, name, reportType, scheduleCron string, recipientEmails []string, filters map[string]any, format string) (*models.ScheduledReport, error) {
	emailsJSON, err := json.Marshal(recipientEmails)
	if err != nil {
		return nil, fmt.Errorf("marshalling recipient emails: %w", err)
	}
	filtersJSON, err := json.Marshal(filters)
	if err != nil {
		return nil, fmt.Errorf("marshalling filters: %w", err)
	}

	row := r.db.QueryRow(ctx,
		`UPDATE scheduled_reports
		 SET name = $2, report_type = $3, schedule_cron = $4, recipient_emails = $5, filters = $6, format = $7, updated_at = now()
		 WHERE id = $1
		 RETURNING id, name, report_type, schedule_cron, recipient_emails, filters, format, last_run_at, created_by, created_at, updated_at`,
		id, name, reportType, scheduleCron, emailsJSON, filtersJSON, format,
	)
	return scanScheduledReport(row)
}

// UpdateScheduledReportRunTime marks when a scheduled report was last run.
func (r *Repository) UpdateScheduledReportRunTime(ctx context.Context, id int64, t time.Time) error {
	_, err := r.db.Exec(ctx,
		`UPDATE scheduled_reports SET last_run_at = $2, updated_at = now() WHERE id = $1`,
		id, t,
	)
	return err
}

// DeleteScheduledReport removes a scheduled report.
func (r *Repository) DeleteScheduledReport(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM scheduled_reports WHERE id = $1`, id)
	return err
}

// scanScheduledReport scans a single row into a ScheduledReport.
func scanScheduledReport(row interface {
	Scan(dest ...any) error
}) (*models.ScheduledReport, error) {
	rpt := &models.ScheduledReport{}
	var emailsJSON, filtersJSON []byte
	err := row.Scan(
		&rpt.ID, &rpt.Name, &rpt.ReportType, &rpt.ScheduleCron,
		&emailsJSON, &filtersJSON,
		&rpt.Format, &rpt.LastRunAt, &rpt.CreatedBy,
		&rpt.CreatedAt, &rpt.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(emailsJSON, &rpt.RecipientEmails); err != nil {
		return nil, fmt.Errorf("unmarshalling recipient_emails: %w", err)
	}
	if err := json.Unmarshal(filtersJSON, &rpt.Filters); err != nil {
		return nil, fmt.Errorf("unmarshalling filters: %w", err)
	}
	return rpt, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Inactive IP reclamation (#224)
// ─────────────────────────────────────────────────────────────────────────────

// GetInactiveIPs returns assigned IPs that have not been seen within the given number of days.
// It excludes reserved IPs and gateway addresses.
func (r *Repository) GetInactiveIPs(ctx context.Context, days int, sectionID *int64) ([]*models.InactiveIPReport, error) {
	query := `
		SELECT
			ip.id,
			ip.address::text,
			ip.hostname,
			host(s.network_address) || '/' || s.prefix_length AS subnet_cidr,
			sec.name AS section_name,
			ip.assigned_to,
			ip.last_seen,
			CASE
				WHEN ip.last_seen IS NULL THEN $1
				ELSE GREATEST(0, EXTRACT(DAY FROM now() - ip.last_seen)::int)
			END AS days_inactive
		FROM ip_addresses ip
		JOIN subnets s ON s.id = ip.subnet_id
		JOIN sections sec ON sec.id = s.section_id
		WHERE ip.status = 'assigned'
		  AND (ip.device_id IS NOT NULL OR ip.assigned_to IS NOT NULL)
		  AND (ip.last_seen IS NULL OR ip.last_seen < now() - ($1 || ' days')::interval)
		  AND ip.address::text != s.gateway
		  AND ($2::bigint IS NULL OR sec.id = $2)
		ORDER BY ip.last_seen ASC NULLS FIRST
	`
	rows, err := r.db.Query(ctx, query, days, sectionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*models.InactiveIPReport
	for rows.Next() {
		rec := &models.InactiveIPReport{}
		if err := rows.Scan(
			&rec.IPID, &rec.IPAddress, &rec.Hostname,
			&rec.SubnetCIDR, &rec.SectionName,
			&rec.AssignedTo, &rec.LastSeen, &rec.DaysInactive,
		); err != nil {
			return nil, err
		}
		results = append(results, rec)
	}
	return results, rows.Err()
}

// BulkReleaseIPs sets multiple IPs to 'available' and clears assignment fields.
// Returns the count of rows actually updated.
func (r *Repository) BulkReleaseIPs(ctx context.Context, ipIDs []int64) (int64, error) {
	tag, err := r.db.Exec(ctx,
		`UPDATE ip_addresses
		 SET status = 'available', assigned_to = NULL, device_id = NULL, updated_at = now()
		 WHERE id = ANY($1) AND status = 'assigned'`,
		ipIDs,
	)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

// GetSubnetsByUtilisationThreshold returns subnets whose latest utilisation exceeds the given percentage.
func (r *Repository) GetSubnetsByUtilisationThreshold(ctx context.Context, thresholdPct float64) ([]*models.SubnetUtilisationTrend, error) {
	rows, err := r.db.Query(ctx, `
		WITH latest AS (
			SELECT DISTINCT ON (subnet_id)
				subnet_id, utilisation_pct
			FROM subnet_utilisation_history
			ORDER BY subnet_id, recorded_at DESC
		)
		SELECT s.id,
		       host(s.network_address) || '/' || s.prefix_length,
		       s.description,
		       l.utilisation_pct,
		       0::numeric,
		       0::numeric
		FROM latest l
		JOIN subnets s ON s.id = l.subnet_id
		WHERE l.utilisation_pct > $1
		ORDER BY l.utilisation_pct DESC
	`, thresholdPct)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*models.SubnetUtilisationTrend
	for rows.Next() {
		t := &models.SubnetUtilisationTrend{}
		if err := rows.Scan(&t.SubnetID, &t.CIDR, &t.Description, &t.CurrentPct, &t.WeekAgoPct, &t.DeltaPct); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// ─────────────────────────────────────────────────────────────────────────────
// Duplicate detection (#425)
// ─────────────────────────────────────────────────────────────────────────────

// GetDuplicates returns a report of duplicate device hostnames and conflicting
// IP assignments (same address used by more than one record).
func (r *Repository) GetDuplicates(ctx context.Context) (*models.DuplicatesReport, error) {
	report := &models.DuplicatesReport{
		DuplicateHostnames: []models.DuplicateHostname{},
		ConflictingIPs:     []models.ConflictingIP{},
	}

	// Duplicate hostnames — collect one row per hostname then fetch ids separately.
	dupRows, err := r.db.Query(ctx, `
		SELECT hostname, COUNT(*) AS cnt
		FROM devices
		WHERE hostname IS NOT NULL AND hostname <> ''
		GROUP BY hostname
		HAVING COUNT(*) > 1
		ORDER BY cnt DESC
		LIMIT 100`)
	if err != nil {
		return nil, err
	}
	type dupMeta struct {
		hostname string
		count    int
	}
	var dups []dupMeta
	for dupRows.Next() {
		var dm dupMeta
		if err := dupRows.Scan(&dm.hostname, &dm.count); err != nil {
			dupRows.Close()
			return nil, err
		}
		dups = append(dups, dm)
	}
	dupRows.Close()
	if err := dupRows.Err(); err != nil {
		return nil, err
	}

	for _, dm := range dups {
		idRows, err := r.db.Query(ctx,
			`SELECT id FROM devices WHERE hostname = $1 ORDER BY id`, dm.hostname)
		if err != nil {
			return nil, err
		}
		var ids []int64
		for idRows.Next() {
			var id int64
			if err := idRows.Scan(&id); err != nil {
				idRows.Close()
				return nil, err
			}
			ids = append(ids, id)
		}
		idRows.Close()
		if err := idRows.Err(); err != nil {
			return nil, err
		}
		report.DuplicateHostnames = append(report.DuplicateHostnames, models.DuplicateHostname{
			Hostname:  dm.hostname,
			Count:     dm.count,
			DeviceIDs: ids,
		})
	}

	// Conflicting IPs — same address assigned to more than one ip_addresses row.
	cipRows, err := r.db.Query(ctx, `
		SELECT
			ip.address::text,
			host(s.network_address) || '/' || s.prefix_length AS cidr,
			COUNT(DISTINCT ip.id) AS cnt
		FROM ip_addresses ip
		JOIN subnets s ON s.id = ip.subnet_id
		WHERE ip.status = 'used'
		GROUP BY ip.address, s.network_address, s.prefix_length
		HAVING COUNT(DISTINCT ip.id) > 1
		ORDER BY cnt DESC
		LIMIT 100`)
	if err != nil {
		return nil, err
	}
	type cipMeta struct {
		address string
		cidr    string
		count   int
	}
	var cips []cipMeta
	for cipRows.Next() {
		var cm cipMeta
		if err := cipRows.Scan(&cm.address, &cm.cidr, &cm.count); err != nil {
			cipRows.Close()
			return nil, err
		}
		cips = append(cips, cm)
	}
	cipRows.Close()
	if err := cipRows.Err(); err != nil {
		return nil, err
	}

	for _, cm := range cips {
		hnRows, err := r.db.Query(ctx,
			`SELECT COALESCE(hostname, '') FROM ip_addresses WHERE address::text = $1 AND status = 'used' ORDER BY id`,
			cm.address)
		if err != nil {
			return nil, err
		}
		var hostnames []string
		for hnRows.Next() {
			var hn string
			if err := hnRows.Scan(&hn); err != nil {
				hnRows.Close()
				return nil, err
			}
			hostnames = append(hostnames, hn)
		}
		hnRows.Close()
		if err := hnRows.Err(); err != nil {
			return nil, err
		}
		report.ConflictingIPs = append(report.ConflictingIPs, models.ConflictingIP{
			IPAddress:  cm.address,
			SubnetCIDR: cm.cidr,
			Count:      cm.count,
			Hostnames:  hostnames,
		})
	}

	return report, nil
}

// SubnetGapRow holds subnet free-space data for the subnet_gaps report.
type SubnetGapRow struct {
	SubnetID    int64
	CIDR        string
	Description string
	TotalIPs    int
	UsedIPs     int
	FreeIPs     int
	UsedPct     float64
}

// GetSubnetGaps returns all subnets with their allocated vs free IP counts.
func (r *Repository) GetSubnetGaps(ctx context.Context) ([]*SubnetGapRow, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			s.id,
			host(s.network_address) || '/' || s.prefix_length AS cidr,
			s.description,
			(1 << (32 - s.prefix_length)) AS total_ips,
			COUNT(ip.id) FILTER (WHERE ip.status <> 'available') AS used_ips,
			(1 << (32 - s.prefix_length)) - COUNT(ip.id) FILTER (WHERE ip.status <> 'available') AS free_ips
		FROM subnets s
		LEFT JOIN ip_addresses ip ON ip.subnet_id = s.id
		GROUP BY s.id
		ORDER BY free_ips DESC, s.network_address
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*SubnetGapRow
	for rows.Next() {
		row := &SubnetGapRow{}
		if err := rows.Scan(&row.SubnetID, &row.CIDR, &row.Description, &row.TotalIPs, &row.UsedIPs, &row.FreeIPs); err != nil {
			return nil, err
		}
		if row.TotalIPs > 0 {
			row.UsedPct = float64(row.UsedIPs) / float64(row.TotalIPs) * 100
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

// VLANAssignmentRow holds VLAN-to-subnet assignment data for the vlan_assignment report.
type VLANAssignmentRow struct {
	VLANID      int64
	VLANName    string
	VLANTag     int
	SubnetCount int
	SubnetCIDRs string
}

// GetVLANAssignment returns all VLANs with their assigned subnets.
func (r *Repository) GetVLANAssignment(ctx context.Context) ([]*VLANAssignmentRow, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			v.id,
			v.name,
			v.vlan_id AS vlan_tag,
			COUNT(s.id) AS subnet_count,
			COALESCE(string_agg(host(s.network_address) || '/' || s.prefix_length, ', ' ORDER BY s.network_address), '') AS subnet_cidrs
		FROM vlans v
		LEFT JOIN subnets s ON s.vlan_id = v.id
		GROUP BY v.id, v.name, v.vlan_id
		ORDER BY v.vlan_id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*VLANAssignmentRow
	for rows.Next() {
		row := &VLANAssignmentRow{}
		if err := rows.Scan(&row.VLANID, &row.VLANName, &row.VLANTag, &row.SubnetCount, &row.SubnetCIDRs); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

// IPAgeRow holds IP address age data for the ip_age report.
type IPAgeRow struct {
	IPID          int64
	Address       string
	Status        string
	AssignedTo    string
	DaysOld       int
	DaysSinceSeen int
}

// GetIPAge returns all IP addresses with their age and days since last seen.
func (r *Repository) GetIPAge(ctx context.Context) ([]*IPAgeRow, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			id,
			address,
			status,
			COALESCE(assigned_to, '') AS assigned_to,
			EXTRACT(DAY FROM now() - created_at)::int AS days_old,
			CASE WHEN last_seen IS NOT NULL THEN EXTRACT(DAY FROM now() - last_seen)::int ELSE -1 END AS days_since_seen
		FROM ip_addresses
		ORDER BY days_old DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*IPAgeRow
	for rows.Next() {
		row := &IPAgeRow{}
		if err := rows.Scan(&row.IPID, &row.Address, &row.Status, &row.AssignedTo, &row.DaysOld, &row.DaysSinceSeen); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

// DNSAuditRow holds DNS audit data for the dns_audit report.
type DNSAuditRow struct {
	IPID           int64  `json:"ip_id"`
	Address        string `json:"address"`
	DNSName        string `json:"dns_name"`
	PTRRecord      string `json:"ptr_record"`
	DNSLastChecked string `json:"dns_last_checked"`
}

// GetDNSAudit returns all IP addresses that have a dns_name set, with their DNS check status.
func (r *Repository) GetDNSAudit(ctx context.Context) ([]*DNSAuditRow, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			id,
			address,
			COALESCE(dns_name, '') AS dns_name,
			COALESCE(ptr_record, '') AS ptr_record,
			COALESCE(dns_last_checked::text, 'never') AS dns_last_checked
		FROM ip_addresses
		WHERE dns_name IS NOT NULL AND dns_name <> ''
		ORDER BY address
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*DNSAuditRow
	for rows.Next() {
		row := &DNSAuditRow{}
		if err := rows.Scan(&row.IPID, &row.Address, &row.DNSName, &row.PTRRecord, &row.DNSLastChecked); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

// GetInactiveDevices returns devices that have not been pinged within the given number of days.
func (r *Repository) GetInactiveDevices(ctx context.Context, days int) ([]*models.InactiveDeviceReport, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			id,
			hostname,
			COALESCE(vendor, '') AS vendor,
			COALESCE(model, '') AS model,
			last_ping_at,
			CASE
				WHEN last_ping_at IS NULL THEN $1
				ELSE GREATEST(0, EXTRACT(DAY FROM now() - last_ping_at)::int)
			END AS days_inactive
		FROM devices
		WHERE last_ping_at IS NULL OR last_ping_at < now() - ($1 || ' days')::interval
		ORDER BY days_inactive DESC
		LIMIT 500
	`, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*models.InactiveDeviceReport
	for rows.Next() {
		d := &models.InactiveDeviceReport{}
		if err := rows.Scan(&d.DeviceID, &d.Hostname, &d.Vendor, &d.Model, &d.LastPingAt, &d.DaysInactive); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

// GetOverdueScanJobs returns active scan jobs that have not run within the expected window.
// It returns jobs that have not run in more than the given number of days (or never).
func (r *Repository) GetOverdueScanJobs(ctx context.Context, days int) ([]*models.FailedScanJobReport, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			id,
			name,
			COALESCE(schedule_cron, '') AS schedule_cron,
			last_run_at,
			CASE
				WHEN last_run_at IS NULL THEN $1
				ELSE GREATEST(0, EXTRACT(DAY FROM now() - last_run_at)::int)
			END AS days_since_run,
			is_active
		FROM scan_jobs
		WHERE is_active = true
		  AND (last_run_at IS NULL OR last_run_at < now() - ($1 || ' days')::interval)
		ORDER BY days_since_run DESC
		LIMIT 500
	`, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*models.FailedScanJobReport
	for rows.Next() {
		j := &models.FailedScanJobReport{}
		if err := rows.Scan(&j.JobID, &j.JobName, &j.ScheduleCron, &j.LastRunAt, &j.DaysSinceRun, &j.IsActive); err != nil {
			return nil, err
		}
		out = append(out, j)
	}
	return out, rows.Err()
}
