package repository

// DB-backed integration test harness. Tests that need a real Postgres call
// testRepo(t), which skips when TEST_DATABASE_URL is unset so the suite stays
// green for offline unit-test runs. CI provides a Postgres service; locally,
// `make test-integration` boots a throwaway container.

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/stretchr/testify/require"
)

var (
	testPoolOnce sync.Once
	testPool     *pgxpool.Pool
	testPoolErr  error
)

// testRepo returns a Repository backed by the TEST_DATABASE_URL database with
// all migrations applied and data tables truncated for isolation. Tests using
// it must not call t.Parallel(): they share one database.
func testRepo(t *testing.T) *Repository {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping DB integration test")
	}

	testPoolOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		testPool, testPoolErr = pgxpool.New(ctx, dsn)
		if testPoolErr != nil {
			return
		}
		if testPoolErr = testPool.Ping(ctx); testPoolErr != nil {
			return
		}

		sqlDB := sql.OpenDB(stdlib.GetConnector(*testPool.Config().ConnConfig))
		defer sqlDB.Close()
		_, testPoolErr = migrate.Exec(sqlDB, "postgres", &migrate.FileMigrationSource{Dir: "../migrations"}, migrate.Up)
		if testPoolErr != nil {
			testPoolErr = fmt.Errorf("applying migrations: %w", testPoolErr)
		}
	})
	require.NoError(t, testPoolErr, "test database setup failed")

	cleanTestTables(t, testPool)
	return NewRepository(testPool)
}

// cleanTestTables truncates the tables integration tests write to. Extend the
// list as new tests need more tables; lookup/seed tables (device_types) are
// intentionally not truncated.
func cleanTestTables(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_, err := pool.Exec(ctx, `
		TRUNCATE custom_field_values, custom_field_definitions,
		         device_interfaces, devices,
		         ip_addresses, subnets, networks, users
		RESTART IDENTITY CASCADE`)
	require.NoError(t, err, "truncating test tables")
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
