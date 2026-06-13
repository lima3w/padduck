package handlers

import (
	"github.com/gofiber/fiber/v2"
	"padduck/services"
)

// GetScanJobHistory handles GET /api/v1/admin/scan-jobs/:id/history
func (h *Handler) GetScanJobHistory(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminRead); err != nil {
		return nil
	}
	jobID, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid job ID")
	}
	// Verify job exists
	if _, err := h.service.Discovery.GetJob(c.Context(), int64(jobID)); err != nil {
		return RespondError(c, fiber.StatusNotFound, ErrNotFound, "scan job not found")
	}
	runs, err := h.service.Discovery.ListScanRuns(c.Context(), int64(jobID))
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}
	return c.JSON(runs)
}

// GetScanRunDetail handles GET /api/v1/admin/scan-jobs/:id/history/:run_id
func (h *Handler) GetScanRunDetail(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminRead); err != nil {
		return nil
	}
	runID, err := c.ParamsInt("run_id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid run ID")
	}
	run, changes, err := h.service.Discovery.GetScanRunWithChanges(c.Context(), int64(runID))
	if err != nil {
		return RespondError(c, fiber.StatusNotFound, ErrNotFound, "scan run not found")
	}
	return c.JSON(fiber.Map{
		"run":     run,
		"changes": changes,
	})
}
