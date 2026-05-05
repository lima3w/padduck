package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"ipam-next/services"
)

// Register handles POST /api/v1/auth/register
func (h *Handler) Register(c *fiber.Ctx) error {
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if req.Username == "" || req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "username, email, and password are required"})
	}

	result, err := h.service.Registration.Register(c.Context(), services.RegisterRequest{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, services.ErrRegistrationDisabled):
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		case errors.Is(err, services.ErrUsernameTaken),
			errors.Is(err, services.ErrEmailTaken):
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
		case errors.Is(err, services.ErrInvalidUsername),
			errors.Is(err, services.ErrInvalidEmail),
			errors.Is(err, services.ErrPasswordTooShort):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "registration failed"})
		}
	}

	message := "Registration successful. You can now log in."
	switch result.State {
	case "pending_email_verification":
		message = "Registration successful. Please check your email to verify your address."
	case "pending_admin_approval":
		message = "Registration successful. Your account is pending admin approval."
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": message,
		"state":   result.State,
		"user": fiber.Map{
			"id":       result.User.ID,
			"username": result.User.Username,
			"email":    result.User.Email,
		},
	})
}

// VerifyEmail handles GET /api/v1/auth/verify-email
func (h *Handler) VerifyEmail(c *fiber.Ctx) error {
	token := c.Query("token")
	if token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "token is required"})
	}

	if err := h.service.Registration.VerifyEmail(c.Context(), token); err != nil {
		switch {
		case errors.Is(err, services.ErrVerificationInvalid):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid or expired verification token"})
		case errors.Is(err, services.ErrVerificationAlreadyUsed):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "verification token already used"})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "verification failed"})
		}
	}

	return c.JSON(fiber.Map{"message": "Email verified successfully. You can now log in."})
}

// ResendVerification handles POST /api/v1/auth/resend-verification
func (h *Handler) ResendVerification(c *fiber.Ctx) error {
	var req struct {
		Email string `json:"email"`
	}
	if err := c.BodyParser(&req); err != nil || req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "email is required"})
	}

	_ = h.service.Registration.ResendVerification(c.Context(), req.Email)

	// Always return success to avoid email enumeration
	return c.JSON(fiber.Map{"message": "If your email is pending verification, a new link has been sent."})
}
