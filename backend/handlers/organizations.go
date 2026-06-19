package handlers

import (
	"github.com/gofiber/fiber/v2"
	"padduck/services"
)

type createOrgRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// ListOrganizations handles GET /api/v1/admin/organizations
func (h *Handler) ListOrganizations(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2OrgRead) {
		return nil
	}
	orgs, err := h.ops.Organizations.List(c.Context())
	if err != nil {
		reqLogger(c).Error("error listing organizations", "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}
	return c.JSON(fiber.Map{"organizations": orgs})
}

// CreateOrganization handles POST /api/v1/admin/organizations
func (h *Handler) CreateOrganization(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2OrgWrite) {
		return nil
	}
	req := new(createOrgRequest)
	if err := c.BodyParser(req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}
	org, err := h.ops.Organizations.Create(c.Context(), req.Name, req.Slug)
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(org)
}

// DeleteOrganization handles DELETE /api/v1/admin/organizations/:id
func (h *Handler) DeleteOrganization(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2OrgWrite) {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid organization ID")
	}
	if err := h.ops.Organizations.Delete(c.Context(), int64(id)); err != nil {
		reqLogger(c).Error("error deleting organization", "id", id, "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}
	return c.SendStatus(fiber.StatusNoContent)
}
