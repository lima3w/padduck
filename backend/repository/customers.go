package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"ipam-next/models"
)

func (r *Repository) CreateCustomer(ctx context.Context, name, description, email, phone, notes string) (*models.Customer, error) {
	query := `INSERT INTO customers (name, description, email, phone, notes)
	          VALUES ($1, $2, $3, $4, $5)
	          RETURNING id, name, description, email, phone, notes, created_at, updated_at`
	c := &models.Customer{}
	err := r.db.QueryRow(ctx, query, name, description, email, phone, notes).Scan(
		&c.ID, &c.Name, &c.Description, &c.Email, &c.Phone, &c.Notes, &c.CreatedAt, &c.UpdatedAt,
	)
	return c, err
}

func (r *Repository) GetCustomerByID(ctx context.Context, id int64) (*models.Customer, error) {
	query := `SELECT id, name, description, email, phone, notes, created_at, updated_at FROM customers WHERE id = $1`
	c := &models.Customer{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&c.ID, &c.Name, &c.Description, &c.Email, &c.Phone, &c.Notes, &c.CreatedAt, &c.UpdatedAt,
	)
	return c, err
}

func (r *Repository) ListAllCustomers(ctx context.Context) ([]*models.Customer, error) {
	query := `SELECT id, name, description, email, phone, notes, created_at, updated_at FROM customers ORDER BY name ASC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	customers := make([]*models.Customer, 0)
	for rows.Next() {
		c := &models.Customer{}
		if err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.Email, &c.Phone, &c.Notes, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		customers = append(customers, c)
	}
	return customers, rows.Err()
}

func (r *Repository) UpdateCustomer(ctx context.Context, id int64, name, description, email, phone, notes string) (*models.Customer, error) {
	query := `UPDATE customers SET name = $1, description = $2, email = $3, phone = $4, notes = $5, updated_at = CURRENT_TIMESTAMP
	          WHERE id = $6
	          RETURNING id, name, description, email, phone, notes, created_at, updated_at`
	c := &models.Customer{}
	err := r.db.QueryRow(ctx, query, name, description, email, phone, notes, id).Scan(
		&c.ID, &c.Name, &c.Description, &c.Email, &c.Phone, &c.Notes, &c.CreatedAt, &c.UpdatedAt,
	)
	return c, err
}

func (r *Repository) DeleteCustomer(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM customers WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
