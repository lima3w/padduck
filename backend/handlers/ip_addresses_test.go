package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateIPAddressRequest_Validation(t *testing.T) {
	req := &CreateIPAddressRequest{
		Address:  "192.168.1.10",
		Hostname: "server1.example.com",
		Status:   "available",
	}

	assert.NotEmpty(t, req.Address)
	assert.Equal(t, "192.168.1.10", req.Address)
	assert.Equal(t, "server1.example.com", req.Hostname)
	assert.Equal(t, "available", req.Status)
}

func TestAssignIPAddressRequest_Validation(t *testing.T) {
	req := &AssignIPAddressRequest{
		AssignedTo: "server1",
	}

	assert.NotEmpty(t, req.AssignedTo)
	assert.Equal(t, "server1", req.AssignedTo)
}

func TestAllocateIPAddressRequest_Validation(t *testing.T) {
	req := &struct {
		AssignedTo string `json:"assigned_to"`
	}{
		AssignedTo: "server1",
	}

	assert.NotEmpty(t, req.AssignedTo)
	assert.Equal(t, "server1", req.AssignedTo)
}

func TestSubnetUtilization_Validation(t *testing.T) {
	util := &struct {
		Total       int64   `json:"total"`
		Available   int64   `json:"available"`
		Assigned    int64   `json:"assigned"`
		Reserved    int64   `json:"reserved"`
		Utilization float64 `json:"utilization_percent"`
	}{
		Total:       100,
		Available:   75,
		Assigned:    20,
		Reserved:    5,
		Utilization: 25.0,
	}

	assert.Equal(t, int64(100), util.Total)
	assert.Equal(t, int64(75), util.Available)
	assert.Equal(t, int64(20), util.Assigned)
	assert.Equal(t, int64(5), util.Reserved)
	assert.Equal(t, 25.0, util.Utilization)
}
