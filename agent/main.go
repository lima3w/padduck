// Package main is the IPAM Next remote scan agent binary.
// It polls the IPAM server for assigned scan jobs, runs ping scans,
// and posts results back.
//
// Configuration via environment variables:
//
//	IPAM_SERVER_URL  — base URL of the IPAM server (e.g. https://ipam.example.com)
//	IPAM_AGENT_TOKEN — raw bearer token issued when creating the agent
//	POLL_INTERVAL    — polling interval in seconds (default: 30)
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
)

// AgentScanResult matches the server-side struct.
type AgentScanResult struct {
	SubnetID       int64  `json:"subnet_id"`
	IPAddressID    int64  `json:"ip_address_id,omitempty"`
	IPAddress      string `json:"ip_address"`
	IsAlive        bool   `json:"is_alive"`
	ResponseTimeMs int64  `json:"response_time_ms,omitempty"`
}

// subnetInfo is a single subnet entry in a job response.
type subnetInfo struct {
	ID   int64  `json:"id"`
	CIDR string `json:"cidr"`
}

// ScanJob is the enriched job payload returned by /scan-agent/jobs.
type ScanJob struct {
	ID              int64        `json:"id"`
	Name            string       `json:"name"`
	Subnets         []subnetInfo `json:"subnets"`
	PingConcurrency int          `json:"ping_concurrency"`
	ScanType        string       `json:"scan_type"`
}

func main() {
	serverURL := os.Getenv("IPAM_SERVER_URL")
	agentToken := os.Getenv("IPAM_AGENT_TOKEN")
	pollIntervalStr := os.Getenv("POLL_INTERVAL")

	if serverURL == "" || agentToken == "" {
		log.Fatal("IPAM_SERVER_URL and IPAM_AGENT_TOKEN must be set")
	}

	pollInterval := 30 * time.Second
	if pollIntervalStr != "" {
		if n, err := strconv.Atoi(pollIntervalStr); err == nil && n > 0 {
			pollInterval = time.Duration(n) * time.Second
		}
	}

	client := &http.Client{Timeout: 30 * time.Second}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Println("shutting down agent")
		cancel()
	}()

	log.Printf("IPAM scan agent started, server=%s poll_interval=%s", serverURL, pollInterval)

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	runCycle(ctx, client, serverURL, agentToken)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			runCycle(ctx, client, serverURL, agentToken)
		}
	}
}

func runCycle(ctx context.Context, client *http.Client, serverURL, token string) {
	if err := doHeartbeat(ctx, client, serverURL, token); err != nil {
		log.Printf("heartbeat error: %v", err)
	}

	jobs, err := fetchJobs(ctx, client, serverURL, token)
	if err != nil {
		log.Printf("fetch jobs error: %v", err)
		return
	}

	for _, job := range jobs {
		log.Printf("running job %d (%s): %d subnet(s)", job.ID, job.Name, len(job.Subnets))
		results := runJob(ctx, job)
		log.Printf("job %d: %d results", job.ID, len(results))
		if err := postResults(ctx, client, serverURL, token, job.ID, results); err != nil {
			log.Printf("post results for job %d error: %v", job.ID, err)
		}
	}
}

func doHeartbeat(ctx context.Context, client *http.Client, serverURL, token string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, serverURL+"/api/v1/scan-agent/heartbeat", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("heartbeat status %d", resp.StatusCode)
	}
	return nil
}

func fetchJobs(ctx context.Context, client *http.Client, serverURL, token string) ([]ScanJob, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, serverURL+"/api/v1/scan-agent/jobs", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch jobs status %d", resp.StatusCode)
	}
	var jobs []ScanJob
	if err := json.NewDecoder(resp.Body).Decode(&jobs); err != nil {
		return nil, err
	}
	return jobs, nil
}

// runJob scans every subnet in the job and returns liveness results.
func runJob(ctx context.Context, job ScanJob) []AgentScanResult {
	concurrency := job.PingConcurrency
	if concurrency <= 0 {
		concurrency = 10
	}

	var results []AgentScanResult
	var mu sync.Mutex

	for _, sn := range job.Subnets {
		ips, err := enumerateCIDR(sn.CIDR)
		if err != nil {
			log.Printf("job %d: enumerate %s: %v", job.ID, sn.CIDR, err)
			continue
		}

		sem := make(chan struct{}, concurrency)
		var wg sync.WaitGroup

		for _, ip := range ips {
			if ctx.Err() != nil {
				break
			}
			wg.Add(1)
			sem <- struct{}{}
			go func(ipStr string) {
				defer wg.Done()
				defer func() { <-sem }()

				alive, ms := pingHost(ipStr, 2*time.Second)
				res := AgentScanResult{
					SubnetID:  sn.ID,
					IPAddress: ipStr,
					IsAlive:   alive,
				}
				if alive {
					res.ResponseTimeMs = ms
				}
				mu.Lock()
				results = append(results, res)
				mu.Unlock()
			}(ip)
		}
		wg.Wait()
	}
	return results
}

// pingHost returns whether the host responded and the round-trip time in ms.
func pingHost(host string, timeout time.Duration) (bool, int64) {
	start := time.Now()
	cmd := exec.Command("ping", "-c", "1", "-W", strconv.Itoa(int(timeout.Seconds())), host)
	err := cmd.Run()
	elapsed := time.Since(start).Milliseconds()
	if err != nil {
		return false, 0
	}
	return true, elapsed
}

// enumerateCIDR returns all host IPs in a subnet (excluding network and broadcast).
func enumerateCIDR(cidr string) ([]string, error) {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}
	base := ipNet.IP.To4()
	if base == nil {
		return nil, fmt.Errorf("only IPv4 supported")
	}
	// Compute broadcast address: base | ~mask
	broadcast := make(net.IP, 4)
	for i := 0; i < 4; i++ {
		broadcast[i] = base[i] | ^ipNet.Mask[i]
	}
	var ips []string
	for cur := cloneIP(base); ipNet.Contains(cur); incrementIP(cur) {
		if cur.Equal(net.IP(ipNet.IP)) || cur.Equal(broadcast) {
			continue
		}
		ips = append(ips, cur.String())
	}
	return ips, nil
}

func cloneIP(ip net.IP) net.IP {
	out := make(net.IP, len(ip))
	copy(out, ip)
	return out
}

func incrementIP(ip net.IP) {
	for i := len(ip) - 1; i >= 0; i-- {
		ip[i]++
		if ip[i] != 0 {
			break
		}
	}
}

func postResults(ctx context.Context, client *http.Client, serverURL, token string, jobID int64, results []AgentScanResult) error {
	payload := struct {
		JobID   int64             `json:"job_id"`
		Results []AgentScanResult `json:"results"`
	}{JobID: jobID, Results: results}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, serverURL+"/api/v1/scan-agent/results", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("post results status %d", resp.StatusCode)
	}
	return nil
}
