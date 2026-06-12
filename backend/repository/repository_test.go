package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRepository(t *testing.T) {
	repo := NewRepository(nil)
	assert.NotNil(t, repo)
}

// DB-backed integration tests live in *_integration_test.go and run when
// TEST_DATABASE_URL points at a Postgres instance (see testdb_test.go).
