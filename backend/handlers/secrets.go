package handlers

import (
	"github.com/gofiber/fiber/v2"
	"padduck/services"
)

func (h *Handler) secretIDParam(c *fiber.Ctx) (int64, error) {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return 0, RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid secret ID")
	}
	return int64(id), nil
}

// ListSecrets handles GET /api/v1/admin/secrets (metadata only).
func (h *Handler) ListSecrets(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminRead) {
		return nil
	}
	secrets, err := h.ops.Secrets.List(c.Context())
	if err != nil {
		reqLogger(c).Error("error listing secrets", "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}
	return c.JSON(secrets)
}

// CreateSecret handles POST /api/v1/admin/secrets.
func (h *Handler) CreateSecret(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminWrite) {
		return nil
	}
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Value       string `json:"value"`
	}
	if err := c.BodyParser(&req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}
	uid, uname := auditUserFromCtx(c)
	sec, err := h.ops.Secrets.Create(c.Context(), req.Name, req.Description, req.Value, uid)
	if err != nil {
		reqLogger(c).Error("error creating secret", "error", err)
		return respondCustomerASError(c, err, "secret")
	}
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "secret_created",
		ResourceType: "secret", ResourceID: &sec.ID, ResourceName: sec.Name,
		NewValues: map[string]any{"name": sec.Name, "version": sec.Version},
	})
	return c.Status(fiber.StatusCreated).JSON(sec)
}

// RotateSecret handles PUT /api/v1/admin/secrets/:id/rotate.
func (h *Handler) RotateSecret(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminWrite) {
		return nil
	}
	id, err := h.secretIDParam(c)
	if err != nil {
		return err
	}
	var req struct {
		Value string `json:"value"`
	}
	if err := c.BodyParser(&req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}
	uid, uname := auditUserFromCtx(c)
	sec, err := h.ops.Secrets.Rotate(c.Context(), id, req.Value, uid)
	if err != nil {
		reqLogger(c).Error("error rotating secret", "id", id, "error", err)
		return respondCustomerASError(c, err, "secret")
	}
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "secret_rotated",
		ResourceType: "secret", ResourceID: &sec.ID, ResourceName: sec.Name,
		NewValues: map[string]any{"name": sec.Name, "version": sec.Version},
	})
	return c.JSON(sec)
}

// ListSecretVersions handles GET /api/v1/admin/secrets/:id/versions.
func (h *Handler) ListSecretVersions(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminRead) {
		return nil
	}
	id, err := h.secretIDParam(c)
	if err != nil {
		return err
	}
	versions, err := h.ops.Secrets.Versions(c.Context(), id)
	if err != nil {
		return respondCustomerASError(c, err, "secret")
	}
	return c.JSON(versions)
}

// DeleteSecretVersion handles DELETE /api/v1/admin/secrets/:id/versions/:version.
func (h *Handler) DeleteSecretVersion(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminWrite) {
		return nil
	}
	id, err := h.secretIDParam(c)
	if err != nil {
		return err
	}
	version, verr := c.ParamsInt("version")
	if verr != nil || version <= 0 {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid version")
	}
	sec, err := h.ops.Secrets.Get(c.Context(), id)
	if err != nil {
		return respondCustomerASError(c, err, "secret")
	}
	if err := h.ops.Secrets.DeleteVersion(c.Context(), id, version); err != nil {
		reqLogger(c).Error("error deleting secret version", "id", id, "version", version, "error", err)
		return respondCustomerASError(c, err, "secret version")
	}
	uid, uname := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "secret_version_deleted",
		ResourceType: "secret", ResourceID: &sec.ID, ResourceName: sec.Name,
		OldValues: map[string]any{"version": version},
	})
	return c.SendStatus(fiber.StatusNoContent)
}

// DeleteSecret handles DELETE /api/v1/admin/secrets/:id.
func (h *Handler) DeleteSecret(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminWrite) {
		return nil
	}
	id, err := h.secretIDParam(c)
	if err != nil {
		return err
	}
	sec, err := h.ops.Secrets.Get(c.Context(), id)
	if err != nil {
		return respondCustomerASError(c, err, "secret")
	}
	if err := h.ops.Secrets.Delete(c.Context(), id); err != nil {
		reqLogger(c).Error("error deleting secret", "id", id, "error", err)
		return respondCustomerASError(c, err, "secret")
	}
	uid, uname := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "secret_deleted",
		ResourceType: "secret", ResourceID: &sec.ID, ResourceName: sec.Name,
		OldValues: map[string]any{"name": sec.Name},
	})
	return c.SendStatus(fiber.StatusNoContent)
}
