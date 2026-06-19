package handlers

import (
	"github.com/gofiber/fiber/v2"
	"padduck/models"
	"padduck/services"
)

type CreateNetworkRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateNetworkRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// CreateNetwork handles POST /api/v1/networks
func (h *Handler) CreateNetwork(c *fiber.Ctx) error {
	req := new(CreateNetworkRequest)
	if err := c.BodyParser(req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}
	if !h.requirePerm(c, services.PermV2NetworkWrite) {
		return nil
	}

	// Always derive createdBy from the authenticated user — never trust caller-supplied values.
	var createdBy int64
	if u, ok := c.Locals("user").(*models.User); ok && u != nil {
		createdBy = u.ID
	} else if userID, ok := c.Locals("userID").(int64); ok {
		createdBy = userID
	}

	section, err := h.ops.IPAM.CreateNetwork(c.Context(), req.Name, req.Description, createdBy)
	if err != nil {
		reqLogger(c).Error("error creating section", "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}

	uid, uname := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "section_created",
		ResourceType: "section", ResourceID: &section.ID, ResourceName: section.Name,
		NewValues: map[string]string{"name": section.Name, "description": section.Description},
	})

	return c.Status(fiber.StatusCreated).JSON(section)
}

// GetNetwork handles GET /api/v1/networks/:id
func (h *Handler) GetNetwork(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2NetworkRead) {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid section ID")
	}

	section, err := h.ops.IPAM.GetNetwork(c.Context(), int64(id))
	if err != nil {
		reqLogger(c).Error("error getting section", "id", id, "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}

	return c.JSON(section)
}

// ListNetworks handles GET /api/v1/networks
// Supports ?page=1&limit=25 for pagination. Without those params it returns all results.
func (h *Handler) ListNetworks(c *fiber.Ctx) error {
	addDeprecationHeaders(c, "/api/v2/networks")
	if !h.requirePerm(c, services.PermV2NetworkList) {
		return nil
	}

	page, limit, opts := parseListOptions(c)
	if c.Query("page") != "" || c.Query("limit") != "" || opts.Sort != "" || opts.Query != "" {
		sections, total, err := h.ops.IPAM.ListNetworksPaginatedWithOptions(c.Context(), page, limit, opts)
		if err != nil {
			reqLogger(c).Error("error listing sections", "error", err)
			return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
		}
		if sections == nil {
			sections = make([]*models.Network, 0)
		}
		return c.JSON(fiber.Map{
			"data":  sections,
			"total": total,
			"page":  page,
			"limit": limit,
		})
	}

	sections, err := h.ops.IPAM.ListNetworks(c.Context())
	if err != nil {
		reqLogger(c).Error("error listing sections", "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}
	if sections == nil {
		sections = make([]*models.Network, 0)
	}
	return c.JSON(sections)
}

// UpdateNetwork handles PUT /api/v1/networks/:id
func (h *Handler) UpdateNetwork(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid section ID")
	}
	if !h.requirePerm(c, services.PermV2NetworkWrite, services.ResourceScope{Type: "section", ID: int64(id)}) {
		return nil
	}

	req := new(UpdateNetworkRequest)
	if err := c.BodyParser(req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}

	oldSection, _ := h.ops.IPAM.GetNetwork(c.Context(), int64(id))

	section, err := h.ops.IPAM.UpdateNetwork(c.Context(), int64(id), req.Name, req.Description)
	if err != nil {
		reqLogger(c).Error("error updating section", "id", id, "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}

	uid, uname := auditUserFromCtx(c)
	var oldVals interface{}
	if oldSection != nil {
		oldVals = map[string]string{"name": oldSection.Name, "description": oldSection.Description}
	}
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "section_updated",
		ResourceType: "section", ResourceID: &section.ID, ResourceName: section.Name,
		OldValues: oldVals,
		NewValues: map[string]string{"name": req.Name, "description": req.Description},
	})

	return c.JSON(section)
}

// DeleteNetwork handles DELETE /api/v1/networks/:id
func (h *Handler) DeleteNetwork(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid section ID")
	}
	if !h.requirePerm(c, services.PermV2NetworkDelete, services.ResourceScope{Type: "section", ID: int64(id)}) {
		return nil
	}

	if err := h.ops.IPAM.DeleteNetwork(c.Context(), int64(id)); err != nil {
		reqLogger(c).Error("error deleting section", "id", id, "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}

	uid, uname := auditUserFromCtx(c)
	sid := int64(id)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "section_deleted",
		ResourceType: "section", ResourceID: &sid,
	})

	return c.SendStatus(fiber.StatusNoContent)
}
