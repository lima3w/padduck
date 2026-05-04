package handlers

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"ipam-next/models"
)

type CreateSectionRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedBy   int64  `json:"created_by"`
}

type UpdateSectionRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// CreateSection handles POST /api/v1/sections
func (h *Handler) CreateSection(c *fiber.Ctx) error {
	req := new(CreateSectionRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	section, err := h.service.CreateSection(c.Context(), req.Name, req.Description, req.CreatedBy)
	if err != nil {
		log.Printf("Error creating section: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(section)
}

// GetSection handles GET /api/v1/sections/:id
func (h *Handler) GetSection(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid section ID"})
	}

	section, err := h.service.GetSection(c.Context(), int64(id))
	if err != nil {
		log.Printf("Error getting section %d: %v", id, err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "section not found"})
	}

	return c.JSON(section)
}

// ListSections handles GET /api/v1/sections
func (h *Handler) ListSections(c *fiber.Ctx) error {
	sections, err := h.service.ListSections(c.Context())
	if err != nil {
		log.Printf("Error listing sections: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Return empty array instead of nil
	if sections == nil {
		sections = make([]*models.Section, 0)
	}

	return c.JSON(sections)
}

// UpdateSection handles PUT /api/v1/sections/:id
func (h *Handler) UpdateSection(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid section ID"})
	}

	req := new(UpdateSectionRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	section, err := h.service.UpdateSection(c.Context(), int64(id), req.Name, req.Description)
	if err != nil {
		log.Printf("Error updating section %d: %v", id, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(section)
}

// DeleteSection handles DELETE /api/v1/sections/:id
func (h *Handler) DeleteSection(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid section ID"})
	}

	if err := h.service.DeleteSection(c.Context(), int64(id)); err != nil {
		log.Printf("Error deleting section %d: %v", id, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
