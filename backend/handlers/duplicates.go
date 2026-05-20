package handlers

import (
	"github.com/gofiber/fiber/v2"
	"padduck/services"
)

// GetDuplicates handles GET /api/v1/admin/reports/duplicates
func (h *Handler) GetDuplicates(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminRead); err != nil {
		return nil
	}
	result, err := h.service.Reports.GetDuplicates(c.Context())
	if err != nil {
		reqLogger(c).Error("get duplicates report failed", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.JSON(result)
}
