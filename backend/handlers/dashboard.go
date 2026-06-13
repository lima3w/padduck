package handlers

import (
	"github.com/gofiber/fiber/v2"
	"padduck/services"
)

// GetDashboardSummary handles GET /api/v1/dashboard/summary
func (h *Handler) GetDashboardSummary(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2NetworkList); err != nil {
		return nil
	}

	summary, err := h.service.GetDashboardSummary(c.Context())
	if err != nil {
		reqLogger(c).Error("GetDashboardSummary failed", "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}

	return c.JSON(summary)
}

// GetDashboardRecentActivity handles GET /api/v1/dashboard/recent-activity
func (h *Handler) GetDashboardRecentActivity(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2NetworkList); err != nil {
		return nil
	}

	activities, err := h.service.GetDashboardRecentActivity(c.Context())
	if err != nil {
		reqLogger(c).Error("GetDashboardRecentActivity failed", "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}

	return c.JSON(activities)
}

// GetSubnetTree handles GET /api/v1/networks/:id/subnets/tree
func (h *Handler) GetSubnetTree(c *fiber.Ctx) error {
	networkID, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid section ID")
	}
	if err := h.permCheck(c, services.PermV2SubnetList, services.ResourceScope{Type: "section", ID: int64(networkID)}); err != nil {
		return nil
	}

	tree, err := h.service.GetSubnetTree(c.Context(), int64(networkID))
	if err != nil {
		reqLogger(c).Error("GetSubnetTree failed", "error", err, "network_id", networkID)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}

	return c.JSON(tree)
}
