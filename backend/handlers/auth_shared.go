package handlers

import (
	"github.com/gofiber/fiber/v2"
	"padduck/models"
	"padduck/services"
)

// ============================================================
// Shared helpers
// ============================================================

// issueSessionResponse creates a web session for user and returns the standard LoginResponse.
func (h *Handler) issueSessionResponse(c *fiber.Ctx, user *models.User) error {
	if err := h.service.UpdateLastLogin(c.Context(), user.ID); err != nil {
		reqLogger(c).Error("error updating last login", "user_id", user.ID, "error", err)
	}

	token, err := h.service.CreateWebSession(c.Context(), user.ID, c.IP(), c.Get("User-Agent"))
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to create session")
	}

	uid := user.ID
	h.auditLog(c, services.AuditEntry{
		UserID: &uid, Username: user.Username, Action: "login",
		ResourceType: "session", Status: "success",
	})

	h.setSessionCookie(c, token)
	return c.JSON(LoginResponse{
		User: UserResponse{
			ID:                     user.ID,
			Username:               user.Username,
			Email:                  user.Email,
			Role:                   user.Role,
			State:                  user.State,
			GravatarURL:            gravatarURL(user.Email, 80),
			PrivacyAcceptedVersion: user.PrivacyAcceptedVersion,
			CreatedAt:              user.CreatedAt.String(),
			UpdatedAt:              user.UpdatedAt.String(),
		},
	})
}

// maskSecret returns "****" if set is true, or "" otherwise.
func maskSecret(set bool) string {
	if set {
		return "****"
	}
	return ""
}

// GetAuthProviders handles GET /api/v1/auth/providers.
// Returns which external auth providers are currently enabled (no auth required).
func (h *Handler) GetAuthProviders(c *fiber.Ctx) error {
	ldapEnabled := false
	oauth2Enabled := false
	samlEnabled := false

	if h.service != nil && h.service.LDAP != nil {
		if cfg, err := h.service.LDAP.GetConfig(c.Context()); err == nil && cfg != nil {
			ldapEnabled = cfg.Enabled
		}
	}
	if h.service != nil && h.service.OAuth2 != nil {
		if cfg, err := h.service.OAuth2.GetConfig(c.Context()); err == nil && cfg != nil {
			oauth2Enabled = cfg.Enabled
		}
	}
	if h.service != nil && h.service.SAML != nil {
		if cfg, err := h.service.SAML.GetConfig(c.Context()); err == nil && cfg != nil {
			samlEnabled = cfg.Enabled
		}
	}

	return c.JSON(fiber.Map{
		"ldap":   ldapEnabled,
		"oauth2": oauth2Enabled,
		"saml":   samlEnabled,
	})
}

// RegisterExternalAuthPublicRoutes adds LDAP/OAuth2/SAML routes that must stay
// reachable before a session exists.
func (h *Handler) RegisterExternalAuthPublicRoutes(auth fiber.Router) {
	// Public authentication endpoints
	auth.Get("/providers", h.GetAuthProviders)
	auth.Get("/ldap/login", h.LDAPStatus)
	auth.Post("/ldap/login", h.LDAPLogin)
	auth.Get("/oauth2/login", h.OAuth2Login)
	auth.Get("/oauth2/callback", h.OAuth2Callback)
	auth.Get("/saml/metadata", h.SAMLMetadata)
	auth.Get("/saml/login", h.SAMLLogin)
	auth.Post("/saml/acs", h.SAMLAssertionConsumerService)
}

// RegisterExternalAuthAdminRoutes adds protected external-auth configuration routes.
func (h *Handler) RegisterExternalAuthAdminRoutes(admin fiber.Router) {
	// Admin configuration endpoints
	admin.Get("/auth/ldap", h.GetLDAPConfig)
	admin.Put("/auth/ldap", h.UpdateLDAPConfig)
	admin.Post("/auth/ldap/test", h.TestLDAPConnection)
	admin.Get("/auth/ldap/group-mappings", h.ListLDAPGroupMappings)
	admin.Post("/auth/ldap/group-mappings", h.CreateLDAPGroupMapping)
	admin.Delete("/auth/ldap/group-mappings/:id", h.DeleteLDAPGroupMapping)
	admin.Get("/auth/oauth2", h.GetOAuth2Config)
	admin.Put("/auth/oauth2", h.UpdateOAuth2Config)
	admin.Get("/auth/saml", h.GetSAMLConfig)
	admin.Put("/auth/saml", h.UpdateSAMLConfig)
}
