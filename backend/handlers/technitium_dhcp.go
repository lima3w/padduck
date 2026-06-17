package handlers

import (
	"github.com/gofiber/fiber/v2"
	"padduck/services"
)

// ListTechnitiumDHCPScopes handles GET /api/v1/admin/dhcp/technitium/scopes
func (h *Handler) ListTechnitiumDHCPScopes(c *fiber.Ctx) error {
	if err := requireAdmin(c); err != nil {
		return nil
	}
	scopes, err := h.service.ListTechnitiumDHCPScopes(c.Context())
	if err != nil {
		return RespondError(c, fiber.StatusBadGateway, ErrBadGateway, err.Error())
	}
	return c.JSON(fiber.Map{"scopes": scopes})
}

// SyncTechnitiumLeases handles POST /api/v1/admin/dhcp/technitium/sync
func (h *Handler) SyncTechnitiumLeases(c *fiber.Ctx) error {
	if err := requireAdmin(c); err != nil {
		return nil
	}
	count, err := h.service.SyncTechnitiumLeases(c.Context())
	if err != nil {
		return RespondError(c, fiber.StatusBadGateway, ErrBadGateway, err.Error())
	}
	adminID, adminName := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: adminID, Username: adminName, Action: "technitium_dhcp_sync",
		ResourceType: "dhcp_lease",
	})
	return c.JSON(fiber.Map{"synced": count})
}

// ImportTechnitiumScope handles POST /api/v1/admin/dhcp/technitium/import-scope
func (h *Handler) ImportTechnitiumScope(c *fiber.Ctx) error {
	if err := requireAdmin(c); err != nil {
		return nil
	}
	var req struct {
		ScopeName string `json:"scope_name"`
		NetworkID int64  `json:"network_id"`
	}
	if err := c.BodyParser(&req); err != nil || req.ScopeName == "" || req.NetworkID == 0 {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "scope_name and network_id are required")
	}
	subnet, err := h.service.ImportTechnitiumScope(c.Context(), req.ScopeName, req.NetworkID)
	if err != nil {
		return RespondError(c, fiber.StatusBadGateway, ErrBadGateway, err.Error())
	}
	adminID, adminName := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: adminID, Username: adminName, Action: "technitium_scope_imported",
		ResourceType: "subnet", ResourceID: &subnet.ID,
	})
	return c.Status(fiber.StatusCreated).JSON(subnet)
}

// PushDHCPReservation handles POST /api/v1/ip-addresses/:id/dhcp-reservation
func (h *Handler) PushDHCPReservation(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid IP address ID")
	}
	if !h.requirePerm(c, services.PermV2IPAssign) {
		return nil
	}
	if err := h.service.PushDHCPReservation(c.Context(), id); err != nil {
		return RespondError(c, fiber.StatusBadGateway, ErrBadGateway, err.Error())
	}
	return c.JSON(fiber.Map{"message": "reservation created"})
}

// RemoveDHCPReservation handles DELETE /api/v1/ip-addresses/:id/dhcp-reservation
func (h *Handler) RemoveDHCPReservation(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid IP address ID")
	}
	if !h.requirePerm(c, services.PermV2IPAssign) {
		return nil
	}
	if err := h.service.RemoveDHCPReservation(c.Context(), id); err != nil {
		return RespondError(c, fiber.StatusBadGateway, ErrBadGateway, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}
