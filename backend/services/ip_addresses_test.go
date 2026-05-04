package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateIPAddress(t *testing.T) {
	svc := NewService(nil)
	ctx := context.Background()

	testCases := []struct {
		name      string
		subnetID  int64
		address   string
		status    string
		shouldErr bool
		errMsg    string
	}{
		{
			name:      "invalid subnet ID",
			subnetID:  0,
			address:   "192.168.0.10",
			status:    "available",
			shouldErr: true,
			errMsg:    "invalid subnet ID",
		},
		{
			name:      "invalid IP address",
			subnetID:  1,
			address:   "999.999.999.999",
			status:    "available",
			shouldErr: true,
			errMsg:    "invalid IP address",
		},
		{
			name:      "invalid status",
			subnetID:  1,
			address:   "192.168.0.10",
			status:    "invalid",
			shouldErr: true,
			errMsg:    "invalid IP status",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.CreateIPAddress(ctx, tc.subnetID, tc.address, "", tc.status)
			if tc.shouldErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestReleaseIPAddress(t *testing.T) {
	svc := NewService(nil)
	ctx := context.Background()

	// Test validation - invalid ID
	_, err := svc.ReleaseIPAddress(ctx, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid IP address ID")
}
