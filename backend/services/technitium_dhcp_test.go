package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScopeToCIDR(t *testing.T) {
	cases := []struct {
		name          string
		startingAddr  string
		subnetMask    string
		wantNetwork   string
		wantPrefix    int
		wantErrSubstr string
	}{
		{
			name:         "/24 — network address equals starting address",
			startingAddr: "192.168.1.0",
			subnetMask:   "255.255.255.0",
			wantNetwork:  "192.168.1.0",
			wantPrefix:   24,
		},
		{
			name:         "/24 — starting address is a host within the range",
			startingAddr: "192.168.1.1",
			subnetMask:   "255.255.255.0",
			wantNetwork:  "192.168.1.0",
			wantPrefix:   24,
		},
		{
			name:         "/16",
			startingAddr: "10.10.0.1",
			subnetMask:   "255.255.0.0",
			wantNetwork:  "10.10.0.0",
			wantPrefix:   16,
		},
		{
			name:         "/8",
			startingAddr: "10.0.0.1",
			subnetMask:   "255.0.0.0",
			wantNetwork:  "10.0.0.0",
			wantPrefix:   8,
		},
		{
			name:         "/25",
			startingAddr: "192.168.0.128",
			subnetMask:   "255.255.255.128",
			wantNetwork:  "192.168.0.128",
			wantPrefix:   25,
		},
		{
			name:          "invalid starting address",
			startingAddr:  "not-an-ip",
			subnetMask:    "255.255.255.0",
			wantErrSubstr: "invalid starting address",
		},
		{
			name:          "invalid subnet mask",
			startingAddr:  "192.168.1.0",
			subnetMask:    "not-a-mask",
			wantErrSubstr: "invalid subnet mask",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			network, prefix, err := scopeToCIDR(tc.startingAddr, tc.subnetMask)
			if tc.wantErrSubstr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErrSubstr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantNetwork, network)
			assert.Equal(t, tc.wantPrefix, prefix)
		})
	}
}
