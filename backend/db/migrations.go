package db

import (
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5/stdlib"
	migrate "github.com/rubenv/sql-migrate"
)

// RunMigrations runs pending database migrations
func (db *DB) RunMigrations(migrationsPath string) error {
	sqlDB := sql.OpenDB(stdlib.GetConnector(*db.pool.Config().ConnConfig))
	defer sqlDB.Close()

	source := &migrate.FileMigrationSource{
		Dir: migrationsPath,
	}

	n, err := migrate.Exec(sqlDB, "postgres", source, migrate.Up)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	if n > 0 {
		fmt.Printf("Applied %d migrations\n", n)
	}

	return nil
}

// GetMigrationStatus returns the status of all migrations
func (db *DB) GetMigrationStatus(migrationsPath string) ([]*migrate.MigrationRecord, error) {
	sqlDB := sql.OpenDB(stdlib.GetConnector(*db.pool.Config().ConnConfig))
	defer sqlDB.Close()

	records, err := migrate.GetMigrationRecords(sqlDB, "postgres")
	if err != nil {
		return nil, fmt.Errorf("failed to get migration status: %w", err)
	}

	return records, nil
}

// DryRunMigrations prints pending migrations and their SQL without applying them.
// Returns the number of pending migrations.
func (db *DB) DryRunMigrations(migrationsPath string) (int, error) {
	sqlDB := sql.OpenDB(stdlib.GetConnector(*db.pool.Config().ConnConfig))
	defer sqlDB.Close()

	source := &migrate.FileMigrationSource{Dir: migrationsPath}
	planned, _, err := migrate.PlanMigration(sqlDB, "postgres", source, migrate.Up, 0)
	if err != nil {
		return 0, fmt.Errorf("planning migrations: %w", err)
	}

	if len(planned) == 0 {
		fmt.Println("No pending migrations.")
		return 0, nil
	}

	fmt.Printf("%d pending migration(s):\n\n", len(planned))
	for _, p := range planned {
		fmt.Printf("--- %s ---\n", p.Migration.Id)
		for _, q := range p.Queries {
			fmt.Println(q)
		}
		fmt.Println()
	}
	return len(planned), nil
}
