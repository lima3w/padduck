package handlers

import (
	"github.com/gofiber/fiber/v2"
	"padduck/models"
	"padduck/services"
	"padduck/utils"
)

type RequestPasswordResetRequest struct {
	Email string `json:"email"`
}

type ResetPasswordRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

type PasswordResetResponse struct {
	Message string `json:"message"`
}

// RequestPasswordReset handles POST /api/v1/auth/request-password-reset
func (h *Handler) RequestPasswordReset(c *fiber.Ctx) error {
	req := new(RequestPasswordResetRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "email is required"})
	}

	if err := h.service.SendPasswordResetEmail(c.Context(), req.Email); err != nil {
		reqLogger(c).Error("error sending password reset email", "error", err)
	}

	// Always return success to prevent email enumeration
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "If that email is registered, a reset link has been sent.",
	})
}

// ResetPassword handles POST /api/v1/auth/reset-password
func (h *Handler) ResetPassword(c *fiber.Ctx) error {
	req := new(ResetPasswordRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if req.Token == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "token and password required"})
	}

	if len(req.Password) < 8 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "password must be at least 8 characters"})
	}

	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		reqLogger(c).Error("error hashing password", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to reset password"})
	}

	userID, err := h.service.ResetPasswordWithToken(c.Context(), req.Token, passwordHash)
	if err != nil {
		reqLogger(c).Warn("password reset failed", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid or expired reset token"})
	}

	if userID > 0 {
		_ = h.service.Notification.Queue(c.Context(), userID, services.NotifPasswordChanged, map[string]interface{}{
			"IP": c.IP(),
		})
	}

	return c.JSON(PasswordResetResponse{
		Message: "Password has been reset successfully",
	})
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// ChangeMyPassword handles POST /api/v1/auth/me/change-password
func (h *Handler) ChangeMyPassword(c *fiber.Ctx) error {
	currentUser, ok := c.Locals("user").(*models.User)
	if !ok || currentUser == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	req := new(ChangePasswordRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if req.CurrentPassword == "" || req.NewPassword == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "current_password and new_password are required"})
	}

	if len(req.NewPassword) < 8 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "new password must be at least 8 characters"})
	}

	// Keep the session making the change alive; every other session for the
	// user is revoked by the service.
	if err := h.service.ChangePassword(c.Context(), currentUser.ID, req.CurrentPassword, req.NewPassword, c.Cookies(sessionCookieName)); err != nil {
		if err.Error() == "current password is incorrect" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
		}
		reqLogger(c).Error("error changing password", "user_id", currentUser.ID, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to change password"})
	}

	uid, uname := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "password_changed",
		ResourceType: "user", ResourceID: &currentUser.ID, ResourceName: currentUser.Username,
	})

	_ = h.service.Notification.Queue(c.Context(), currentUser.ID, services.NotifPasswordChanged, map[string]interface{}{
		"IP": c.IP(),
	})

	return c.Status(fiber.StatusNoContent).Send(nil)
}
