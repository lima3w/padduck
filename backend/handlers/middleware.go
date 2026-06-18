package handlers

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// AnonymousAPIMiddleware allows unauthenticated read-only access when the
// anonymous_api_enabled config key is "true"; otherwise requires normal auth.
func (h *Handler) AnonymousAPIMiddleware(c *fiber.Ctx) error {
	val, _ := h.service.Config.GetCtx(c.Context(), "anonymous_api_enabled")
	if val == "true" {
		return h.OptionalAuthMiddleware(c)
	}
	return h.AuthMiddleware(c)
}

// AuthMiddleware validates session cookies (web) or Bearer API tokens (scripts).
func (h *Handler) AuthMiddleware(c *fiber.Ctx) error {
	// Try session cookie first (web browser requests)
	if cookieToken := c.Cookies(sessionCookieName); cookieToken != "" {
		user, session, err := h.ops.Identity.ValidateSession(c.Context(), cookieToken)
		if err == nil {
			c.Locals("user", user)
			c.Locals("userID", user.ID)
			c.Locals("sessionID", session.ID)
			return c.Next()
		}
	}

	// Fall back to Bearer token (API clients)
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return RespondError(c, fiber.StatusUnauthorized, ErrUnauthorized, "Missing credentials")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return RespondError(c, fiber.StatusUnauthorized, ErrUnauthorized, "Invalid authorization header format")
	}

	token := parts[1]

	// Bearer tokens are API tokens only
	user, apiToken, err := h.ops.Identity.ValidateAPIToken(c.Context(), token, c.IP())
	if err != nil {
		reqLogger(c).Error("auth error", "error", err)
		return RespondError(c, fiber.StatusUnauthorized, ErrUnauthorized, "Invalid or expired token")
	}

	if user.LastLoginAt != nil && h.ops.Identity.IsSessionExpired(user.LastLoginAt) {
		return RespondError(c, fiber.StatusUnauthorized, ErrUnauthorized, "Session expired, please login again")
	}

	c.Locals("user", user)
	c.Locals("userID", user.ID)

	// Store token scope and enforce rate limit
	if apiToken != nil {
		c.Locals("tokenScope", apiToken.Scope)

		// Rate limit check
		limitStr, _ := h.service.Config.GetCtx(c.Context(), "api_token_rate_limit_per_minute")
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

// RequireBearerAuth rejects requests that authenticated via session cookie instead of a Bearer token.
// Use on endpoints designed for server-to-server API calls (e.g., Grafana datasource) to prevent CSRF.
func (h *Handler) RequireBearerAuth(c *fiber.Ctx) error {
	auth := c.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return RespondError(c, fiber.StatusUnauthorized, ErrUnauthorized, "Bearer token required for this endpoint")
	}
	return c.Next()
}

// OptionalAuthMiddleware validates credentials if present but does not require them.
func (h *Handler) OptionalAuthMiddleware(c *fiber.Ctx) error {
	// Try session cookie first
	if cookieToken := c.Cookies(sessionCookieName); cookieToken != "" {
		user, session, err := h.ops.Identity.ValidateSession(c.Context(), cookieToken)
		if err == nil {
			c.Locals("user", user)
			c.Locals("userID", user.ID)
			c.Locals("sessionID", session.ID)
			return c.Next()
		}
	}

	// Try Bearer API token
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Next()
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return c.Next()
	}

	user, _, err := h.ops.Identity.ValidateAPIToken(c.Context(), parts[1], c.IP())
	if err == nil {
		c.Locals("user", user)
		c.Locals("userID", user.ID)
	}

	return c.Next()
}
