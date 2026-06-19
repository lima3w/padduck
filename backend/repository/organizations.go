package repository

import (
	"context"

	"padduck/models"
)

func (r *Repository) CreateOrganization(ctx context.Context, name, slug string) (*models.Organization, error) {
	org := &models.Organization{}
	err := r.db.QueryRow(ctx, `
		INSERT INTO organizations (name, slug)
		VALUES ($1, $2)
		RETURNING id, name, slug, created_at`,
		name, slug,
	).Scan(&org.ID, &org.Name, &org.Slug, &org.CreatedAt)
	return org, err
}

func (r *Repository) GetOrganization(ctx context.Context, id int64) (*models.Organization, error) {
	org := &models.Organization{}
	err := r.db.QueryRow(ctx, `
		SELECT id, name, slug, created_at FROM organizations WHERE id = $1`,
		id,
	).Scan(&org.ID, &org.Name, &org.Slug, &org.CreatedAt)
	return org, err
}

func (r *Repository) GetOrganizationBySlug(ctx context.Context, slug string) (*models.Organization, error) {
	org := &models.Organization{}
	err := r.db.QueryRow(ctx, `
		SELECT id, name, slug, created_at FROM organizations WHERE slug = $1`,
		slug,
	).Scan(&org.ID, &org.Name, &org.Slug, &org.CreatedAt)
	return org, err
}

func (r *Repository) ListOrganizations(ctx context.Context) ([]*models.Organization, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, name, slug, created_at FROM organizations ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var orgs []*models.Organization
	for rows.Next() {
		org := &models.Organization{}
		if err := rows.Scan(&org.ID, &org.Name, &org.Slug, &org.CreatedAt); err != nil {
			return nil, err
		}
		orgs = append(orgs, org)
	}
	return orgs, rows.Err()
}

func (r *Repository) DeleteOrganization(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM organizations WHERE id = $1`, id)
	return err
}

// OrganizationExists returns true if at least one organization row exists.
func (r *Repository) OrganizationExists(ctx context.Context) (bool, error) {
	var count int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM organizations`).Scan(&count)
	return count > 0, err
}

// EnsureDefaultOrganization creates the default organization if none exists and
// assigns all users without an organization to it. Returns the default org ID.
func (r *Repository) EnsureDefaultOrganization(ctx context.Context) (int64, error) {
	exists, err := r.OrganizationExists(ctx)
	if err != nil {
		return 0, err
	}
	if exists {
		org, err := r.GetOrganizationBySlug(ctx, "default")
		if err != nil {
			return 0, err
		}
		return org.ID, nil
	}

	org, err := r.CreateOrganization(ctx, "Default", "default")
	if err != nil {
		return 0, err
	}
	_, err = r.db.Exec(ctx, `UPDATE users SET organization_id = $1 WHERE organization_id IS NULL`, org.ID)
	return org.ID, err
}
