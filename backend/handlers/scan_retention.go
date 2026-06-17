package handlers

import (
	"github.com/gofiber/fiber/v2"
	"padduck/services"
)

// GetScanRetentionSettings handles GET /api/v1/admin/scan-retention
func (h *Handler) GetScanRetentionSettings(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminRead) {
		return nil
	}
	settings, err := h.ops.Discovery.GetRetentionSettings(c.Context())
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}
	return c.JSON(settings)
}

// UpdateScanRetentionSettings handles PUT /api/v1/admin/scan-retention
func (h *Handler) UpdateScanRetentionSettings(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminWrite) {
		return nil
	}
	var req struct {
		RawHistoryDays  int  `json:"raw_history_days"`
		RollupEnabled   bool `json:"rollup_enabled"`
		RollupAfterDays int  `json:"rollup_after_days"`
	}
	if err := c.BodyParser(&req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}
	settings, err := h.ops.Discovery.UpdateRetentionSettings(c.Context(), req.RawHistoryDays, req.RollupAfterDays, req.RollupEnabled)
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, err.Error())
	}
	return c.JSON(settings)
}

// RunScanRetentionPrune handles POST /api/v1/admin/scan-retention/prune
func (h *Handler) RunScanRetentionPrune(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminWrite) {
		return nil
	}
	pruned, err := h.ops.Discovery.RunRetentionPrune(c.Context())
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, err.Error())
	}
	return c.JSON(fiber.Map{"pruned": pruned})
}
