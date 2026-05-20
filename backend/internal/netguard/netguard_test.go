package netguard

import (
	"context"
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsSafeOutboundAddr(t *testing.T) {
	tests := []struct {
		name string
		addr string
		want bool
	}{
		{name: "public IPv4", addr: "8.8.8.8", want: true},
		{name: "loopback", addr: "127.0.0.1", want: false},
		{name: "private", addr: "10.0.0.1", want: false},
		{name: "link local", addr: "169.254.1.1", want: false},
		{name: "unspecified", addr: "0.0.0.0", want: false},
		{name: "public IPv6", addr: "2001:4860:4860::8888", want: true},
		{name: "unique local IPv6", addr: "fd00::1", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsSafeOutboundAddr(netip.MustParseAddr(tt.addr)))
		})
	}
}

func TestValidateURLRejectsUnsafeTargets(t *testing.T) {
	tests := []string{
		"http://127.0.0.1/hook",
		"http://10.0.0.1/hook",
		"http://[::1]/hook",
		"ftp://8.8.8.8/hook",
		"https://user:pass@8.8.8.8/hook",
	}
	for _, rawURL := range tests {
		t.Run(rawURL, func(t *testing.T) {
			assert.Error(t, ValidateURL(context.Background(), rawURL))
		})
	}
}

func TestValidateURLAllowsPublicIPLiteral(t *testing.T) {
	assert.NoError(t, ValidateURL(context.Background(), "https://8.8.8.8/hook"))
}
