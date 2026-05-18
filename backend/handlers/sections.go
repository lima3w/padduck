package handlers

import (
	"github.com/gofiber/fiber/v2"
	"ipam-next/models"
	"ipam-next/services"
)

type CreateSectionRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
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
	if err := h.permCheck(c, services.PermV2SectionWrite); err != nil {
		return nil
	}

	// Always derive createdBy from the authenticated user — never trust caller-supplied values.
	var createdBy int64
	if u, ok := c.Locals("user").(*models.User); ok && u != nil {
		createdBy = u.ID
	} else if userID, ok := c.Locals("userID").(int64); ok {
		createdBy = userID
	}

	section, err := h.service.CreateSection(c.Context(), req.Name, req.Description, createdBy)
	if err != nil {
		reqLogger(c).Error("error creating section", "error", err)
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
	if err := h.permCheck(c, services.PermV2SectionRead); err != nil {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid section ID"})
	}

	section, err := h.service.GetSection(c.Context(), int64(id))
	if err != nil {
		reqLogger(c).Error("error getting section", "id", id, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.JSON(section)
}

// ListSections handles GET /api/v1/sections
// Supports ?page=1&limit=25 for pagination. Without those params it returns all results.
func (h *Handler) ListSections(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2SectionList); err != nil {
		return nil
	}

	page := c.QueryInt("page", 0)
	limit := c.QueryInt("limit", 0)

	// If pagination params are provided, use paginated version
	if page > 0 || limit > 0 {
		if page < 1 {
			page = 1
		}
		if limit < 1 {
			limit = 25
		}
		sections, total, err := h.service.ListSectionsPaginated(c.Context(), page, limit)
		if err != nil {
			reqLogger(c).Error("error listing sections", "error", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
		}
		if sections == nil {
			sections = make([]*models.Section, 0)
		}
		return c.JSON(fiber.Map{
			"data":  sections,
			"total": total,
			"page":  page,
			"limit": limit,
		})
	}

	sections, err := h.service.ListSections(c.Context())
	if err != nil {
		reqLogger(c).Error("error listing sections", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
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
	if err := h.permCheck(c, services.PermV2SectionWrite, services.ResourceScope{Type: "section", ID: int64(id)}); err != nil {
		return nil
	}

	req := new(UpdateSectionRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	section, err := h.service.UpdateSection(c.Context(), int64(id), req.Name, req.Description)
	if err != nil {
		reqLogger(c).Error("error updating section", "id", id, "error", err)
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
	if err := h.permCheck(c, services.PermV2SectionDelete, services.ResourceScope{Type: "section", ID: int64(id)}); err != nil {
		return nil
	}

	if err := h.service.DeleteSection(c.Context(), int64(id)); err != nil {
		reqLogger(c).Error("error deleting section", "id", id, "error", err)
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
