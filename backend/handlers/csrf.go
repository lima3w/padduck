package handlers

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/gofiber/fiber/v2"
)

const (
	CSRFHeaderName = "X-CSRF-Token"
	CSRFCookieName = "csrf-token"
)

// newCSRFToken returns a signed CSRF token: "<64-hex-random>.<32-hex-HMAC>".
// The HMAC prevents an attacker with cookie-injection capability from forging
// a token that passes server-side validation (signed double-submit cookie pattern).
func (h *Handler) newCSRFToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	rawHex := hex.EncodeToString(raw)
	return rawHex + "." + h.csrfMAC(rawHex), nil
}

// csrfMAC returns a 32-char hex HMAC-SHA256 of data using the handler's CSRF secret.
func (h *Handler) csrfMAC(data string) string {
	m := hmac.New(sha256.New, h.csrfSecret)
	m.Write([]byte(data))
	return hex.EncodeToString(m.Sum(nil))[:32]
}

// validCSRFToken checks both the double-submit invariant (cookie == header) and
// the HMAC signature on the random component.
func (h *Handler) validCSRFToken(cookie, header string) bool {
	if cookie == "" || header == "" {
		return false
	}
	// Constant-time comparison prevents timing attacks on the full token.
	if !hmac.Equal([]byte(cookie), []byte(header)) {
		return false
	}
	parts := strings.SplitN(cookie, ".", 2)
	if len(parts) != 2 {
		return false
	}
	expected := h.csrfMAC(parts[0])
	return hmac.Equal([]byte(parts[1]), []byte(expected))
}

// CSRFMiddleware adds CSRF protection to state-mutating requests (POST/PUT/DELETE/PATCH).
// On safe methods (GET/HEAD/OPTIONS) it issues a signed cookie only if the client
// has none yet — this prevents rotation attacks via attacker-controlled GET requests.
func (h *Handler) CSRFMiddleware(c *fiber.Ctx) error {
	method := c.Method()

	if method == "GET" || method == "HEAD" || method == "OPTIONS" {
		if c.Cookies(CSRFCookieName) == "" {
			if token, err := h.newCSRFToken(); err == nil {
				c.Cookie(&fiber.Cookie{
					Name:     CSRFCookieName,
					Value:    token,
					MaxAge:   3600,
					Secure:   h.isProduction,
					HTTPOnly: false,
					SameSite: "Strict",
				})
			}
		}
		return c.Next()
	}

	if !h.validCSRFToken(c.Cookies(CSRFCookieName), c.Get(CSRFHeaderName)) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "csrf validation failed"})
	}

	return c.Next()
}

// GetCSRFToken handles GET /api/v1/csrf-token — issues a fresh signed token and
// sets it in the cookie so the client's JS can read it for subsequent requests.
func (h *Handler) GetCSRFToken(c *fiber.Ctx) error {
	token, err := h.newCSRFToken()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to generate token"})
	}

	c.Cookie(&fiber.Cookie{
		Name:     CSRFCookieName,
		Value:    token,
		MaxAge:   3600,
		Secure:   h.isProduction,
		HTTPOnly: false,
		SameSite: "Strict",
	})

	return c.JSON(fiber.Map{"csrf_token": token})
}
