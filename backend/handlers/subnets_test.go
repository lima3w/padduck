package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateSubnetRequest_Validation(t *testing.T) {
	req := &CreateSubnetRequest{
		NetworkAddress: "192.168.0.0",
		PrefixLength:   24,
		Description:    "Test Subnet",
	}

	assert.NotEmpty(t, req.NetworkAddress)
	assert.Equal(t, "192.168.0.0", req.NetworkAddress)
	assert.Equal(t, 24, req.PrefixLength)
	assert.Equal(t, "Test Subnet", req.Description)
}

func TestUpdateSubnetRequest_Validation(t *testing.T) {
	req := &UpdateSubnetRequest{
		Description: "Updated Description",
	}

	assert.NotEmpty(t, req.Description)
	assert.Equal(t, "Updated Description", req.Description)
}
