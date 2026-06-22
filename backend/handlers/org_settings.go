package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"padduck/models"
	"padduck/services"
)

type orgSettingsRequest struct {
	MaxSubnets         *int    `json:"max_subnets"`
	MaxIPAddresses     *int    `json:"max_ip_addresses"`
	MaxUsers           *int    `json:"max_users"`
	MaxWebhooks        *int    `json:"max_webhooks"`
	MaxAPITokens       *int    `json:"max_api_tokens"`
	AuditRetentionDays *int    `json:"audit_retention_days"`
	SMTPHost           *string `json:"smtp_host"`
	SMTPPort           *int    `json:"smtp_port"`
	SMTPFrom           *string `json:"smtp_from"`
}

// GetOrgSettings handles GET /api/v1/admin/organization/settings
// Org admins can view their own settings (quotas visible but not raiseable).
func (h *Handler) GetOrgSettings(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminWrite) {
		return nil
	}
	orgID := orgIDFromCtx(c)
	if orgID == nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "no organization in context")
	}
	s, err := h.ops.OrgSettings.GetSettings(c.Context(), *orgID)
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to fetch org settings")
	}
	return c.JSON(s)
}

// UpdateOrgSettings handles PUT /api/v1/admin/organization/settings
// Org admins may update SMTP overrides only; quota fields are ignored (platform-admin only).
func (h *Handler) UpdateOrgSettings(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminWrite) {
		return nil
	}
	orgID := orgIDFromCtx(c)
	if orgID == nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "no organization in context")
	}
	var req orgSettingsRequest
	if err := c.BodyParser(&req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}
	// Fetch existing settings first so quotas set by platform admins are preserved.
	existing, _ := h.ops.OrgSettings.GetSettings(c.Context(), *orgID)
	existing.SMTPHost = req.SMTPHost
	existing.SMTPPort = req.SMTPPort
	existing.SMTPFrom = req.SMTPFrom
	s, err := h.ops.OrgSettings.UpsertSettings(c.Context(), existing)
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to update org settings")
	}
	uid, uname := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "org_settings_updated",
		ResourceType: "org_settings",
		NewValues: map[string]interface{}{
			"smtp_host": req.SMTPHost,
			"smtp_port": req.SMTPPort,
			"smtp_from": req.SMTPFrom,
		},
	})
	return c.JSON(s)
}

// GetPlatformOrgSettings handles GET /api/v1/platform/organizations/:id/settings
func (h *Handler) GetPlatformOrgSettings(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid organization ID")
	}
	s, err := h.ops.OrgSettings.GetSettings(c.Context(), id)
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to fetch org settings")
	}
	return c.JSON(s)
}

// UpdatePlatformOrgSettings handles PUT /api/v1/platform/organizations/:id/settings
// Platform admins may update all quota and integration fields.
func (h *Handler) UpdatePlatformOrgSettings(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid organization ID")
	}
	var req orgSettingsRequest
	if err := c.BodyParser(&req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}
	s, err := h.ops.OrgSettings.UpsertSettings(c.Context(), &models.OrganizationSettings{
		OrganizationID:     id,
		MaxSubnets:         req.MaxSubnets,
		MaxIPAddresses:     req.MaxIPAddresses,
		MaxUsers:           req.MaxUsers,
		MaxWebhooks:        req.MaxWebhooks,
		MaxAPITokens:       req.MaxAPITokens,
		AuditRetentionDays: req.AuditRetentionDays,
		SMTPHost:           req.SMTPHost,
		SMTPPort:           req.SMTPPort,
		SMTPFrom:           req.SMTPFrom,
	})
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to update org settings")
	}
	uid, uname := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "org_settings_updated",
		ResourceType: "org_settings",
		NewValues: map[string]interface{}{
			"organization_id":     id,
			"max_users":           req.MaxUsers,
			"max_webhooks":        req.MaxWebhooks,
			"max_api_tokens":      req.MaxAPITokens,
			"audit_retention_days": req.AuditRetentionDays,
		},
	})
	return c.JSON(s)
}
