package db

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	migrate "github.com/rubenv/sql-migrate"
)

const migrationsDir = "../migrations"

// migrationTestDB creates a scratch database on the TEST_DATABASE_URL server
// so the up/down cycle cannot interfere with other test packages (the
// repository integration tests share the same server and run in parallel).
func migrationTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping migration integration test")
	}

	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatalf("parsing TEST_DATABASE_URL: %v", err)
	}

	admin, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("connecting to admin database: %v", err)
	}
	const scratchName = "padduck_migration_test"
	if _, err := admin.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s WITH (FORCE)", scratchName)); err != nil {
		t.Fatalf("dropping stale scratch database: %v", err)
	}
	if _, err := admin.Exec("CREATE DATABASE " + scratchName); err != nil {
		t.Fatalf("creating scratch database: %v", err)
	}
	t.Cleanup(func() {
		_, _ = admin.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s WITH (FORCE)", scratchName))
		admin.Close()
	})

	u.Path = "/" + scratchName
	scratch, err := sql.Open("pgx", u.String())
	if err != nil {
		t.Fatalf("connecting to scratch database: %v", err)
	}
	scratch.SetConnMaxLifetime(time.Minute)
	t.Cleanup(func() { scratch.Close() })
	return scratch
}

// TestMigrationsUpDownUp applies every migration on a clean schema, rolls
// every one of them back down to zero, then applies them all again. The
// re-up pass catches down migrations that leave residue (orphaned tables,
// types, or indexes) that would break a rebuild.
func TestMigrationsUpDownUp(t *testing.T) {
	sqlDB := migrationTestDB(t)
	source := &migrate.FileMigrationSource{Dir: migrationsDir}

	found, err := source.FindMigrations()
	if err != nil {
		t.Fatalf("loading migrations from %s: %v", migrationsDir, err)
	}
	if len(found) == 0 {
		t.Fatalf("no migrations found in %s", migrationsDir)
	}
	t.Logf("found %d migration files", len(found))

	up1, err := migrate.Exec(sqlDB, "postgres", source, migrate.Up)
	if err != nil {
		t.Fatalf("up pass on clean schema failed: %v", err)
	}
	if up1 != len(found) {
		t.Fatalf("up pass applied %d of %d migrations", up1, len(found))
	}
	assertTableExists(t, sqlDB, "users", true)

	down, err := migrate.Exec(sqlDB, "postgres", source, migrate.Down)
	if err != nil {
		t.Fatalf("down pass failed: %v", err)
	}
	if down != len(found) {
		t.Fatalf("down pass reverted %d of %d migrations", down, len(found))
	}
	assertTableExists(t, sqlDB, "users", false)
	assertSchemaEmpty(t, sqlDB)

	up2, err := migrate.Exec(sqlDB, "postgres", source, migrate.Up)
	if err != nil {
		t.Fatalf("re-up pass after full rollback failed: %v", err)
	}
	if up2 != len(found) {
		t.Fatalf("re-up pass applied %d of %d migrations", up2, len(found))
	}
	assertTableExists(t, sqlDB, "users", true)
}

// TestMigrationFilesArePaired verifies every up migration has a down
// counterpart and vice versa, without needing a database.
func TestMigrationFilesArePaired(t *testing.T) {
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		t.Fatalf("reading %s: %v", migrationsDir, err)
	}
	ups := map[string]bool{}
	downs := map[string]bool{}
	for _, e := range entries {
		name := e.Name()
		switch {
		case strings.HasSuffix(name, ".up.sql"):
			ups[strings.TrimSuffix(name, ".up.sql")] = true
		case strings.HasSuffix(name, ".down.sql"):
			downs[strings.TrimSuffix(name, ".down.sql")] = true
		}
	}
	for base := range ups {
		if !downs[base] {
			t.Errorf("migration %s has no .down.sql counterpart", base)
		}
	}
	for base := range downs {
		if !ups[base] {
			t.Errorf("migration %s has no .up.sql counterpart", base)
		}
	}
}

func assertTableExists(t *testing.T, db *sql.DB, table string, want bool) {
	t.Helper()
	var exists bool
	err := db.QueryRow(
		`SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = $1)`,
		table).Scan(&exists)
	if err != nil {
		t.Fatalf("checking table %s: %v", table, err)
	}
	if exists != want {
		t.Fatalf("table %s exists=%v, want %v", table, exists, want)
	}
}

// assertSchemaEmpty verifies that after a full rollback the public schema
// holds nothing but sql-migrate's own bookkeeping table.
func assertSchemaEmpty(t *testing.T, db *sql.DB) {
	t.Helper()
	rows, err := db.Query(
		`SELECT table_name FROM information_schema.tables
		 WHERE table_schema = 'public' AND table_name <> 'gorp_migrations'`)
	if err != nil {
		t.Fatalf("listing leftover tables: %v", err)
	}
	defer rows.Close()
	var leftovers []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatal(err)
		}
		leftovers = append(leftovers, name)
	}
	if len(leftovers) > 0 {
		t.Errorf("tables left behind after full rollback: %s", strings.Join(leftovers, ", "))
	}
}
