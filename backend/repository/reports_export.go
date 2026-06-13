package repository

import (
	"context"

	"padduck/models"
)

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
	DeviceID      *int64
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
			device_id,
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
		if err := rows.Scan(&row.IPID, &row.Address, &row.Status, &row.DeviceID, &row.DaysOld, &row.DaysSinceSeen); err != nil {
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
		WHERE last_ping_at IS NULL OR last_ping_at < now() - ($1 * INTERVAL '1 day')
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
		  AND (last_run_at IS NULL OR last_run_at < now() - ($1 * INTERVAL '1 day'))
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
