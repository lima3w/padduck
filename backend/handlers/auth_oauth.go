package handlers

import (
	"github.com/gofiber/fiber/v2"
	"padduck/models"
	"padduck/services"
)

// ============================================================
// Public OAuth2 / OIDC auth endpoints
// ============================================================

// OAuth2Login handles GET /api/v1/auth/oauth2/login.
// Redirects the browser to the provider's authorization URL.
func (h *Handler) OAuth2Login(c *fiber.Ctx) error {
	cfg, err := h.service.OAuth2.GetConfig(c.Context())
	if err != nil || cfg == nil || !cfg.Enabled {
		return RespondError(c, fiber.StatusNotFound, ErrNotFound, "OAuth2 authentication is not enabled")
	}

	redirectBack := c.Query("redirect", "/")
	authURL, _, err := h.service.OAuth2.GetAuthURL(c.Context(), redirectBack)
	if err != nil {
		reqLogger(c).Error("OAuth2 GetAuthURL error", "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to build authorization URL")
	}
	return c.Redirect(authURL, fiber.StatusFound)
}

// OAuth2Callback handles GET /api/v1/auth/oauth2/callback.
// Exchanges the authorization code for a session and redirects to the frontend.
func (h *Handler) OAuth2Callback(c *fiber.Ctx) error {
	code := c.Query("code")
	state := c.Query("state")
	if code == "" || state == "" {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "missing code or state parameter")
	}

	user, err := h.service.OAuth2.Exchange(c.Context(), code, state)
	if err != nil {
		reqLogger(c).Error("OAuth2 exchange error", "error", err)
		return RespondError(c, fiber.StatusUnauthorized, ErrUnauthorized, "OAuth2 authentication failed")
	}

	if !h.issueSessionCookie(c, user, "oauth2") {
		return nil
	}
	return c.Redirect("/", fiber.StatusFound)
}

// ============================================================
// Admin OAuth2 config endpoints
// ============================================================

// GetOAuth2Config handles GET /api/v1/admin/auth/oauth2.
func (h *Handler) GetOAuth2Config(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminWrite) {
		return nil
	}

	cfg, err := h.service.OAuth2.GetConfig(c.Context())
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to load OAuth2 config")
	}
	if cfg == nil {
		return c.JSON(fiber.Map{})
	}

	return c.JSON(fiber.Map{
		"id":                cfg.ID,
		"enabled":           cfg.Enabled,
		"provider_name":     cfg.ProviderName,
		"client_id":         cfg.ClientID,
		"client_secret":     maskSecret(len(cfg.ClientSecretEnc) > 0),
		"discovery_url":     cfg.DiscoveryURL,
		"authorization_url": cfg.AuthorizationURL,
		"token_url":         cfg.TokenURL,
		"userinfo_url":      cfg.UserinfoURL,
		"scopes":            cfg.Scopes,
		"redirect_uri":      cfg.RedirectURI,
		"created_at":        cfg.CreatedAt,
		"updated_at":        cfg.UpdatedAt,
	})
}

// UpdateOAuth2Config handles PUT /api/v1/admin/auth/oauth2.
func (h *Handler) UpdateOAuth2Config(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminWrite) {
		return nil
	}

	var req struct {
		Enabled          bool   `json:"enabled"`
		ProviderName     string `json:"provider_name"`
		ClientID         string `json:"client_id"`
		ClientSecret     string `json:"client_secret"`
		DiscoveryURL     string `json:"discovery_url"`
		AuthorizationURL string `json:"authorization_url"`
		TokenURL         string `json:"token_url"`
		UserinfoURL      string `json:"userinfo_url"`
		Scopes           string `json:"scopes"`
		RedirectURI      string `json:"redirect_uri"`
	}
	if err := c.BodyParser(&req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}

	if req.Scopes == "" {
		req.Scopes = "openid email profile"
	}

	if req.Enabled {
		if req.ClientID == "" {
			return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "client_id is required when OAuth2 is enabled")
		}
		hasDiscovery := req.DiscoveryURL != ""
		hasManual := req.AuthorizationURL != "" && req.TokenURL != "" && req.UserinfoURL != ""
		if !hasDiscovery && !hasManual {
			return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "either discovery_url or all of authorization_url, token_url, and userinfo_url are required when OAuth2 is enabled")
		}
		if req.RedirectURI == "" {
			return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "redirect_uri is required when OAuth2 is enabled")
		}
	}

	cfg := &models.OAuth2Config{
		Enabled:          req.Enabled,
		ProviderName:     req.ProviderName,
		ClientID:         req.ClientID,
		DiscoveryURL:     req.DiscoveryURL,
		AuthorizationURL: req.AuthorizationURL,
		TokenURL:         req.TokenURL,
		UserinfoURL:      req.UserinfoURL,
		Scopes:           req.Scopes,
		RedirectURI:      req.RedirectURI,
	}
	if req.ClientSecret != "" && req.ClientSecret != "****" {
		cfg.ClientSecretEnc = []byte(req.ClientSecret)
	}

	if err := h.service.OAuth2.SaveConfig(c.Context(), cfg); err != nil {
		reqLogger(c).Error("UpdateOAuth2Config error", "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to save OAuth2 config")
	}

	uid, uname := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "update",
		ResourceType: "oauth2_config",
	})
	return c.SendStatus(fiber.StatusNoContent)
}
