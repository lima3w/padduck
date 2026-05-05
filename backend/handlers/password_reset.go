package handlers

import (
	"log"

	"github.com/gofiber/fiber/v2"
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

	// This is a placeholder - in production, you'd send an email with reset link
	// For now, we'll generate and return the token (this should be sent via email)
	token, err := h.service.CreatePasswordResetToken(c.Context(), req.Email)
	if err != nil {
		log.Printf("Error creating password reset token: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to request password reset"})
	}

	// In production, send this token via email instead of returning it
	// For development, we return it so it can be tested
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Password reset link has been sent to your email",
		"token":   token, // Remove this in production - only for testing
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
		log.Printf("Error hashing password: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to reset password"})
	}

	if err := h.service.ResetPasswordWithToken(c.Context(), req.Token, passwordHash); err != nil {
		log.Printf("Error resetting password: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid or expired reset token"})
	}

	return c.JSON(PasswordResetResponse{
		Message: "Password has been reset successfully",
	})
}
