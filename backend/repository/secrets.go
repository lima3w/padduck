package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"padduck/models"
)

const secretCols = `id, organization_id, name, description, current_version, created_by, created_at, rotated_at`

func scanSecret(row pgx.Row) (*models.Secret, error) {
	s := &models.Secret{Exists: true}
	err := row.Scan(&s.ID, &s.OrganizationID, &s.Name, &s.Description, &s.Version, &s.CreatedBy, &s.CreatedAt, &s.RotatedAt)
	return s, err
}

// CreateSecret inserts a secret and its first version atomically.
func (r *Repository) CreateSecret(ctx context.Context, name, description string, createdBy *int64, encDEK, encValue []byte) (*models.Secret, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	s, err := scanSecret(txQueryRow(ctx, tx, `
		INSERT INTO secrets (organization_id, name, description, created_by)
		VALUES (@orgval, $1, $2, $3)
		RETURNING `+secretCols, name, description, createdBy))
	if err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO secret_versions (secret_id, version, encrypted_dek, encrypted_value, created_by)
		VALUES ($1, 1, $2, $3, $4)`, s.ID, encDEK, encValue, createdBy); err != nil {
		return nil, err
	}
	return s, tx.Commit(ctx)
}

// GetSecret returns secret metadata by ID.
func (r *Repository) GetSecret(ctx context.Context, id int64) (*models.Secret, error) {
	return scanSecret(r.sQueryRow(ctx, `SELECT `+secretCols+` FROM secrets WHERE @org(organization_id) AND id = $1`, id))
}

// GetSecretByName returns secret metadata by name.
func (r *Repository) GetSecretByName(ctx context.Context, name string) (*models.Secret, error) {
	return scanSecret(r.sQueryRow(ctx, `SELECT `+secretCols+` FROM secrets WHERE @org(organization_id) AND name = $1`, name))
}

// ListSecrets returns metadata for every secret in scope, ordered by name.
func (r *Repository) ListSecrets(ctx context.Context) ([]*models.Secret, error) {
	rows, err := r.sQuery(ctx, `SELECT `+secretCols+` FROM secrets WHERE @org(organization_id) ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]*models.Secret, 0)
	for rows.Next() {
		s, err := scanSecret(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// RotateSecret adds a new version and makes it current. The previous versions
// are kept until deleted explicitly.
func (r *Repository) RotateSecret(ctx context.Context, id int64, createdBy *int64, encDEK, encValue []byte) (*models.Secret, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// MAX(version) rather than current_version so a deleted old version can
	// never cause a number to be reused.
	s, err := scanSecret(txQueryRow(ctx, tx, `
		UPDATE secrets SET
		  current_version = (SELECT COALESCE(MAX(version), 0) + 1 FROM secret_versions WHERE secret_id = secrets.id),
		  rotated_at = now()
		WHERE @org(organization_id) AND id = $1
		RETURNING `+secretCols, id))
	if err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO secret_versions (secret_id, version, encrypted_dek, encrypted_value, created_by)
		VALUES ($1, $2, $3, $4, $5)`, s.ID, s.Version, encDEK, encValue, createdBy); err != nil {
		return nil, err
	}
	return s, tx.Commit(ctx)
}

// GetCurrentSecretCiphertext returns the encrypted DEK and value of the secret's
// current version, with the secret's metadata.
func (r *Repository) GetCurrentSecretCiphertext(ctx context.Context, name string) (*models.Secret, []byte, []byte, error) {
	s, err := r.GetSecretByName(ctx, name)
	if err != nil {
		return nil, nil, nil, err
	}
	var dek, val []byte
	err = r.db.QueryRow(ctx,
		`SELECT encrypted_dek, encrypted_value FROM secret_versions WHERE secret_id = $1 AND version = $2`,
		s.ID, s.Version).Scan(&dek, &val)
	return s, dek, val, err
}

// ListSecretVersions returns version metadata for a secret, newest first.
func (r *Repository) ListSecretVersions(ctx context.Context, id int64) ([]*models.SecretVersion, error) {
	s, err := r.GetSecret(ctx, id) // enforces org scope
	if err != nil {
		return nil, err
	}
	rows, err := r.db.Query(ctx,
		`SELECT version, created_by, created_at FROM secret_versions WHERE secret_id = $1 ORDER BY version DESC`, s.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]*models.SecretVersion, 0)
	for rows.Next() {
		v := &models.SecretVersion{}
		if err := rows.Scan(&v.Version, &v.CreatedBy, &v.CreatedAt); err != nil {
			return nil, err
		}
		v.Current = v.Version == s.Version
		out = append(out, v)
	}
	return out, rows.Err()
}

// DeleteSecretVersion removes one non-current version. It returns pgx.ErrNoRows
// when the version does not exist.
func (r *Repository) DeleteSecretVersion(ctx context.Context, id int64, version int) error {
	s, err := r.GetSecret(ctx, id) // enforces org scope
	if err != nil {
		return err
	}
	if s.Version == version {
		return fmt.Errorf("invalid version: the current version cannot be deleted (rotate first or delete the secret)")
	}
	tag, err := r.db.Exec(ctx, `DELETE FROM secret_versions WHERE secret_id = $1 AND version = $2`, s.ID, version)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// DeleteSecret removes a secret and all of its versions.
func (r *Repository) DeleteSecret(ctx context.Context, id int64) error {
	tag, err := r.sExec(ctx, `DELETE FROM secrets WHERE @org(organization_id) AND id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
