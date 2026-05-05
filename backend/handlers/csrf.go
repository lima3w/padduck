package handlers

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/gofiber/fiber/v2"
)

const (
	CSRFHeaderName = "X-CSRF-Token"
	CSRFCookieName = "csrf-token"
)

// GenerateCSRFToken generates a new CSRF token
func GenerateCSRFToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// CSRFMiddleware adds CSRF protection to POST, PUT, DELETE requests
func (h *Handler) CSRFMiddleware(c *fiber.Ctx) error {
	method := c.Method()

	// Skip CSRF check for GET, HEAD, OPTIONS
	if method == "GET" || method == "HEAD" || method == "OPTIONS" {
		token, err := GenerateCSRFToken()
		if err == nil {
			c.Cookie(&fiber.Cookie{
				Name:     CSRFCookieName,
				Value:    token,
				MaxAge:   3600,
				Secure:   false,
				HTTPOnly: false,
				SameSite: "Strict",
			})
		}
		return c.Next()
	}

	// Get CSRF token from cookie (this is the server's token)
	cookieToken := c.Cookies(CSRFCookieName)

	// Get CSRF token from header (this is from the client)
	headerToken := c.Get(CSRFHeaderName)

	// If there's no cookie token, generate one
	if cookieToken == "" {
		token, err := GenerateCSRFToken()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to generate csrf token"})
		}
		c.Cookie(&fiber.Cookie{
			Name:     CSRFCookieName,
			Value:    token,
			MaxAge:   3600,
			Secure:   false,
			HTTPOnly: false,
			SameSite: "Strict",
		})
		cookieToken = token
	}

	// Validate that header token matches cookie token
	if headerToken == "" || headerToken != cookieToken {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "csrf validation failed"})
	}

	return c.Next()
}

// GetCSRFToken handles GET /api/v1/csrf-token to get a CSRF token
func (h *Handler) GetCSRFToken(c *fiber.Ctx) error {
	token, err := GenerateCSRFToken()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to generate token"})
	}

	c.Cookie(&fiber.Cookie{
		Name:     CSRFCookieName,
		Value:    token,
		MaxAge:   3600,
		Secure:   false,
		HTTPOnly: false,
		SameSite: "Strict",
	})

	return c.JSON(fiber.Map{"csrf_token": token})
}
