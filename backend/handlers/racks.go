package handlers

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"ipam-next/repository"
	"ipam-next/services"
)

// ListRacks handles GET /api/v1/racks
func (h *Handler) ListRacks(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2LocationList); err != nil {
		return nil
	}

	var locationID *int64
	if v := c.QueryInt("location_id", 0); v > 0 {
		id := int64(v)
		locationID = &id
	}

	racks, err := h.service.ListRacks(c.Context(), locationID)
	if err != nil {
		reqLogger(c).Error("error listing racks", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.JSON(racks)
}

// CreateRack handles POST /api/v1/racks
func (h *Handler) CreateRack(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2LocationWrite); err != nil {
		return nil
	}

	req := new(repository.RackParams)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "rack name is required"})
	}

	rack, err := h.service.CreateRack(c.Context(), req)
	if err != nil {
		reqLogger(c).Error("error creating rack", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.Status(fiber.StatusCreated).JSON(rack)
}

// GetRack handles GET /api/v1/racks/:id
func (h *Handler) GetRack(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2LocationRead); err != nil {
		return nil
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid rack ID"})
	}

	rack, err := h.service.GetRack(c.Context(), int64(id))
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "rack not found"})
		}
		reqLogger(c).Error("error getting rack", "id", id, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.JSON(rack)
}

// UpdateRack handles PUT /api/v1/racks/:id
func (h *Handler) UpdateRack(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2LocationWrite); err != nil {
		return nil
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid rack ID"})
	}

	req := new(repository.RackParams)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "rack name is required"})
	}

	rack, err := h.service.UpdateRack(c.Context(), int64(id), req)
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "rack not found"})
		}
		reqLogger(c).Error("error updating rack", "id", id, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.JSON(rack)
}

// DeleteRack handles DELETE /api/v1/racks/:id
func (h *Handler) DeleteRack(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2LocationDelete); err != nil {
		return nil
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid rack ID"})
	}

	if err := h.service.DeleteRack(c.Context(), int64(id)); err != nil {
		if errors.Is(err, services.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "rack not found"})
		}
		reqLogger(c).Error("error deleting rack", "id", id, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// ListDevicesInRack handles GET /api/v1/racks/:id/devices
func (h *Handler) ListDevicesInRack(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2DeviceRead); err != nil {
		return nil
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid rack ID"})
	}

	devices, err := h.service.ListDevicesInRack(c.Context(), int64(id))
	if err != nil {
		reqLogger(c).Error("error listing devices in rack", "id", id, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.JSON(devices)
}
