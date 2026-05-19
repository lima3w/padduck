package handlers

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"ipam-next/services"
)

// GetIdentityPolicies handles GET /api/v1/admin/identity-policies
// Returns current identity and session policy configuration.
func (h *Handler) GetIdentityPolicies(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminWrite); err != nil {
		return nil
	}

	ctx := c.Context()

	enforceMFA := false
	if v, err := h.service.Config.GetCtx(ctx, "enforce_mfa"); err == nil && v == "true" {
		enforceMFA = true
	}

	sessionMaxAgeHours := 24
	if v, err := h.service.Config.GetCtx(ctx, "session_max_age_hours"); err == nil && v != "" {
		var n int
		if _, err2 := fmt.Sscanf(v, "%d", &n); err2 == nil && n > 0 {
			sessionMaxAgeHours = n
		}
	}

	apiTokenMaxAgeDays := 0
	if v, err := h.service.Config.GetCtx(ctx, "api_token_max_age_days"); err == nil && v != "" {
		var n int
		if _, err2 := fmt.Sscanf(v, "%d", &n); err2 == nil && n >= 0 {
			apiTokenMaxAgeDays = n
		}
	}

	inactiveUserDays := 0
	if v, err := h.service.Config.GetCtx(ctx, "inactive_user_days"); err == nil && v != "" {
		var n int
		if _, err2 := fmt.Sscanf(v, "%d", &n); err2 == nil && n >= 0 {
			inactiveUserDays = n
		}
	}

	return c.JSON(fiber.Map{
		"enforce_mfa":             enforceMFA,
		"session_max_age_hours":   sessionMaxAgeHours,
		"api_token_max_age_days":  apiTokenMaxAgeDays,
		"inactive_user_days":      inactiveUserDays,
	})
}

// UpdateIdentityPolicies handles PUT /api/v1/admin/identity-policies
// Validates and saves identity/session policy configuration.
func (h *Handler) UpdateIdentityPolicies(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminWrite); err != nil {
		return nil
	}

	var req struct {
		EnforceMFA            bool `json:"enforce_mfa"`
		SessionMaxAgeHours    int  `json:"session_max_age_hours"`
		APITokenMaxAgeDays    int  `json:"api_token_max_age_days"`
		InactiveUserDays      int  `json:"inactive_user_days"`
	}
	if err := c.BodyParser(&req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}

	// Validate ranges
	if req.SessionMaxAgeHours < 1 || req.SessionMaxAgeHours > 8760 {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "session_max_age_hours must be between 1 and 8760")
	}
	if req.APITokenMaxAgeDays < 0 || req.APITokenMaxAgeDays > 3650 {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "api_token_max_age_days must be between 0 and 3650")
	}
	if req.InactiveUserDays < 0 || req.InactiveUserDays > 3650 {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "inactive_user_days must be between 0 and 3650")
	}

	enforceMFAStr := "false"
	if req.EnforceMFA {
		enforceMFAStr = "true"
	}

	err := h.service.Config.SetMultiple(map[string]string{
		"enforce_mfa":            enforceMFAStr,
		"session_max_age_hours":  fmt.Sprintf("%d", req.SessionMaxAgeHours),
		"api_token_max_age_days": fmt.Sprintf("%d", req.APITokenMaxAgeDays),
		"inactive_user_days":     fmt.Sprintf("%d", req.InactiveUserDays),
	})
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to save identity policies", err.Error())
	}

	return c.JSON(fiber.Map{
		"enforce_mfa":             req.EnforceMFA,
		"session_max_age_hours":   req.SessionMaxAgeHours,
		"api_token_max_age_days":  req.APITokenMaxAgeDays,
		"inactive_user_days":      req.InactiveUserDays,
	})
}

// ListSessionRisk handles GET /api/v1/admin/session-risk
// Returns all active sessions with risk signals.
func (h *Handler) ListSessionRisk(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminWrite); err != nil {
		return nil
	}

	sessions, err := h.service.GetRepository().ListAllActiveSessions(c.Context())
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to list active sessions", err.Error())
	}

	type sessionRiskItem struct {
		UserID          int64      `json:"user_id"`
		Username        string     `json:"username"`
		IPAddress       string     `json:"ip_address"`
		LastUsedAt      time.Time  `json:"last_used_at"`
		CreatedAt       time.Time  `json:"created_at"`
		IsImpersonation bool       `json:"is_impersonation"`
		RiskFlags       []string   `json:"risk_flags"`
	}

	now := time.Now()
	sevenDaysAgo := now.Add(-7 * 24 * time.Hour)

	result := make([]sessionRiskItem, 0, len(sessions))
	for _, s := range sessions {
		var flags []string
		if s.IsImpersonation {
			flags = append(flags, "impersonation")
		}
		if s.CreatedAt.Before(sevenDaysAgo) {
			flags = append(flags, "long_lived")
		}
		if flags == nil {
			flags = []string{}
		}

		result = append(result, sessionRiskItem{
			UserID:          s.UserID,
			Username:        s.Username,
			IPAddress:       s.IPAddress,
			LastUsedAt:      s.LastUsedAt,
			CreatedAt:       s.CreatedAt,
			IsImpersonation: s.IsImpersonation,
			RiskFlags:       flags,
		})
	}

	return c.JSON(fiber.Map{"sessions": result})
}
