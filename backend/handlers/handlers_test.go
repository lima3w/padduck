package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"ipam-next/services"
)

func TestNewHandler(t *testing.T) {
	var svc *services.Service
	handler := NewHandler(svc)

	assert.NotNil(t, handler)
	assert.Equal(t, svc, handler.service)
}
