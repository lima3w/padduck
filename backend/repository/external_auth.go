package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"padduck/models"
)

// ---- LDAP config ----

// GetLDAPConfig retrieves the single-row LDAP configuration (id=1).
// Returns nil, nil if no row exists yet.
func (r *Repository) GetLDAPConfig(ctx context.Context) (*models.LDAPConfig, error) {
	query := `
		SELECT id, enabled, host, port, bind_dn, bind_password_enc, base_dn,
		       user_filter, username_attr, email_attr, tls_mode, tls_skip_verify,
		       created_at, updated_at
		FROM ldap_configs
		ORDER BY id
		LIMIT 1`
	row := r.db.QueryRow(ctx, query)
	cfg := &models.LDAPConfig{}
	err := row.Scan(
		&cfg.ID, &cfg.Enabled, &cfg.Host, &cfg.Port, &cfg.BindDN, &cfg.BindPasswordEnc,
		&cfg.BaseDN, &cfg.UserFilter, &cfg.UsernameAttr, &cfg.EmailAttr,
		&cfg.TLSMode, &cfg.TLSSkipVerify, &cfg.CreatedAt, &cfg.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return cfg, nil
}

// UpsertLDAPConfig inserts or updates the single LDAP config row.
func (r *Repository) UpsertLDAPConfig(ctx context.Context, cfg *models.LDAPConfig) error {
	query := `
		INSERT INTO ldap_configs
			(id, enabled, host, port, bind_dn, bind_password_enc, base_dn,
			 user_filter, username_attr, email_attr, tls_mode, tls_skip_verify, updated_at)
		VALUES (1, $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW())
		ON CONFLICT (id) DO UPDATE SET
			enabled           = EXCLUDED.enabled,
			host              = EXCLUDED.host,
			port              = EXCLUDED.port,
			bind_dn           = EXCLUDED.bind_dn,
			bind_password_enc = EXCLUDED.bind_password_enc,
			base_dn           = EXCLUDED.base_dn,
			user_filter       = EXCLUDED.user_filter,
			username_attr     = EXCLUDED.username_attr,
			email_attr        = EXCLUDED.email_attr,
			tls_mode          = EXCLUDED.tls_mode,
			tls_skip_verify   = EXCLUDED.tls_skip_verify,
			updated_at        = NOW()`
	_, err := r.db.Exec(ctx, query,
		cfg.Enabled, cfg.Host, cfg.Port, cfg.BindDN, cfg.BindPasswordEnc,
		cfg.BaseDN, cfg.UserFilter, cfg.UsernameAttr, cfg.EmailAttr,
		cfg.TLSMode, cfg.TLSSkipVerify,
	)
	return err
}

// ---- LDAP group mappings ----

// GetLDAPGroupMappings returns all LDAP group → role mappings.
func (r *Repository) GetLDAPGroupMappings(ctx context.Context) ([]*models.LDAPGroupRoleMapping, error) {
	query := `SELECT id, ldap_group_dn, role_id, created_at FROM ldap_group_role_mappings ORDER BY id`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*models.LDAPGroupRoleMapping
	for rows.Next() {
		m := &models.LDAPGroupRoleMapping{}
		if err := rows.Scan(&m.ID, &m.LDAPGroupDN, &m.RoleID, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// CreateLDAPGroupMapping inserts a new LDAP group → role mapping.
func (r *Repository) CreateLDAPGroupMapping(ctx context.Context, m *models.LDAPGroupRoleMapping) error {
	query := `
		INSERT INTO ldap_group_role_mappings (ldap_group_dn, role_id)
		VALUES ($1, $2)
		RETURNING id, created_at`
	return r.db.QueryRow(ctx, query, m.LDAPGroupDN, m.RoleID).Scan(&m.ID, &m.CreatedAt)
}

// DeleteLDAPGroupMapping removes a mapping by ID.
func (r *Repository) DeleteLDAPGroupMapping(ctx context.Context, id int64) error {
	result, err := r.db.Exec(ctx, `DELETE FROM ldap_group_role_mappings WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// ---- OAuth2 config ----

// GetOAuth2Config retrieves the single-row OAuth2 configuration.
// Returns nil, nil if no row exists.
func (r *Repository) GetOAuth2Config(ctx context.Context) (*models.OAuth2Config, error) {
	query := `
		SELECT id, enabled, provider_name, client_id, client_secret_enc,
		       discovery_url, authorization_url, token_url, userinfo_url,
		       scopes, redirect_uri, created_at, updated_at
		FROM oauth2_configs
		ORDER BY id
		LIMIT 1`
	row := r.db.QueryRow(ctx, query)
	cfg := &models.OAuth2Config{}
	err := row.Scan(
		&cfg.ID, &cfg.Enabled, &cfg.ProviderName, &cfg.ClientID, &cfg.ClientSecretEnc,
		&cfg.DiscoveryURL, &cfg.AuthorizationURL, &cfg.TokenURL, &cfg.UserinfoURL,
		&cfg.Scopes, &cfg.RedirectURI, &cfg.CreatedAt, &cfg.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return cfg, nil
}

// UpsertOAuth2Config inserts or updates the single OAuth2 config row.
func (r *Repository) UpsertOAuth2Config(ctx context.Context, cfg *models.OAuth2Config) error {
	query := `
		INSERT INTO oauth2_configs
			(id, enabled, provider_name, client_id, client_secret_enc,
			 discovery_url, authorization_url, token_url, userinfo_url,
			 scopes, redirect_uri, updated_at)
		VALUES (1, $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
		ON CONFLICT (id) DO UPDATE SET
			enabled           = EXCLUDED.enabled,
			provider_name     = EXCLUDED.provider_name,
			client_id         = EXCLUDED.client_id,
			client_secret_enc = EXCLUDED.client_secret_enc,
			discovery_url     = EXCLUDED.discovery_url,
			authorization_url = EXCLUDED.authorization_url,
			token_url         = EXCLUDED.token_url,
			userinfo_url      = EXCLUDED.userinfo_url,
			scopes            = EXCLUDED.scopes,
			redirect_uri      = EXCLUDED.redirect_uri,
			updated_at        = NOW()`
	_, err := r.db.Exec(ctx, query,
		cfg.Enabled, cfg.ProviderName, cfg.ClientID, cfg.ClientSecretEnc,
		cfg.DiscoveryURL, cfg.AuthorizationURL, cfg.TokenURL, cfg.UserinfoURL,
		cfg.Scopes, cfg.RedirectURI,
	)
	return err
}

// ---- SAML config ----

// GetSAMLConfig retrieves the single-row SAML configuration.
// Returns nil, nil if no row exists.
func (r *Repository) GetSAMLConfig(ctx context.Context) (*models.SAMLConfig, error) {
	query := `
		SELECT id, enabled, idp_metadata_url, idp_metadata_xml, sp_cert_pem, sp_key_pem,
		       entity_id, acs_url, name_id_format, created_at, updated_at
		FROM saml_configs
		ORDER BY id
		LIMIT 1`
	row := r.db.QueryRow(ctx, query)
	cfg := &models.SAMLConfig{}
	err := row.Scan(
		&cfg.ID, &cfg.Enabled, &cfg.IDPMetadataURL, &cfg.IDPMetadataXML,
		&cfg.SPCertPEM, &cfg.SPKeyPEM, &cfg.EntityID, &cfg.ACSURL,
		&cfg.NameIDFormat, &cfg.CreatedAt, &cfg.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return cfg, nil
}

// UpsertSAMLConfig inserts or updates the single SAML config row.
func (r *Repository) UpsertSAMLConfig(ctx context.Context, cfg *models.SAMLConfig) error {
	query := `
		INSERT INTO saml_configs
			(id, enabled, idp_metadata_url, idp_metadata_xml, sp_cert_pem, sp_key_pem,
			 entity_id, acs_url, name_id_format, updated_at)
		VALUES (1, $1, $2, $3, $4, $5, $6, $7, $8, NOW())
		ON CONFLICT (id) DO UPDATE SET
			enabled          = EXCLUDED.enabled,
			idp_metadata_url = EXCLUDED.idp_metadata_url,
			idp_metadata_xml = EXCLUDED.idp_metadata_xml,
			sp_cert_pem      = EXCLUDED.sp_cert_pem,
			sp_key_pem       = EXCLUDED.sp_key_pem,
			entity_id        = EXCLUDED.entity_id,
			acs_url          = EXCLUDED.acs_url,
			name_id_format   = EXCLUDED.name_id_format,
			updated_at       = NOW()`
	_, err := r.db.Exec(ctx, query,
		cfg.Enabled, cfg.IDPMetadataURL, cfg.IDPMetadataXML,
		cfg.SPCertPEM, cfg.SPKeyPEM, cfg.EntityID, cfg.ACSURL, cfg.NameIDFormat,
	)
	return err
}

// ---- External auth users ----

// FindUserByExternalAuth returns the user associated with the given provider + externalID.
// Returns nil, nil if not found.
func (r *Repository) FindUserByExternalAuth(ctx context.Context, provider, externalID string) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, role, state,
		       last_login_at, suspended_at, suspended_by, suspension_reason,
		       privacy_accepted_at, privacy_accepted_version,
		       deletion_requested_at, anonymized_at,
		       external_auth_provider, external_auth_id,
		       created_at, updated_at
		FROM users
		WHERE external_auth_provider = $1 AND external_auth_id = $2`
	row := r.db.QueryRow(ctx, query, provider, externalID)
	user := &models.User{}
	err := row.Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.State,
		&user.LastLoginAt, &user.SuspendedAt, &user.SuspendedBy, &user.SuspensionReason,
		&user.PrivacyAcceptedAt, &user.PrivacyAcceptedVersion,
		&user.DeletionRequestedAt, &user.AnonymizedAt,
		&user.ExternalAuthProvider, &user.ExternalAuthID,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

// SetUserExternalAuth associates an external auth provider and ID with a user.
func (r *Repository) SetUserExternalAuth(ctx context.Context, userID int64, provider, externalID string) error {
	query := `UPDATE users SET external_auth_provider = $1, external_auth_id = $2 WHERE id = $3`
	result, err := r.db.Exec(ctx, query, provider, externalID, userID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// CreateExternalUser creates a new user for an external auth provider.
func (r *Repository) CreateExternalUser(ctx context.Context, username, email, provider, externalID string) (*models.User, error) {
	query := `
		INSERT INTO users (username, email, role, state, external_auth_provider, external_auth_id)
		VALUES ($1, $2, 'user', 'active', $3, $4)
		RETURNING id, username, email, password_hash, role, state,
		          last_login_at, suspended_at, suspended_by, suspension_reason,
		          privacy_accepted_at, privacy_accepted_version,
		          deletion_requested_at, anonymized_at,
		          external_auth_provider, external_auth_id,
		          created_at, updated_at`
	row := r.db.QueryRow(ctx, query, username, email, provider, externalID)
	user := &models.User{}
	err := row.Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.State,
		&user.LastLoginAt, &user.SuspendedAt, &user.SuspendedBy, &user.SuspensionReason,
		&user.PrivacyAcceptedAt, &user.PrivacyAcceptedVersion,
		&user.DeletionRequestedAt, &user.AnonymizedAt,
		&user.ExternalAuthProvider, &user.ExternalAuthID,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// ---- OAuth2 state ----

// SaveOAuth2State persists a short-lived OAuth2 state token.
func (r *Repository) SaveOAuth2State(ctx context.Context, state, redirectURI string, expiresAt time.Time) error {
	// Purge stale states on each write to keep table clean
	_, _ = r.db.Exec(ctx, `DELETE FROM oauth2_states WHERE expires_at < NOW()`)

	query := `INSERT INTO oauth2_states (state, redirect_uri, expires_at) VALUES ($1, $2, $3)
	          ON CONFLICT (state) DO UPDATE SET redirect_uri = EXCLUDED.redirect_uri, expires_at = EXCLUDED.expires_at`
	_, err := r.db.Exec(ctx, query, state, redirectURI, expiresAt)
	return err
}

// ConsumeOAuth2State atomically retrieves and deletes a state token.
// Returns ErrNotFound if the state is missing or expired.
func (r *Repository) ConsumeOAuth2State(ctx context.Context, state string) (redirectURI string, err error) {
	query := `
		DELETE FROM oauth2_states
		WHERE state = $1 AND expires_at >= NOW()
		RETURNING redirect_uri`
	row := r.db.QueryRow(ctx, query, state)
	if err = row.Scan(&redirectURI); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", fmt.Errorf("invalid or expired OAuth2 state")
		}
		return "", err
	}
	return redirectURI, nil
}
