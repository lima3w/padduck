package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"ipam-next/repository"
	"ipam-next/services"
)

// ListCustomFieldDefinitions handles GET /api/v1/admin/custom-fields
func (h *Handler) ListCustomFieldDefinitions(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminRead); err != nil {
		return nil
	}
	entityType := c.Query("entity_type")
	defs, err := h.service.ListCustomFieldDefinitions(c.Context(), entityType)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(defs)
}

// CreateCustomFieldDefinition handles POST /api/v1/admin/custom-fields
func (h *Handler) CreateCustomFieldDefinition(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminWrite); err != nil {
		return nil
	}
	p := new(repository.CustomFieldDefinitionParams)
	if err := c.BodyParser(p); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	def, err := h.service.CreateCustomFieldDefinition(c.Context(), p)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(def)
}

// GetCustomFieldDefinition handles GET /api/v1/admin/custom-fields/:id
func (h *Handler) GetCustomFieldDefinition(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}
	if err := h.permCheck(c, services.PermV2AdminRead); err != nil {
		return nil
	}

	def, err := h.service.GetCustomFieldDefinition(c.Context(), int64(id))
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		reqLogger(c).Error("error getting custom field definition", "id", id, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.JSON(def)
}

// UpdateCustomFieldDefinition handles PUT /api/v1/admin/custom-fields/:id
func (h *Handler) UpdateCustomFieldDefinition(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}
	if err := h.permCheck(c, services.PermV2AdminWrite); err != nil {
		return nil
	}

	p := new(repository.CustomFieldDefinitionParams)
	if err := c.BodyParser(p); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	def, err := h.service.UpdateCustomFieldDefinition(c.Context(), int64(id), p)
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(def)
}

// DeleteCustomFieldDefinition handles DELETE /api/v1/admin/custom-fields/:id
func (h *Handler) DeleteCustomFieldDefinition(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}
	if err := h.permCheck(c, services.PermV2AdminWrite); err != nil {
		return nil
	}

	if err := h.service.DeleteCustomFieldDefinition(c.Context(), int64(id)); err != nil {
		if errors.Is(err, services.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		reqLogger(c).Error("error deleting custom field definition", "id", id, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// ReorderCustomFieldDefinitions handles PUT /api/v1/admin/custom-fields/reorder
func (h *Handler) ReorderCustomFieldDefinitions(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminWrite); err != nil {
		return nil
	}
	var body struct {
		IDs []int64 `json:"ids"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if err := h.service.ReorderCustomFieldDefinitions(c.Context(), body.IDs); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}
