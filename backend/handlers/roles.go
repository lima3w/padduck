package handlers

import (
	"github.com/gofiber/fiber/v2"
	"padduck/models"
	"padduck/services"
)

// PermissionPreset is a static named set of permissions that can be applied to a role.
type PermissionPreset struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

func builtinPresets() []PermissionPreset {
	return []PermissionPreset{
		{
			ID:          "read-only",
			Name:        "Read Only",
			Description: "View all resources, no modifications.",
			Permissions: []string{
				"ipam:section:list", "ipam:section:read",
				"ipam:subnet:list", "ipam:subnet:read",
				"ipam:ip_address:list", "ipam:ip_address:read",
				"devices:read",
				"ipam:vlan:list", "ipam:vlan:read",
				"ipam:vrf:list", "ipam:vrf:read",
			},
		},
		{
			ID:          "network-ops",
			Name:        "Network Ops",
			Description: "Manage subnets, IPs, and devices. Cannot manage users or roles.",
			Permissions: []string{
				"ipam:section:list", "ipam:section:read", "ipam:section:write",
				"ipam:subnet:list", "ipam:subnet:read", "ipam:subnet:write", "ipam:subnet:delete",
				"ipam:ip_address:list", "ipam:ip_address:read", "ipam:ip_address:assign", "ipam:ip_address:release",
				"devices:read", "devices:write",
			},
		},
		{
			ID:          "helpdesk",
			Name:        "Helpdesk",
			Description: "View and assign IPs. Submit allocation requests.",
			Permissions: []string{
				"ipam:section:list", "ipam:section:read",
				"ipam:subnet:list", "ipam:subnet:read",
				"ipam:ip_address:list", "ipam:ip_address:read", "ipam:ip_address:assign",
				"ipam:subnet_request:submit",
			},
		},
		{
			ID:          "full-admin",
			Name:        "Full Admin",
			Description: "All permissions.",
			Permissions: []string{
				"ipam:section:list", "ipam:section:read", "ipam:section:write", "ipam:section:delete",
				"ipam:subnet:list", "ipam:subnet:read", "ipam:subnet:write", "ipam:subnet:delete",
				"ipam:ip_address:list", "ipam:ip_address:read", "ipam:ip_address:assign", "ipam:ip_address:release",
				"devices:read", "devices:write", "devices:delete", "devices:admin",
				"ipam:vlan:list", "ipam:vlan:read", "ipam:vlan:write", "ipam:vlan:delete",
				"ipam:vrf:list", "ipam:vrf:read", "ipam:vrf:write", "ipam:vrf:delete",
				"auth:user:list", "auth:user:read", "auth:user:write",
				"auth:audit:read", "auth:admin:read", "auth:admin:write",
			},
		},
	}
}

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
	RoleID     int64  `json:"role_id"`
	LocationID *int64 `json:"location_id"`
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
	if err := h.service.AssignRoleToUser(c.Context(), int64(id), req.RoleID, req.LocationID); err != nil {
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

// ---- Permission presets ----

// ListRolePresets returns the static list of built-in permission presets.
func (h *Handler) ListRolePresets(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminWrite); err != nil {
		return nil
	}
	return c.JSON(builtinPresets())
}

// GetRolePresetDiff compares a role's current permissions against a named preset.
func (h *Handler) GetRolePresetDiff(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminWrite); err != nil {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid role ID")
	}
	presetID := c.Query("preset")
	if presetID == "" {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "preset query parameter is required")
	}

	// Find the requested preset.
	var preset *PermissionPreset
	for _, p := range builtinPresets() {
		p := p
		if p.ID == presetID {
			preset = &p
			break
		}
	}
	if preset == nil {
		return RespondError(c, fiber.StatusNotFound, ErrNotFound, "preset not found")
	}

	// Load the role from the database.
	role, err := h.service.GetRole(c.Context(), int64(id))
	if err != nil {
		return RespondError(c, fiber.StatusNotFound, ErrNotFound, "role not found")
	}

	// Build a set of the role's current permissions.
	rolePerms := make(map[string]struct{}, len(role.Permissions))
	for _, p := range role.Permissions {
		rolePerms[p.Permission] = struct{}{}
	}

	// Build a set of the preset's permissions.
	presetPerms := make(map[string]struct{}, len(preset.Permissions))
	for _, p := range preset.Permissions {
		presetPerms[p] = struct{}{}
	}

	// Compute diff.
	var added, removed, unchanged []string
	for _, p := range preset.Permissions {
		if _, ok := rolePerms[p]; ok {
			unchanged = append(unchanged, p)
		} else {
			added = append(added, p)
		}
	}
	for p := range rolePerms {
		if _, ok := presetPerms[p]; !ok {
			removed = append(removed, p)
		}
	}
	if added == nil {
		added = []string{}
	}
	if removed == nil {
		removed = []string{}
	}
	if unchanged == nil {
		unchanged = []string{}
	}

	return c.JSON(fiber.Map{
		"preset":    preset,
		"role":      fiber.Map{"id": role.ID, "name": role.Name},
		"added":     added,
		"removed":   removed,
		"unchanged": unchanged,
	})
}
