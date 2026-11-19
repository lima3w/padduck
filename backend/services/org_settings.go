package services

import (
	"context"

	"padduck/models"
	"padduck/repository"
)

// QuotaExceededError is returned when an org's resource quota would be breached.
type QuotaExceededError = repository.QuotaExceededError

type OrgSettingsService struct {
	repo *repository.Repository
}

func NewOrgSettingsService(repo *repository.Repository) *OrgSettingsService {
	return &OrgSettingsService{repo: repo}
}

func (s *OrgSettingsService) GetSettings(ctx context.Context, orgID int64) (*models.OrganizationSettings, error) {
	settings, err := s.repo.GetOrganizationSettings(ctx, orgID)
	if err != nil {
		// No settings row yet — return empty defaults.
		return &models.OrganizationSettings{OrganizationID: orgID}, nil
	}
	return settings, nil
}

func (s *OrgSettingsService) UpsertSettings(ctx context.Context, in *models.OrganizationSettings) (*models.OrganizationSettings, error) {
	return s.repo.UpsertOrganizationSettings(ctx, in)
}

// CheckQuota verifies whether creating one more resource of the given type
// would exceed the org's configured limit. Nil orgID means no org scoping —
// quota is skipped. resource must be one of: "users", "webhooks", "api_tokens",
// "subnets", "ip_addresses".
func (s *OrgSettingsService) CheckQuota(ctx context.Context, orgID *int64, resource string) error {
	if orgID == nil {
		return nil
	}
	settings, err := s.repo.GetOrganizationSettings(ctx, *orgID)
	if err != nil {
		return nil // no settings = no limit
	}

	var limit *int
	var current int64

	switch resource {
	case "users":
		limit = settings.MaxUsers
		if limit == nil {
			return nil
		}
		current, err = s.repo.CountUsersInOrg(ctx, *orgID)
	case "webhooks":
		limit = settings.MaxWebhooks
		if limit == nil {
			return nil
		}
		current, err = s.repo.CountWebhooksByOrg(ctx, *orgID)
	case "api_tokens":
		limit = settings.MaxAPITokens
		if limit == nil {
			return nil
		}
		current, err = s.repo.CountAPITokensByOrg(ctx, *orgID)
	case "subnets":
		limit = settings.MaxSubnets
		if limit == nil {
			return nil
		}
		current, err = s.repo.CountSubnetsByOrg(ctx, *orgID)
	case "ip_addresses":
		limit = settings.MaxIPAddresses
		if limit == nil {
			return nil
		}
		current, err = s.repo.CountIPAddressesByOrg(ctx, *orgID)
	default:
		return nil
	}

	if err != nil {
		return nil // count failure = allow the operation
	}
	if int(current) >= *limit {
		return &QuotaExceededError{Resource: resource, Limit: *limit}
	}
	return nil
}

// CheckQuotaCtx is CheckQuota for callers that only have a context. The org is
// taken from the request scope; system contexts (no org) are not limited.
func (s *OrgSettingsService) CheckQuotaCtx(ctx context.Context, resource string) error {
	org, ok := repository.OrgFromContext(ctx)
	if !ok {
		return nil
	}
	return s.CheckQuota(ctx, &org, resource)
}

// GetOrgSMTPOverride returns per-org SMTP settings if configured, else nil values.
func (s *OrgSettingsService) GetOrgSMTPOverride(ctx context.Context, orgID *int64) (host *string, port *int, from *string) {
	if orgID == nil {
		return nil, nil, nil
	}
	settings, err := s.repo.GetOrganizationSettings(ctx, *orgID)
	if err != nil {
		return nil, nil, nil
	}
	return settings.SMTPHost, settings.SMTPPort, settings.SMTPFrom
}

// ListAll returns all organization settings rows.
func (s *OrgSettingsService) ListAll(ctx context.Context) ([]*models.OrganizationSettings, error) {
	return s.repo.ListOrganizationSettings(ctx)
}

// GetOrgAuditRetention returns the per-org audit retention days, or 0 if not set.
func (s *OrgSettingsService) GetOrgAuditRetention(ctx context.Context, orgID int64) int {
	settings, err := s.repo.GetOrganizationSettings(ctx, orgID)
	if err != nil || settings.AuditRetentionDays == nil {
		return 0
	}
	return *settings.AuditRetentionDays
}
