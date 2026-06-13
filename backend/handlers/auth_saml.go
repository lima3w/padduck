package handlers

import (
	"github.com/gofiber/fiber/v2"
	"padduck/models"
	"padduck/services"
)

// ============================================================
// Public SAML endpoints
// ============================================================

// SAMLMetadata handles GET /api/v1/auth/saml/metadata.
// Returns the SP metadata XML document.
func (h *Handler) SAMLMetadata(c *fiber.Ctx) error {
	xml, err := h.service.SAML.GetSPMetadata(c.Context())
	if err != nil {
		reqLogger(c).Error("SAML metadata error", "error", err)
		return RespondError(c, fiber.StatusServiceUnavailable, ErrServiceUnavailable, "SAML not configured")
	}
	c.Set("Content-Type", "application/xml")
	return c.Send(xml)
}

// SAMLLogin handles GET /api/v1/auth/saml/login.
// Redirects to the IdP login page.
func (h *Handler) SAMLLogin(c *fiber.Ctx) error {
	cfg, err := h.service.SAML.GetConfig(c.Context())
	if err != nil || cfg == nil || !cfg.Enabled {
		return RespondError(c, fiber.StatusNotFound, ErrNotFound, "SAML authentication is not enabled")
	}

	relayState := c.Query("relay_state", "/")
	loginURL, err := h.service.SAML.GetLoginURL(c.Context(), relayState)
	if err != nil {
		reqLogger(c).Error("SAML GetLoginURL error", "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to build SAML login URL")
	}
	return c.Redirect(loginURL, fiber.StatusFound)
}

// SAMLAssertionConsumerService handles POST /api/v1/auth/saml/acs.
// Validates the IdP assertion, creates a session, and redirects to the frontend.
func (h *Handler) SAMLAssertionConsumerService(c *fiber.Ctx) error {
	samlResponse := c.FormValue("SAMLResponse")
	if samlResponse == "" {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "missing SAMLResponse parameter")
	}

	acsURL := c.BaseURL() + "/api/v1/auth/saml/acs"

	user, err := h.service.SAML.ProcessAssertion(c.Context(), samlResponse, acsURL)
	if err != nil {
		reqLogger(c).Error("SAML ACS error", "error", err)
		return RespondError(c, fiber.StatusUnauthorized, ErrUnauthorized, "SAML authentication failed")
	}

	token, err := h.service.CreateWebSession(c.Context(), user.ID, c.IP(), c.Get("User-Agent"))
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to create session")
	}

	uid := user.ID
	h.auditLog(c, services.AuditEntry{
		UserID:       &uid,
		Username:     user.Username,
		Action:       "login",
		ResourceType: "session",
		ResourceName: "saml",
	})

	h.setSessionCookie(c, token)
	return c.Redirect("/", fiber.StatusFound)
}

// ============================================================
// Admin SAML config endpoints
// ============================================================

// GetSAMLConfig handles GET /api/v1/admin/auth/saml.
func (h *Handler) GetSAMLConfig(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminWrite); err != nil {
		return nil
	}

	cfg, err := h.service.SAML.GetConfig(c.Context())
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to load SAML config")
	}
	if cfg == nil {
		return c.JSON(fiber.Map{})
	}

	// Never return the SP private key — replace with a presence indicator.
	return c.JSON(fiber.Map{
		"id":                cfg.ID,
		"enabled":           cfg.Enabled,
		"idp_metadata_url":  cfg.IDPMetadataURL,
		"idp_metadata_xml":  cfg.IDPMetadataXML,
		"sp_cert_pem":       cfg.SPCertPEM,
		"sp_key_configured": cfg.SPKeyPEM != "",
		"entity_id":         cfg.EntityID,
		"acs_url":           cfg.ACSURL,
		"name_id_format":    cfg.NameIDFormat,
		"created_at":        cfg.CreatedAt,
		"updated_at":        cfg.UpdatedAt,
	})
}

// UpdateSAMLConfig handles PUT /api/v1/admin/auth/saml.
func (h *Handler) UpdateSAMLConfig(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminWrite); err != nil {
		return nil
	}

	var req models.SAMLConfig
	if err := c.BodyParser(&req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}

	if req.NameIDFormat == "" {
		req.NameIDFormat = "urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress"
	}

	if err := h.service.SAML.SaveConfig(c.Context(), &req); err != nil {
		reqLogger(c).Error("UpdateSAMLConfig error", "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to save SAML config")
	}

	uid, uname := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "update",
		ResourceType: "saml_config",
	})
	return c.SendStatus(fiber.StatusNoContent)
}
