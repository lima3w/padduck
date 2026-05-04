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
