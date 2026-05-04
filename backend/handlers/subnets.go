package handlers

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"ipam-next/models"
)

type CreateSubnetRequest struct {
	NetworkAddress string `json:"network_address"`
	PrefixLength   int    `json:"prefix_length"`
	Description    string `json:"description"`
}

type UpdateSubnetRequest struct {
	Description string `json:"description"`
}

// CreateSubnet handles POST /api/v1/sections/:sectionID/subnets
func (h *Handler) CreateSubnet(c *fiber.Ctx) error {
	sectionID, err := c.ParamsInt("sectionID")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid section ID"})
	}

	req := new(CreateSubnetRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	subnet, err := h.service.CreateSubnet(c.Context(), int64(sectionID), req.NetworkAddress, req.PrefixLength, req.Description)
	if err != nil {
		log.Printf("Error creating subnet: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.Status(fiber.StatusCreated).JSON(subnet)
}

// GetSubnet handles GET /api/v1/subnets/:id
func (h *Handler) GetSubnet(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid subnet ID"})
	}

	subnet, err := h.service.GetSubnet(c.Context(), int64(id))
	if err != nil {
		log.Printf("Error getting subnet %d: %v", id, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.JSON(subnet)
}

// ListSubnets handles GET /api/v1/sections/:sectionID/subnets
func (h *Handler) ListSubnets(c *fiber.Ctx) error {
	sectionID, err := c.ParamsInt("sectionID")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid section ID"})
	}

	subnets, err := h.service.ListSubnets(c.Context(), int64(sectionID))
	if err != nil {
		log.Printf("Error listing subnets for section %d: %v", sectionID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	if subnets == nil {
		subnets = make([]*models.Subnet, 0)
	}

	return c.JSON(subnets)
}

// UpdateSubnet handles PUT /api/v1/subnets/:id
func (h *Handler) UpdateSubnet(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid subnet ID"})
	}

	req := new(UpdateSubnetRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	subnet, err := h.service.UpdateSubnet(c.Context(), int64(id), req.Description)
	if err != nil {
		log.Printf("Error updating subnet %d: %v", id, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.JSON(subnet)
}

// DeleteSubnet handles DELETE /api/v1/subnets/:id
func (h *Handler) DeleteSubnet(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid subnet ID"})
	}

	if err := h.service.DeleteSubnet(c.Context(), int64(id)); err != nil {
		log.Printf("Error deleting subnet %d: %v", id, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
