package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// collectHostIPs gathers the iterated IPs for assertions.
func collectHostIPs(cidr string) ([]string, error) {
	var ips []string
	err := forEachHostIP(cidr, func(ip string) bool {
		ips = append(ips, ip)
		return true
	})
	return ips, err
}

func TestForEachHostIP(t *testing.T) {
	tests := []struct {
		cidr    string
		wantLen int
		wantErr bool
	}{
		{"192.168.1.0/30", 2, false}, // .1 and .2 (excludes .0 and .3)
		{"10.0.0.0/29", 6, false},    // .1–.6
		{"10.0.0.0/24", 254, false},
		{"10.0.0.0/16", 65534, false}, // exactly at the limit
		{"invalid", 0, true},
	}

	for _, tt := range tests {
		ips, err := collectHostIPs(tt.cidr)
		if tt.wantErr {
			if err == nil {
				t.Errorf("forEachHostIP(%q): expected error, got none", tt.cidr)
			}
			continue
		}
		if err != nil {
			t.Errorf("forEachHostIP(%q): unexpected error: %v", tt.cidr, err)
			continue
		}
		if len(ips) != tt.wantLen {
			t.Errorf("forEachHostIP(%q): got %d IPs, want %d", tt.cidr, len(ips), tt.wantLen)
		}
	}
}

func TestForEachHostIPRejectsBroadPrefixes(t *testing.T) {
	for _, cidr := range []string{"10.0.0.0/8", "0.0.0.0/0", "172.16.0.0/12", "10.0.0.0/15"} {
		called := false
		err := forEachHostIP(cidr, func(string) bool {
			called = true
			return true
		})
		if err == nil {
			t.Errorf("forEachHostIP(%q): expected error for prefix broader than /%d", cidr, minCIDRPrefix)
		}
		if called {
			t.Errorf("forEachHostIP(%q): callback invoked despite rejected prefix", cidr)
		}
	}
}

func TestForEachHostIPStopsEarly(t *testing.T) {
	count := 0
	err := forEachHostIP("10.0.0.0/24", func(string) bool {
		count++
		return count < 10
	})
	if err != nil {
		t.Fatal(err)
	}
	if count != 10 {
		t.Errorf("expected iteration to stop after 10 callbacks, got %d", count)
	}
}

func TestForEachHostIPExcludesNetworkAndBroadcast(t *testing.T) {
	ips, err := collectHostIPs("192.168.1.0/24")
	if err != nil {
		t.Fatal(err)
	}
	for _, ip := range ips {
		if ip == "192.168.1.0" || ip == "192.168.1.255" {
			t.Errorf("forEachHostIP: included network/broadcast address %q", ip)
		}
	}
}

func TestRunJobEmptySubnets(t *testing.T) {
	job := ScanJob{
		ID:              1,
		Name:            "test",
		Subnets:         nil,
		PingConcurrency: 4,
	}
	results := runJob(context.Background(), job)
	if len(results) != 0 {
		t.Errorf("runJob with no subnets: expected 0 results, got %d", len(results))
	}
}

func TestRunJobInvalidCIDR(t *testing.T) {
	job := ScanJob{
		ID:   2,
		Name: "bad",
		Subnets: []subnetInfo{
			{ID: 99, CIDR: "not-a-cidr"},
		},
		PingConcurrency: 2,
	}
	results := runJob(context.Background(), job)
	if len(results) != 0 {
		t.Errorf("runJob with invalid CIDR: expected 0 results, got %d", len(results))
	}
}

func TestRunJobCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // pre-cancel

	job := ScanJob{
		ID:   3,
		Name: "cancelled",
		Subnets: []subnetInfo{
			{ID: 1, CIDR: "10.255.255.0/24"},
		},
		PingConcurrency: 10,
	}
	// Should return quickly without scanning due to cancelled context.
	results := runJob(ctx, job)
	_ = results // result count is non-deterministic, just ensure no panic/hang
}

func TestValidateServerURL(t *testing.T) {
	tests := []struct {
		url           string
		allowInsecure bool
		wantErr       bool
	}{
		{"https://padduck.example.com", false, false},
		{"http://padduck.example.com", false, true},  // cleartext token requires opt-in
		{"http://padduck.example.com", true, false},  // explicit opt-in
		{"ftp://padduck.example.com", false, true},
		{"ftp://padduck.example.com", true, true}, // opt-in does not unlock other schemes
		{"padduck.example.com", false, true},      // not absolute
		{"", false, true},
	}
	for _, tt := range tests {
		err := validateServerURL(tt.url, tt.allowInsecure)
		if (err != nil) != tt.wantErr {
			t.Errorf("validateServerURL(%q, %v): err=%v, wantErr=%v", tt.url, tt.allowInsecure, err, tt.wantErr)
		}
	}
}

// ---------------------------------------------------------------------------
// agentEndpoint — path joining and scheme validation
// ---------------------------------------------------------------------------

func TestAgentEndpoint_PathJoining(t *testing.T) {
	cases := []struct {
		baseURL  string
		path     string
		wantSufx string // suffix the result must end with
		wantErr  bool
	}{
		// Base URL without trailing slash.
		{"https://example.com", "/api/v1/scan-agent/heartbeat", "/api/v1/scan-agent/heartbeat", false},
		// Base URL with trailing slash — should not produce double slash.
		{"https://example.com/", "/api/v1/scan-agent/heartbeat", "/api/v1/scan-agent/heartbeat", false},
		// Base URL with a path prefix.
		{"https://example.com/padduck", "/api/v1/scan-agent/heartbeat", "/padduck/api/v1/scan-agent/heartbeat", false},
		// Base URL with path prefix and trailing slash.
		{"https://example.com/padduck/", "/api/v1/scan-agent/heartbeat", "/padduck/api/v1/scan-agent/heartbeat", false},
		// Query string and fragment must be stripped.
		{"https://example.com?foo=bar#baz", "/api/v1/scan-agent/heartbeat", "/api/v1/scan-agent/heartbeat", false},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("%s+%s", tc.baseURL, tc.path), func(t *testing.T) {
			got, err := agentEndpoint(tc.baseURL, tc.path)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error, got %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.HasSuffix(got, tc.wantSufx) {
				t.Errorf("got %q, want suffix %q", got, tc.wantSufx)
			}
			// No query string or fragment in result.
			if strings.ContainsAny(got, "?#") {
				t.Errorf("result %q must not contain query or fragment", got)
			}
		})
	}
}

func TestAgentEndpoint_SchemeRejection(t *testing.T) {
	// Only http and https are accepted; anything else is an error.
	_, err := agentEndpoint("ftp://example.com", "/path")
	if err == nil {
		t.Error("expected error for ftp:// scheme, got nil")
	}
}

func TestAgentEndpoint_MissingHost(t *testing.T) {
	_, err := agentEndpoint("https://", "/path")
	if err == nil {
		t.Error("expected error when host is empty, got nil")
	}
}

// ---------------------------------------------------------------------------
// doHeartbeat
// ---------------------------------------------------------------------------

func TestDoHeartbeat_Success(t *testing.T) {
	var gotMethod, gotPath, gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := srv.Client()
	err := doHeartbeat(context.Background(), client, srv.URL, "test-token")
	if err != nil {
		t.Fatalf("doHeartbeat returned unexpected error: %v", err)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("method: got %q, want POST", gotMethod)
	}
	if gotPath != "/api/v1/scan-agent/heartbeat" {
		t.Errorf("path: got %q, want /api/v1/scan-agent/heartbeat", gotPath)
	}
	if gotAuth != "Bearer test-token" {
		t.Errorf("Authorization: got %q, want \"Bearer test-token\"", gotAuth)
	}
}

func TestDoHeartbeat_NonOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	client := srv.Client()
	err := doHeartbeat(context.Background(), client, srv.URL, "tok")
	if err == nil {
		t.Fatal("expected error for non-200 status, got nil")
	}
	if !strings.Contains(err.Error(), "503") {
		t.Errorf("error should mention status code 503, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// fetchJobs
// ---------------------------------------------------------------------------

func TestFetchJobs_ValidJSON(t *testing.T) {
	payload := `[
		{
			"id": 42,
			"name": "nightly",
			"subnets": [{"id": 7, "cidr": "10.0.1.0/24"}],
			"ping_concurrency": 5,
			"scan_type": "icmp"
		}
	]`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, payload)
	}))
	defer srv.Close()

	jobs, err := fetchJobs(context.Background(), srv.Client(), srv.URL, "tok")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}
	j := jobs[0]
	if j.ID != 42 {
		t.Errorf("job ID: got %d, want 42", j.ID)
	}
	if j.Name != "nightly" {
		t.Errorf("job Name: got %q, want \"nightly\"", j.Name)
	}
	if j.PingConcurrency != 5 {
		t.Errorf("ping_concurrency: got %d, want 5", j.PingConcurrency)
	}
	if j.ScanType != "icmp" {
		t.Errorf("scan_type: got %q, want \"icmp\"", j.ScanType)
	}
	if len(j.Subnets) != 1 || j.Subnets[0].CIDR != "10.0.1.0/24" || j.Subnets[0].ID != 7 {
		t.Errorf("subnets: got %+v", j.Subnets)
	}
}

func TestFetchJobs_EmptyArray(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `[]`)
	}))
	defer srv.Close()

	jobs, err := fetchJobs(context.Background(), srv.Client(), srv.URL, "tok")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(jobs) != 0 {
		t.Errorf("expected empty slice, got %d jobs", len(jobs))
	}
}

func TestFetchJobs_NonOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	_, err := fetchJobs(context.Background(), srv.Client(), srv.URL, "bad-tok")
	if err == nil {
		t.Fatal("expected error for non-200 status, got nil")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("error should mention 401, got: %v", err)
	}
}

func TestFetchJobs_MalformedJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{not valid json`)
	}))
	defer srv.Close()

	_, err := fetchJobs(context.Background(), srv.Client(), srv.URL, "tok")
	if err == nil {
		t.Fatal("expected JSON decode error, got nil")
	}
}

func TestFetchJobs_BearerHeader(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `[]`)
	}))
	defer srv.Close()

	_, _ = fetchJobs(context.Background(), srv.Client(), srv.URL, "my-secret")
	if gotAuth != "Bearer my-secret" {
		t.Errorf("Authorization header: got %q, want \"Bearer my-secret\"", gotAuth)
	}
}

// ---------------------------------------------------------------------------
// postResults
// ---------------------------------------------------------------------------

func TestPostResults_Payload(t *testing.T) {
	type requestBody struct {
		JobID   int64             `json:"job_id"`
		Results []AgentScanResult `json:"results"`
	}

	var got requestBody
	var gotContentType, gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotContentType = r.Header.Get("Content-Type")
		gotAuth = r.Header.Get("Authorization")
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			http.Error(w, "bad body", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	results := []AgentScanResult{
		{SubnetID: 3, IPAddress: "10.0.0.1", IsAlive: true, ResponseTimeMs: 12},
		{SubnetID: 3, IPAddress: "10.0.0.2", IsAlive: false},
	}
	err := postResults(context.Background(), srv.Client(), srv.URL, "post-token", 99, results)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotContentType != "application/json" {
		t.Errorf("Content-Type: got %q, want application/json", gotContentType)
	}
	if gotAuth != "Bearer post-token" {
		t.Errorf("Authorization: got %q, want \"Bearer post-token\"", gotAuth)
	}
	if got.JobID != 99 {
		t.Errorf("job_id: got %d, want 99", got.JobID)
	}
	if len(got.Results) != 2 {
		t.Fatalf("results count: got %d, want 2", len(got.Results))
	}
	if got.Results[0].IPAddress != "10.0.0.1" || !got.Results[0].IsAlive {
		t.Errorf("first result: %+v", got.Results[0])
	}
	if got.Results[1].IPAddress != "10.0.0.2" || got.Results[1].IsAlive {
		t.Errorf("second result: %+v", got.Results[1])
	}
}

func TestPostResults_NonOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	err := postResults(context.Background(), srv.Client(), srv.URL, "tok", 1, nil)
	if err == nil {
		t.Fatal("expected error for non-200 status, got nil")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("error should mention 500, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// runCycle — integration test using TEST-NET (192.0.2.0/30) so ping reports
// "not alive" but terminates quickly without hanging or touching real hosts.
// ---------------------------------------------------------------------------

func TestRunCycle_FullCycle(t *testing.T) {
	const token = "cycle-token"

	// Track which endpoints were hit so we can assert all three were called.
	var heartbeatCalled, jobsCalled, resultsCalled bool

	// Capture the results payload posted back.
	type resultPayload struct {
		JobID   int64             `json:"job_id"`
		Results []AgentScanResult `json:"results"`
	}
	var posted resultPayload

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the bearer token on every request.
		if r.Header.Get("Authorization") != "Bearer "+token {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		switch r.URL.Path {
		case "/api/v1/scan-agent/heartbeat":
			heartbeatCalled = true
			w.WriteHeader(http.StatusOK)

		case "/api/v1/scan-agent/jobs":
			jobsCalled = true
			// Return a single job with a /30 in TEST-NET — 2 host IPs, all dead.
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, `[{
				"id": 1,
				"name": "cycle-test",
				"subnets": [{"id": 10, "cidr": "192.0.2.0/30"}],
				"ping_concurrency": 2,
				"scan_type": "icmp"
			}]`)

		case "/api/v1/scan-agent/results":
			resultsCalled = true
			if err := json.NewDecoder(r.Body).Decode(&posted); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusOK)

		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	client := srv.Client()
	runCycle(context.Background(), client, srv.URL, token)

	if !heartbeatCalled {
		t.Error("heartbeat endpoint was not called")
	}
	if !jobsCalled {
		t.Error("jobs endpoint was not called")
	}
	if !resultsCalled {
		t.Error("results endpoint was not called")
	}

	// 192.0.2.0/30 has 2 host addresses: .1 and .2.
	if posted.JobID != 1 {
		t.Errorf("posted job_id: got %d, want 1", posted.JobID)
	}
	if len(posted.Results) != 2 {
		t.Errorf("posted results count: got %d, want 2", len(posted.Results))
	}
	for _, r := range posted.Results {
		if r.SubnetID != 10 {
			t.Errorf("result subnet_id: got %d, want 10", r.SubnetID)
		}
		// All TEST-NET IPs must be unreachable.
		if r.IsAlive {
			t.Errorf("TEST-NET IP %s reported alive (unexpected)", r.IPAddress)
		}
	}
}
