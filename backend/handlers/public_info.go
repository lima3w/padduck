package handlers

import "github.com/gofiber/fiber/v2"

// GetPublicInfo handles GET /api/v1/public-info
// Returns publicly visible instance metadata (no auth required).
func (h *Handler) GetPublicInfo(c *fiber.Ctx) error {
	appURL, _ := h.service.Config.GetCtx(c.Context(), "app_url")
	return c.JSON(fiber.Map{
		"app_url": appURL,
	})
}
