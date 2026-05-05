package handlers

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"ipam-next/models"
)

type CreateVRFRequest struct {
	Name               string `json:"name"`
	RouteDistinguisher string `json:"route_distinguisher"`
	Description        string `json:"description"`
}

type UpdateVRFRequest struct {
	Name               string `json:"name"`
	RouteDistinguisher string `json:"route_distinguisher"`
	Description        string `json:"description"`
}

func (h *Handler) ListVRFs(c *fiber.Ctx) error {
	vrfs, err := h.service.ListVRFs(c.Context())
	if err != nil {
		log.Printf("Error listing VRFs: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	if vrfs == nil {
		vrfs = make([]*models.VRF, 0)
	}

	return c.JSON(vrfs)
}

func (h *Handler) GetVRF(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid VRF ID"})
	}

	vrf, err := h.service.GetVRF(c.Context(), int64(id))
	if err != nil {
		log.Printf("Error getting VRF %d: %v", id, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.JSON(vrf)
}

func (h *Handler) CreateVRF(c *fiber.Ctx) error {
	req := new(CreateVRFRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	vrf, err := h.service.CreateVRF(c.Context(), req.Name, req.RouteDistinguisher, req.Description)
	if err != nil {
		log.Printf("Error creating VRF: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.Status(fiber.StatusCreated).JSON(vrf)
}

func (h *Handler) UpdateVRF(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid VRF ID"})
	}

	req := new(UpdateVRFRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	vrf, err := h.service.UpdateVRF(c.Context(), int64(id), req.Name, req.RouteDistinguisher, req.Description)
	if err != nil {
		log.Printf("Error updating VRF %d: %v", id, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.JSON(vrf)
}

func (h *Handler) DeleteVRF(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid VRF ID"})
	}

	if err := h.service.DeleteVRF(c.Context(), int64(id)); err != nil {
		log.Printf("Error deleting VRF %d: %v", id, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
