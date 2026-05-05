package handlers

import (
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// AuthMiddleware validates API tokens on protected routes
func (h *Handler) AuthMiddleware(c *fiber.Ctx) error {
	// Extract token from Authorization header
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return RespondError(c, fiber.StatusUnauthorized, ErrUnauthorized, "Missing authorization header")
	}

	// Token format: "Bearer <token>"
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return RespondError(c, fiber.StatusUnauthorized, ErrUnauthorized, "Invalid authorization header format")
	}

	token := parts[1]

	// Validate token and get user
	user, err := h.service.ValidateAPIToken(c.Context(), token)
	if err != nil {
		log.Printf("Token validation error: %v", err)
		return RespondError(c, fiber.StatusUnauthorized, ErrUnauthorized, "Invalid or expired token", err.Error())
	}

	// Store user in context for handlers to access
	c.Locals("user", user)
	c.Locals("userID", user.ID)

	return c.Next()
}

// OptionalAuthMiddleware validates API tokens if present, but doesn't require them
func (h *Handler) OptionalAuthMiddleware(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Next()
	}

	// Token format: "Bearer <token>"
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return c.Next()
	}

	token := parts[1]

	// Validate token and get user
	user, err := h.service.ValidateAPIToken(c.Context(), token)
	if err == nil {
		c.Locals("user", user)
		c.Locals("userID", user.ID)
	}

	return c.Next()
}
