package handlers

import (
	"github.com/gofiber/fiber/v2"
	"ipam-next/models"
	"ipam-next/services"
)

// auditLog writes an audit log entry, injecting IP and user-agent from the request.
func (h *Handler) auditLog(c *fiber.Ctx, entry services.AuditEntry) {
	entry.IPAddress = c.IP()
	entry.UserAgent = c.Get("User-Agent")
	h.service.Audit.Log(c.Context(), entry)
}

// auditUserFromCtx extracts the authenticated user's ID and username from context.
func auditUserFromCtx(c *fiber.Ctx) (userID *int64, username string) {
	if u, ok := c.Locals("user").(*models.User); ok {
		id := u.ID
		return &id, u.Username
	}
	return nil, ""
}
