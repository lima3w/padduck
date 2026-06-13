package handlers

import (
	"github.com/gofiber/fiber/v2"
	"padduck/models"
	"padduck/services"
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
