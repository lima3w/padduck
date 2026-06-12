package main

import (
	"context"
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
