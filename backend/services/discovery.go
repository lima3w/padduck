package services

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"ipam-next/internal/scanner"
	"ipam-next/models"
)

// DiscoveryService handles network scanning and IP detection
type DiscoveryService struct {
	repository discoveryRepo
	config     *ConfigService
}

type discoveryRepo interface {
	GetSubnetByID(ctx context.Context, id int64) (*models.Subnet, error)
	ListIPAddressesBySubnet(ctx context.Context, subnetID int64) ([]*models.IPAddress, error)
	CreateScanJob(ctx context.Context, name string, subnetIDs []int64, scheduleCron *string, createdBy int64) (*models.ScanJob, error)
	GetScanJobByID(ctx context.Context, id int64) (*models.ScanJob, error)
	ListScanJobs(ctx context.Context) ([]*models.ScanJob, error)
	ListActiveScanJobs(ctx context.Context) ([]*models.ScanJob, error)
	UpdateScanJob(ctx context.Context, id int64, name string, subnetIDs []int64, scheduleCron *string, isActive bool) (*models.ScanJob, error)
	UpdateScanJobRunTime(ctx context.Context, id int64, nextRunAt *time.Time) error
	DeleteScanJob(ctx context.Context, id int64) error
	CreateScanResult(ctx context.Context, jobID, subnetID int64, ipAddressID *int64, ipAddress string, isAlive bool, responseTimeMs *int64, ptrRecord *string, fwdRevMismatch bool) (*models.ScanResult, error)
	ListScanResultsByJob(ctx context.Context, jobID int64, limit int) ([]*models.ScanResult, error)
	ListScanResultsBySubnet(ctx context.Context, subnetID int64, limit int) ([]*models.ScanResult, error)
	SetIPAddressPTRFromScan(ctx context.Context, ipID int64, ptrRecord string) error
}

func NewDiscoveryService(repo discoveryRepo, configSvc *ConfigService) *DiscoveryService {
	return &DiscoveryService{repository: repo, config: configSvc}
}

// PingHost checks if a host responds to ICMP ping, returning response time in ms
func PingHost(host string, timeout time.Duration) (bool, int64) {
	start := time.Now()
	cmd := exec.Command("ping", "-c", "1", "-W", strconv.Itoa(int(timeout.Seconds())), host)
	err := cmd.Run()
	elapsed := time.Since(start).Milliseconds()
	if err != nil {
		return false, 0
	}
	return true, elapsed
}

// ScanSubnet scans all IPs in a subnet CIDR range for liveness
func (d *DiscoveryService) ScanSubnet(ctx context.Context, jobID, subnetID int64, networkAddr string, prefixLen int, existingIPs map[string]int64, concurrency int) ([]*models.ScanResult, error) {
	if concurrency <= 0 {
		concurrency = 20
	}

	resolveHostnames := true
	if d.config != nil {
		if v, _ := d.config.Get("scanner_resolve_hostnames"); v == "false" {
			resolveHostnames = false
		}
	}

	ips, err := enumerateCIDR(networkAddr, prefixLen)
	if err != nil {
		return nil, fmt.Errorf("enumerate CIDR: %w", err)
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
			// Propagate PTR to the ip_addresses row when it exists.
			if r.ptr != nil && ipAddressID != nil {
				_ = d.repository.SetIPAddressPTRFromScan(ctx, *ipAddressID, *r.ptr)
			}
		}
	}
	return scanResults, nil
}

// RunJob executes a scan job immediately
func (d *DiscoveryService) RunJob(ctx context.Context, job *models.ScanJob) error {
	for _, subnetID := range job.SubnetIDs {
		subnet, err := d.repository.GetSubnetByID(ctx, subnetID)
		if err != nil {
			continue
		}
		ips, err := d.repository.ListIPAddressesBySubnet(ctx, subnetID)
		if err != nil {
			continue
		}
		existingIPs := make(map[string]int64, len(ips))
		for _, ip := range ips {
			existingIPs[ip.Address] = ip.ID
		}
		_, _ = d.ScanSubnet(ctx, job.ID, subnetID, subnet.NetworkAddress, subnet.PrefixLength, existingIPs, 20)
	}
	return d.repository.UpdateScanJobRunTime(ctx, job.ID, nil)
}

// CreateJob creates a new scan job
func (d *DiscoveryService) CreateJob(ctx context.Context, name string, subnetIDs []int64, scheduleCron *string, createdBy int64) (*models.ScanJob, error) {
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
	return d.repository.CreateScanJob(ctx, name, subnetIDs, scheduleCron, createdBy)
}

// GetJob retrieves a scan job by ID
func (d *DiscoveryService) GetJob(ctx context.Context, id int64) (*models.ScanJob, error) {
	return d.repository.GetScanJobByID(ctx, id)
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

// StartScheduler starts the background scheduler that runs active jobs on their cron schedule
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
