package repository

import (
	"context"
	"fmt"

	"ipam-next/models"
)

// Config operations

func (r *Repository) GetConfig(ctx context.Context, key string) (*models.Config, error) {
	query := `SELECT key, value, created_at, updated_at FROM configs WHERE key = $1`
	cfg := &models.Config{}
	err := r.db.QueryRow(ctx, query, key).Scan(&cfg.Key, &cfg.Value, &cfg.CreatedAt, &cfg.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (r *Repository) ListConfigs(ctx context.Context) ([]*models.Config, error) {
	query := `SELECT key, value, created_at, updated_at FROM configs ORDER BY key ASC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	configs := make([]*models.Config, 0)
	for rows.Next() {
		cfg := &models.Config{}
		if err := rows.Scan(&cfg.Key, &cfg.Value, &cfg.CreatedAt, &cfg.UpdatedAt); err != nil {
			return nil, err
		}
		configs = append(configs, cfg)
	}
	return configs, rows.Err()
}

func (r *Repository) SetConfig(ctx context.Context, key, value string) error {
	query := `INSERT INTO configs (key, value) VALUES ($1, $2)
	          ON CONFLICT (key) DO UPDATE SET value = $2, updated_at = CURRENT_TIMESTAMP`
	_, err := r.db.Exec(ctx, query, key, value)
	return err
}

// SetConfigMultiple applies all key-value pairs atomically within a single transaction.
// If any write fails the entire update is rolled back.
func (r *Repository) SetConfigMultiple(ctx context.Context, pairs map[string]string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("config: begin transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	query := `INSERT INTO configs (key, value) VALUES ($1, $2)
	          ON CONFLICT (key) DO UPDATE SET value = $2, updated_at = CURRENT_TIMESTAMP`
	for key, value := range pairs {
		if _, err := tx.Exec(ctx, query, key, value); err != nil {
			return fmt.Errorf("config: set %q: %w", key, err)
		}
	}
	return tx.Commit(ctx)
}
