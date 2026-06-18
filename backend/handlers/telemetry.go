package handlers

import "github.com/gofiber/fiber/v2"

// SendTelemetryNow handles POST /api/v1/admin/telemetry/send.
// Collects and sends a snapshot immediately, regardless of the opt-in flag.
// Returns an error if PocketBase is not configured or the POST fails.
func (h *Handler) SendTelemetryNow(c *fiber.Ctx) error {
	if err := requireAdmin(c); err != nil {
		return nil
	}
	if err := h.ops.Telemetry.SendNow(c.Context()); err != nil {
		return RespondError(c, fiber.StatusBadGateway, "telemetry_send_failed", err.Error())
	}
	return c.JSON(fiber.Map{"ok": true})
}
