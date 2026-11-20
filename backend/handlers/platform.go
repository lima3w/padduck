package handlers

import (
	"context"
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"padduck/services"
)

// ListPlatformOrganizations handles GET /api/v1/platform/organizations
func (h *Handler) ListPlatformOrganizations(c *fiber.Ctx) error {
	orgs, err := h.ops.Organizations.List(c.Context())
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to list organizations")
	}
	return c.JSON(orgs)
}

// GetPlatformOrganization handles GET /api/v1/platform/organizations/:id
func (h *Handler) GetPlatformOrganization(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid organization ID")
	}
	org, err := h.ops.Organizations.Get(c.Context(), id)
	if err != nil {
		return RespondError(c, fiber.StatusNotFound, ErrNotFound, "organization not found")
	}
	repo := h.service.GetRepository()
	ctx := c.Context()
	stats := fiber.Map{}
	counters := []struct {
		key string
		fn  func(context.Context, int64) (int64, error)
	}{
		{"user_count", repo.CountUsersInOrg},
		{"subnet_count", repo.CountSubnetsByOrg},
		{"ip_address_count", repo.CountIPAddressesByOrg},
		{"api_token_count", repo.CountAPITokensByOrg},
		{"webhook_count", repo.CountWebhooksByOrg},
	}
	for _, ct := range counters {
		n, err := ct.fn(ctx, id)
		if err != nil {
			reqLogger(c).Error("platform org stats", "stat", ct.key, "error", err)
			return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to compute organization stats")
		}
		stats[ct.key] = n
	}
	return c.JSON(fiber.Map{"organization": org, "stats": stats})
}

// GetPlatformAuditLog handles GET /api/v1/platform/audit-log
// Supports ?org_id=, ?limit=, ?offset=, and all filters from buildAuditFilter.
func (h *Handler) GetPlatformAuditLog(c *fiber.Ctx) error {
	filter := buildAuditFilter(c)
	filter.AllOrgs = true
	filter.OrgID = nil
	if orgStr := c.Query("org_id"); orgStr != "" {
		orgID, err := strconv.ParseInt(orgStr, 10, 64)
		if err != nil || orgID <= 0 {
			return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid org_id")
		}
		filter.OrgID = &orgID
		filter.AllOrgs = false
	}

	logs, err := h.service.Audit.ListAuditLogs(c.Context(), filter)
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to fetch audit logs")
	}
	total, err := h.service.Audit.CountAuditLogs(c.Context(), filter)
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to count audit logs")
	}
	return c.JSON(fiber.Map{
		"logs":   formatAuditLogs(logs),
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

// PlatformImpersonate handles POST /api/v1/platform/impersonate
// Body: {"organization_id": 5}
// Returns a 1-hour API token scoped to the target org.
func (h *Handler) PlatformImpersonate(c *fiber.Ctx) error {
	var req struct {
		OrganizationID int64 `json:"organization_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}
	if req.OrganizationID <= 0 {
		return RespondValidationError(c, "validation failed", []ValidationField{
			{Field: "organization_id", Message: "organization_id must be greater than zero"},
		})
	}

	org, err := h.ops.Organizations.Get(c.Context(), req.OrganizationID)
	if err != nil {
		return RespondError(c, fiber.StatusNotFound, ErrNotFound, "organization not found")
	}

	adminUserID, _ := c.Locals("userID").(int64)
	rawToken, apiToken, err := h.ops.Identity.GenerateImpersonationToken(c.Context(), adminUserID, req.OrganizationID)
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to create impersonation token")
	}

	uid, username := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: username, Action: "platform_impersonate",
		ResourceType: "organization", ResourceID: &req.OrganizationID, ResourceName: org.Name,
		NewValues: map[string]interface{}{"impersonated_org_id": req.OrganizationID, "token_id": apiToken.ID},
	})

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"token":        rawToken,
		"token_id":     apiToken.ID,
		"expires_at":   apiToken.ExpiresAt,
		"organization": fiber.Map{"id": org.ID, "name": org.Name, "slug": org.Slug},
		"note":         "Use this token as a Bearer token. All actions are audit-logged under your identity.",
	})
}

// SetPlatformAdmin handles PUT /api/v1/platform/users/:id/platform-admin
// Body: {"is_platform_admin": true}
func (h *Handler) SetPlatformAdmin(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid user ID")
	}
	var req struct {
		IsPlatformAdmin bool `json:"is_platform_admin"`
	}
	if err := c.BodyParser(&req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}
	if caller, _ := c.Locals("userID").(int64); caller == id && !req.IsPlatformAdmin {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "you cannot remove your own platform admin access")
	}
	if err := h.service.GetRepository().SetPlatformAdmin(c.Context(), id, req.IsPlatformAdmin); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return RespondError(c, fiber.StatusNotFound, ErrNotFound, "user not found")
		}
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to update user")
	}
	uid, username := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: username, Action: "platform_admin_set",
		ResourceType: "user", ResourceID: &id,
		NewValues: map[string]interface{}{"is_platform_admin": req.IsPlatformAdmin},
	})
	return c.JSON(fiber.Map{"user_id": id, "is_platform_admin": req.IsPlatformAdmin})
}
