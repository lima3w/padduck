package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgconn"
	"padduck/models"
	"padduck/services"
)

type createOrgRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// isInstanceAdmin reports whether the caller may manage organizations
// themselves: a platform admin, or an admin of the default organization (the
// instance owner on a single-tenant install). Admins of any other organization
// manage only their own organization's data.
func isInstanceAdmin(c *fiber.Ctx) bool {
	if ok, _ := c.Locals("isPlatformAdmin").(bool); ok {
		return true
	}
	u, ok := c.Locals("user").(*models.User)
	if !ok || u == nil || u.Role != "admin" {
		return false
	}
	org := orgIDFromCtx(c)
	return org != nil && *org == 1
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
	if !isInstanceAdmin(c) {
		// A regular admin sees only their own organization.
		own := orgs[:0:0]
		if mine := orgIDFromCtx(c); mine != nil {
			for _, o := range orgs {
				if o.ID == *mine {
					own = append(own, o)
				}
			}
		}
		orgs = own
	}
	return c.JSON(fiber.Map{"organizations": orgs})
}

// CreateOrganization handles POST /api/v1/admin/organizations
func (h *Handler) CreateOrganization(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2OrgWrite) {
		return nil
	}
	if !isInstanceAdmin(c) {
		return RespondError(c, fiber.StatusForbidden, ErrForbidden, "only a platform admin can create organizations")
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
	if !isInstanceAdmin(c) {
		return RespondError(c, fiber.StatusForbidden, ErrForbidden, "only a platform admin can delete organizations")
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid organization ID")
	}
	if id == 1 {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "the default organization cannot be deleted")
	}
	if err := h.ops.Organizations.Delete(c.Context(), int64(id)); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return RespondError(c, fiber.StatusConflict, ErrConflict, "organization still has users or data; remove them first")
		}
		reqLogger(c).Error("error deleting organization", "id", id, "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}
	return c.SendStatus(fiber.StatusNoContent)
}
