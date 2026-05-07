package handlers

import (
	"log"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// AuthMiddleware validates session tokens (or legacy API tokens) on protected routes.
func (h *Handler) AuthMiddleware(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return RespondError(c, fiber.StatusUnauthorized, ErrUnauthorized, "Missing authorization header")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return RespondError(c, fiber.StatusUnauthorized, ErrUnauthorized, "Invalid authorization header format")
	}

	token := parts[1]

	// Try session-based auth first
	user, session, err := h.service.ValidateSession(c.Context(), token)
	if err == nil {
		c.Locals("user", user)
		c.Locals("userID", user.ID)
		c.Locals("sessionID", session.ID)
		return c.Next()
	}

	// Fall back to API token auth
	user, apiToken, err := h.service.ValidateAPIToken(c.Context(), token, c.IP())
	if err != nil {
		log.Printf("Auth error: %v", err)
		return RespondError(c, fiber.StatusUnauthorized, ErrUnauthorized, "Invalid or expired token")
	}

	if user.LastLoginAt != nil && h.service.IsSessionExpired(user.LastLoginAt) {
		return RespondError(c, fiber.StatusUnauthorized, ErrUnauthorized, "Session expired, please login again")
	}

	c.Locals("user", user)
	c.Locals("userID", user.ID)

	// Store token scope and enforce rate limit
	if apiToken != nil {
		c.Locals("tokenScope", apiToken.Scope)

		// Rate limit check
		limitStr, _ := h.service.Config.Get("api_token_rate_limit_per_minute")
		limit := 100
		if n, err2 := strconv.Atoi(limitStr); err2 == nil && n >= 0 {
			limit = n
		}
		if !h.tokenLimiter.Allow(apiToken.ID, limit) {
			c.Set("Retry-After", "60")
			return RespondError(c, fiber.StatusTooManyRequests, ErrServiceUnavailable, "rate limit exceeded")
		}

		// Enforce token scope
		if scope, ok := c.Locals("tokenScope").(string); ok {
			path := c.Path()
			method := c.Method()
			if scope == "read" && method != "GET" && method != "HEAD" && method != "OPTIONS" {
				return RespondError(c, fiber.StatusForbidden, ErrForbidden, "token is read-only")
			}
			if scope == "write" && strings.HasPrefix(path, "/api/v1/admin") {
				return RespondError(c, fiber.StatusForbidden, ErrForbidden, "token scope does not allow admin operations")
			}
			// "admin" scope: allow everything
		}
	}

	return c.Next()
}

// OptionalAuthMiddleware validates tokens if present but does not require them.
func (h *Handler) OptionalAuthMiddleware(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Next()
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return c.Next()
	}

	token := parts[1]

	user, session, err := h.service.ValidateSession(c.Context(), token)
	if err == nil {
		c.Locals("user", user)
		c.Locals("userID", user.ID)
		c.Locals("sessionID", session.ID)
		return c.Next()
	}

	user, _, err = h.service.ValidateAPIToken(c.Context(), token, c.IP())
	if err == nil {
		c.Locals("user", user)
		c.Locals("userID", user.ID)
	}

	return c.Next()
}
