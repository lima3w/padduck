package main

import (
	"context"
	"testing"
)

func TestEnumerateCIDR(t *testing.T) {
	tests := []struct {
		cidr    string
		wantLen int
		wantErr bool
	}{
		{"192.168.1.0/30", 2, false},  // .1 and .2 (excludes .0 and .3)
		{"10.0.0.0/29", 6, false},     // .1–.6
		{"10.0.0.0/24", 254, false},
		{"invalid", 0, true},
	}

	for _, tt := range tests {
		ips, err := enumerateCIDR(tt.cidr)
		if tt.wantErr {
			if err == nil {
				t.Errorf("enumerateCIDR(%q): expected error, got none", tt.cidr)
			}
			continue
		}
		if err != nil {
			t.Errorf("enumerateCIDR(%q): unexpected error: %v", tt.cidr, err)
			continue
		}
		if len(ips) != tt.wantLen {
			t.Errorf("enumerateCIDR(%q): got %d IPs, want %d", tt.cidr, len(ips), tt.wantLen)
		}
	}
}

func TestEnumerateCIDRExcludesNetworkAndBroadcast(t *testing.T) {
	ips, err := enumerateCIDR("192.168.1.0/24")
	if err != nil {
		t.Fatal(err)
	}
	for _, ip := range ips {
		if ip == "192.168.1.0" || ip == "192.168.1.255" {
			t.Errorf("enumerateCIDR: included network/broadcast address %q", ip)
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
