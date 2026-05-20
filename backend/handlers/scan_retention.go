package handlers

import (
	"github.com/gofiber/fiber/v2"
	"padduck/services"
)

// GetScanRetentionSettings handles GET /api/v1/admin/scan-retention
func (h *Handler) GetScanRetentionSettings(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminRead); err != nil {
		return nil
	}
	settings, err := h.service.Discovery.GetRetentionSettings(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.JSON(settings)
}

// UpdateScanRetentionSettings handles PUT /api/v1/admin/scan-retention
func (h *Handler) UpdateScanRetentionSettings(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminWrite); err != nil {
		return nil
	}
	var req struct {
		RawHistoryDays  int  `json:"raw_history_days"`
		RollupEnabled   bool `json:"rollup_enabled"`
		RollupAfterDays int  `json:"rollup_after_days"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	settings, err := h.service.Discovery.UpdateRetentionSettings(c.Context(), req.RawHistoryDays, req.RollupAfterDays, req.RollupEnabled)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(settings)
}

// RunScanRetentionPrune handles POST /api/v1/admin/scan-retention/prune
func (h *Handler) RunScanRetentionPrune(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminWrite); err != nil {
		return nil
	}
	pruned, err := h.service.Discovery.RunRetentionPrune(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"pruned": pruned})
}
