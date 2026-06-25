package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	pool *pgxpool.Pool
}

var (
	connectMaxAttempts = 10
	connectBaseDelay   = 1 * time.Second
	connectMaxDelay    = 16 * time.Second
)

// Connect creates a new PostgreSQL connection pool, retrying the initial ping
// with exponential backoff to handle the race between the backend and
// PostgreSQL starting simultaneously on a fresh install.
func Connect(ctx context.Context, connString string) (*DB, error) {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	delay := connectBaseDelay
	for attempt := 1; attempt <= connectMaxAttempts; attempt++ {
		if err = pool.Ping(ctx); err == nil {
			return &DB{pool: pool}, nil
		}
		if attempt == connectMaxAttempts {
			break
		}
		log.Printf("database not ready (attempt %d/%d): %v — retrying in %s",
			attempt, connectMaxAttempts, err, delay)
		select {
		case <-ctx.Done():
			pool.Close()
			return nil, ctx.Err()
		case <-time.After(delay):
		}
		if delay < connectMaxDelay {
			delay *= 2
			if delay > connectMaxDelay {
				delay = connectMaxDelay
			}
		}
	}

	pool.Close()
	return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", connectMaxAttempts, err)
}

// Pool returns the underlying pgxpool.Pool
func (db *DB) Pool() *pgxpool.Pool {
	return db.pool
}

// Close closes the connection pool
func (db *DB) Close() {
	if db.pool != nil {
		db.pool.Close()
	}
}

// Ping verifies the database connection is alive
func (db *DB) Ping(ctx context.Context) error {
	return db.pool.Ping(ctx)
}
