package repository

import (
	"context"
	"fmt"

	"padduck/models"
)

// QuotaExceededError is returned when an org's resource quota would be breached.
type QuotaExceededError struct {
	Resource string
	Limit    int
}

func (e *QuotaExceededError) Error() string {
	return fmt.Sprintf("quota exceeded: %s limit is %d", e.Resource, e.Limit)
}

// enforceQuota rejects an insert that would take the caller's org past its
// configured subnet or IP address limit. It lives in the repository so every
// creation path (API, discovery, DNS sync, import) is covered. System contexts
// with no org, and orgs with no limit set, are unrestricted. Check-then-insert
// is not atomic, so concurrent creates can overshoot by a few rows.
func (r *Repository) enforceQuota(ctx context.Context, resource string) error {
	org, ok := OrgFromContext(ctx)
	if !ok {
		return nil
	}
	settings, err := r.GetOrganizationSettings(ctx, org)
	if err != nil {
		return nil // no settings row = no limits
	}
	var limit *int
	var count func(context.Context, int64) (int64, error)
	switch resource {
	case "subnets":
		limit, count = settings.MaxSubnets, r.CountSubnetsByOrg
	case "ip_addresses":
		limit, count = settings.MaxIPAddresses, r.CountIPAddressesByOrg
	default:
		return nil
	}
	if limit == nil {
		return nil
	}
	n, err := count(ctx, org)
	if err != nil {
		return nil
	}
	if int(n) >= *limit {
		return &QuotaExceededError{Resource: resource, Limit: *limit}
	}
	return nil
}

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

func (r *Repository) CountSubnetsByOrg(ctx context.Context, orgID int64) (int64, error) {
	var n int64
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM subnets WHERE organization_id = $1`, orgID).Scan(&n)
	return n, err
}

func (r *Repository) CountIPAddressesByOrg(ctx context.Context, orgID int64) (int64, error) {
	var n int64
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM ip_addresses WHERE organization_id = $1`, orgID).Scan(&n)
	return n, err
}
