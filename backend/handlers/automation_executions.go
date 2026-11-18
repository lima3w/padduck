package handlers

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"padduck/repository"
	"padduck/services"
)

// ListAutomationExecutions handles GET /api/v1/admin/automation/executions
// Query: policy_id, resource_type, resource_id, limit, offset.
func (h *Handler) ListAutomationExecutions(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminRead) {
		return nil
	}
	f := repository.PolicyExecutionLogFilter{ResourceType: c.Query("resource_type")}
	for name, dst := range map[string]*int64{"policy_id": &f.PolicyID, "resource_id": &f.ResourceID} {
		if v := c.Query(name); v != "" {
			n, err := strconv.ParseInt(v, 10, 64)
			if err != nil || n <= 0 {
				return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid "+name)
			}
			*dst = n
		}
	}
	f.Limit = c.QueryInt("limit", 50)
	f.Offset = c.QueryInt("offset", 0)
	if f.Offset < 0 {
		f.Offset = 0
	}
	logs, total, err := h.ops.Automation.ListExecutions(c.Context(), f)
	if err != nil {
		reqLogger(c).Error("listing automation executions", "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to list executions")
	}
	return c.JSON(fiber.Map{"executions": logs, "total": total})
}

// GetAutomationExecution handles GET /api/v1/admin/automation/executions/:id
func (h *Handler) GetAutomationExecution(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminRead) {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid execution ID")
	}
	l, err := h.ops.Automation.GetExecution(c.Context(), int64(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return RespondError(c, fiber.StatusNotFound, ErrNotFound, "execution not found")
		}
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to load execution")
	}
	return c.JSON(l)
}

// RollbackAutomationExecution handles POST /api/v1/admin/automation/executions/:id/rollback
func (h *Handler) RollbackAutomationExecution(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminWrite) {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid execution ID")
	}
	uid, uname := auditUserFromCtx(c)
	l, err := h.ops.Automation.RollbackExecution(c.Context(), int64(id), uid)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return RespondError(c, fiber.StatusNotFound, ErrNotFound, "execution not found")
	case errors.Is(err, services.ErrExecutionAlreadyRolledBack), errors.Is(err, services.ErrNothingToRollBack):
		return RespondError(c, fiber.StatusConflict, ErrConflict, err.Error())
	case err != nil:
		reqLogger(c).Error("rolling back automation execution", "id", id, "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to roll back execution")
	}
	outcomes := map[string]string{"policy": l.PolicyName}
	for i, a := range l.ActionsTaken {
		if a.Rollback != nil {
			outcomes[strconv.Itoa(i)+":"+a.Type] = a.Rollback.Status
		}
	}
	execID := l.ID
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "automation_execution_rolled_back",
		ResourceType: "policy_execution", ResourceID: &execID, ResourceName: l.PolicyName,
		NewValues: outcomes,
	})
	return c.JSON(l)
}
