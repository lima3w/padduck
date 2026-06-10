package services

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"padduck/internal/scanner"
	"padduck/models"
)

// DiscoveryService handles network scanning and IP detection
type DiscoveryService struct {
	repository    discoveryRepo
	config        *ConfigService
	encryptionKey string
	// in-flight tracks running job IDs (value is struct{})
	inFlight sync.Map
	// semaphore channel limits concurrent RunJob executions
	semaphore chan struct{}
}

type discoveryRepo interface {
	GetSubnetByID(ctx context.Context, id int64) (*models.Subnet, error)
	ListIPAddressesBySubnet(ctx context.Context, subnetID int64) ([]*models.IPAddress, error)
	CreateScanJob(ctx context.Context, name string, subnetIDs []int64, scheduleCron *string, createdBy int64, autoAddIPs bool) (*models.ScanJob, error)
	GetScanJobByID(ctx context.Context, id int64) (*models.ScanJob, error)
	ListScanJobs(ctx context.Context) ([]*models.ScanJob, error)
	ListActiveScanJobs(ctx context.Context) ([]*models.ScanJob, error)
	UpdateScanJob(ctx context.Context, id int64, name string, subnetIDs []int64, scheduleCron *string, isActive bool) (*models.ScanJob, error)
	UpdateScanJobFull(ctx context.Context, id int64, name string, subnetIDs []int64, scheduleCron *string, isActive bool, pingConcurrency int, notifyOnChange bool, scanType string, agentID *int64, autoAddIPs bool, discoverDNS bool, dnsOverwrite bool) (*models.ScanJob, error)
	UpdateScanJobRunTime(ctx context.Context, id int64, nextRunAt *time.Time) error
	DeleteScanJob(ctx context.Context, id int64) error
	CreateScanResult(ctx context.Context, jobID, subnetID int64, ipAddressID *int64, ipAddress string, isAlive bool, responseTimeMs *int64, ptrRecord *string, fwdRevMismatch bool) (*models.ScanResult, error)
	ListScanResultsByJob(ctx context.Context, jobID int64, limit int) ([]*models.ScanResult, error)
	ListScanResultsBySubnet(ctx context.Context, subnetID int64, limit int) ([]*models.ScanResult, error)
	SetIPAddressPTRFromScan(ctx context.Context, ipID int64, ptrRecord string, overwrite bool) error
	// Port scan (#214)
	UpdateIPPortScan(ctx context.Context, ipID int64, ports map[string]bool) error
	// Scan runs (#211)
	CreateScanRun(ctx context.Context, scanJobID int64) (*models.ScanRun, error)
	FinishScanRun(ctx context.Context, runID int64, newCount, goneCount, changedCount int) error
	CreateScanRunChange(ctx context.Context, runID int64, ipAddress, changeType string) (*models.ScanRunChange, error)
	ListScanRuns(ctx context.Context, jobID int64, limit int) ([]*models.ScanRun, error)
	GetScanRun(ctx context.Context, runID int64) (*models.ScanRun, error)
	ListScanRunChanges(ctx context.Context, runID int64) ([]*models.ScanRunChange, error)
	GetLastAliveIPsForJob(ctx context.Context, jobID int64) (map[string]bool, error)
	// SNMP (#210)
	UpdateIPFromSNMP(ctx context.Context, ipID int64, macAddress, hostname string) error
	GetDeviceSNMPByIPID(ctx context.Context, ipID int64) (*models.DeviceSNMP, error)
	// Scan agents (#212)
	CreateScanAgent(ctx context.Context, name, tokenHash string, ttlDays int) (*models.ScanAgent, error)
	GetScanAgentByToken(ctx context.Context, tokenHash string) (*models.ScanAgent, error)
	GetScanAgentByID(ctx context.Context, id int64) (*models.ScanAgent, error)
	ListScanAgents(ctx context.Context) ([]*models.ScanAgent, error)
	HeartbeatAgent(ctx context.Context, id int64, version *string, capabilities []string, status string, lastError *string) error
	UpdateScanAgentActive(ctx context.Context, id int64, isActive bool) (*models.ScanAgent, error)
	UpdateScanAgentToken(ctx context.Context, id int64, newTokenHash string, ttlDays int) (*models.ScanAgent, error)
	DeleteScanAgent(ctx context.Context, id int64) error
	ListScanJobsForAgent(ctx context.Context, agentID int64) ([]*models.ScanJob, error)
	// Scan profiles (#432)
	CreateScanProfile(ctx context.Context, name, scanType string, desc *string, pingConcurrency int, tcpPorts *string, dnsLookup bool, snmpCommunity *string, snmpVersion string) (*models.ScanProfile, error)
	ListScanProfiles(ctx context.Context) ([]*models.ScanProfile, error)
	GetScanProfileByID(ctx context.Context, id int64) (*models.ScanProfile, error)
	UpdateScanProfile(ctx context.Context, id int64, name, scanType string, desc *string, pingConcurrency int, tcpPorts *string, dnsLookup bool, snmpCommunity *string, snmpVersion string) (*models.ScanProfile, error)
	DeleteScanProfile(ctx context.Context, id int64) error
	SetSubnetScanProfile(ctx context.Context, subnetID int64, profileID *int64) error
	// Scan retention (#435)
	GetScanRetentionSettings(ctx context.Context) (*models.ScanRetentionSettings, error)
	UpdateScanRetentionSettings(ctx context.Context, rawHistoryDays, rollupAfterDays int, rollupEnabled bool) (*models.ScanRetentionSettings, error)
	PruneScanHistory(ctx context.Context, olderThanDays int) (int64, error)
	// Device fingerprints (#430)
	GetDeviceFingerprint(ctx context.Context, deviceID int64) (*models.DeviceFingerprint, error)
	UpsertDeviceFingerprint(ctx context.Context, deviceID int64, openPorts []int, osGuess, vendorGuess *string, confidenceScore float64, evidence []string) (*models.DeviceFingerprint, error)
	// Discovery conflicts (#431)
	ListDiscoveryConflicts(ctx context.Context, status string) ([]*models.DiscoveryConflict, error)
	GetDiscoveryConflict(ctx context.Context, id int64) (*models.DiscoveryConflict, error)
	CreateDiscoveryConflict(ctx context.Context, deviceID int64, fieldName, discoveredValue string, currentValue *string, confidenceScore float64, source string) (*models.DiscoveryConflict, error)
	ResolveDiscoveryConflict(ctx context.Context, id int64, action string, reviewedBy string) (*models.DiscoveryConflict, error)
	// Auto-add IPs (#item5)
	CreateIPAddress(ctx context.Context, subnetID int64, address, hostname string, status string, assignedTo *string, tagID *int64, macAddress, ptrRecord *string) (*models.IPAddress, error)
}

// maxConcurrentJobsFromEnv reads SCAN_MAX_CONCURRENT_JOBS (default 4, min 1).
func maxConcurrentJobsFromEnv() int {
	if v := os.Getenv("SCAN_MAX_CONCURRENT_JOBS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 1 {
			return n
		}
	}
	return 4
}

func NewDiscoveryService(repo discoveryRepo, configSvc *ConfigService, encryptionKey string) *DiscoveryService {
	maxJobs := maxConcurrentJobsFromEnv()
	return &DiscoveryService{
		repository:    repo,
		config:        configSvc,
		encryptionKey: encryptionKey,
		semaphore:     make(chan struct{}, maxJobs),
	}
}

// PingHost checks if a host responds to ICMP ping, returning response time in ms
func PingHost(host string, timeout time.Duration) (bool, int64) {
	if net.ParseIP(host) == nil {
		return false, 0
	}
	start := time.Now()
	cmd := exec.Command("ping", "-c", "1", "-W", strconv.Itoa(int(timeout.Seconds())), host) // #nosec G204 -- host is validated as an IP address.
	err := cmd.Run()
	elapsed := time.Since(start).Milliseconds()
	if err != nil {
		return false, 0
	}
	return true, elapsed
}

// portScanConcurrency returns the configured per-job port scan concurrency.
func (d *DiscoveryService) portScanConcurrency() int {
	if d.config != nil {
		if v, _ := d.config.Get("scanner_port_scan_concurrency"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				return n
			}
		}
	}
	return 10
}

// parsePorts splits a comma-separated port list string into a slice of strings.
func parsePorts(portList string) []string {
	raw := strings.Split(portList, ",")
	out := make([]string, 0, len(raw))
	for _, p := range raw {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// ScanSubnet scans all IPs in a subnet CIDR range for liveness.
// #248: concurrency comes from job.PingConcurrency.
// #214: performs port scan on alive IPs when enabled.
// #210: performs SNMP scan when scan_type includes snmp.
func (d *DiscoveryService) ScanSubnet(ctx context.Context, jobID, subnetID int64, networkAddr string, prefixLen int, existingIPs map[string]int64, concurrency int, runID int64, prevAlive map[string]bool, job *models.ScanJob) ([]*models.ScanResult, int, int, int, error) {
	if concurrency <= 0 {
		concurrency = 20
	}

	resolveHostnames := true
	portScanEnabled := false
	portList := "22,80,443,3306,5432,8080,8443"

	if d.config != nil {
		if v, _ := d.config.Get("scanner_resolve_hostnames"); v == "false" {
			resolveHostnames = false
		}
		if v, _ := d.config.Get("scanner_port_scan_enabled"); v == "true" {
			portScanEnabled = true
		}
		if v, _ := d.config.Get("scanner_port_list"); v != "" {
			portList = v
		}
	}
	// Per-job DNS setting takes precedence over global config.
	if job != nil {
		resolveHostnames = job.DiscoverDNS
	}
	dnsOverwrite := job != nil && job.DNSOverwrite

	ports := parsePorts(portList)
	portConcurrency := d.portScanConcurrency()

	doSNMP := job != nil && (job.ScanType == "snmp" || job.ScanType == "ping+snmp")
	snmpCommunity := "public"
	snmpVersion := "2c"
	if d.config != nil {
		if v, _ := d.config.Get("scanner_snmp_community"); v != "" {
			snmpCommunity = v
		}
		if v, _ := d.config.Get("scanner_snmp_version"); v != "" {
			snmpVersion = v
		}
	}

	ips, err := enumerateCIDR(networkAddr, prefixLen)
	if err != nil {
		return nil, 0, 0, 0, fmt.Errorf("enumerate CIDR: %w", err)
	}

	type work struct {
		ip string
	}
	type result struct {
		ip             string
		alive          bool
		responseTimeMs int64
		ptr            *string
		fwdRevMismatch bool
	}

	workCh := make(chan work, len(ips))
	resultCh := make(chan result, len(ips))

	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for w := range workCh {
				select {
				case <-ctx.Done():
					return
				default:
				}
				alive, ms := PingHost(w.ip, 2*time.Second)
				r := result{ip: w.ip, alive: alive, responseTimeMs: ms}
				if alive && resolveHostnames {
					dns := scanner.ResolveHostname(w.ip, 2*time.Second)
					if dns.PTR != "" {
						r.ptr = &dns.PTR
						r.fwdRevMismatch = dns.FwdRevMismatch
					}
				}
				resultCh <- r
			}
		}()
	}

	for _, ip := range ips {
		workCh <- work{ip: ip}
	}
	close(workCh)

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	var newCount, goneCount, changedCount int
	scanResults := make([]*models.ScanResult, 0, len(ips))
	for r := range resultCh {
		var ipAddressID *int64
		if id, ok := existingIPs[r.ip]; ok {
			ipAddressID = &id
		}
		var ms *int64
		if r.responseTimeMs > 0 {
			ms = &r.responseTimeMs
		}
		sr, err := d.repository.CreateScanResult(ctx, jobID, subnetID, ipAddressID, r.ip, r.alive, ms, r.ptr, r.fwdRevMismatch)
		if err == nil {
			scanResults = append(scanResults, sr)
			if r.ptr != nil && ipAddressID != nil {
				_ = d.repository.SetIPAddressPTRFromScan(ctx, *ipAddressID, *r.ptr, dnsOverwrite)
			}
		}

		// --- Auto-add discovered IPs (#item5) ---
		if r.alive && ipAddressID == nil && job != nil && job.AutoAddIPs {
			hostname := ""
			if r.ptr != nil {
				hostname = *r.ptr
			}
			newIP, createErr := d.repository.CreateIPAddress(ctx, subnetID, r.ip, hostname, "assigned", nil, nil, nil, r.ptr)
			if createErr != nil {
				log.Printf("[discovery] auto-add IP %s in subnet %d: %v", r.ip, subnetID, createErr)
			} else {
				id := newIP.ID
				ipAddressID = &id
				existingIPs[r.ip] = id
				log.Printf("[discovery] auto-added IP %s to subnet %d (job=%d)", r.ip, subnetID, jobID)
				if r.ptr != nil {
					_ = d.repository.SetIPAddressPTRFromScan(ctx, id, *r.ptr, dnsOverwrite)
				}
			}
		}

		// --- Port scan (#214) ---
		if portScanEnabled && r.alive && ipAddressID != nil {
			portResult := scanner.ScanPorts(ctx, r.ip, ports, portConcurrency, time.Second)
			if err2 := d.repository.UpdateIPPortScan(ctx, *ipAddressID, portResult); err2 != nil {
				log.Printf("port scan update error ip=%s: %v", r.ip, err2)
			}
		}

		// --- SNMP scan (#210) ---
		if doSNMP && r.alive && ipAddressID != nil {
			community, version, v3 := d.snmpCredsForIP(ctx, *ipAddressID, snmpCommunity, snmpVersion)
			snmpResult, err2 := scanner.ScanSNMP(ctx, r.ip, community, version, v3, 5*time.Second)
			if err2 == nil && snmpResult != nil {
				mac := ""
				if len(snmpResult.Interfaces) > 0 {
					mac = snmpResult.Interfaces[0].MACAddress
				}
				_ = d.repository.UpdateIPFromSNMP(ctx, *ipAddressID, mac, snmpResult.SysName)
			}
		}

		// --- Change detection (#211) ---
		if runID > 0 && prevAlive != nil {
			wasAlive, hadPrev := prevAlive[r.ip]
			if !hadPrev && r.alive {
				newCount++
				_, _ = d.repository.CreateScanRunChange(ctx, runID, r.ip, "new")
			} else if hadPrev && wasAlive && !r.alive {
				goneCount++
				_, _ = d.repository.CreateScanRunChange(ctx, runID, r.ip, "gone")
			} else if hadPrev && !wasAlive && r.alive {
				changedCount++
				_, _ = d.repository.CreateScanRunChange(ctx, runID, r.ip, "changed")
			}
		}
	}
	return scanResults, newCount, goneCount, changedCount, nil
}

// RunJob executes a scan job immediately.
// #248: acquires semaphore slot and records in-flight state.
func (d *DiscoveryService) RunJob(ctx context.Context, job *models.ScanJob) error {
	// Check in-flight (skip if already running)
	if _, loaded := d.inFlight.LoadOrStore(job.ID, struct{}{}); loaded {
		log.Printf("scan job %d already running, skipping", job.ID)
		return nil
	}
	defer d.inFlight.Delete(job.ID)

	// Acquire semaphore slot
	select {
	case d.semaphore <- struct{}{}:
	case <-ctx.Done():
		return ctx.Err()
	}
	defer func() { <-d.semaphore }()

	log.Printf("scan job %d started", job.ID)
	defer log.Printf("scan job %d finished", job.ID)

	concurrency := job.PingConcurrency
	if concurrency <= 0 {
		concurrency = 20
	}

	// Create scan run (#211)
	var runID int64
	var prevAlive map[string]bool
	run, err := d.repository.CreateScanRun(ctx, job.ID)
	if err == nil {
		runID = run.ID
		prevAlive, _ = d.repository.GetLastAliveIPsForJob(ctx, job.ID)
	} else {
		log.Printf("scan job %d: create scan run error: %v", job.ID, err)
	}

	var totalNew, totalGone, totalChanged int

	for _, subnetID := range job.SubnetIDs {
		subnet, err := d.repository.GetSubnetByID(ctx, subnetID)
		if err != nil {
			log.Printf("scan job %d: get subnet %d error: %v", job.ID, subnetID, err)
			continue
		}
		ips, err := d.repository.ListIPAddressesBySubnet(ctx, subnetID)
		if err != nil {
			log.Printf("scan job %d: list IPs for subnet %d error: %v", job.ID, subnetID, err)
			continue
		}
		existingIPs := make(map[string]int64, len(ips))
		for _, ip := range ips {
			existingIPs[ip.Address] = ip.ID
		}

		// Apply scan profile overrides: if the subnet has a profile assigned, use its settings
		// as the base; the job's own settings act as the fallback when no profile is set.
		effectiveConcurrency := concurrency
		effectiveJob := job
		if subnet.ScanProfileID != nil {
			if profile, profileErr := d.repository.GetScanProfileByID(ctx, *subnet.ScanProfileID); profileErr == nil {
				effectiveConcurrency = profile.PingConcurrency
				if effectiveConcurrency <= 0 {
					effectiveConcurrency = concurrency
				}
				// Build a shallow copy of job with profile-derived scan type
				jobCopy := *job
				jobCopy.ScanType = profile.ScanType
				effectiveJob = &jobCopy
			} else {
				log.Printf("scan job %d: load profile %d for subnet %d error: %v", job.ID, *subnet.ScanProfileID, subnetID, profileErr)
			}
		}

		_, n, g, ch, err := d.ScanSubnet(ctx, job.ID, subnetID, subnet.NetworkAddress, subnet.PrefixLength, existingIPs, effectiveConcurrency, runID, prevAlive, effectiveJob)
		if err != nil {
			log.Printf("scan job %d: subnet %d scan error: %v", job.ID, subnetID, err)
		}
		totalNew += n
		totalGone += g
		totalChanged += ch
	}

	// Finish scan run (#211)
	if runID > 0 {
		if err := d.repository.FinishScanRun(ctx, runID, totalNew, totalGone, totalChanged); err != nil {
			log.Printf("scan job %d: finish scan run error: %v", job.ID, err)
		}
		// notify_on_change: log (full email notification would need notification service wiring)
		if job.NotifyOnChange && (totalNew+totalGone+totalChanged) > 0 {
			log.Printf("scan job %d: changes detected: new=%d gone=%d changed=%d (notify_on_change=true)", job.ID, totalNew, totalGone, totalChanged)
		}
	}

	return d.repository.UpdateScanJobRunTime(ctx, job.ID, nil)
}

// CreateJob creates a new scan job
func (d *DiscoveryService) CreateJob(ctx context.Context, name string, subnetIDs []int64, scheduleCron *string, createdBy int64, autoAddIPs bool) (*models.ScanJob, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if len(subnetIDs) == 0 {
		return nil, fmt.Errorf("at least one subnet ID is required")
	}
	if scheduleCron != nil && *scheduleCron != "" {
		if err := validateCron(*scheduleCron); err != nil {
			return nil, fmt.Errorf("invalid cron expression: %w", err)
		}
	}
	return d.repository.CreateScanJob(ctx, name, subnetIDs, scheduleCron, createdBy, autoAddIPs)
}

// GetJob retrieves a scan job by ID
func (d *DiscoveryService) GetJob(ctx context.Context, id int64) (*models.ScanJob, error) {
	return d.repository.GetScanJobByID(ctx, id)
}

// IsRunning reports whether a scan job is currently executing.
func (d *DiscoveryService) IsRunning(jobID int64) bool {
	_, ok := d.inFlight.Load(jobID)
	return ok
}

// ListJobs returns all scan jobs
func (d *DiscoveryService) ListJobs(ctx context.Context) ([]*models.ScanJob, error) {
	return d.repository.ListScanJobs(ctx)
}

// UpdateJob updates a scan job
func (d *DiscoveryService) UpdateJob(ctx context.Context, id int64, name string, subnetIDs []int64, scheduleCron *string, isActive bool) (*models.ScanJob, error) {
	if scheduleCron != nil && *scheduleCron != "" {
		if err := validateCron(*scheduleCron); err != nil {
			return nil, fmt.Errorf("invalid cron expression: %w", err)
		}
	}
	return d.repository.UpdateScanJob(ctx, id, name, subnetIDs, scheduleCron, isActive)
}

// UpdateJobFull updates all mutable fields of a scan job.
func (d *DiscoveryService) UpdateJobFull(ctx context.Context, id int64, name string, subnetIDs []int64, scheduleCron *string, isActive bool, pingConcurrency int, notifyOnChange bool, scanType string, agentID *int64, autoAddIPs bool, discoverDNS bool, dnsOverwrite bool) (*models.ScanJob, error) {
	if scheduleCron != nil && *scheduleCron != "" {
		if err := validateCron(*scheduleCron); err != nil {
			return nil, fmt.Errorf("invalid cron expression: %w", err)
		}
	}
	if pingConcurrency <= 0 {
		pingConcurrency = 20
	}
	if pingConcurrency > 100 {
		pingConcurrency = 100
	}
	validTypes := map[string]bool{"ping": true, "snmp": true, "ping+snmp": true}
	if scanType == "" {
		scanType = "ping"
	}
	if !validTypes[scanType] {
		return nil, fmt.Errorf("invalid scan_type: must be ping, snmp, or ping+snmp")
	}
	return d.repository.UpdateScanJobFull(ctx, id, name, subnetIDs, scheduleCron, isActive, pingConcurrency, notifyOnChange, scanType, agentID, autoAddIPs, discoverDNS, dnsOverwrite)
}

// DeleteJob deletes a scan job
func (d *DiscoveryService) DeleteJob(ctx context.Context, id int64) error {
	return d.repository.DeleteScanJob(ctx, id)
}

// ListResults returns scan results for a job
func (d *DiscoveryService) ListResults(ctx context.Context, jobID int64, limit int) ([]*models.ScanResult, error) {
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	return d.repository.ListScanResultsByJob(ctx, jobID, limit)
}

// ListSubnetResults returns scan results for a subnet
func (d *DiscoveryService) ListSubnetResults(ctx context.Context, subnetID int64, limit int) ([]*models.ScanResult, error) {
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	return d.repository.ListScanResultsBySubnet(ctx, subnetID, limit)
}

// ListScanRuns returns the last 50 scan runs for a job.
func (d *DiscoveryService) ListScanRuns(ctx context.Context, jobID int64) ([]*models.ScanRun, error) {
	return d.repository.ListScanRuns(ctx, jobID, 50)
}

// GetScanRunWithChanges returns a scan run and its changes.
func (d *DiscoveryService) GetScanRunWithChanges(ctx context.Context, runID int64) (*models.ScanRun, []*models.ScanRunChange, error) {
	run, err := d.repository.GetScanRun(ctx, runID)
	if err != nil {
		return nil, nil, err
	}
	changes, err := d.repository.ListScanRunChanges(ctx, runID)
	if err != nil {
		return nil, nil, err
	}
	return run, changes, nil
}

// StartScheduler starts the background scheduler that runs active jobs on their cron schedule.
// #248: uses bounded worker pool via semaphore + in-flight tracking.
func (d *DiscoveryService) StartScheduler(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case now := <-ticker.C:
				jobs, err := d.repository.ListActiveScanJobs(ctx)
				if err != nil {
					continue
				}
				for _, job := range jobs {
					if shouldRunJob(job, now) {
						jobCopy := job
						go func() {
							_ = d.RunJob(ctx, jobCopy)
						}()
					}
				}
			}
		}
	}()
}

// shouldRunJob checks if a job should run at the given time based on its cron schedule
func shouldRunJob(job *models.ScanJob, now time.Time) bool {
	if job.ScheduleCron == nil || *job.ScheduleCron == "" {
		return false
	}
	if job.NextRunAt != nil && now.Before(*job.NextRunAt) {
		return false
	}
	return matchesCron(*job.ScheduleCron, now)
}

// matchesCron checks if the current time matches a simple cron expression
// Supports: "* * * * *" style (minute hour dom month dow)
func matchesCron(cron string, t time.Time) bool {
	parts := strings.Fields(cron)
	if len(parts) != 5 {
		return false
	}
	checks := []struct {
		field string
		value int
	}{
		{parts[0], t.Minute()},
		{parts[1], t.Hour()},
		{parts[2], t.Day()},
		{parts[3], int(t.Month())},
		{parts[4], int(t.Weekday())},
	}
	for _, c := range checks {
		if c.field == "*" {
			continue
		}
		v, err := strconv.Atoi(c.field)
		if err != nil || v != c.value {
			return false
		}
	}
	return true
}

// validateCron validates that a cron expression has 5 fields
func validateCron(cron string) error {
	parts := strings.Fields(cron)
	if len(parts) != 5 {
		return fmt.Errorf("cron must have 5 fields (min hour dom month dow), got %d", len(parts))
	}
	return nil
}

// enumerateCIDR returns all host IPs in a subnet (excluding network and broadcast)
func enumerateCIDR(networkAddr string, prefixLen int) ([]string, error) {
	cidr := fmt.Sprintf("%s/%d", networkAddr, prefixLen)
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	ip := ipNet.IP.To4()
	if ip == nil {
		return nil, fmt.Errorf("only IPv4 supported")
	}

	start := binary.BigEndian.Uint32(ip)
	mask := binary.BigEndian.Uint32(ipNet.Mask)
	network := start & mask
	broadcast := network | ^mask

	ips := make([]string, 0, int(broadcast-network-1))
	for i := network + 1; i < broadcast; i++ {
		b := make([]byte, 4)
		binary.BigEndian.PutUint32(b, i)
		ips = append(ips, net.IP(b).String())
	}
	return ips, nil
}

// ---------------------------------------------------------------------------
// Scan agent service methods (#212)
// ---------------------------------------------------------------------------

// CreateAgent creates a new scan agent and returns the raw token (shown once).
// ttlDays == 0 means the token never expires.
func (d *DiscoveryService) CreateAgent(ctx context.Context, name string, ttlDays int) (*models.ScanAgent, string, error) {
	if name == "" {
		return nil, "", fmt.Errorf("agent name is required")
	}
	rawToken, tokenHash, err := generateAgentToken()
	if err != nil {
		return nil, "", fmt.Errorf("generate token: %w", err)
	}
	agent, err := d.repository.CreateScanAgent(ctx, name, tokenHash, ttlDays)
	if err != nil {
		return nil, "", err
	}
	return agent, rawToken, nil
}

// ListAgents returns all scan agents.
func (d *DiscoveryService) ListAgents(ctx context.Context) ([]*models.ScanAgent, error) {
	return d.repository.ListScanAgents(ctx)
}

// GetAgent retrieves a scan agent by ID.
func (d *DiscoveryService) GetAgent(ctx context.Context, id int64) (*models.ScanAgent, error) {
	return d.repository.GetScanAgentByID(ctx, id)
}

// RotateToken issues a new token for an agent and returns the raw token.
// ttlDays == 0 clears the expiry; ttlDays < 0 preserves the existing expiry.
func (d *DiscoveryService) RotateToken(ctx context.Context, id int64, ttlDays int) (*models.ScanAgent, string, error) {
	rawToken, tokenHash, err := generateAgentToken()
	if err != nil {
		return nil, "", fmt.Errorf("generate token: %w", err)
	}
	agent, err := d.repository.UpdateScanAgentToken(ctx, id, tokenHash, ttlDays)
	if err != nil {
		return nil, "", err
	}
	return agent, rawToken, nil
}

// DeleteAgent removes a scan agent.
func (d *DiscoveryService) DeleteAgent(ctx context.Context, id int64) error {
	return d.repository.DeleteScanAgent(ctx, id)
}

// MarkOfflineStale marks agents offline if last_seen > 15 minutes ago.
func (d *DiscoveryService) MarkOfflineStale(ctx context.Context) error {
	agents, err := d.repository.ListScanAgents(ctx)
	if err != nil {
		return err
	}
	threshold := time.Now().Add(-15 * time.Minute)
	for _, a := range agents {
		if a.IsActive && a.LastSeen != nil && a.LastSeen.Before(threshold) {
			if _, err := d.repository.UpdateScanAgentActive(ctx, a.ID, false); err != nil {
				log.Printf("mark agent %d offline error: %v", a.ID, err)
			}
		}
	}
	return nil
}

// AuthenticateAgent validates a Bearer token and returns the agent.
func (d *DiscoveryService) AuthenticateAgent(ctx context.Context, rawToken string) (*models.ScanAgent, error) {
	tokenHash := hashAgentToken(rawToken)
	return d.repository.GetScanAgentByToken(ctx, tokenHash)
}

// GetJobsForAgent returns active scan jobs assigned to an agent.
func (d *DiscoveryService) GetJobsForAgent(ctx context.Context, agentID int64) ([]*models.ScanJob, error) {
	return d.repository.ListScanJobsForAgent(ctx, agentID)
}

// HeartbeatAgent records that an agent is alive and stores optional health metadata.
func (d *DiscoveryService) HeartbeatAgent(ctx context.Context, agentID int64, version *string, capabilities []string, status string, lastError *string) error {
	return d.repository.HeartbeatAgent(ctx, agentID, version, capabilities, status, lastError)
}

// AcceptAgentResults processes scan results submitted by a remote agent.
func (d *DiscoveryService) AcceptAgentResults(ctx context.Context, agentID int64, jobID int64, results []AgentScanResult) error {
	job, err := d.repository.GetScanJobByID(ctx, jobID)
	if err != nil {
		return fmt.Errorf("job not found: %w", err)
	}
	if job.AgentID == nil || *job.AgentID != agentID {
		return fmt.Errorf("job not assigned to this agent")
	}
	for _, res := range results {
		var ipAddrID *int64
		if res.IPAddressID > 0 {
			v := res.IPAddressID
			ipAddrID = &v
		}
		var ms *int64
		if res.ResponseTimeMs > 0 {
			ms = &res.ResponseTimeMs
		}
		_, err := d.repository.CreateScanResult(ctx, jobID, res.SubnetID, ipAddrID, res.IPAddress, res.IsAlive, ms, nil, false)
		if err != nil {
			log.Printf("agent %d: store result for %s: %v", agentID, res.IPAddress, err)
		}
		// Auto-add: if alive and no existing IP record, create one.
		if res.IsAlive && ipAddrID == nil && job.AutoAddIPs && res.SubnetID > 0 {
			newIP, createErr := d.repository.CreateIPAddress(ctx, res.SubnetID, res.IPAddress, "", "assigned", nil, nil, nil, nil)
			if createErr != nil {
				log.Printf("agent %d: auto-add IP %s in subnet %d: %v", agentID, res.IPAddress, res.SubnetID, createErr)
			} else {
				log.Printf("agent %d: auto-added IP %s to subnet %d", agentID, res.IPAddress, res.SubnetID)
				_ = newIP
			}
		}
	}
	return d.repository.UpdateScanJobRunTime(ctx, jobID, nil)
}

// AgentScanResult is the payload submitted by an agent for a single IP.
type AgentScanResult struct {
	SubnetID       int64  `json:"subnet_id"`
	IPAddressID    int64  `json:"ip_address_id,omitempty"`
	IPAddress      string `json:"ip_address"`
	IsAlive        bool   `json:"is_alive"`
	ResponseTimeMs int64  `json:"response_time_ms,omitempty"`
}

// snmpCredsForIP returns the SNMP community, version, and SNMPv3 params to use
// when scanning ipID. If the IP is linked to a device with stored credentials
// those are preferred; otherwise the globalCommunity/globalVersion fallbacks are
// returned with a nil v3 param.
func (d *DiscoveryService) snmpCredsForIP(ctx context.Context, ipID int64, globalCommunity, globalVersion string) (community, version string, v3 *scanner.SNMPv3Params) {
	community = globalCommunity
	version = globalVersion

	creds, err := d.repository.GetDeviceSNMPByIPID(ctx, ipID)
	if err != nil {
		log.Printf("[discovery] snmpCredsForIP ip_id=%d: %v", ipID, err)
		return
	}
	if creds == nil {
		return // no device linked
	}

	if creds.SNMPCommunity != nil && *creds.SNMPCommunity != "" {
		if dec, err := DecryptString(d.encryptionKey, *creds.SNMPCommunity); err == nil {
			community = dec
		}
	}
	if creds.SNMPVersion != "" {
		version = creds.SNMPVersion
	}
	if version == "v3" || version == "3" {
		version = "3"
		p := &scanner.SNMPv3Params{}
		if creds.SNMPV3User != nil {
			p.User = *creds.SNMPV3User
		}
		if creds.SNMPV3AuthProto != nil {
			p.AuthProto = *creds.SNMPV3AuthProto
		}
		if creds.SNMPV3AuthPass != nil {
			if dec, err := DecryptString(d.encryptionKey, *creds.SNMPV3AuthPass); err == nil {
				p.AuthPass = dec
			}
		}
		if creds.SNMPV3PrivProto != nil {
			p.PrivProto = *creds.SNMPV3PrivProto
		}
		if creds.SNMPV3PrivPass != nil {
			if dec, err := DecryptString(d.encryptionKey, *creds.SNMPV3PrivPass); err == nil {
				p.PrivPass = dec
			}
		}
		v3 = p
	}
	return
}

// ---------------------------------------------------------------------------
// Scan profile service methods (#432)
// ---------------------------------------------------------------------------

// CreateScanProfile creates a new scan profile.
func (d *DiscoveryService) CreateScanProfile(ctx context.Context, name, scanType string, desc *string, pingConcurrency int, tcpPorts *string, dnsLookup bool, snmpCommunity *string, snmpVersion string) (*models.ScanProfile, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	validTypes := map[string]bool{"ping": true, "snmp": true, "ping+snmp": true}
	if scanType == "" {
		scanType = "ping"
	}
	if !validTypes[scanType] {
		return nil, fmt.Errorf("invalid scan_type: must be ping, snmp, or ping+snmp")
	}
	if pingConcurrency <= 0 {
		pingConcurrency = 20
	}
	if snmpVersion == "" {
		snmpVersion = "v2c"
	}
	return d.repository.CreateScanProfile(ctx, name, scanType, desc, pingConcurrency, tcpPorts, dnsLookup, snmpCommunity, snmpVersion)
}

// ListScanProfiles returns all scan profiles.
func (d *DiscoveryService) ListScanProfiles(ctx context.Context) ([]*models.ScanProfile, error) {
	return d.repository.ListScanProfiles(ctx)
}

// GetScanProfileByID retrieves a scan profile by ID.
func (d *DiscoveryService) GetScanProfileByID(ctx context.Context, id int64) (*models.ScanProfile, error) {
	return d.repository.GetScanProfileByID(ctx, id)
}

// UpdateScanProfile updates a scan profile.
func (d *DiscoveryService) UpdateScanProfile(ctx context.Context, id int64, name, scanType string, desc *string, pingConcurrency int, tcpPorts *string, dnsLookup bool, snmpCommunity *string, snmpVersion string) (*models.ScanProfile, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	validTypes := map[string]bool{"ping": true, "snmp": true, "ping+snmp": true}
	if scanType == "" {
		scanType = "ping"
	}
	if !validTypes[scanType] {
		return nil, fmt.Errorf("invalid scan_type: must be ping, snmp, or ping+snmp")
	}
	if pingConcurrency <= 0 {
		pingConcurrency = 20
	}
	if snmpVersion == "" {
		snmpVersion = "v2c"
	}
	return d.repository.UpdateScanProfile(ctx, id, name, scanType, desc, pingConcurrency, tcpPorts, dnsLookup, snmpCommunity, snmpVersion)
}

// DeleteScanProfile removes a scan profile.
func (d *DiscoveryService) DeleteScanProfile(ctx context.Context, id int64) error {
	return d.repository.DeleteScanProfile(ctx, id)
}

// SetSubnetScanProfile assigns or clears the scan profile for a subnet.
func (d *DiscoveryService) SetSubnetScanProfile(ctx context.Context, subnetID int64, profileID *int64) error {
	return d.repository.SetSubnetScanProfile(ctx, subnetID, profileID)
}

// ---------------------------------------------------------------------------
// Scan retention service methods (#435)
// ---------------------------------------------------------------------------

// GetRetentionSettings returns the current scan retention settings.
func (d *DiscoveryService) GetRetentionSettings(ctx context.Context) (*models.ScanRetentionSettings, error) {
	return d.repository.GetScanRetentionSettings(ctx)
}

// UpdateRetentionSettings validates and persists updated retention settings.
func (d *DiscoveryService) UpdateRetentionSettings(ctx context.Context, rawHistoryDays, rollupAfterDays int, rollupEnabled bool) (*models.ScanRetentionSettings, error) {
	if rawHistoryDays < 1 {
		return nil, fmt.Errorf("raw_history_days must be >= 1")
	}
	if rollupAfterDays < 1 {
		return nil, fmt.Errorf("rollup_after_days must be >= 1")
	}
	return d.repository.UpdateScanRetentionSettings(ctx, rawHistoryDays, rollupAfterDays, rollupEnabled)
}

// RunRetentionPrune reads retention settings and deletes scan data older than the configured threshold.
func (d *DiscoveryService) RunRetentionPrune(ctx context.Context) (int64, error) {
	settings, err := d.repository.GetScanRetentionSettings(ctx)
	if err != nil {
		return 0, fmt.Errorf("get retention settings: %w", err)
	}
	return d.repository.PruneScanHistory(ctx, settings.RawHistoryDays)
}

// ---------------------------------------------------------------------------
// Device fingerprints (#430)
// ---------------------------------------------------------------------------

// GetDeviceFingerprint returns the stored fingerprint for a device.
func (d *DiscoveryService) GetDeviceFingerprint(ctx context.Context, deviceID int64) (*models.DeviceFingerprint, error) {
	return d.repository.GetDeviceFingerprint(ctx, deviceID)
}

// BuildDeviceFingerprint derives a fingerprint from the device's scan data and persists it.
func (d *DiscoveryService) BuildDeviceFingerprint(ctx context.Context, deviceID int64, deviceIP string, isAlive bool, ptrRecord *string, openPortsStr *string) (*models.DeviceFingerprint, error) {
	var openPorts []int
	var evidence []string
	var confidence float64

	if isAlive {
		confidence += 0.4
		evidence = append(evidence, "responds to ping")
	}
	if ptrRecord != nil && *ptrRecord != "" {
		confidence += 0.2
		evidence = append(evidence, "has PTR record: "+*ptrRecord)
	}
	if openPortsStr != nil && *openPortsStr != "" {
		parts := strings.Split(*openPortsStr, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if n, err := strconv.Atoi(p); err == nil {
				openPorts = append(openPorts, n)
			}
		}
		if len(openPorts) > 0 {
			confidence += 0.2
			evidence = append(evidence, fmt.Sprintf("%d open ports detected", len(openPorts)))
		}
	}

	var osGuess *string
	var vendorGuess *string
	portSet := make(map[int]bool)
	for _, p := range openPorts {
		portSet[p] = true
	}
	if portSet[22] && portSet[80] {
		g := "Linux/Unix"
		osGuess = &g
		confidence += 0.1
	} else if portSet[3389] || portSet[445] {
		g := "Windows"
		osGuess = &g
		confidence += 0.1
	}
	if portSet[161] {
		g := "Network device"
		vendorGuess = &g
		evidence = append(evidence, "SNMP port open")
	}

	if confidence > 1.0 {
		confidence = 1.0
	}

	return d.repository.UpsertDeviceFingerprint(ctx, deviceID, openPorts, osGuess, vendorGuess, confidence, evidence)
}

// StartRetentionPruner starts a background goroutine that runs RunRetentionPrune once per day.
func (d *DiscoveryService) StartRetentionPruner(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if _, err := d.RunRetentionPrune(ctx); err != nil {
					log.Printf("[retention] prune error: %v", err)
				}
			}
		}
	}()
}

// ---------------------------------------------------------------------------
// Discovery conflicts (#431)
// ---------------------------------------------------------------------------

// ListDiscoveryConflicts returns all discovery conflicts, optionally filtered by status.
func (d *DiscoveryService) ListDiscoveryConflicts(ctx context.Context, status string) ([]*models.DiscoveryConflict, error) {
	return d.repository.ListDiscoveryConflicts(ctx, status)
}

// GetDiscoveryConflict retrieves a single discovery conflict by ID.
func (d *DiscoveryService) GetDiscoveryConflict(ctx context.Context, id int64) (*models.DiscoveryConflict, error) {
	return d.repository.GetDiscoveryConflict(ctx, id)
}

// CreateDiscoveryConflict records a new conflict between discovered and manual data.
func (d *DiscoveryService) CreateDiscoveryConflict(ctx context.Context, deviceID int64, fieldName, discoveredValue string, currentValue *string, confidence float64, source string) (*models.DiscoveryConflict, error) {
	return d.repository.CreateDiscoveryConflict(ctx, deviceID, fieldName, discoveredValue, currentValue, confidence, source)
}

// ResolveDiscoveryConflict accepts or rejects a pending conflict.
func (d *DiscoveryService) ResolveDiscoveryConflict(ctx context.Context, id int64, action string, reviewedBy string) (*models.DiscoveryConflict, error) {
	if action != "accepted" && action != "rejected" {
		return nil, fmt.Errorf("action must be 'accepted' or 'rejected'")
	}
	return d.repository.ResolveDiscoveryConflict(ctx, id, action, reviewedBy)
}
