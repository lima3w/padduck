package handlers

import (
	"github.com/gofiber/fiber/v2"
	"padduck/services"
)

type createGrantRequest struct {
	UserID     int64   `json:"user_id"`
	Permission string  `json:"permission"`
	ScopeType  *string `json:"scope_type"`
	ScopeID    *int64  `json:"scope_id"`
}

// ListUserGrants handles GET /api/v1/admin/users/:id/grants
func (h *Handler) ListUserGrants(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2OrgRead) {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid user ID")
	}
	grants, err := h.ops.Identity.ListUserGrants(c.Context(), int64(id))
	if err != nil {
		reqLogger(c).Error("error listing user grants", "user_id", id, "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}
	return c.JSON(fiber.Map{"grants": grants})
}

// CreateGrant handles POST /api/v1/admin/role-grants
func (h *Handler) CreateGrant(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2OrgWrite) {
		return nil
	}
	grantorID, ok := c.Locals("userID").(int64)
	if !ok || grantorID <= 0 {
		return RespondError(c, fiber.StatusUnauthorized, ErrUnauthorized, "not authenticated")
	}
	req := new(createGrantRequest)
	if err := c.BodyParser(req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}
	if req.UserID <= 0 {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "user_id is required")
	}
	if req.Permission == "" {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "permission is required")
	}
	grant, err := h.ops.Identity.CreateGrant(c.Context(), grantorID, services.CreateGrantRequest{
		UserID:     req.UserID,
		Permission: req.Permission,
		ScopeType:  req.ScopeType,
		ScopeID:    req.ScopeID,
	})
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(grant)
}

// RevokeGrant handles DELETE /api/v1/admin/role-grants/:id
func (h *Handler) RevokeGrant(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2OrgWrite) {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid grant ID")
	}
	if err := h.ops.Identity.RevokeGrant(c.Context(), int64(id)); err != nil {
		reqLogger(c).Error("error revoking grant", "id", id, "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}
	return c.SendStatus(fiber.StatusNoContent)
}
