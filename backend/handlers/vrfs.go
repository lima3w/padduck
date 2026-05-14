package handlers

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"ipam-next/models"
	"ipam-next/services"
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

func (h *Handler) ListVRFs(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2VRFList); err != nil {
		return nil
	}
	vrfs, err := h.service.ListVRFs(c.Context())
	if err != nil {
		log.Printf("Error listing VRFs: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	if vrfs == nil {
		vrfs = make([]*models.VRF, 0)
	}

	return c.JSON(vrfs)
}

func (h *Handler) GetVRF(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid VRF ID"})
	}
	if err := h.permCheck(c, services.PermV2VRFRead); err != nil {
		return nil
	}

	vrf, err := h.service.GetVRF(c.Context(), int64(id))
	if err != nil {
		log.Printf("Error getting VRF %d: %v", id, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.JSON(vrf)
}

func (h *Handler) CreateVRF(c *fiber.Ctx) error {
	req := new(CreateVRFRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := h.permCheck(c, services.PermV2VRFWrite); err != nil {
		return nil
	}

	vrf, err := h.service.CreateVRF(c.Context(), req.Name, req.RouteDistinguisher, req.Description)
	if err != nil {
		log.Printf("Error creating VRF: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid VRF ID"})
	}
	if err := h.permCheck(c, services.PermV2VRFWrite, services.ResourceScope{Type: "vrf", ID: int64(id)}); err != nil {
		return err
	}

	req := new(UpdateVRFRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	vrf, err := h.service.UpdateVRF(c.Context(), int64(id), req.Name, req.RouteDistinguisher, req.Description)
	if err != nil {
		log.Printf("Error updating VRF %d: %v", id, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid VRF ID"})
	}
	if err := h.permCheck(c, services.PermV2VRFDelete, services.ResourceScope{Type: "vrf", ID: int64(id)}); err != nil {
		return err
	}

	if err := h.service.DeleteVRF(c.Context(), int64(id)); err != nil {
		log.Printf("Error deleting VRF %d: %v", id, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	uid, uname := auditUserFromCtx(c)
	vid := int64(id)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "vrf_deleted",
		ResourceType: "vrf", ResourceID: &vid,
	})

	return c.SendStatus(fiber.StatusNoContent)
}
