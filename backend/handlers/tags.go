package handlers

import (
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"ipam-next/models"
	"ipam-next/services"
)

func requireAdmin(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*models.User)
	if !ok || user.Role != "admin" {
		_ = c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "admin access required"})
		return errResponseWritten
	}
	return nil
}

type CreateTagRequest struct {
	Name        string  `json:"name"`
	Colour      string  `json:"colour"`
	Description *string `json:"description"`
}

type UpdateTagRequest struct {
	Name        string  `json:"name"`
	Colour      string  `json:"colour"`
	Description *string `json:"description"`
}

// ListTags handles GET /api/v1/tags
func (h *Handler) ListTags(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2IPRead); err != nil {
		return nil
	}

	tags, err := h.service.ListIPTags(c.Context())
	if err != nil {
		log.Printf("Error listing tags: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	if tags == nil {
		tags = make([]*models.IPTag, 0)
	}

	return c.JSON(tags)
}

// CreateTag handles POST /api/v1/tags
func (h *Handler) CreateTag(c *fiber.Ctx) error {
	if err := requireAdmin(c); err != nil {
		return nil
	}

	req := new(CreateTagRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name is required"})
	}

	tag, err := h.service.CreateIPTag(c.Context(), req.Name, req.Colour, req.Description)
	if err != nil {
		log.Printf("Error creating tag: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(tag)
}

// UpdateTag handles PUT /api/v1/tags/:id
func (h *Handler) UpdateTag(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid tag ID"})
	}
	if err := requireAdmin(c); err != nil {
		return nil
	}

	req := new(UpdateTagRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	tag, err := h.service.UpdateIPTag(c.Context(), int64(id), req.Name, req.Colour, req.Description)
	if err != nil {
		log.Printf("Error updating tag %d: %v", id, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(tag)
}

// DeleteTag handles DELETE /api/v1/tags/:id
func (h *Handler) DeleteTag(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid tag ID"})
	}
	if err := requireAdmin(c); err != nil {
		return nil
	}

	if err := h.service.DeleteIPTag(c.Context(), int64(id)); err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "system tag") || strings.Contains(errMsg, "in use") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": errMsg})
		}
		if strings.Contains(errMsg, "not found") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": errMsg})
		}
		log.Printf("Error deleting tag %d: %v", id, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
