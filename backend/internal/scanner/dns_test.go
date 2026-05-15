package scanner

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestResolveHostname_SubMillisecondTimeout_ReturnsEmpty(t *testing.T) {
	// An impossibly short timeout must not block or panic.
	result := ResolveHostname("1.2.3.4", 1*time.Nanosecond)
	assert.Empty(t, result.PTR)
	assert.False(t, result.FwdRevMismatch)
}

func TestResolveHostname_TestNetIP_ReturnsEmpty(t *testing.T) {
	// RFC 5737 TEST-NET addresses have no PTR records.
	result := ResolveHostname("192.0.2.1", 2*time.Second)
	assert.Empty(t, result.PTR)
	assert.False(t, result.FwdRevMismatch)
}

func TestResolveHostname_InvalidIP_ReturnsEmpty(t *testing.T) {
	result := ResolveHostname("not-an-ip", 2*time.Second)
	assert.Empty(t, result.PTR)
}

func TestDNSResult_ZeroValue(t *testing.T) {
	var r DNSResult
	assert.Empty(t, r.PTR)
	assert.False(t, r.FwdRevMismatch)
}
