package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// buildPTR — pure function tests
// ---------------------------------------------------------------------------

func TestBuildPTR_ValidIPv4(t *testing.T) {
	zone, name := buildPTR("192.168.1.5")
	assert.Equal(t, "1.168.192.in-addr.arpa.", zone)
	assert.Equal(t, "5.1.168.192.in-addr.arpa.", name)
}

func TestBuildPTR_AnotherIPv4(t *testing.T) {
	zone, name := buildPTR("10.0.0.1")
	assert.Equal(t, "0.0.10.in-addr.arpa.", zone)
	assert.Equal(t, "1.0.0.10.in-addr.arpa.", name)
}

func TestBuildPTR_InvalidAddress_ReturnsEmpty(t *testing.T) {
	zone, name := buildPTR("not-an-ip")
	assert.Empty(t, zone)
	assert.Empty(t, name)
}

func TestBuildPTR_IPv6_ReturnsEmpty(t *testing.T) {
	zone, name := buildPTR("2001:db8::1")
	assert.Empty(t, zone)
	assert.Empty(t, name)
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
