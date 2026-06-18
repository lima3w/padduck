package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"padduck/services"
)

// Register handles POST /api/v1/auth/register
func (h *Handler) Register(c *fiber.Ctx) error {
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}

	if req.Username == "" || req.Email == "" || req.Password == "" {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "username, email, and password are required")
	}

	result, err := h.auth.Registration.Register(c.Context(), services.RegisterRequest{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, services.ErrRegistrationDisabled):
			return RespondError(c, fiber.StatusForbidden, ErrForbidden, err.Error())
		case errors.Is(err, services.ErrUsernameTaken),
			errors.Is(err, services.ErrEmailTaken):
			return RespondError(c, fiber.StatusConflict, ErrConflict, "an account with those details already exists")
		case errors.Is(err, services.ErrInvalidUsername),
			errors.Is(err, services.ErrInvalidEmail),
			errors.Is(err, services.ErrPasswordTooShort):
			return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, err.Error())
		default:
			return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "registration failed")
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
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "token is required")
	}

	if err := h.auth.Registration.VerifyEmail(c.Context(), token); err != nil {
		switch {
		case errors.Is(err, services.ErrVerificationInvalid):
			return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid or expired verification token")
		case errors.Is(err, services.ErrVerificationAlreadyUsed):
			return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "verification token already used")
		default:
			return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "verification failed")
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
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "email is required")
	}

	_ = h.auth.Registration.ResendVerification(c.Context(), req.Email)

	// Always return success to avoid email enumeration
	return c.JSON(fiber.Map{"message": "If your email is pending verification, a new link has been sent."})
}
