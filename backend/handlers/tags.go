package handlers

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"padduck/models"
	"padduck/services"
)

func requireAdmin(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*models.User)
	if !ok || user.Role != "admin" {
		_ = RespondError(c, fiber.StatusForbidden, ErrForbidden, "admin access required")
		return errResponseWritten
	}
	// API tokens with non-admin scope must not reach admin-only handlers.
	if scope, ok := c.Locals("tokenScope").(string); ok && scope != "admin" {
		_ = RespondError(c, fiber.StatusForbidden, ErrForbidden, "token scope does not allow admin operations")
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
	if !h.requirePerm(c, services.PermV2IPRead) {
		return nil
	}

	tags, err := h.service.ListIPTags(c.Context())
	if err != nil {
		reqLogger(c).Error("error listing tags", "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
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
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "name is required")
	}

	tag, err := h.service.CreateIPTag(c.Context(), req.Name, req.Colour, req.Description)
	if err != nil {
		reqLogger(c).Error("error creating tag", "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}

	return c.Status(fiber.StatusCreated).JSON(tag)
}

// UpdateTag handles PUT /api/v1/tags/:id
func (h *Handler) UpdateTag(c *fiber.Ctx) error {
	if err := requireAdmin(c); err != nil {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid tag ID")
	}

	req := new(UpdateTagRequest)
	if err := c.BodyParser(req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}

	tag, err := h.service.UpdateIPTag(c.Context(), int64(id), req.Name, req.Colour, req.Description)
	if err != nil {
		reqLogger(c).Error("error updating tag", "id", id, "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}

	return c.JSON(tag)
}

// DeleteTag handles DELETE /api/v1/tags/:id
func (h *Handler) DeleteTag(c *fiber.Ctx) error {
	if err := requireAdmin(c); err != nil {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid tag ID")
	}

	if err := h.service.DeleteIPTag(c.Context(), int64(id)); err != nil {
		if errors.Is(err, services.ErrSystemTag) || errors.Is(err, services.ErrTagInUse) {
			return RespondError(c, fiber.StatusConflict, ErrConflict, err.Error())
		}
		if errors.Is(err, services.ErrNotFound) {
			return RespondError(c, fiber.StatusNotFound, ErrNotFound, err.Error())
		}
		reqLogger(c).Error("error deleting tag", "id", id, "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}

	return c.SendStatus(fiber.StatusNoContent)
}
