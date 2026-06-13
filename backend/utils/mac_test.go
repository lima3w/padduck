package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeMAC(t *testing.T) {
	tests := []struct {
		input   string
		want    string
		wantErr bool
	}{
		// colon-separated
		{"aa:bb:cc:dd:ee:ff", "aa:bb:cc:dd:ee:ff", false},
		{"AA:BB:CC:DD:EE:FF", "aa:bb:cc:dd:ee:ff", false},
		// dash-separated
		{"aa-bb-cc-dd-ee-ff", "aa:bb:cc:dd:ee:ff", false},
		{"AA-BB-CC-DD-EE-FF", "aa:bb:cc:dd:ee:ff", false},
		// Cisco dot-separated
		{"aabb.ccdd.eeff", "aa:bb:cc:dd:ee:ff", false},
		{"AABB.CCDD.EEFF", "aa:bb:cc:dd:ee:ff", false},
		// no separator
		{"aabbccddeeff", "aa:bb:cc:dd:ee:ff", false},
		{"AABBCCDDEEFF", "aa:bb:cc:dd:ee:ff", false},
		// whitespace stripped
		{"  aa:bb:cc:dd:ee:ff  ", "aa:bb:cc:dd:ee:ff", false},
		// empty — allowed (no MAC set)
		{"", "", false},
		// invalid
		{"aa:bb:cc:dd:ee", "", true},
		{"zz:bb:cc:dd:ee:ff", "", true},
		{"aa:bb:cc:dd:ee:ff:00", "", true},
		{"not-a-mac", "", true},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got, err := NormalizeMAC(tc.input)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.want, got)
			}
		})
	}
}
