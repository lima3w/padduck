package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"ipam-next/repository"
)

func TestNewService(t *testing.T) {
	repo := repository.NewRepository(nil) // nil pool for testing structure
	svc := NewService(repo)

	assert.NotNil(t, svc)
	assert.Equal(t, repo, svc.GetRepository())
}
