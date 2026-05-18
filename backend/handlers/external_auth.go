package handlers

import (
	"github.com/gofiber/fiber/v2"
	"ipam-next/models"
	"ipam-next/services"
)

// ============================================================
// Public LDAP auth endpoints
// ============================================================

// LDAPStatus handles GET /api/v1/auth/ldap/login.
// Returns 404 when LDAP is disabled, or {"ldap_enabled": true} when active.
func (h *Handler) LDAPStatus(c *fiber.Ctx) error {
	cfg, err := h.service.LDAP.GetConfig(c.Context())
	if err != nil || cfg == nil || !cfg.Enabled {
		return RespondError(c, fiber.StatusNotFound, ErrNotFound, "LDAP authentication is not enabled")
	}
	return c.JSON(fiber.Map{"ldap_enabled": true})
}

// LDAPLogin handles POST /api/v1/auth/ldap/login.
// Body: {"username":"...", "password":"..."}
func (h *Handler) LDAPLogin(c *fiber.Ctx) error {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}
	if req.Username == "" || req.Password == "" {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "username and password required")
	}

	cfg, err := h.service.LDAP.GetConfig(c.Context())
	if err != nil || cfg == nil || !cfg.Enabled {
		return RespondError(c, fiber.StatusNotFound, ErrNotFound, "LDAP authentication is not enabled")
	}

	user, err := h.service.LDAP.Authenticate(c.Context(), req.Username, req.Password)
	if err != nil {
		reqLogger(c).Error("LDAP authentication failed", "username", req.Username, "error", err)
		return RespondError(c, fiber.StatusUnauthorized, ErrUnauthorized, "invalid credentials")
	}

	return h.issueSessionResponse(c, user)
}

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
		ResourceName: "oauth2",
	})

	h.setSessionCookie(c, token)
	return c.Redirect("/", fiber.StatusFound)
}

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
// Admin LDAP config endpoints (require admin:write)
// ============================================================

// GetLDAPConfig handles GET /api/v1/admin/auth/ldap.
func (h *Handler) GetLDAPConfig(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminWrite); err != nil {
		return nil
	}

	cfg, err := h.service.LDAP.GetConfig(c.Context())
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to load LDAP config")
	}
	if cfg == nil {
		return c.JSON(fiber.Map{})
	}

	// Mask the bind password
	masked := fiber.Map{
		"id":              cfg.ID,
		"enabled":         cfg.Enabled,
		"host":            cfg.Host,
		"port":            cfg.Port,
		"bind_dn":         cfg.BindDN,
		"bind_password":   maskSecret(len(cfg.BindPasswordEnc) > 0),
		"base_dn":         cfg.BaseDN,
		"user_filter":     cfg.UserFilter,
		"username_attr":   cfg.UsernameAttr,
		"email_attr":      cfg.EmailAttr,
		"tls_mode":        cfg.TLSMode,
		"tls_skip_verify": cfg.TLSSkipVerify,
		"created_at":      cfg.CreatedAt,
		"updated_at":      cfg.UpdatedAt,
	}
	return c.JSON(masked)
}

// UpdateLDAPConfig handles PUT /api/v1/admin/auth/ldap.
func (h *Handler) UpdateLDAPConfig(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminWrite); err != nil {
		return nil
	}

	var req struct {
		Enabled       bool   `json:"enabled"`
		Host          string `json:"host"`
		Port          int    `json:"port"`
		BindDN        string `json:"bind_dn"`
		BindPassword  string `json:"bind_password"` // raw password from UI
		BaseDN        string `json:"base_dn"`
		UserFilter    string `json:"user_filter"`
		UsernameAttr  string `json:"username_attr"`
		EmailAttr     string `json:"email_attr"`
		TLSMode       string `json:"tls_mode"`
		TLSSkipVerify bool   `json:"tls_skip_verify"`
	}
	if err := c.BodyParser(&req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}

	if req.Port == 0 {
		req.Port = 389
	}
	if req.UserFilter == "" {
		req.UserFilter = "(sAMAccountName=%s)"
	}
	if req.UsernameAttr == "" {
		req.UsernameAttr = "sAMAccountName"
	}
	if req.EmailAttr == "" {
		req.EmailAttr = "mail"
	}
	if req.TLSMode == "" {
		req.TLSMode = "none"
	}

	cfg := &models.LDAPConfig{
		Enabled:       req.Enabled,
		Host:          req.Host,
		Port:          req.Port,
		BindDN:        req.BindDN,
		BaseDN:        req.BaseDN,
		UserFilter:    req.UserFilter,
		UsernameAttr:  req.UsernameAttr,
		EmailAttr:     req.EmailAttr,
		TLSMode:       req.TLSMode,
		TLSSkipVerify: req.TLSSkipVerify,
	}
	// If caller sends "****" (masked), preserve the existing password
	if req.BindPassword != "" && req.BindPassword != "****" {
		cfg.BindPasswordEnc = []byte(req.BindPassword)
	}

	if err := h.service.LDAP.SaveConfig(c.Context(), cfg); err != nil {
		reqLogger(c).Error("UpdateLDAPConfig error", "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to save LDAP config")
	}

	uid, uname := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "update",
		ResourceType: "ldap_config",
	})
	return c.SendStatus(fiber.StatusNoContent)
}

// TestLDAPConnection handles POST /api/v1/admin/auth/ldap/test.
func (h *Handler) TestLDAPConnection(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminWrite); err != nil {
		return nil
	}

	if err := h.service.LDAP.TestConnection(c.Context()); err != nil {
		reqLogger(c).Error("LDAP connection test failed", "error", err)
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"ok": false, "error": "LDAP connection failed"})
	}
	return c.JSON(fiber.Map{"ok": true})
}

// ListLDAPGroupMappings handles GET /api/v1/admin/auth/ldap/group-mappings.
func (h *Handler) ListLDAPGroupMappings(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminWrite); err != nil {
		return nil
	}

	mappings, err := h.service.GetRepository().GetLDAPGroupMappings(c.Context())
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to load group mappings")
	}
	if mappings == nil {
		mappings = []*models.LDAPGroupRoleMapping{}
	}
	return c.JSON(mappings)
}

// CreateLDAPGroupMapping handles POST /api/v1/admin/auth/ldap/group-mappings.
func (h *Handler) CreateLDAPGroupMapping(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminWrite); err != nil {
		return nil
	}

	var req struct {
		LDAPGroupDN string `json:"ldap_group_dn"`
		RoleID      int64  `json:"role_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}
	if req.LDAPGroupDN == "" || req.RoleID == 0 {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "ldap_group_dn and role_id are required")
	}

	m := &models.LDAPGroupRoleMapping{
		LDAPGroupDN: req.LDAPGroupDN,
		RoleID:      req.RoleID,
	}
	if err := h.service.GetRepository().CreateLDAPGroupMapping(c.Context(), m); err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to create group mapping")
	}

	uid, uname := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "create",
		ResourceType: "ldap_group_mapping",
	})
	return c.Status(fiber.StatusCreated).JSON(m)
}

// DeleteLDAPGroupMapping handles DELETE /api/v1/admin/auth/ldap/group-mappings/:id.
func (h *Handler) DeleteLDAPGroupMapping(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminWrite); err != nil {
		return nil
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid mapping ID")
	}

	if err := h.service.GetRepository().DeleteLDAPGroupMapping(c.Context(), int64(id)); err != nil {
		return RespondError(c, fiber.StatusNotFound, ErrNotFound, "mapping not found")
	}

	uid, uname := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "delete",
		ResourceType: "ldap_group_mapping",
	})
	return c.SendStatus(fiber.StatusNoContent)
}

// ============================================================
// Admin OAuth2 config endpoints
// ============================================================

// GetOAuth2Config handles GET /api/v1/admin/auth/oauth2.
func (h *Handler) GetOAuth2Config(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminWrite); err != nil {
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
	if err := h.permCheck(c, services.PermV2AdminWrite); err != nil {
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
		"id":               cfg.ID,
		"enabled":          cfg.Enabled,
		"idp_metadata_url": cfg.IDPMetadataURL,
		"idp_metadata_xml": cfg.IDPMetadataXML,
		"sp_cert_pem":      cfg.SPCertPEM,
		"sp_key_configured": cfg.SPKeyPEM != "",
		"entity_id":        cfg.EntityID,
		"acs_url":          cfg.ACSURL,
		"name_id_format":   cfg.NameIDFormat,
		"created_at":       cfg.CreatedAt,
		"updated_at":       cfg.UpdatedAt,
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

// ============================================================
// Helpers
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

// RegisterExternalAuthRoutes adds the LDAP/OAuth2/SAML routes to the app.
func (h *Handler) RegisterExternalAuthRoutes(auth fiber.Router, admin fiber.Router) {
	// Public authentication endpoints
	auth.Get("/ldap/login", h.LDAPStatus)
	auth.Post("/ldap/login", h.LDAPLogin)
	auth.Get("/oauth2/login", h.OAuth2Login)
	auth.Get("/oauth2/callback", h.OAuth2Callback)
	auth.Get("/saml/metadata", h.SAMLMetadata)
	auth.Get("/saml/login", h.SAMLLogin)
	auth.Post("/saml/acs", h.SAMLAssertionConsumerService)

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
