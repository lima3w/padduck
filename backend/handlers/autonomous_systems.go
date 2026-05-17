package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"ipam-next/models"
	"ipam-next/services"
)

type CreateAutonomousSystemRequest struct {
	ASN         int64  `json:"asn"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	RIR         string `json:"rir"`
}

type UpdateAutonomousSystemRequest struct {
	ASN         int64  `json:"asn"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	RIR         string `json:"rir"`
}

func (h *Handler) ListAutonomousSystems(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2ASList); err != nil {
		return nil
	}
	items, err := h.service.ListAutonomousSystems(c.Context())
	if err != nil {
		reqLogger(c).Error("error listing autonomous systems", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	if items == nil {
		items = make([]*models.AutonomousSystem, 0)
	}
	return c.JSON(items)
}

func (h *Handler) GetAutonomousSystem(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid ID"})
	}
	if err := h.permCheck(c, services.PermV2ASRead); err != nil {
		return nil
	}
	item, err := h.service.GetAutonomousSystem(c.Context(), int64(id))
	if err != nil {
		reqLogger(c).Error("error getting autonomous system", "id", id, "error", err)
		return respondCustomerASError(c, err, "autonomous system")
	}
	return c.JSON(item)
}

func (h *Handler) CreateAutonomousSystem(c *fiber.Ctx) error {
	req := new(CreateAutonomousSystemRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := h.permCheck(c, services.PermV2ASWrite); err != nil {
		return nil
	}
	item, err := h.service.CreateAutonomousSystem(c.Context(), req.ASN, req.Name, req.Description, req.Type, req.RIR)
	if err != nil {
		reqLogger(c).Error("error creating autonomous system", "error", err)
		return respondCustomerASError(c, err, "autonomous system")
	}
	uid, uname := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "autonomous_system_created",
		ResourceType: "autonomous_system", ResourceID: &item.ID,
		NewValues: map[string]string{"asn": fmt.Sprintf("%d", item.ASN), "name": item.Name},
	})
	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *Handler) UpdateAutonomousSystem(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid ID"})
	}
	if err := h.permCheck(c, services.PermV2ASWrite, services.ResourceScope{Type: "autonomous_system", ID: int64(id)}); err != nil {
		return nil
	}
	req := new(UpdateAutonomousSystemRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	item, err := h.service.UpdateAutonomousSystem(c.Context(), int64(id), req.ASN, req.Name, req.Description, req.Type, req.RIR)
	if err != nil {
		reqLogger(c).Error("error updating autonomous system", "id", id, "error", err)
		return respondCustomerASError(c, err, "autonomous system")
	}
	uid, uname := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "autonomous_system_updated",
		ResourceType: "autonomous_system", ResourceID: &item.ID,
		NewValues: map[string]string{"asn": fmt.Sprintf("%d", req.ASN), "name": req.Name},
	})
	return c.JSON(item)
}

func (h *Handler) DeleteAutonomousSystem(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid ID"})
	}
	if err := h.permCheck(c, services.PermV2ASDelete, services.ResourceScope{Type: "autonomous_system", ID: int64(id)}); err != nil {
		return nil
	}
	if err := h.service.DeleteAutonomousSystem(c.Context(), int64(id)); err != nil {
		reqLogger(c).Error("error deleting autonomous system", "id", id, "error", err)
		return respondCustomerASError(c, err, "autonomous system")
	}
	uid, uname := auditUserFromCtx(c)
	aid := int64(id)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "autonomous_system_deleted",
		ResourceType: "autonomous_system", ResourceID: &aid,
	})
	return c.SendStatus(fiber.StatusNoContent)
}
