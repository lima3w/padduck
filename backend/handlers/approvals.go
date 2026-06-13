package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"padduck/models"
	"padduck/services"
)

// ListPendingApprovals handles GET /api/v1/admin/approvals
func (h *Handler) ListPendingApprovals(c *fiber.Ctx) error {
	currentUser, ok := c.Locals("user").(*models.User)
	if !ok || currentUser.Role != "admin" {
		return RespondError(c, fiber.StatusForbidden, ErrForbidden, "admin access required")
	}

	approvals, err := h.service.Registration.ListPendingApprovals(c.Context())
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to list approvals")
	}

	type approvalWithUser struct {
		ID        int64  `json:"id"`
		UserID    int64  `json:"user_id"`
		Username  string `json:"username"`
		Email     string `json:"email"`
		Status    string `json:"status"`
		CreatedAt string `json:"created_at"`
	}

	result := make([]approvalWithUser, 0, len(approvals))
	for _, a := range approvals {
		user, err := h.service.GetRepository().GetUserByID(c.Context(), a.UserID)
		username, email := "", ""
		if err == nil {
			username = user.Username
			email = user.Email
		}
		result = append(result, approvalWithUser{
			ID:        a.ID,
			UserID:    a.UserID,
			Username:  username,
			Email:     email,
			Status:    a.Status,
			CreatedAt: a.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return c.JSON(fiber.Map{"approvals": result})
}

// ApproveUser handles POST /api/v1/admin/approvals/:id/approve
func (h *Handler) ApproveUser(c *fiber.Ctx) error {
	currentUser, ok := c.Locals("user").(*models.User)
	if !ok || currentUser.Role != "admin" {
		return RespondError(c, fiber.StatusForbidden, ErrForbidden, "admin access required")
	}

	approvalID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid approval ID")
	}

	reviewerID := currentUser.ID

	if err := h.service.Registration.ApproveUser(c.Context(), approvalID, reviewerID); err != nil {
		reqLogger(c).Error("approve user failed", "approval_id", approvalID, "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}

	adminID, adminName := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: adminID, Username: adminName, Action: "user_approved",
		ResourceType: "user_approval", ResourceID: &approvalID,
	})

	return c.JSON(fiber.Map{"message": "User approved"})
}

// RejectUser handles POST /api/v1/admin/approvals/:id/reject
func (h *Handler) RejectUser(c *fiber.Ctx) error {
	currentUser, ok := c.Locals("user").(*models.User)
	if !ok || currentUser.Role != "admin" {
		return RespondError(c, fiber.StatusForbidden, ErrForbidden, "admin access required")
	}

	approvalID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid approval ID")
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.BodyParser(&req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}

	reviewerID := currentUser.ID

	if err := h.service.Registration.RejectUser(c.Context(), approvalID, reviewerID, req.Reason); err != nil {
		reqLogger(c).Error("reject user failed", "approval_id", approvalID, "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}

	adminID, adminName := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: adminID, Username: adminName, Action: "user_rejected",
		ResourceType: "user_approval", ResourceID: &approvalID,
		NewValues: map[string]string{"reason": req.Reason},
	})

	return c.JSON(fiber.Map{"message": "User rejected"})
}
