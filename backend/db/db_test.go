package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConnect(t *testing.T) {
	// Test with invalid connection string
	ctx := context.Background()
	_, err := Connect(ctx, "invalid://connection")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse connection string")

	// Valid format but unreachable database
	_, err = Connect(ctx, "postgres://user:pass@localhost:9999/db")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to ping database")
}

func TestDB_Close(t *testing.T) {
	// Close should not panic even if pool is nil
	db := &DB{pool: nil}
	assert.NotPanics(t, func() {
		db.Close()
	})
}
