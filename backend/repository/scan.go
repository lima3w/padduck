package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"ipam-next/models"
)

// scanJobCols is the column list for scan_jobs SELECT queries.
const scanJobCols = `id, name, subnet_ids, schedule_cron, is_active, last_run_at, next_run_at, created_by, created_at, updated_at, ping_concurrency, notify_on_change, scan_type, agent_id`

func scanScanJob(row interface{ Scan(dest ...any) error }) (*models.ScanJob, error) {
	j := &models.ScanJob{}
	return j, row.Scan(&j.ID, &j.Name, &j.SubnetIDs, &j.ScheduleCron, &j.IsActive, &j.LastRunAt, &j.NextRunAt, &j.CreatedBy, &j.CreatedAt, &j.UpdatedAt, &j.PingConcurrency, &j.NotifyOnChange, &j.ScanType, &j.AgentID)
}

// CreateScanJob creates a new discovery scan job
func (r *Repository) CreateScanJob(ctx context.Context, name string, subnetIDs []int64, scheduleCron *string, createdBy int64) (*models.ScanJob, error) {
	query := `INSERT INTO scan_jobs (name, subnet_ids, schedule_cron, created_by)
		VALUES ($1, $2, $3, $4)
		RETURNING ` + scanJobCols
	return scanScanJob(r.db.QueryRow(ctx, query, name, subnetIDs, scheduleCron, createdBy))
}

// GetScanJobByID retrieves a scan job by ID
func (r *Repository) GetScanJobByID(ctx context.Context, id int64) (*models.ScanJob, error) {
	return scanScanJob(r.db.QueryRow(ctx, `SELECT `+scanJobCols+` FROM scan_jobs WHERE id = $1`, id))
}

// ListScanJobs returns all scan jobs
func (r *Repository) ListScanJobs(ctx context.Context) ([]*models.ScanJob, error) {
	rows, err := r.db.Query(ctx, `SELECT `+scanJobCols+` FROM scan_jobs ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	jobs := make([]*models.ScanJob, 0)
	for rows.Next() {
		j, err := scanScanJob(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}
	return jobs, rows.Err()
}

// ListActiveScanJobs returns all active scan jobs with a schedule
func (r *Repository) ListActiveScanJobs(ctx context.Context) ([]*models.ScanJob, error) {
	rows, err := r.db.Query(ctx, `SELECT `+scanJobCols+` FROM scan_jobs WHERE is_active = TRUE AND schedule_cron IS NOT NULL`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	jobs := make([]*models.ScanJob, 0)
	for rows.Next() {
		j, err := scanScanJob(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}
	return jobs, rows.Err()
}

// UpdateScanJob updates a scan job's configuration
func (r *Repository) UpdateScanJob(ctx context.Context, id int64, name string, subnetIDs []int64, scheduleCron *string, isActive bool) (*models.ScanJob, error) {
	query := `UPDATE scan_jobs SET name = $2, subnet_ids = $3, schedule_cron = $4, is_active = $5, updated_at = CURRENT_TIMESTAMP WHERE id = $1
		RETURNING ` + scanJobCols
	return scanScanJob(r.db.QueryRow(ctx, query, id, name, subnetIDs, scheduleCron, isActive))
}

// UpdateScanJobFull updates all mutable fields of a scan job.
func (r *Repository) UpdateScanJobFull(ctx context.Context, id int64, name string, subnetIDs []int64, scheduleCron *string, isActive bool, pingConcurrency int, notifyOnChange bool, scanType string, agentID *int64) (*models.ScanJob, error) {
	query := `UPDATE scan_jobs
		SET name=$2, subnet_ids=$3, schedule_cron=$4, is_active=$5, ping_concurrency=$6,
		    notify_on_change=$7, scan_type=$8, agent_id=$9, updated_at=CURRENT_TIMESTAMP
		WHERE id=$1 RETURNING ` + scanJobCols
	return scanScanJob(r.db.QueryRow(ctx, query, id, name, subnetIDs, scheduleCron, isActive, pingConcurrency, notifyOnChange, scanType, agentID))
}

// UpdateScanJobRunTime updates last_run_at and next_run_at after a scan
func (r *Repository) UpdateScanJobRunTime(ctx context.Context, id int64, nextRunAt *time.Time) error {
	query := `UPDATE scan_jobs SET last_run_at = CURRENT_TIMESTAMP, next_run_at = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id, nextRunAt)
	return err
}

// DeleteScanJob deletes a scan job
func (r *Repository) DeleteScanJob(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM scan_jobs WHERE id = $1`, id)
	return err
}

// CreateScanResult records the result of scanning a single IP
func (r *Repository) CreateScanResult(ctx context.Context, jobID, subnetID int64, ipAddressID *int64, ipAddress string, isAlive bool, responseTimeMs *int64, ptrRecord *string, fwdRevMismatch bool) (*models.ScanResult, error) {
	query := `INSERT INTO scan_results (job_id, subnet_id, ip_address_id, ip_address, is_alive, response_time_ms, ptr_record, fwd_rev_mismatch)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, job_id, subnet_id, ip_address_id, ip_address, is_alive, response_time_ms, ptr_record, fwd_rev_mismatch, scanned_at`
	row := r.db.QueryRow(ctx, query, jobID, subnetID, ipAddressID, ipAddress, isAlive, responseTimeMs, ptrRecord, fwdRevMismatch)

	sr := &models.ScanResult{}
	err := row.Scan(&sr.ID, &sr.JobID, &sr.SubnetID, &sr.IPAddressID, &sr.IPAddress, &sr.IsAlive, &sr.ResponseTimeMs, &sr.PTRRecord, &sr.FwdRevMismatch, &sr.ScannedAt)
	if err != nil {
		return nil, err
	}
	return sr, nil
}

// ListScanResultsByJob returns recent scan results for a job
func (r *Repository) ListScanResultsByJob(ctx context.Context, jobID int64, limit int) ([]*models.ScanResult, error) {
	query := `SELECT id, job_id, subnet_id, ip_address_id, ip_address, is_alive, response_time_ms, ptr_record, fwd_rev_mismatch, scanned_at FROM scan_results WHERE job_id = $1 ORDER BY scanned_at DESC LIMIT $2`
	rows, err := r.db.Query(ctx, query, jobID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]*models.ScanResult, 0)
	for rows.Next() {
		sr := &models.ScanResult{}
		if err := rows.Scan(&sr.ID, &sr.JobID, &sr.SubnetID, &sr.IPAddressID, &sr.IPAddress, &sr.IsAlive, &sr.ResponseTimeMs, &sr.PTRRecord, &sr.FwdRevMismatch, &sr.ScannedAt); err != nil {
			return nil, err
		}
		results = append(results, sr)
	}
	return results, rows.Err()
}

// SetIPAddressPTRFromScan updates ptr_record on an IP address row from scan data.
// It also sets dns_name if dns_name is currently empty, without overwriting existing values.
func (r *Repository) SetIPAddressPTRFromScan(ctx context.Context, ipID int64, ptrRecord string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE ip_addresses
		SET ptr_record = $2,
		    dns_name   = CASE WHEN (dns_name IS NULL OR dns_name = '') THEN $2 ELSE dns_name END,
		    updated_at = now()
		WHERE id = $1`, ipID, ptrRecord)
	return err
}

// ListScanResultsBySubnet returns recent scan results for a subnet
func (r *Repository) ListScanResultsBySubnet(ctx context.Context, subnetID int64, limit int) ([]*models.ScanResult, error) {
	query := `SELECT id, job_id, subnet_id, ip_address_id, ip_address, is_alive, response_time_ms, ptr_record, fwd_rev_mismatch, scanned_at FROM scan_results WHERE subnet_id = $1 ORDER BY scanned_at DESC LIMIT $2`
	rows, err := r.db.Query(ctx, query, subnetID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]*models.ScanResult, 0)
	for rows.Next() {
		sr := &models.ScanResult{}
		if err := rows.Scan(&sr.ID, &sr.JobID, &sr.SubnetID, &sr.IPAddressID, &sr.IPAddress, &sr.IsAlive, &sr.ResponseTimeMs, &sr.PTRRecord, &sr.FwdRevMismatch, &sr.ScannedAt); err != nil {
			return nil, err
		}
		results = append(results, sr)
	}
	return results, rows.Err()
}

// CreateScanRun inserts a new scan_run and returns it.
func (r *Repository) CreateScanRun(ctx context.Context, scanJobID int64) (*models.ScanRun, error) {
	row := r.db.QueryRow(ctx,
		`INSERT INTO scan_runs (scan_job_id) VALUES ($1) RETURNING id, scan_job_id, started_at, finished_at, new_count, gone_count, changed_count`,
		scanJobID,
	)
	sr := &models.ScanRun{}
	return sr, row.Scan(&sr.ID, &sr.ScanJobID, &sr.StartedAt, &sr.FinishedAt, &sr.NewCount, &sr.GoneCount, &sr.ChangedCount)
}

// FinishScanRun marks a scan_run as finished with final counts.
func (r *Repository) FinishScanRun(ctx context.Context, runID int64, newCount, goneCount, changedCount int) error {
	_, err := r.db.Exec(ctx,
		`UPDATE scan_runs SET finished_at=now(), new_count=$2, gone_count=$3, changed_count=$4 WHERE id=$1`,
		runID, newCount, goneCount, changedCount,
	)
	return err
}

// CreateScanRunChange records a single IP change detected during a scan run.
func (r *Repository) CreateScanRunChange(ctx context.Context, runID int64, ipAddress, changeType string) (*models.ScanRunChange, error) {
	row := r.db.QueryRow(ctx,
		`INSERT INTO scan_run_changes (run_id, ip_address, change_type) VALUES ($1, $2, $3) RETURNING id, run_id, ip_address, change_type, scanned_at`,
		runID, ipAddress, changeType,
	)
	ch := &models.ScanRunChange{}
	return ch, row.Scan(&ch.ID, &ch.RunID, &ch.IPAddress, &ch.ChangeType, &ch.ScannedAt)
}

// ListScanRuns returns the last `limit` scan runs for a given job.
func (r *Repository) ListScanRuns(ctx context.Context, jobID int64, limit int) ([]*models.ScanRun, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, scan_job_id, started_at, finished_at, new_count, gone_count, changed_count FROM scan_runs WHERE scan_job_id=$1 ORDER BY started_at DESC LIMIT $2`,
		jobID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]*models.ScanRun, 0)
	for rows.Next() {
		sr := &models.ScanRun{}
		if err := rows.Scan(&sr.ID, &sr.ScanJobID, &sr.StartedAt, &sr.FinishedAt, &sr.NewCount, &sr.GoneCount, &sr.ChangedCount); err != nil {
			return nil, err
		}
		result = append(result, sr)
	}
	return result, rows.Err()
}

// GetScanRun retrieves a single scan run by ID.
func (r *Repository) GetScanRun(ctx context.Context, runID int64) (*models.ScanRun, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, scan_job_id, started_at, finished_at, new_count, gone_count, changed_count FROM scan_runs WHERE id=$1`,
		runID,
	)
	sr := &models.ScanRun{}
	err := row.Scan(&sr.ID, &sr.ScanJobID, &sr.StartedAt, &sr.FinishedAt, &sr.NewCount, &sr.GoneCount, &sr.ChangedCount)
	if err != nil {
		return nil, err
	}
	return sr, nil
}

// ListScanRunChanges returns all changes recorded for a scan run.
func (r *Repository) ListScanRunChanges(ctx context.Context, runID int64) ([]*models.ScanRunChange, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, run_id, ip_address, change_type, scanned_at FROM scan_run_changes WHERE run_id=$1 ORDER BY scanned_at`,
		runID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]*models.ScanRunChange, 0)
	for rows.Next() {
		ch := &models.ScanRunChange{}
		if err := rows.Scan(&ch.ID, &ch.RunID, &ch.IPAddress, &ch.ChangeType, &ch.ScannedAt); err != nil {
			return nil, err
		}
		result = append(result, ch)
	}
	return result, rows.Err()
}

// GetLastAliveIPsForJob returns the set of IPs that were alive in the most recent
// scan run for this job, used to compute changes.
func (r *Repository) GetLastAliveIPsForJob(ctx context.Context, jobID int64) (map[string]bool, error) {
	// Find the most recent finished run for this job
	row := r.db.QueryRow(ctx,
		`SELECT id FROM scan_runs WHERE scan_job_id=$1 AND finished_at IS NOT NULL ORDER BY finished_at DESC LIMIT 1`,
		jobID,
	)
	var prevRunID int64
	if err := row.Scan(&prevRunID); err != nil {
		// No previous run — return empty map
		return map[string]bool{}, nil
	}

	rows, err := r.db.Query(ctx,
		`SELECT ip_address, is_alive FROM scan_results WHERE job_id=$1 AND scanned_at >= (SELECT started_at FROM scan_runs WHERE id=$2) ORDER BY ip_address`,
		jobID, prevRunID,
	)
	if err != nil {
		return map[string]bool{}, nil
	}
	defer rows.Close()
	m := map[string]bool{}
	for rows.Next() {
		var ip string
		var alive bool
		if err := rows.Scan(&ip, &alive); err != nil {
			continue
		}
		m[ip] = alive
	}
	return m, rows.Err()
}

// UpdateIPFromSNMP stores the MAC address and hostname discovered via SNMP.
// Hostname is only updated when the existing value is empty.
func (r *Repository) UpdateIPFromSNMP(ctx context.Context, ipID int64, macAddress, hostname string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE ip_addresses
		SET mac_address = CASE WHEN $2 != '' THEN $2 ELSE mac_address END,
		    hostname    = CASE WHEN (hostname IS NULL OR hostname = '') AND $3 != '' THEN $3 ELSE hostname END,
		    updated_at  = now()
		WHERE id = $1`,
		ipID, macAddress, hostname,
	)
	return err
}

// scanAgentCols is the column list for scan_agents SELECT queries.
const scanAgentCols = `id, name, token_hash, last_seen, is_active, created_at, version, capabilities, status, last_error`

// CreateScanAgent inserts a new scan agent record.
func (r *Repository) CreateScanAgent(ctx context.Context, name, tokenHash string) (*models.ScanAgent, error) {
	row := r.db.QueryRow(ctx,
		`INSERT INTO scan_agents (name, token_hash) VALUES ($1, $2) RETURNING `+scanAgentCols,
		name, tokenHash,
	)
	return scanScanAgent(row)
}

// GetScanAgentByToken retrieves an active scan agent by token hash.
func (r *Repository) GetScanAgentByToken(ctx context.Context, tokenHash string) (*models.ScanAgent, error) {
	row := r.db.QueryRow(ctx,
		`SELECT `+scanAgentCols+` FROM scan_agents WHERE token_hash=$1 AND is_active=true`,
		tokenHash,
	)
	return scanScanAgent(row)
}

// GetScanAgentByID retrieves a scan agent by its primary key.
func (r *Repository) GetScanAgentByID(ctx context.Context, id int64) (*models.ScanAgent, error) {
	row := r.db.QueryRow(ctx,
		`SELECT `+scanAgentCols+` FROM scan_agents WHERE id=$1`,
		id,
	)
	return scanScanAgent(row)
}

// ListScanAgents returns all scan agents ordered by name.
func (r *Repository) ListScanAgents(ctx context.Context) ([]*models.ScanAgent, error) {
	rows, err := r.db.Query(ctx,
		`SELECT `+scanAgentCols+` FROM scan_agents ORDER BY name`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]*models.ScanAgent, 0)
	for rows.Next() {
		a, err := scanScanAgent(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, a)
	}
	return result, rows.Err()
}

// HeartbeatAgent updates last_seen and optional health fields for an agent.
func (r *Repository) HeartbeatAgent(ctx context.Context, id int64, version *string, capabilities []string, status string, lastError *string) error {
	var capJSON *string
	if capabilities != nil {
		b, err := json.Marshal(capabilities)
		if err != nil {
			return fmt.Errorf("marshal capabilities: %w", err)
		}
		s := string(b)
		capJSON = &s
	}
	_, err := r.db.Exec(ctx, `
		UPDATE scan_agents
		SET last_seen = now(),
		    version = COALESCE($2, version),
		    capabilities = CASE WHEN $3::jsonb IS NOT NULL THEN $3::jsonb ELSE capabilities END,
		    status = COALESCE($4, status),
		    last_error = $5
		WHERE id = $1`,
		id, version, capJSON, status, lastError,
	)
	return err
}

// UpdateScanAgentActive sets is_active on a scan agent.
func (r *Repository) UpdateScanAgentActive(ctx context.Context, id int64, isActive bool) (*models.ScanAgent, error) {
	row := r.db.QueryRow(ctx,
		`UPDATE scan_agents SET is_active=$2 WHERE id=$1 RETURNING `+scanAgentCols,
		id, isActive,
	)
	return scanScanAgent(row)
}

// UpdateScanAgentToken replaces the token hash for an agent.
func (r *Repository) UpdateScanAgentToken(ctx context.Context, id int64, newTokenHash string) (*models.ScanAgent, error) {
	row := r.db.QueryRow(ctx,
		`UPDATE scan_agents SET token_hash=$2 WHERE id=$1 RETURNING `+scanAgentCols,
		id, newTokenHash,
	)
	return scanScanAgent(row)
}

// DeleteScanAgent removes a scan agent by ID.
func (r *Repository) DeleteScanAgent(ctx context.Context, id int64) error {
	ct, err := r.db.Exec(ctx, `DELETE FROM scan_agents WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("scan agent not found")
	}
	return nil
}

// ListScanJobsForAgent returns active scan jobs assigned to a specific agent.
func (r *Repository) ListScanJobsForAgent(ctx context.Context, agentID int64) ([]*models.ScanJob, error) {
	rows, err := r.db.Query(ctx,
		`SELECT `+scanJobCols+` FROM scan_jobs WHERE agent_id=$1 AND is_active=true ORDER BY id`,
		agentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	jobs := make([]*models.ScanJob, 0)
	for rows.Next() {
		j, err := scanScanJob(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}
	return jobs, rows.Err()
}

func scanScanAgent(row interface{ Scan(dest ...any) error }) (*models.ScanAgent, error) {
	a := &models.ScanAgent{}
	var capJSON []byte
	err := row.Scan(&a.ID, &a.Name, &a.TokenHash, &a.LastSeen, &a.IsActive, &a.CreatedAt, &a.Version, &capJSON, &a.Status, &a.LastError)
	if err != nil {
		return nil, err
	}
	if len(capJSON) > 0 {
		_ = json.Unmarshal(capJSON, &a.Capabilities)
	}
	return a, nil
}

// ---------------------------------------------------------------------------
// Scan profiles (#432)
// ---------------------------------------------------------------------------

const scanProfileCols = `id, name, description, scan_type, ping_concurrency, tcp_ports, dns_lookup, snmp_community, snmp_version, created_at, updated_at`

func scanScanProfile(row interface{ Scan(dest ...any) error }) (*models.ScanProfile, error) {
	p := &models.ScanProfile{}
	return p, row.Scan(&p.ID, &p.Name, &p.Description, &p.ScanType, &p.PingConcurrency, &p.TCPPorts, &p.DNSLookup, &p.SNMPCommunity, &p.SNMPVersion, &p.CreatedAt, &p.UpdatedAt)
}

// CreateScanProfile inserts a new scan profile.
func (r *Repository) CreateScanProfile(ctx context.Context, name, scanType string, desc *string, pingConcurrency int, tcpPorts *string, dnsLookup bool, snmpCommunity *string, snmpVersion string) (*models.ScanProfile, error) {
	query := `INSERT INTO scan_profiles (name, description, scan_type, ping_concurrency, tcp_ports, dns_lookup, snmp_community, snmp_version)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING ` + scanProfileCols
	return scanScanProfile(r.db.QueryRow(ctx, query, name, desc, scanType, pingConcurrency, tcpPorts, dnsLookup, snmpCommunity, snmpVersion))
}

// ListScanProfiles returns all scan profiles ordered by name.
func (r *Repository) ListScanProfiles(ctx context.Context) ([]*models.ScanProfile, error) {
	rows, err := r.db.Query(ctx, `SELECT `+scanProfileCols+` FROM scan_profiles ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]*models.ScanProfile, 0)
	for rows.Next() {
		p, err := scanScanProfile(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, rows.Err()
}

// GetScanProfileByID retrieves a scan profile by primary key.
func (r *Repository) GetScanProfileByID(ctx context.Context, id int64) (*models.ScanProfile, error) {
	return scanScanProfile(r.db.QueryRow(ctx, `SELECT `+scanProfileCols+` FROM scan_profiles WHERE id=$1`, id))
}

// UpdateScanProfile updates a scan profile.
func (r *Repository) UpdateScanProfile(ctx context.Context, id int64, name, scanType string, desc *string, pingConcurrency int, tcpPorts *string, dnsLookup bool, snmpCommunity *string, snmpVersion string) (*models.ScanProfile, error) {
	query := `UPDATE scan_profiles
		SET name=$2, description=$3, scan_type=$4, ping_concurrency=$5, tcp_ports=$6, dns_lookup=$7, snmp_community=$8, snmp_version=$9, updated_at=now()
		WHERE id=$1 RETURNING ` + scanProfileCols
	return scanScanProfile(r.db.QueryRow(ctx, query, id, name, desc, scanType, pingConcurrency, tcpPorts, dnsLookup, snmpCommunity, snmpVersion))
}

// DeleteScanProfile removes a scan profile by ID.
func (r *Repository) DeleteScanProfile(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM scan_profiles WHERE id=$1`, id)
	return err
}

// SetSubnetScanProfile sets or clears the scan_profile_id for a subnet.
func (r *Repository) SetSubnetScanProfile(ctx context.Context, subnetID int64, profileID *int64) error {
	_, err := r.db.Exec(ctx, `UPDATE subnets SET scan_profile_id=$2 WHERE id=$1`, subnetID, profileID)
	return err
}
