package handlers

import (
	"github.com/gofiber/fiber/v2"
	"ipam-next/models"
	"ipam-next/services"
)

// ---- Role management ----

func (h *Handler) ListRoles(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*models.User)
	if !ok || user.Role != "admin" {
		return RespondError(c, fiber.StatusForbidden, ErrForbidden, "admin access required")
	}
	roles, err := h.service.ListRoles(c.Context())
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to list roles")
	}
	return c.JSON(roles)
}

func (h *Handler) GetRole(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*models.User)
	if !ok || user.Role != "admin" {
		return RespondError(c, fiber.StatusForbidden, ErrForbidden, "admin access required")
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid role ID")
	}
	role, err := h.service.GetRole(c.Context(), int64(id))
	if err != nil {
		return RespondError(c, fiber.StatusNotFound, ErrNotFound, "role not found")
	}
	return c.JSON(role)
}

type CreateRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (h *Handler) CreateRole(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*models.User)
	if !ok || user.Role != "admin" {
		return RespondError(c, fiber.StatusForbidden, ErrForbidden, "admin access required")
	}
	req := new(CreateRoleRequest)
	if err := c.BodyParser(req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}
	role, err := h.service.CreateRole(c.Context(), req.Name, req.Description)
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(role)
}

type UpdateRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (h *Handler) UpdateRole(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*models.User)
	if !ok || user.Role != "admin" {
		return RespondError(c, fiber.StatusForbidden, ErrForbidden, "admin access required")
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid role ID")
	}
	req := new(UpdateRoleRequest)
	if err := c.BodyParser(req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}
	role, err := h.service.UpdateRole(c.Context(), int64(id), req.Name, req.Description)
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, err.Error())
	}
	return c.JSON(role)
}

func (h *Handler) DeleteRole(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*models.User)
	if !ok || user.Role != "admin" {
		return RespondError(c, fiber.StatusForbidden, ErrForbidden, "admin access required")
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid role ID")
	}
	if err := h.service.DeleteRole(c.Context(), int64(id)); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// ---- Permission management ----

type AddPermissionRequest struct {
	Permission   string  `json:"permission"`
	ResourceType *string `json:"resource_type"`
	ResourceID   *int64  `json:"resource_id"`
}

func (h *Handler) AddPermissionToRole(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*models.User)
	if !ok || user.Role != "admin" {
		return RespondError(c, fiber.StatusForbidden, ErrForbidden, "admin access required")
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid role ID")
	}
	req := new(AddPermissionRequest)
	if err := c.BodyParser(req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}
	perm, err := h.service.AddPermissionToRole(c.Context(), int64(id), req.Permission, req.ResourceType, req.ResourceID)
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(perm)
}

func (h *Handler) RemovePermissionFromRole(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*models.User)
	if !ok || user.Role != "admin" {
		return RespondError(c, fiber.StatusForbidden, ErrForbidden, "admin access required")
	}
	permID, err := c.ParamsInt("perm_id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid permission ID")
	}
	if err := h.service.RemovePermissionFromRole(c.Context(), int64(permID)); err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to remove permission")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) ListAvailablePermissions(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*models.User)
	if !ok || user.Role != "admin" {
		return RespondError(c, fiber.StatusForbidden, ErrForbidden, "admin access required")
	}
	return c.JSON(services.AllPermissions)
}

// ---- User-role assignment ----

func (h *Handler) GetUserRoles(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*models.User)
	if !ok || user.Role != "admin" {
		return RespondError(c, fiber.StatusForbidden, ErrForbidden, "admin access required")
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid user ID")
	}
	roles, err := h.service.GetUserRoles(c.Context(), int64(id))
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to get user roles")
	}
	return c.JSON(roles)
}

type AssignRoleRequest struct {
	RoleID int64 `json:"role_id"`
}

func (h *Handler) AssignRoleToUser(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*models.User)
	if !ok || user.Role != "admin" {
		return RespondError(c, fiber.StatusForbidden, ErrForbidden, "admin access required")
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid user ID")
	}
	req := new(AssignRoleRequest)
	if err := c.BodyParser(req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}
	if err := h.service.AssignRoleToUser(c.Context(), int64(id), req.RoleID); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, err.Error())
	}

	_ = h.service.Notification.Queue(c.Context(), int64(id), services.NotifRoleChanged, map[string]interface{}{
		"ChangedBy": user.Username,
	})

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) RemoveRoleFromUser(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*models.User)
	if !ok || user.Role != "admin" {
		return RespondError(c, fiber.StatusForbidden, ErrForbidden, "admin access required")
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid user ID")
	}
	roleID, err := c.ParamsInt("role_id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid role ID")
	}
	if err := h.service.RemoveRoleFromUser(c.Context(), int64(id), int64(roleID)); err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to remove role")
	}
	return c.SendStatus(fiber.StatusNoContent)
}
