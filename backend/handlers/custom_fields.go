package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"padduck/repository"
	"padduck/services"
)

var allowedEntityTypes = map[string]bool{"subnet": true, "ip_address": true, "device": true}
var allowedFieldTypes = map[string]bool{"text": true, "number": true, "textarea": true, "dropdown": true, "checkbox": true, "date": true, "url": true, "email": true}

func validateCustomFieldParams(p *repository.CustomFieldDefinitionParams) []ValidationField {
	var fields []ValidationField
	if !allowedEntityTypes[p.EntityType] {
		fields = append(fields, ValidationField{Field: "entity_type", Message: "entity_type must be one of: subnet, ip_address, device"})
	}
	if !allowedFieldTypes[p.FieldType] {
		fields = append(fields, ValidationField{Field: "field_type", Message: "field_type must be one of: text, number, textarea, dropdown, checkbox, date, url, email"})
	}
	return fields
}

// ListCustomFieldDefinitions handles GET /api/v1/admin/custom-fields
func (h *Handler) ListCustomFieldDefinitions(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminRead) {
		return nil
	}
	entityType := c.Query("entity_type")
	defs, err := h.ops.Workflow.ListCustomFieldDefinitions(c.Context(), entityType)
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, err.Error())
	}
	return c.JSON(defs)
}

// CreateCustomFieldDefinition handles POST /api/v1/admin/custom-fields
func (h *Handler) CreateCustomFieldDefinition(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminWrite) {
		return nil
	}
	p := new(repository.CustomFieldDefinitionParams)
	if err := c.BodyParser(p); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}
	if vf := validateCustomFieldParams(p); len(vf) > 0 {
		return RespondValidationError(c, "validation failed", vf)
	}

	def, err := h.ops.Workflow.CreateCustomFieldDefinition(c.Context(), p)
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(def)
}

// GetCustomFieldDefinition handles GET /api/v1/admin/custom-fields/:id
func (h *Handler) GetCustomFieldDefinition(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid id")
	}
	if !h.requirePerm(c, services.PermV2AdminRead) {
		return nil
	}

	def, err := h.ops.Workflow.GetCustomFieldDefinition(c.Context(), int64(id))
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			return RespondError(c, fiber.StatusNotFound, ErrNotFound, err.Error())
		}
		reqLogger(c).Error("error getting custom field definition", "id", id, "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}
	return c.JSON(def)
}

// UpdateCustomFieldDefinition handles PUT /api/v1/admin/custom-fields/:id
func (h *Handler) UpdateCustomFieldDefinition(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid id")
	}
	if !h.requirePerm(c, services.PermV2AdminWrite) {
		return nil
	}

	p := new(repository.CustomFieldDefinitionParams)
	if err := c.BodyParser(p); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}
	if vf := validateCustomFieldParams(p); len(vf) > 0 {
		return RespondValidationError(c, "validation failed", vf)
	}

	def, err := h.ops.Workflow.UpdateCustomFieldDefinition(c.Context(), int64(id), p)
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			return RespondError(c, fiber.StatusNotFound, ErrNotFound, err.Error())
		}
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, err.Error())
	}
	return c.JSON(def)
}

// DeleteCustomFieldDefinition handles DELETE /api/v1/admin/custom-fields/:id
func (h *Handler) DeleteCustomFieldDefinition(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid id")
	}
	if !h.requirePerm(c, services.PermV2AdminWrite) {
		return nil
	}

	if err := h.ops.Workflow.DeleteCustomFieldDefinition(c.Context(), int64(id)); err != nil {
		if errors.Is(err, services.ErrNotFound) {
			return RespondError(c, fiber.StatusNotFound, ErrNotFound, err.Error())
		}
		reqLogger(c).Error("error deleting custom field definition", "id", id, "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// ReorderCustomFieldDefinitions handles PUT /api/v1/admin/custom-fields/reorder
func (h *Handler) ReorderCustomFieldDefinitions(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminWrite) {
		return nil
	}
	var body struct {
		IDs []int64 `json:"ids"`
	}
	if err := c.BodyParser(&body); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}

	if err := h.ops.Workflow.ReorderCustomFieldDefinitions(c.Context(), body.IDs); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}
