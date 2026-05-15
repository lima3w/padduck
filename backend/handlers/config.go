package handlers

import (
	"github.com/gofiber/fiber/v2"
	"ipam-next/models"
	"ipam-next/services"
)

// GetConfig handles GET /api/v1/admin/config
func (h *Handler) GetConfig(c *fiber.Ctx) error {
	currentUser, ok := c.Locals("user").(*models.User)
	if !ok || currentUser.Role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "admin access required"})
	}

	configs, err := h.service.Config.List()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to load config"})
	}

	result := make(map[string]string)
	for _, cfg := range configs {
		// Redact SMTP password in response
		if cfg.Key == "smtp_password" {
			if cfg.Value != "" {
				result[cfg.Key] = "********"
			} else {
				result[cfg.Key] = ""
			}
			continue
		}
		result[cfg.Key] = cfg.Value
	}

	return c.JSON(fiber.Map{"config": result})
}

// UpdateConfig handles PUT /api/v1/admin/config
func (h *Handler) UpdateConfig(c *fiber.Ctx) error {
	currentUser, ok := c.Locals("user").(*models.User)
	if !ok || currentUser.Role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "admin access required"})
	}

	var updates map[string]string
	if err := c.BodyParser(&updates); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	allowed := map[string]bool{
		"app_url":                      true,
		"registration_enabled":         true,
		"require_email_verification":   true,
		"require_admin_approval":       true,
		"smtp_host":                    true,
		"smtp_port":                    true,
		"smtp_username":                true,
		"smtp_password":                true,
		"smtp_from":                    true,
		"smtp_tls":                     true,
		"audit_log_retention_days":     true,
		"allow_subnet_overlaps":        true,
		"default_auto_reserve_first":   true,
		"default_auto_reserve_last":    true,
		"scanner_resolve_hostnames":    true,
	}

	for key, value := range updates {
		if !allowed[key] {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "unknown config key: " + key})
		}
		// Don't overwrite password if placeholder sent
		if key == "smtp_password" && value == "********" {
			continue
		}
		if err := h.service.Config.Set(key, value); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update config"})
		}
	}

	// Redact sensitive values before logging
	loggableUpdates := make(map[string]string, len(updates))
	for k, v := range updates {
		if k == "smtp_password" {
			loggableUpdates[k] = "***"
		} else {
			loggableUpdates[k] = v
		}
	}
	adminID, adminName := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: adminID, Username: adminName, Action: "config_updated",
		ResourceType: "config", NewValues: loggableUpdates,
	})

	return c.JSON(fiber.Map{"message": "config updated"})
}

// TestSMTP handles POST /api/v1/admin/config/test-email
func (h *Handler) TestSMTP(c *fiber.Ctx) error {
	currentUser, ok := c.Locals("user").(*models.User)
	if !ok || currentUser.Role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "admin access required"})
	}

	var req struct {
		To string `json:"to"`
	}
	if err := c.BodyParser(&req); err != nil || req.To == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "to address is required"})
	}

	if err := h.service.Email.Send(req.To, "IPAM SMTP Test", "This is a test email from IPAM."); err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "SMTP test failed: " + err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Test email sent successfully"})
}
