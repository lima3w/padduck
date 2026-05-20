// Package main is the Padduck remote scan agent binary.
// It polls the Padduck server for assigned scan jobs, runs ping scans,
// and posts results back.
//
// Configuration via environment variables:
//
//	PADDUCK_SERVER_URL  — base URL of the Padduck server (e.g. https://padduck.example.com)
//	PADDUCK_AGENT_TOKEN — raw bearer token issued when creating the agent
//	POLL_INTERVAL       — polling interval in seconds (default: 30)
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
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
	serverURL := os.Getenv("PADDUCK_SERVER_URL")
	agentToken := os.Getenv("PADDUCK_AGENT_TOKEN")
	pollIntervalStr := os.Getenv("POLL_INTERVAL")

	if serverURL == "" || agentToken == "" {
		log.Fatal("PADDUCK_SERVER_URL and PADDUCK_AGENT_TOKEN must be set")
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

	log.Printf("Padduck scan agent started, server=%q poll_interval=%s", serverURL, pollInterval) // #nosec G706 -- operator-provided URL is quoted in a local startup log.

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
	endpoint, err := agentEndpoint(serverURL, "/api/v1/scan-agent/heartbeat")
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, nil) // #nosec G704 -- server URL is operator configured and scheme validated.
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := client.Do(req) // #nosec G704 -- request target is the configured IPAM server.
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("heartbeat status %d", resp.StatusCode)
	}
	return nil
}

func fetchJobs(ctx context.Context, client *http.Client, serverURL, token string) ([]ScanJob, error) {
	endpoint, err := agentEndpoint(serverURL, "/api/v1/scan-agent/jobs")
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil) // #nosec G704 -- server URL is operator configured and scheme validated.
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := client.Do(req) // #nosec G704 -- request target is the configured IPAM server.
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
	if net.ParseIP(host) == nil {
		return false, 0
	}
	start := time.Now()
	cmd := exec.Command("ping", "-c", "1", "-W", strconv.Itoa(int(timeout.Seconds())), host) // #nosec G204 -- host is generated from parsed CIDR IPs.
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
	endpoint, err := agentEndpoint(serverURL, "/api/v1/scan-agent/results")
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body)) // #nosec G704 -- server URL is operator configured and scheme validated.
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req) // #nosec G704 -- request target is the configured IPAM server.
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("post results status %d", resp.StatusCode)
	}
	return nil
}

func agentEndpoint(serverURL, path string) (string, error) {
	base, err := url.Parse(strings.TrimSpace(serverURL))
	if err != nil {
		return "", err
	}
	if base.Scheme != "http" && base.Scheme != "https" {
		return "", fmt.Errorf("unsupported server URL scheme %q", base.Scheme)
	}
	if base.Host == "" {
		return "", fmt.Errorf("server URL host is required")
	}
	base.Path = strings.TrimRight(base.Path, "/") + path
	base.RawQuery = ""
	base.Fragment = ""
	return base.String(), nil
}
