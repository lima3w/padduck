package db

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConnect(t *testing.T) {
	// Speed up retries for tests.
	origAttempts, origBase, origMax := connectMaxAttempts, connectBaseDelay, connectMaxDelay
	connectMaxAttempts = 2
	connectBaseDelay = 1 * time.Millisecond
	connectMaxDelay = 2 * time.Millisecond
	t.Cleanup(func() {
		connectMaxAttempts = origAttempts
		connectBaseDelay = origBase
		connectMaxDelay = origMax
	})

	ctx := context.Background()

	_, err := Connect(ctx, "invalid://connection")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse connection string")

	_, err = Connect(ctx, "postgres://user:pass@localhost:9999/db")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect to database after")
}

func TestConnectSurfacesHintForUnescapedUserinfo(t *testing.T) {
	ctx := context.Background()

	// '<' is not a valid userinfo character per net/url and isn't otherwise
	// a URL delimiter, so this reliably fails with "invalid userinfo" during
	// ParseConfig rather than falling through to a real connection attempt.
	_, err := Connect(ctx, "postgres://padduck:p<ss@db:5432/padduck")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse connection string")
	assert.Contains(t, err.Error(), "percent-encode it")
}

func TestDB_Close(t *testing.T) {
	// Close should not panic even if pool is nil
	db := &DB{pool: nil}
	assert.NotPanics(t, func() {
		db.Close()
	})
}
