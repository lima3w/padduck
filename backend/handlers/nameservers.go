package handlers

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"padduck/repository"
	"padduck/services"
)

// ListNameservers handles GET /api/v1/nameservers
func (h *Handler) ListNameservers(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2NameserverList); err != nil {
		return nil
	}
	ns, err := h.service.ListNameservers(c.Context())
	if err != nil {
		reqLogger(c).Error("error listing nameservers", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.JSON(ns)
}

// GetNameserver handles GET /api/v1/nameservers/:id
func (h *Handler) GetNameserver(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2NameserverRead); err != nil {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid nameserver ID"})
	}
	ns, err := h.service.GetNameserver(c.Context(), int64(id))
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "nameserver not found"})
		}
		reqLogger(c).Error("error getting nameserver", "id", id, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.JSON(ns)
}

// CreateNameserver handles POST /api/v1/nameservers
func (h *Handler) CreateNameserver(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2NameserverWrite); err != nil {
		return nil
	}
	req := new(repository.NameserverParams)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Server1 = strings.TrimSpace(req.Server1)
	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "nameserver name is required"})
	}
	if req.Server1 == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "server1 is required"})
	}
	ns, err := h.service.CreateNameserver(c.Context(), req)
	if err != nil {
		reqLogger(c).Error("error creating nameserver", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	uid, uname := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "nameserver_created",
		ResourceType: "nameserver", ResourceID: &ns.ID, ResourceName: ns.Name,
	})
	return c.Status(fiber.StatusCreated).JSON(ns)
}

// UpdateNameserver handles PUT /api/v1/nameservers/:id
func (h *Handler) UpdateNameserver(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2NameserverWrite); err != nil {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid nameserver ID"})
	}
	req := new(repository.NameserverParams)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Server1 = strings.TrimSpace(req.Server1)
	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "nameserver name is required"})
	}
	if req.Server1 == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "server1 is required"})
	}
	ns, err := h.service.UpdateNameserver(c.Context(), int64(id), req)
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "nameserver not found"})
		}
		reqLogger(c).Error("error updating nameserver", "id", id, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	uid, uname := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "nameserver_updated",
		ResourceType: "nameserver", ResourceID: &ns.ID, ResourceName: ns.Name,
	})
	return c.JSON(ns)
}

// DeleteNameserver handles DELETE /api/v1/nameservers/:id
func (h *Handler) DeleteNameserver(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2NameserverDelete); err != nil {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid nameserver ID"})
	}
	if err := h.service.DeleteNameserver(c.Context(), int64(id)); err != nil {
		if errors.Is(err, services.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "nameserver not found"})
		}
		reqLogger(c).Error("error deleting nameserver", "id", id, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	uid, uname := auditUserFromCtx(c)
	id64 := int64(id)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "nameserver_deleted",
		ResourceType: "nameserver", ResourceID: &id64,
	})
	return c.SendStatus(fiber.StatusNoContent)
}
