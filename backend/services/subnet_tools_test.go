package services

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"ipam-next/models"
)

// ---------------------------------------------------------------------------
// cloneIP
// ---------------------------------------------------------------------------

func TestCloneIP_DoesNotMutateOriginal(t *testing.T) {
	ip := net.IP{10, 0, 0, 0}
	cloned := cloneIP(ip)
	assert.Equal(t, ip, cloned)
	cloned[3] = 255
	assert.Equal(t, byte(0), ip[3], "original should not be mutated by clone")
}

// ---------------------------------------------------------------------------
// incrementIP
// ---------------------------------------------------------------------------

func TestIncrementIP(t *testing.T) {
	tests := []struct {
		name     string
		ip       net.IP
		hostBits int
		expected net.IP
	}{
		{
			name:     "increment by 1 (hostBits=0)",
			ip:       net.IP{10, 0, 0, 0},
			hostBits: 0,
			expected: net.IP{10, 0, 0, 1},
		},
		{
			name:     "increment by 256 (hostBits=8, /24 block)",
			ip:       net.IP{10, 0, 0, 0},
			hostBits: 8,
			expected: net.IP{10, 0, 1, 0},
		},
		{
			name:     "increment by 65536 (hostBits=16, /16 block)",
			ip:       net.IP{10, 0, 0, 0},
			hostBits: 16,
			expected: net.IP{10, 1, 0, 0},
		},
		{
			name:     "increment /25 block (hostBits=7 -> +128)",
			ip:       net.IP{192, 168, 1, 0},
			hostBits: 7,
			expected: net.IP{192, 168, 1, 128},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := cloneIP(tt.ip)
			incrementIP(ip, tt.hostBits)
			assert.Equal(t, tt.expected, ip)
		})
	}
}

// ---------------------------------------------------------------------------
// ipLess
// ---------------------------------------------------------------------------

func TestIPLess(t *testing.T) {
	tests := []struct {
		name     string
		a, b     net.IP
		expected bool
	}{
		{"10.0.0.0 < 10.0.0.1", net.IP{10, 0, 0, 0}, net.IP{10, 0, 0, 1}, true},
		{"10.0.0.1 > 10.0.0.0", net.IP{10, 0, 0, 1}, net.IP{10, 0, 0, 0}, false},
		{"equal", net.IP{10, 0, 0, 0}, net.IP{10, 0, 0, 0}, false},
		{"192.168.0.0 < 192.168.1.0", net.IP{192, 168, 0, 0}, net.IP{192, 168, 1, 0}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, ipLess(tt.a, tt.b))
		})
	}
}

// ---------------------------------------------------------------------------
// SplitSubnet enumeration logic (child CIDR generation)
// ---------------------------------------------------------------------------

func TestSplitSubnet_Enumeration_ChildCIDRs(t *testing.T) {
	// 192.168.1.0/24 → /26 gives 4 blocks
	startIP := net.IP{192, 168, 1, 0}
	hostBits := 32 - 26 // = 6 (each block is 64 addresses)
	expected := []net.IP{
		{192, 168, 1, 0},
		{192, 168, 1, 64},
		{192, 168, 1, 128},
		{192, 168, 1, 192},
	}

	ip := cloneIP(startIP)
	for i, exp := range expected {
		assert.Equal(t, exp, ip, "block %d mismatch", i)
		if i < len(expected)-1 {
			incrementIP(ip, hostBits)
		}
	}
}

// ---------------------------------------------------------------------------
// MergeSubnets power-of-2 validation
// ---------------------------------------------------------------------------

func TestMergeSubnets_PowerOf2Check(t *testing.T) {
	isPow2 := func(n int) bool {
		return n > 0 && (n&(n-1)) == 0
	}

	assert.True(t, isPow2(2), "2 should be valid")
	assert.True(t, isPow2(4), "4 should be valid")
	assert.True(t, isPow2(8), "8 should be valid")
	assert.False(t, isPow2(3), "3 should not be valid")
	assert.False(t, isPow2(5), "5 should not be valid")
	assert.False(t, isPow2(0), "0 should not be valid")
}

// ---------------------------------------------------------------------------
// SubnetResizeConflictError
// ---------------------------------------------------------------------------

func TestSubnetResizeConflictError_NoConflict(t *testing.T) {
	err := &SubnetResizeConflictError{}
	assert.Contains(t, err.Error(), "conflict")
}

func TestSubnetResizeConflictError_WithConflictingIPs(t *testing.T) {
	err := &SubnetResizeConflictError{
		ConflictingIPs: []*models.IPAddress{
			{Address: "192.168.1.200"},
			{Address: "192.168.1.201"},
		},
	}
	assert.Contains(t, err.Error(), "2 IP address")
}

func TestSubnetResizeConflictError_WithConflictingSubnets(t *testing.T) {
	err := &SubnetResizeConflictError{
		ConflictingSubnets: []*models.Subnet{
			{NetworkAddress: "10.0.0.0", PrefixLength: 24},
		},
	}
	assert.Contains(t, err.Error(), "1 existing subnet")
}

// ---------------------------------------------------------------------------
// isExpiredNow
// ---------------------------------------------------------------------------

func TestIsExpiredNow_NilExpiresAt(t *testing.T) {
	assert.False(t, isExpiredNow(nil))
}

func TestIsExpiredNow_FutureExpiresAt(t *testing.T) {
	future := time.Now().Add(1 * time.Hour)
	assert.False(t, isExpiredNow(&future))
}

func TestIsExpiredNow_PastExpiresAt(t *testing.T) {
	past := time.Now().Add(-1 * time.Hour)
	assert.True(t, isExpiredNow(&past))
}
