package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateSectionRequest_Validation(t *testing.T) {
	req := &CreateSectionRequest{
		Name:        "Test Section",
		Description: "Test Description",
		CreatedBy:   1,
	}

	assert.NotEmpty(t, req.Name)
	assert.Equal(t, "Test Section", req.Name)
	assert.Equal(t, int64(1), req.CreatedBy)
}

func TestUpdateSectionRequest_Validation(t *testing.T) {
	req := &UpdateSectionRequest{
		Name:        "Updated Section",
		Description: "Updated Description",
	}

	assert.NotEmpty(t, req.Name)
	assert.Equal(t, "Updated Section", req.Name)
}
