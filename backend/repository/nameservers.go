package repository

import (
	"context"
	"fmt"

	"padduck/models"
)

// ---------------------------------------------------------------------------
// Nameserver operations
// ---------------------------------------------------------------------------

const nsSelectCols = `id, name, server1, server2, server3, description, created_at, updated_at`

func scanNameserver(row interface {
	Scan(dest ...any) error
}) (*models.Nameserver, error) {
	ns := &models.Nameserver{}
	return ns, row.Scan(&ns.ID, &ns.Name, &ns.Server1, &ns.Server2, &ns.Server3, &ns.Description, &ns.CreatedAt, &ns.UpdatedAt)
}

// NameserverParams holds fields for creating or updating a nameserver.
type NameserverParams struct {
	Name        string  `json:"name"`
	Server1     string  `json:"server1"`
	Server2     *string `json:"server2"`
	Server3     *string `json:"server3"`
	Description *string `json:"description"`
}

// CreateNameserver inserts a new nameserver record.
func (r *Repository) CreateNameserver(ctx context.Context, p *NameserverParams) (*models.Nameserver, error) {
	query := `INSERT INTO nameservers (name, server1, server2, server3, description)
	          VALUES ($1, $2, $3, $4, $5) RETURNING ` + nsSelectCols
	return scanNameserver(r.db.QueryRow(ctx, query, p.Name, p.Server1, p.Server2, p.Server3, p.Description))
}

// GetNameserverByID returns a single nameserver.
func (r *Repository) GetNameserverByID(ctx context.Context, id int64) (*models.Nameserver, error) {
	ns, err := scanNameserver(r.db.QueryRow(ctx, `SELECT `+nsSelectCols+` FROM nameservers WHERE id=$1`, id))
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, fmt.Errorf("nameserver not found")
		}
		return nil, err
	}
	return ns, nil
}

// ListNameservers returns all nameservers ordered by name.
func (r *Repository) ListNameservers(ctx context.Context) ([]*models.Nameserver, error) {
	rows, err := r.db.Query(ctx, `SELECT `+nsSelectCols+` FROM nameservers ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]*models.Nameserver, 0)
	for rows.Next() {
		ns, err := scanNameserver(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, ns)
	}
	return result, rows.Err()
}

// UpdateNameserver updates an existing nameserver.
func (r *Repository) UpdateNameserver(ctx context.Context, id int64, p *NameserverParams) (*models.Nameserver, error) {
	query := `UPDATE nameservers SET name=$1, server1=$2, server2=$3, server3=$4, description=$5, updated_at=now()
	          WHERE id=$6 RETURNING ` + nsSelectCols
	ns, err := scanNameserver(r.db.QueryRow(ctx, query, p.Name, p.Server1, p.Server2, p.Server3, p.Description, id))
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, fmt.Errorf("nameserver not found")
		}
		return nil, err
	}
	return ns, nil
}

// DeleteNameserver removes a nameserver by ID.
func (r *Repository) DeleteNameserver(ctx context.Context, id int64) error {
	ct, err := r.db.Exec(ctx, `DELETE FROM nameservers WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("nameserver not found")
	}
	return nil
}
