package services

import "testing"

func TestPolicyMatches(t *testing.T) {
	tests := []struct {
		name       string
		condition  map[string]string
		values     map[string]string
		shouldPass bool
	}{
		{
			name:       "empty conditions allow",
			condition:  map[string]string{},
			values:     map[string]string{"hostname": "edge-01"},
			shouldPass: true,
		},
		{
			name:       "exact match",
			condition:  map[string]string{"subnet_id": "42"},
			values:     map[string]string{"subnet_id": "42"},
			shouldPass: true,
		},
		{
			name:       "prefix match",
			condition:  map[string]string{"hostname": "prod-*"},
			values:     map[string]string{"hostname": "prod-router-01"},
			shouldPass: true,
		},
		{
			name:       "missing value denies",
			condition:  map[string]string{"hostname": "prod-*"},
			values:     map[string]string{"subnet_id": "42"},
			shouldPass: false,
		},
		{
			name:       "different exact value denies",
			condition:  map[string]string{"subnet_id": "42"},
			values:     map[string]string{"subnet_id": "43"},
			shouldPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := policyMatches(tt.condition, tt.values); got != tt.shouldPass {
				t.Fatalf("policyMatches() = %v, want %v", got, tt.shouldPass)
			}
		})
	}
}
