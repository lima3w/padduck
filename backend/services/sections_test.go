package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateSection(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	// Test validation - missing name
	_, err := svc.CreateSection(ctx, "", "description", 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")
}

func TestGetSection(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	// Test validation - invalid ID
	_, err := svc.GetSection(ctx, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid section ID")
}

func TestUpdateSection(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	// Test validation - empty name
	_, err := svc.UpdateSection(ctx, 1, "", "description")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")
}

func TestDeleteSection(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	// Test validation - invalid ID
	err := svc.DeleteSection(ctx, -1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid section ID")
}
