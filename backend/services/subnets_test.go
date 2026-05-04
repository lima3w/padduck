package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateCIDR(t *testing.T) {
	testCases := []struct {
		name          string
		address       string
		prefixLength  int
		shouldErr     bool
		errorContains string
	}{
		{
			name:         "valid CIDR",
			address:      "192.168.0.0",
			prefixLength: 24,
			shouldErr:    false,
		},
		{
			name:          "invalid prefix length - negative",
			address:       "192.168.0.0",
			prefixLength:  -1,
			shouldErr:     true,
			errorContains: "invalid prefix length",
		},
		{
			name:          "invalid prefix length - too large",
			address:       "192.168.0.0",
			prefixLength:  33,
			shouldErr:     true,
			errorContains: "invalid prefix length",
		},
		{
			name:          "invalid IP address",
			address:       "999.999.999.999",
			prefixLength:  24,
			shouldErr:     true,
			errorContains: "invalid network address",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateCIDR(tc.address, tc.prefixLength)
			if tc.shouldErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
