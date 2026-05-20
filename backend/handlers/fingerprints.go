package handlers

import (
	"github.com/gofiber/fiber/v2"
	"padduck/services"
)

// GetDeviceFingerprint handles GET /api/v1/admin/devices/:id/fingerprint
func (h *Handler) GetDeviceFingerprint(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminRead); err != nil {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid device ID"})
	}
	fp, err := h.service.Discovery.GetDeviceFingerprint(c.Context(), int64(id))
	if err != nil {
		return c.JSON(fiber.Map{"fingerprint": nil})
	}
	return c.JSON(fiber.Map{"fingerprint": fp})
}

// BuildDeviceFingerprint handles POST /api/v1/admin/devices/:id/fingerprint
func (h *Handler) BuildDeviceFingerprint(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminWrite); err != nil {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid device ID"})
	}
	var req struct {
		DeviceIP  string  `json:"device_ip"`
		IsAlive   bool    `json:"is_alive"`
		PTRRecord *string `json:"ptr_record"`
		OpenPorts *string `json:"open_ports"` // comma-separated
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	fp, err := h.service.Discovery.BuildDeviceFingerprint(c.Context(), int64(id), req.DeviceIP, req.IsAlive, req.PTRRecord, req.OpenPorts)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"fingerprint": fp})
}
