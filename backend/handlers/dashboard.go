package handlers

import (
	"github.com/gofiber/fiber/v2"
	"padduck/services"
)

// GetDashboardSummary handles GET /api/v1/dashboard/summary
func (h *Handler) GetDashboardSummary(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2NetworkList) {
		return nil
	}

	summary, err := h.ops.IPAM.GetDashboardSummary(c.Context(), orgIDFromCtx(c))
	if err != nil {
		reqLogger(c).Error("GetDashboardSummary failed", "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}

	return c.JSON(summary)
}

// GetDashboardRecentActivity handles GET /api/v1/dashboard/recent-activity
func (h *Handler) GetDashboardRecentActivity(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2NetworkList) {
		return nil
	}

	activities, err := h.ops.IPAM.GetDashboardRecentActivity(c.Context())
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
	if !h.requirePerm(c, services.PermV2SubnetList, services.ResourceScope{Type: "section", ID: int64(networkID)}) {
		return nil
	}

	tree, err := h.ops.IPAM.GetSubnetTree(c.Context(), int64(networkID))
	if err != nil {
		reqLogger(c).Error("GetSubnetTree failed", "error", err, "network_id", networkID)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}

	return c.JSON(tree)
}
