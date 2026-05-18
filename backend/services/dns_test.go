package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"ipam-next/models"
)

// ---------------------------------------------------------------------------
// buildPTR — pure function tests
// ---------------------------------------------------------------------------

func TestBuildPTR_ValidIPv4(t *testing.T) {
	zone, name := buildPTR("192.168.1.5", 0)
	assert.Equal(t, "1.168.192.in-addr.arpa.", zone)
	assert.Equal(t, "5.1.168.192.in-addr.arpa.", name)
}

func TestBuildPTR_AnotherIPv4(t *testing.T) {
	zone, name := buildPTR("10.0.0.1", 0)
	assert.Equal(t, "0.0.10.in-addr.arpa.", zone)
	assert.Equal(t, "1.0.0.10.in-addr.arpa.", name)
}

func TestBuildPTR_InvalidAddress_ReturnsEmpty(t *testing.T) {
	zone, name := buildPTR("not-an-ip", 0)
	assert.Empty(t, zone)
	assert.Empty(t, name)
}

func TestBuildPTR_IPv6_Prefix48(t *testing.T) {
	// 2001:db8::1 with /48 → 12-nibble zone, full 32-nibble name
	zone, name := buildPTR("2001:db8::1", 48)
	assert.Equal(t, "0.0.0.0.8.b.d.0.1.0.0.2.ip6.arpa.", zone)
	assert.Equal(t, "1.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.8.b.d.0.1.0.0.2.ip6.arpa.", name)
}

func TestBuildPTR_IPv6_Prefix64(t *testing.T) {
	// 2001:db8::1 with /64 → 16-nibble zone
	zone, name := buildPTR("2001:db8::1", 64)
	assert.Equal(t, "0.0.0.0.0.0.0.0.8.b.d.0.1.0.0.2.ip6.arpa.", zone)
	assert.Equal(t, "1.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.8.b.d.0.1.0.0.2.ip6.arpa.", name)
}

func TestBuildPTR_IPv6_PrefixZero_ClampedToOneNibble(t *testing.T) {
	// prefixLen=0 should clamp to 1-nibble zone minimum
	zone, name := buildPTR("2001:db8::1", 0)
	assert.Equal(t, "2.ip6.arpa.", zone)
	assert.NotEmpty(t, name)
}

// ---------------------------------------------------------------------------
// containsZone — pure function tests
// ---------------------------------------------------------------------------

func TestContainsZone_Found(t *testing.T) {
	assert.True(t, containsZone("1.168.192.in-addr.arpa., 0.0.10.in-addr.arpa.", "0.0.10.in-addr.arpa."))
}

func TestContainsZone_NotFound(t *testing.T) {
	assert.False(t, containsZone("1.168.192.in-addr.arpa.", "0.0.10.in-addr.arpa."))
}

func TestContainsZone_EmptyList(t *testing.T) {
	assert.False(t, containsZone("", "0.0.10.in-addr.arpa."))
}

// ---------------------------------------------------------------------------
// NewDNSService — smoke test (no DB/config needed for pure functions)
// ---------------------------------------------------------------------------

func TestNewDNSService_NotNil(t *testing.T) {
	svc := &Service{}
	dns := NewDNSService(svc)
	assert.NotNil(t, dns)
}

// ---------------------------------------------------------------------------
// errString helper
// ---------------------------------------------------------------------------

func TestErrString_Nil(t *testing.T) {
	assert.Equal(t, "", errString(nil))
}

func TestErrString_NonNil(t *testing.T) {
	assert.Equal(t, "boom", errString(errTestSentinel))
}

type testError struct{ msg string }

func (e *testError) Error() string { return e.msg }

var errTestSentinel = &testError{msg: "boom"}

// ---------------------------------------------------------------------------
// ListDNSZones / GetDNSZoneRecords — no provider configured
// ---------------------------------------------------------------------------

func TestListDNSZones_NoneConfigured_ReturnsNotConfigured(t *testing.T) {
	dns := NewDNSService(&Service{Config: NewConfigService(nil)})
	zones, configured, err := dns.ListDNSZones(context.Background())
	assert.NoError(t, err)
	assert.False(t, configured)
	assert.Nil(t, zones)
}

func TestGetDNSZoneRecords_NoneConfigured_ReturnsError(t *testing.T) {
	dns := NewDNSService(&Service{Config: NewConfigService(nil)})
	_, err := dns.GetDNSZoneRecords(context.Background(), "example.com", "")
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// SyncIPToDNS / RemoveIPFromDNS — no-op when no provider configured
// ---------------------------------------------------------------------------

func TestSyncIPToDNS_NoneConfigured_NoOp(t *testing.T) {
	dns := NewDNSService(&Service{Config: NewConfigService(nil)})
	dnsName := "host.example.com"
	ip := &models.IPAddress{ID: 1, Address: "192.168.1.1", DNSName: &dnsName}
	assert.NotPanics(t, func() {
		dns.SyncIPToDNS(context.Background(), ip)
	})
}

func TestRemoveIPFromDNS_NoneConfigured_NoOp(t *testing.T) {
	dns := NewDNSService(&Service{Config: NewConfigService(nil)})
	dnsName := "host.example.com"
	ip := &models.IPAddress{ID: 1, Address: "192.168.1.1", DNSName: &dnsName}
	assert.NotPanics(t, func() {
		dns.RemoveIPFromDNS(context.Background(), ip)
	})
}

func TestSyncIPToTechnitium_NilDNSName_NoOp(t *testing.T) {
	dns := NewDNSService(&Service{Config: NewConfigService(nil)})
	ip := &models.IPAddress{ID: 1, Address: "192.168.1.1", DNSName: nil}
	assert.NotPanics(t, func() {
		dns.SyncIPToTechnitium(context.Background(), ip)
	})
}

func TestRemoveIPFromTechnitium_NilDNSName_NoOp(t *testing.T) {
	dns := NewDNSService(&Service{Config: NewConfigService(nil)})
	ip := &models.IPAddress{ID: 1, Address: "192.168.1.1", DNSName: nil}
	assert.NotPanics(t, func() {
		dns.RemoveIPFromTechnitium(context.Background(), ip)
	})
}
