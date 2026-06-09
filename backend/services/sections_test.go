package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateSection_EmptyName(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	_, err := svc.CreateNetwork(ctx, "", "some description", 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "section name is required")
}

func TestCreateSection_ValidName_PassesValidation(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	// A non-empty name passes validation and panics at the nil repo call,
	// which means the validation guard was passed successfully.
	assert.Panics(t, func() {
		_, _ = svc.CreateNetwork(ctx, "My Network", "description", 1)
	})
}

func TestGetSection_InvalidID(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	tests := []struct {
		name string
		id   int64
	}{
		{"zero", 0},
		{"negative five", -5},
		{"negative one", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.GetNetwork(ctx, tt.id)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid section ID")
		})
	}
}

func TestUpdateSection_InvalidID(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	tests := []struct {
		name string
		id   int64
	}{
		{"zero", 0},
		{"negative", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.UpdateNetwork(ctx, tt.id, "Some Name", "desc")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid section ID")
		})
	}
}

func TestUpdateSection_EmptyName(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	// Valid id but empty name should return name-required error before hitting repo
	_, err := svc.UpdateNetwork(ctx, 1, "", "desc")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "section name is required")
}

func TestDeleteSection_InvalidID(t *testing.T) {
	svc := NewService(nil, "0000000000000000000000000000000000000000000000000000000000000000")
	ctx := context.Background()

	tests := []struct {
		name string
		id   int64
	}{
		{"zero", 0},
		{"negative one", -1},
		{"large negative", -100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.DeleteNetwork(ctx, tt.id)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid section ID")
		})
	}
}
