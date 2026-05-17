package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// sanitizeRequestID strips any characters from s that are not alphanumeric,
// hyphens, or underscores, and truncates the result to 64 characters.
// This prevents log injection via a malicious X-Request-ID header.
func sanitizeRequestID(s string) string {
	const max = 64
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			b.WriteRune(r)
			if b.Len() >= max {
				break
			}
		}
	}
	return b.String()
}

// RequestIDMiddleware generates a unique request ID for each request,
// inheriting X-Request-ID from the client if provided.
func RequestIDMiddleware(c *fiber.Ctx) error {
	rid := sanitizeRequestID(c.Get("X-Request-ID"))
	if rid == "" {
		b := make([]byte, 8)
		if _, err := rand.Read(b); err == nil {
			rid = hex.EncodeToString(b)
		}
	}
	c.Locals("requestID", rid)
	c.Set("X-Request-ID", rid)
	return c.Next()
}

// reqLogger returns a slog.Logger with request_id and user_id attached from the Fiber context.
func reqLogger(c *fiber.Ctx) *slog.Logger {
	attrs := []any{"request_id", c.Locals("requestID")}
	if uid := c.Locals("userID"); uid != nil {
		attrs = append(attrs, "user_id", uid)
	}
	return slog.Default().With(attrs...)
}
