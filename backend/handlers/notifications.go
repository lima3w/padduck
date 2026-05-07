package handlers

import (
	"github.com/gofiber/fiber/v2"
	"ipam-next/models"
)

// GetNotificationPreferences handles GET /api/v1/user/notification-preferences
func (h *Handler) GetNotificationPreferences(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*models.User)
	if !ok {
		return RespondError(c, fiber.StatusUnauthorized, ErrUnauthorized, "not authenticated")
	}
	prefs, err := h.service.GetNotificationPreferences(c.Context(), user.ID)
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to load preferences")
	}
	return c.JSON(prefs)
}

type UpdateNotificationPrefsRequest struct {
	LoginSuccess    *bool `json:"login_success"`
	LoginFailed     *bool `json:"login_failed"`
	AccountLocked   *bool `json:"account_locked"`
	PasswordChanged *bool `json:"password_changed"`
	MFAChanges      *bool `json:"mfa_changes"`
	APITokenChanges *bool `json:"api_token_changes"`
	RoleChanges     *bool `json:"role_changes"`
	SessionRevoked  *bool `json:"session_revoked"`
}

// UpdateNotificationPreferences handles PUT /api/v1/user/notification-preferences
func (h *Handler) UpdateNotificationPreferences(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*models.User)
	if !ok {
		return RespondError(c, fiber.StatusUnauthorized, ErrUnauthorized, "not authenticated")
	}

	req := new(UpdateNotificationPrefsRequest)
	if err := c.BodyParser(req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}

	// Load current prefs (defaults all-true if not yet set)
	current, err := h.service.GetNotificationPreferences(c.Context(), user.ID)
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to load preferences")
	}

	// Apply only the fields the caller provided
	if req.LoginSuccess != nil {
		current.LoginSuccess = *req.LoginSuccess
	}
	if req.LoginFailed != nil {
		current.LoginFailed = *req.LoginFailed
	}
	if req.AccountLocked != nil {
		current.AccountLocked = *req.AccountLocked
	}
	if req.PasswordChanged != nil {
		current.PasswordChanged = *req.PasswordChanged
	}
	if req.MFAChanges != nil {
		current.MFAChanges = *req.MFAChanges
	}
	if req.APITokenChanges != nil {
		current.APITokenChanges = *req.APITokenChanges
	}
	if req.RoleChanges != nil {
		current.RoleChanges = *req.RoleChanges
	}
	if req.SessionRevoked != nil {
		current.SessionRevoked = *req.SessionRevoked
	}

	updated, err := h.service.UpsertNotificationPreferences(c.Context(), current)
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to save preferences")
	}
	return c.JSON(updated)
}

// GetNotificationStats handles GET /api/v1/admin/notification-stats
func (h *Handler) GetNotificationStats(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*models.User)
	if !ok || user.Role != "admin" {
		return RespondError(c, fiber.StatusForbidden, ErrForbidden, "admin access required")
	}
	stats, err := h.service.GetNotificationStats(c.Context())
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to load stats")
	}
	return c.JSON(stats)
}
