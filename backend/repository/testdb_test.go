package repository

// DB-backed integration test harness. Tests that need a real Postgres call
// testRepo(t), which skips when TEST_DATABASE_URL is unset so the suite stays
// green for offline unit-test runs. CI provides a Postgres service; locally,
// `make test-integration` boots a throwaway container.

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"padduck/internal/testdb"
)

// testRepo returns a Repository backed by a scratch database with all
// migrations applied and data tables truncated for isolation. Tests using
// it must not call t.Parallel(): they share one database.
func testRepo(t *testing.T) *Repository {
	t.Helper()
	pool := testdb.Connect(t, "repository")
	// Lookup/seed tables (device_types) are intentionally not truncated.
	testdb.Truncate(t, pool,
		"custom_field_values", "custom_field_definitions",
		"device_interfaces", "devices",
		"ip_addresses", "subnets", "networks", "users")
	return NewRepository(pool)
}

// --- shared fixtures ---

func createTestUser(t *testing.T, r *Repository) int64 {
	t.Helper()
	u, err := r.CreateUser(context.Background(), "tester", "tester@example.com")
	require.NoError(t, err)
	return u.ID
}

func createTestNetwork(t *testing.T, r *Repository) int64 {
	t.Helper()
	n, err := r.CreateNetwork(context.Background(), "test-network", "integration fixture", createTestUser(t, r))
	require.NoError(t, err)
	return n.ID
}

func createTestSubnet(t *testing.T, r *Repository, networkID int64, cidrBase string, prefix int) int64 {
	t.Helper()
	s, err := r.CreateSubnet(context.Background(), networkID, cidrBase, prefix, "integration fixture", nil, false, false)
	require.NoError(t, err)
	return s.ID
}

func strPtr(s string) *string { return &s }
