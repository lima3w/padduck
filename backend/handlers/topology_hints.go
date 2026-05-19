package handlers

import (
	"github.com/gofiber/fiber/v2"
	"ipam-next/models"
	"ipam-next/services"
)

// ListTopologyHints handles GET /api/v1/admin/topology/hints?status=suggested
func (h *Handler) ListTopologyHints(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminRead); err != nil {
		return nil
	}
	status := c.Query("status")
	hints, err := h.service.Topology.ListHints(c.Context(), status)
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to list topology hints", err.Error())
	}
	if hints == nil {
		hints = []*models.TopologyHint{}
	}
	return c.JSON(fiber.Map{"hints": hints, "total": len(hints)})
}

// GetTopologyHint handles GET /api/v1/admin/topology/hints/:id
func (h *Handler) GetTopologyHint(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminRead); err != nil {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid hint ID")
	}
	hint, err := h.service.Topology.GetHint(c.Context(), int64(id))
	if err != nil {
		return RespondError(c, fiber.StatusNotFound, ErrNotFound, "topology hint not found")
	}
	return c.JSON(fiber.Map{"hint": hint})
}

// UpdateTopologyHintStatus handles PUT /api/v1/admin/topology/hints/:id/status
// Body: {"status":"confirmed"|"dismissed"}
func (h *Handler) UpdateTopologyHintStatus(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminWrite); err != nil {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid hint ID")
	}
	var req struct {
		Status string `json:"status"`
	}
	if err := c.BodyParser(&req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}
	hint, err := h.service.Topology.UpdateHintStatus(c.Context(), int64(id), req.Status)
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, err.Error())
	}
	return c.JSON(fiber.Map{"hint": hint})
}
