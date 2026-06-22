package repository

import (
	"context"

	"padduck/models"
)

func (r *Repository) GetOrganizationSettings(ctx context.Context, orgID int64) (*models.OrganizationSettings, error) {
	query := `SELECT organization_id, max_subnets, max_ip_addresses, max_users, max_webhooks, max_api_tokens,
	                 audit_retention_days, smtp_host, smtp_port, smtp_from, created_at, updated_at
	          FROM organization_settings WHERE organization_id = $1`
	row := r.db.QueryRow(ctx, query, orgID)
	s := &models.OrganizationSettings{}
	err := row.Scan(&s.OrganizationID, &s.MaxSubnets, &s.MaxIPAddresses, &s.MaxUsers, &s.MaxWebhooks,
		&s.MaxAPITokens, &s.AuditRetentionDays, &s.SMTPHost, &s.SMTPPort, &s.SMTPFrom,
		&s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *Repository) UpsertOrganizationSettings(ctx context.Context, s *models.OrganizationSettings) (*models.OrganizationSettings, error) {
	query := `INSERT INTO organization_settings
	            (organization_id, max_subnets, max_ip_addresses, max_users, max_webhooks, max_api_tokens,
	             audit_retention_days, smtp_host, smtp_port, smtp_from)
	          VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
	          ON CONFLICT (organization_id) DO UPDATE SET
	            max_subnets          = EXCLUDED.max_subnets,
	            max_ip_addresses     = EXCLUDED.max_ip_addresses,
	            max_users            = EXCLUDED.max_users,
	            max_webhooks         = EXCLUDED.max_webhooks,
	            max_api_tokens       = EXCLUDED.max_api_tokens,
	            audit_retention_days = EXCLUDED.audit_retention_days,
	            smtp_host            = EXCLUDED.smtp_host,
	            smtp_port            = EXCLUDED.smtp_port,
	            smtp_from            = EXCLUDED.smtp_from,
	            updated_at           = NOW()
	          RETURNING organization_id, max_subnets, max_ip_addresses, max_users, max_webhooks, max_api_tokens,
	                    audit_retention_days, smtp_host, smtp_port, smtp_from, created_at, updated_at`
	row := r.db.QueryRow(ctx, query, s.OrganizationID, s.MaxSubnets, s.MaxIPAddresses, s.MaxUsers,
		s.MaxWebhooks, s.MaxAPITokens, s.AuditRetentionDays, s.SMTPHost, s.SMTPPort, s.SMTPFrom)
	out := &models.OrganizationSettings{}
	err := row.Scan(&out.OrganizationID, &out.MaxSubnets, &out.MaxIPAddresses, &out.MaxUsers, &out.MaxWebhooks,
		&out.MaxAPITokens, &out.AuditRetentionDays, &out.SMTPHost, &out.SMTPPort, &out.SMTPFrom,
		&out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (r *Repository) ListOrganizationSettings(ctx context.Context) ([]*models.OrganizationSettings, error) {
	query := `SELECT organization_id, max_subnets, max_ip_addresses, max_users, max_webhooks, max_api_tokens,
	                 audit_retention_days, smtp_host, smtp_port, smtp_from, created_at, updated_at
	          FROM organization_settings`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*models.OrganizationSettings
	for rows.Next() {
		s := &models.OrganizationSettings{}
		if err := rows.Scan(&s.OrganizationID, &s.MaxSubnets, &s.MaxIPAddresses, &s.MaxUsers, &s.MaxWebhooks,
			&s.MaxAPITokens, &s.AuditRetentionDays, &s.SMTPHost, &s.SMTPPort, &s.SMTPFrom,
			&s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *Repository) CountWebhooksByOrg(ctx context.Context, orgID int64) (int64, error) {
	var n int64
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM webhook_endpoints WHERE organization_id = $1`, orgID).Scan(&n)
	return n, err
}

func (r *Repository) CountAPITokensByOrg(ctx context.Context, orgID int64) (int64, error) {
	var n int64
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM api_tokens WHERE organization_id = $1 AND (rotation_grace_expires_at IS NULL OR rotation_grace_expires_at > NOW())`,
		orgID).Scan(&n)
	return n, err
}
