package handlers

import (
	"github.com/gofiber/fiber/v2"
	"padduck/services"
)

type RequestUnlockRequest struct {
	Username string `json:"username"`
}

type LoginHistoryResponse struct {
	ID            int64  `json:"id"`
	IPAddress     string `json:"ip_address"`
	UserAgent     string `json:"user_agent"`
	Success       bool   `json:"success"`
	FailureReason string `json:"failure_reason,omitempty"`
	CreatedAt     string `json:"created_at"`
}

// RequestUnlock handles POST /api/v1/auth/unlock
// Sends an unlock email to the account owner.
func (h *Handler) RequestUnlock(c *fiber.Ctx) error {
	req := new(RequestUnlockRequest)
	if err := c.BodyParser(req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}

	// Always return success to prevent username enumeration
	if req.Username != "" {
		if err := h.service.RequestUnlockEmail(c.Context(), req.Username); err != nil {
			reqLogger(c).Error("error sending unlock email", "error", err)
		}
	}

	return c.JSON(fiber.Map{"message": "If the account exists and is locked, an unlock email has been sent"})
}

// VerifyUnlock handles GET /api/v1/auth/unlock?token=xxx
// Unlocks the account via token and redirects to login.
func (h *Handler) VerifyUnlock(c *fiber.Ctx) error {
	token := c.Query("token")
	if token == "" {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "token is required")
	}

	if err := h.service.UnlockAccountByToken(c.Context(), token); err != nil {
		if err == services.ErrInvalidUnlockToken {
			return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid or expired unlock token")
		}
		reqLogger(c).Error("error unlocking account", "error", err)
		return h.StatusInternalServerError(c, "Failed to unlock account", err.Error())
	}

	return c.JSON(fiber.Map{"message": "Account unlocked successfully"})
}

// GetLoginHistory handles GET /api/v1/user/login-history
func (h *Handler) GetLoginHistory(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(int64)
	if !ok {
		return RespondError(c, fiber.StatusUnauthorized, ErrUnauthorized, "user ID not found in context")
	}

	attempts, err := h.service.GetLoginHistory(c.Context(), userID, 20)
	if err != nil {
		reqLogger(c).Error("error fetching login history", "error", err)
		return h.StatusInternalServerError(c, "Failed to fetch login history", err.Error())
	}

	response := make([]LoginHistoryResponse, 0, len(attempts))
	for _, a := range attempts {
		response = append(response, LoginHistoryResponse{
			ID:            a.ID,
			IPAddress:     a.IPAddress,
			UserAgent:     a.UserAgent,
			Success:       a.Success,
			FailureReason: a.FailureReason,
			CreatedAt:     a.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	return c.JSON(response)
}

// AdminUnlockUser handles POST /api/v1/admin/users/:id/unlock
func (h *Handler) AdminUnlockUser(c *fiber.Ctx) error {
	adminID, ok := c.Locals("userID").(int64)
	if !ok {
		return RespondError(c, fiber.StatusUnauthorized, ErrUnauthorized, "user ID not found in context")
	}

	targetID, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid user ID")
	}

	if err := h.service.UnlockAccountByAdmin(c.Context(), int64(targetID), adminID); err != nil {
		reqLogger(c).Error("error admin-unlocking user", "target_id", targetID, "error", err)
		return h.StatusInternalServerError(c, "Failed to unlock account", err.Error())
	}

	uid, uname := auditUserFromCtx(c)
	tid := int64(targetID)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "account_unlocked",
		ResourceType: "user", ResourceID: &tid,
	})

	return c.JSON(fiber.Map{"message": "Account unlocked successfully"})
}
