package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// Ping verifies database connectivity
func (r *Repository) Ping(ctx context.Context) error {
	return r.db.Ping(ctx)
}
