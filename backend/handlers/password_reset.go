package handlers

import (
	"github.com/gofiber/fiber/v2"
	"ipam-next/services"
	"ipam-next/utils"
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
