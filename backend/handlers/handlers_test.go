package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"ipam-next/services"
)

func TestNewHandler(t *testing.T) {
	repo := nil // Would be a mock in real tests
	svc := services.NewService(repo)
	handler := NewHandler(svc)

	assert.NotNil(t, handler)
	assert.Equal(t, svc, handler.service)
}
