package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"ipam-next/models"
	"ipam-next/services"
)

// VerifyMFA handles POST /api/v1/auth/verify-mfa
// Completes an MFA challenge and returns a full session token.
func (h *Handler) VerifyMFA(c *fiber.Ctx) error {
	var req struct {
		Challenge string `json:"mfa_challenge"`
		Code      string `json:"code"`
	}
	if err := c.BodyParser(&req); err != nil || req.Challenge == "" || req.Code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "mfa_challenge and code are required"})
	}

	userID, err := h.service.MFA.CompleteChallenge(c.Context(), req.Challenge, req.Code)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidChallenge), errors.Is(err, services.ErrChallengeExpired):
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid or expired MFA challenge"})
		case errors.Is(err, services.ErrChallengeCompleted):
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "MFA challenge already used"})
		case errors.Is(err, services.ErrInvalidTOTPCode):
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid MFA code"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "MFA verification failed"})
	}

	user, err := h.service.GetUserByID(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "user not found"})
	}

	if err := h.service.UpdateLastLogin(c.Context(), user.ID); err != nil {
		reqLogger(c).Warn("error updating last login", "user_id", user.ID, "error", err)
	}

	token, err := h.service.CreateWebSession(c.Context(), user.ID, c.IP(), c.Get("User-Agent"))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create session"})
	}

	uid := user.ID
	h.auditLog(c, services.AuditEntry{
		UserID: &uid, Username: user.Username, Action: "login_mfa",
		ResourceType: "session", Status: "success",
	})

	h.setSessionCookie(c, token)
	return c.JSON(LoginResponse{
		User: UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			Role:      user.Role,
			State:     user.State,
			CreatedAt: user.CreatedAt.String(),
			UpdatedAt: user.UpdatedAt.String(),
		},
	})
}

// GetMFAStatus handles GET /api/v1/auth/me/mfa
func (h *Handler) GetMFAStatus(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	enabled, backupRemaining := h.service.MFA.GetMFAStatus(c.Context(), user.ID)
	return c.JSON(fiber.Map{
		"totp_enabled":      enabled,
		"backup_codes_left": backupRemaining,
	})
}

// SetupTOTP handles POST /api/v1/auth/me/mfa/setup
// Returns the QR code and secret for the user to scan.
func (h *Handler) SetupTOTP(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)

	secret, qrDataURL, err := h.service.MFA.SetupTOTP(c.Context(), user.ID, user.Username, user.Email)
	if err != nil {
		reqLogger(c).Error("TOTP setup error", "user_id", user.ID, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to set up MFA"})
	}

	return c.JSON(fiber.Map{
		"secret":  secret,
		"qr_code": qrDataURL,
		"message": "Scan the QR code with your authenticator app, then confirm with a 6-digit code.",
	})
}

// ConfirmTOTP handles POST /api/v1/auth/me/mfa/confirm
// Verifies the first code and enables TOTP; returns backup codes.
func (h *Handler) ConfirmTOTP(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)

	var req struct {
		Code string `json:"code"`
	}
	if err := c.BodyParser(&req); err != nil || req.Code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "code is required"})
	}

	backupCodes, err := h.service.MFA.ConfirmTOTP(c.Context(), user.ID, req.Code)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrMFANotSetup):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "MFA setup not started — call /mfa/setup first"})
		case errors.Is(err, services.ErrInvalidTOTPCode):
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid code"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to confirm MFA"})
	}

	uid := user.ID
	h.auditLog(c, services.AuditEntry{
		UserID: &uid, Username: user.Username, Action: "mfa_enabled", ResourceType: "user", ResourceID: &uid,
	})

	_ = h.service.Notification.Queue(c.Context(), user.ID, services.NotifMFAEnabled, map[string]interface{}{
		"IP":     c.IP(),
		"Device": c.Get("User-Agent"),
	})

	return c.JSON(fiber.Map{
		"message":      "MFA enabled. Save these backup codes — they will not be shown again.",
		"backup_codes": backupCodes,
	})
}

// DisableTOTP handles DELETE /api/v1/auth/me/mfa
func (h *Handler) DisableTOTP(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)

	var req struct {
		Code string `json:"code"`
	}
	if err := c.BodyParser(&req); err != nil || req.Code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "code is required to disable MFA"})
	}

	if err := h.service.MFA.DisableTOTP(c.Context(), user.ID, req.Code); err != nil {
		switch {
		case errors.Is(err, services.ErrMFANotEnabled):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "MFA is not enabled"})
		case errors.Is(err, services.ErrInvalidTOTPCode):
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid code"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to disable MFA"})
	}

	uid := user.ID
	h.auditLog(c, services.AuditEntry{
		UserID: &uid, Username: user.Username, Action: "mfa_disabled", ResourceType: "user", ResourceID: &uid,
	})

	_ = h.service.Notification.Queue(c.Context(), user.ID, services.NotifMFADisabled, map[string]interface{}{
		"IP":     c.IP(),
		"Device": c.Get("User-Agent"),
	})

	return c.JSON(fiber.Map{"message": "MFA disabled"})
}

// RegenerateBackupCodes handles POST /api/v1/auth/me/mfa/backup-codes
func (h *Handler) RegenerateBackupCodes(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)

	var req struct {
		Code string `json:"code"`
	}
	if err := c.BodyParser(&req); err != nil || req.Code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "current MFA code is required"})
	}

	codes, err := h.service.MFA.RegenerateBackupCodes(c.Context(), user.ID, req.Code)
	if err != nil {
		if errors.Is(err, services.ErrInvalidTOTPCode) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid code"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to regenerate codes"})
	}

	uid := user.ID
	h.auditLog(c, services.AuditEntry{
		UserID: &uid, Username: user.Username, Action: "backup_codes_regenerated", ResourceType: "user", ResourceID: &uid,
	})

	return c.JSON(fiber.Map{
		"message":      "Backup codes regenerated. Save these — they will not be shown again.",
		"backup_codes": codes,
	})
}
