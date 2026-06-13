package handlers

import (
	"github.com/gofiber/fiber/v2"
	"padduck/models"
	"padduck/services"
)

type CreateVRFRequest struct {
	Name               string `json:"name"`
	RouteDistinguisher string `json:"route_distinguisher"`
	Description        string `json:"description"`
}

type UpdateVRFRequest struct {
	Name               string `json:"name"`
	RouteDistinguisher string `json:"route_distinguisher"`
	Description        string `json:"description"`
}

// ListVRFs handles GET /api/v1/vrfs
// Supports ?page=1&limit=25 for pagination. Without those params it returns all results.
func (h *Handler) ListVRFs(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2VRFList); err != nil {
		return nil
	}

	page := c.QueryInt("page", 0)
	limit := c.QueryInt("limit", 0)

	if page > 0 || limit > 0 {
		if page < 1 {
			page = 1
		}
		if limit < 1 {
			limit = 25
		}
		vrfs, total, err := h.service.ListVRFsPaginated(c.Context(), page, limit)
		if err != nil {
			reqLogger(c).Error("error listing VRFs", "error", err)
			return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
		}
		if vrfs == nil {
			vrfs = make([]*models.VRF, 0)
		}
		return c.JSON(fiber.Map{
			"data":  vrfs,
			"total": total,
			"page":  page,
			"limit": limit,
		})
	}

	vrfs, err := h.service.ListVRFs(c.Context())
	if err != nil {
		reqLogger(c).Error("error listing VRFs", "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}

	if vrfs == nil {
		vrfs = make([]*models.VRF, 0)
	}

	return c.JSON(vrfs)
}

func (h *Handler) GetVRF(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2VRFRead); err != nil {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid VRF ID")
	}

	vrf, err := h.service.GetVRF(c.Context(), int64(id))
	if err != nil {
		reqLogger(c).Error("error getting VRF", "id", id, "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}

	return c.JSON(vrf)
}

func (h *Handler) CreateVRF(c *fiber.Ctx) error {
	req := new(CreateVRFRequest)
	if err := c.BodyParser(req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}
	if err := h.permCheck(c, services.PermV2VRFWrite); err != nil {
		return nil
	}

	vrf, err := h.service.CreateVRF(c.Context(), req.Name, req.RouteDistinguisher, req.Description)
	if err != nil {
		reqLogger(c).Error("error creating VRF", "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}

	uid, uname := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "vrf_created",
		ResourceType: "vrf", ResourceID: &vrf.ID, ResourceName: vrf.Name,
		NewValues: map[string]string{"name": vrf.Name, "rd": vrf.RouteDistinguisher},
	})

	return c.Status(fiber.StatusCreated).JSON(vrf)
}

func (h *Handler) UpdateVRF(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid VRF ID")
	}
	if err := h.permCheck(c, services.PermV2VRFWrite, services.ResourceScope{Type: "vrf", ID: int64(id)}); err != nil {
		return nil
	}

	req := new(UpdateVRFRequest)
	if err := c.BodyParser(req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}

	vrf, err := h.service.UpdateVRF(c.Context(), int64(id), req.Name, req.RouteDistinguisher, req.Description)
	if err != nil {
		reqLogger(c).Error("error updating VRF", "id", id, "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}

	uid, uname := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "vrf_updated",
		ResourceType: "vrf", ResourceID: &vrf.ID, ResourceName: vrf.Name,
		NewValues: map[string]string{"name": req.Name, "rd": req.RouteDistinguisher, "description": req.Description},
	})

	return c.JSON(vrf)
}

func (h *Handler) DeleteVRF(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid VRF ID")
	}
	if err := h.permCheck(c, services.PermV2VRFDelete, services.ResourceScope{Type: "vrf", ID: int64(id)}); err != nil {
		return nil
	}

	if err := h.service.DeleteVRF(c.Context(), int64(id)); err != nil {
		reqLogger(c).Error("error deleting VRF", "id", id, "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}

	uid, uname := auditUserFromCtx(c)
	vid := int64(id)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "vrf_deleted",
		ResourceType: "vrf", ResourceID: &vid,
	})

	return c.SendStatus(fiber.StatusNoContent)
}
