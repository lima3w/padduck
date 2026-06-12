// Package testdb provides a shared harness for DB-backed integration tests.
// Each test package gets its own scratch database on the TEST_DATABASE_URL
// server, so packages running in parallel under `go test ./...` cannot
// interfere with each other's data or truncation.
package testdb

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	migrate "github.com/rubenv/sql-migrate"
)

// MigrationsDir is resolved relative to the test package's directory.
// Both repository and services tests sit one level below backend/.
const MigrationsDir = "../migrations"

var (
	mu    sync.Mutex
	pools = map[string]*pgxpool.Pool{}

	validName = regexp.MustCompile(`^[a-z_]+$`)
)

// Connect returns a pool to the scratch database padduck_test_<name>,
// created fresh with all migrations applied on first use in the process.
// The test is skipped when TEST_DATABASE_URL is unset. name must be a
// short lowercase identifier, normally the test package name.
func Connect(t *testing.T, name string) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping DB integration test")
	}
	if !validName.MatchString(name) {
		t.Fatalf("testdb.Connect: invalid scratch database name %q", name)
	}

	mu.Lock()
	defer mu.Unlock()
	if pool, ok := pools[name]; ok {
		return pool
	}

	pool, err := createScratch(dsn, "padduck_test_"+name)
	if err != nil {
		t.Fatalf("testdb: %v", err)
	}
	pools[name] = pool
	return pool
}

func createScratch(adminDSN, dbName string) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	u, err := url.Parse(adminDSN)
	if err != nil {
		return nil, fmt.Errorf("parsing TEST_DATABASE_URL: %w", err)
	}

	admin, err := sql.Open("pgx", adminDSN)
	if err != nil {
		return nil, fmt.Errorf("connecting to admin database: %w", err)
	}
	defer admin.Close()

	// dbName is validated against ^[a-z_]+$ plus a fixed prefix.
	if _, err := admin.ExecContext(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s WITH (FORCE)", dbName)); err != nil {
		return nil, fmt.Errorf("dropping stale scratch database: %w", err)
	}
	if _, err := admin.ExecContext(ctx, "CREATE DATABASE "+dbName); err != nil {
		return nil, fmt.Errorf("creating scratch database: %w", err)
	}

	u.Path = "/" + dbName
	pool, err := pgxpool.New(ctx, u.String())
	if err != nil {
		return nil, fmt.Errorf("connecting to scratch database: %w", err)
	}

	sqlDB := sql.OpenDB(stdlib.GetConnector(*pool.Config().ConnConfig))
	defer sqlDB.Close()
	if _, err := migrate.Exec(sqlDB, "postgres", &migrate.FileMigrationSource{Dir: MigrationsDir}, migrate.Up); err != nil {
		pool.Close()
		return nil, fmt.Errorf("applying migrations: %w", err)
	}
	return pool, nil
}

// Truncate empties the given tables (with identity restart and cascade) for
// test isolation. Call it at the start of each test that writes data.
func Truncate(t *testing.T, pool *pgxpool.Pool, tables ...string) {
	t.Helper()
	if len(tables) == 0 {
		return
	}
	for _, table := range tables {
		if !validName.MatchString(table) {
			t.Fatalf("testdb.Truncate: invalid table name %q", table)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	stmt := "TRUNCATE "
	for i, table := range tables {
		if i > 0 {
			stmt += ", "
		}
		stmt += table
	}
	stmt += " RESTART IDENTITY CASCADE"
	if _, err := pool.Exec(ctx, stmt); err != nil {
		t.Fatalf("testdb.Truncate: %v", err)
	}
}
