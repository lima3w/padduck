package handlers

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"ipam-next/models"
	"ipam-next/services"
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

	// Use authenticated user ID if available, otherwise default to admin (1)
	createdBy := req.CreatedBy
	if createdBy == 0 {
		if userID, ok := c.Locals("userID").(int64); ok {
			createdBy = userID
		} else {
			createdBy = 1
		}
	}

	section, err := h.service.CreateSection(c.Context(), req.Name, req.Description, createdBy)
	if err != nil {
		log.Printf("Error creating section: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	uid, uname := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "section_created",
		ResourceType: "section", ResourceID: &section.ID, ResourceName: section.Name,
		NewValues: map[string]string{"name": section.Name, "description": section.Description},
	})

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
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.JSON(section)
}

// ListSections handles GET /api/v1/sections
func (h *Handler) ListSections(c *fiber.Ctx) error {
	sections, err := h.service.ListSections(c.Context())
	if err != nil {
		log.Printf("Error listing sections: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
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
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	uid, uname := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "section_updated",
		ResourceType: "section", ResourceID: &section.ID, ResourceName: section.Name,
		NewValues: map[string]string{"name": req.Name, "description": req.Description},
	})

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
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	uid, uname := auditUserFromCtx(c)
	sid := int64(id)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "section_deleted",
		ResourceType: "section", ResourceID: &sid,
	})

	return c.SendStatus(fiber.StatusNoContent)
}
