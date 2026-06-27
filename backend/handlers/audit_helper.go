package handlers

import (
	"github.com/gofiber/fiber/v2"
	"padduck/models"
	"padduck/services"
)

// auditLog writes an audit log entry, injecting IP, user-agent, and orgID from the request.
func (h *Handler) auditLog(c *fiber.Ctx, entry services.AuditEntry) {
	entry.IPAddress = c.IP()
	entry.UserAgent = c.Get("User-Agent")
	if entry.OrgID == nil {
		entry.OrgID = orgIDFromCtx(c)
	}
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

// orgIDFromCtx extracts the caller's organization ID from context (set by AuthMiddleware).
func orgIDFromCtx(c *fiber.Ctx) *int64 {
	v, _ := c.Locals("orgID").(*int64)
	return v
}

// callerID extracts the authenticated user's ID from context.
func callerID(c *fiber.Ctx) *int64 {
	id, _ := auditUserFromCtx(c)
	return id
}

// bypassPolicyFromCtx returns true when the request's API token has bypass_policy=true.
func bypassPolicyFromCtx(c *fiber.Ctx) bool {
	v, _ := c.Locals("bypassPolicy").(bool)
	return v
}
