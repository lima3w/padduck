package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateCIDR(t *testing.T) {
	tests := []struct {
		name          string
		address       string
		prefixLength  int
		wantErr       bool
		errorContains string
	}{
		// Valid cases
		{
			name:         "valid /24",
			address:      "192.168.0.0",
			prefixLength: 24,
			wantErr:      false,
		},
		{
			name:         "valid /8",
			address:      "10.0.0.0",
			prefixLength: 8,
			wantErr:      false,
		},
		{
			name:         "valid /0 edge case",
			address:      "0.0.0.0",
			prefixLength: 0,
			wantErr:      false,
		},
		{
			name:         "valid /32 edge case",
			address:      "255.255.255.255",
			prefixLength: 32,
			wantErr:      false,
		},
		// Invalid prefix length
		{
			name:          "invalid prefix length -1",
			address:       "192.168.0.0",
			prefixLength:  -1,
			wantErr:       true,
			errorContains: "invalid prefix length",
		},
		{
			name:          "invalid prefix length 33",
			address:       "192.168.0.0",
			prefixLength:  33,
			wantErr:       true,
			errorContains: "invalid prefix length",
		},
		// Invalid IP address
		{
			name:          "invalid IP 999.999.999.999",
			address:       "999.999.999.999",
			prefixLength:  24,
			wantErr:       true,
			errorContains: "invalid network address",
		},
		{
			name:          "invalid IP not-an-ip",
			address:       "not-an-ip",
			prefixLength:  24,
			wantErr:       true,
			errorContains: "invalid network address",
		},
		{
			name:          "invalid IP empty string",
			address:       "",
			prefixLength:  24,
			wantErr:       true,
			errorContains: "invalid network address",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCIDR(tt.address, tt.prefixLength)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreateSubnet_Validation(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	tests := []struct {
		name           string
		sectionID      int64
		networkAddress string
		prefixLength   int
		wantErr        bool
		errorContains  string
	}{
		{
			name:           "sectionID zero",
			sectionID:      0,
			networkAddress: "192.168.0.0",
			prefixLength:   24,
			wantErr:        true,
			errorContains:  "invalid section ID",
		},
		{
			name:           "sectionID negative",
			sectionID:      -1,
			networkAddress: "192.168.0.0",
			prefixLength:   24,
			wantErr:        true,
			errorContains:  "invalid section ID",
		},
		{
			name:           "invalid network address",
			sectionID:      1,
			networkAddress: "not-an-ip",
			prefixLength:   24,
			wantErr:        true,
			errorContains:  "invalid network address",
		},
		{
			name:           "invalid prefix length negative",
			sectionID:      1,
			networkAddress: "192.168.0.0",
			prefixLength:   -1,
			wantErr:        true,
			errorContains:  "invalid prefix length",
		},
		{
			name:           "invalid prefix length too large",
			sectionID:      1,
			networkAddress: "192.168.0.0",
			prefixLength:   129,
			wantErr:        true,
			errorContains:  "invalid prefix length",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.CreateSubnet(ctx, tt.sectionID, tt.networkAddress, tt.prefixLength, "", nil, false, false, nil, nil, nil)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetSubnet_InvalidID(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	tests := []struct {
		name string
		id   int64
	}{
		{"zero ID", 0},
		{"negative ID", -1},
		{"large negative", -100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.GetSubnet(ctx, tt.id)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid subnet ID")
		})
	}
}

func TestListSubnets_InvalidSectionID(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	tests := []struct {
		name      string
		sectionID int64
	}{
		{"zero sectionID", 0},
		{"negative sectionID", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.ListSubnets(ctx, tt.sectionID)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid section ID")
		})
	}
}

func TestUpdateSubnet_InvalidID(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	tests := []struct {
		name string
		id   int64
	}{
		{"zero ID", 0},
		{"negative ID", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.UpdateSubnet(ctx, tt.id, "new description", nil, false, false, nil, nil, nil)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid subnet ID")
		})
	}
}

func TestDeleteSubnet_InvalidID(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	tests := []struct {
		name string
		id   int64
	}{
		{"zero ID", 0},
		{"negative ID", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.DeleteSubnet(ctx, tt.id)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid subnet ID")
		})
	}
}
