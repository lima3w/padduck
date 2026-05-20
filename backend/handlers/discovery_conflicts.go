package handlers

import (
	"github.com/gofiber/fiber/v2"
	"padduck/services"
)

// ListDiscoveryConflicts handles GET /api/v1/admin/discovery/conflicts?status=pending
func (h *Handler) ListDiscoveryConflicts(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminRead); err != nil {
		return nil
	}
	status := c.Query("status")
	conflicts, err := h.service.Discovery.ListDiscoveryConflicts(c.Context(), status)
	if err != nil {
		reqLogger(c).Error("error listing discovery conflicts", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	if conflicts == nil {
		return c.JSON([]interface{}{})
	}
	return c.JSON(conflicts)
}

// GetDiscoveryConflict handles GET /api/v1/admin/discovery/conflicts/:id
func (h *Handler) GetDiscoveryConflict(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminRead); err != nil {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid conflict ID"})
	}
	conflict, err := h.service.Discovery.GetDiscoveryConflict(c.Context(), int64(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "conflict not found"})
	}
	return c.JSON(conflict)
}

// ResolveDiscoveryConflict handles POST /api/v1/admin/discovery/conflicts/:id/resolve
// Body: {"action":"accepted"|"rejected"}
func (h *Handler) ResolveDiscoveryConflict(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminWrite); err != nil {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid conflict ID"})
	}

	var req struct {
		Action string `json:"action"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	reviewedBy := "operator"
	if email, ok := c.Locals("userEmail").(string); ok && email != "" {
		reviewedBy = email
	}

	conflict, err := h.service.Discovery.ResolveDiscoveryConflict(c.Context(), int64(id), req.Action, reviewedBy)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(conflict)
}
