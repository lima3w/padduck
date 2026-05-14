package handlers

import (
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"ipam-next/repository"
	"ipam-next/services"
)

// ListLocations handles GET /api/v1/locations
func (h *Handler) ListLocations(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2LocationList); err != nil {
		return nil
	}
	locs, err := h.service.ListLocations(c.Context())
	if err != nil {
		log.Printf("Error listing locations: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.JSON(locs)
}

// GetLocationTree handles GET /api/v1/locations/tree
func (h *Handler) GetLocationTree(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2LocationList); err != nil {
		return nil
	}
	tree, err := h.service.GetLocationTree(c.Context())
	if err != nil {
		log.Printf("Error getting location tree: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.JSON(tree)
}

// CreateLocation handles POST /api/v1/locations
func (h *Handler) CreateLocation(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2LocationWrite); err != nil {
		return nil
	}
	req := new(repository.LocationParams)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "location name is required"})
	}
	loc, err := h.service.CreateLocation(c.Context(), req)
	if err != nil {
		log.Printf("Error creating location: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(loc)
}

// GetLocation handles GET /api/v1/locations/:id
func (h *Handler) GetLocation(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2LocationRead); err != nil {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid location ID"})
	}
	loc, err := h.service.GetLocation(c.Context(), int64(id))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "location not found"})
		}
		log.Printf("Error getting location %d: %v", id, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.JSON(loc)
}

// UpdateLocation handles PUT /api/v1/locations/:id
func (h *Handler) UpdateLocation(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2LocationWrite); err != nil {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid location ID"})
	}
	req := new(repository.LocationParams)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "location name is required"})
	}
	loc, err := h.service.UpdateLocation(c.Context(), int64(id), req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "location not found"})
		}
		log.Printf("Error updating location %d: %v", id, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(loc)
}

// DeleteLocation handles DELETE /api/v1/locations/:id
func (h *Handler) DeleteLocation(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2LocationDelete); err != nil {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid location ID"})
	}
	if err := h.service.DeleteLocation(c.Context(), int64(id)); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "location not found"})
		}
		log.Printf("Error deleting location %d: %v", id, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}
