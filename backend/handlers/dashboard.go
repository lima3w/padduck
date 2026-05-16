package handlers

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"ipam-next/services"
)

// GetDashboardSummary handles GET /api/v1/dashboard/summary
func (h *Handler) GetDashboardSummary(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2SectionList); err != nil {
		return nil
	}

	summary, err := h.service.GetDashboardSummary(c.Context())
	if err != nil {
		slog.Error("GetDashboardSummary failed", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.JSON(summary)
}

// GetDashboardRecentActivity handles GET /api/v1/dashboard/recent-activity
func (h *Handler) GetDashboardRecentActivity(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2SectionList); err != nil {
		return nil
	}

	activities, err := h.service.GetDashboardRecentActivity(c.Context())
	if err != nil {
		slog.Error("GetDashboardRecentActivity failed", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.JSON(activities)
}

// GetSubnetTree handles GET /api/v1/sections/:id/subnets/tree
func (h *Handler) GetSubnetTree(c *fiber.Ctx) error {
	sectionID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid section ID"})
	}
	if err := h.permCheck(c, services.PermV2SubnetList, services.ResourceScope{Type: "section", ID: int64(sectionID)}); err != nil {
		return nil
	}

	tree, err := h.service.GetSubnetTree(c.Context(), int64(sectionID))
	if err != nil {
		slog.Error("GetSubnetTree failed", "error", err, "section_id", sectionID)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.JSON(tree)
}
