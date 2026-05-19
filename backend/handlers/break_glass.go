package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"ipam-next/models"
	"ipam-next/services"
)

// GetBreakGlassStatus handles GET /api/v1/admin/break-glass
// Returns the current active session (or null) plus full session history.
func (h *Handler) GetBreakGlassStatus(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminWrite); err != nil {
		return nil
	}

	active, err := h.service.GetRepository().GetActiveBreakGlassSession(c.Context())
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to get active break-glass session", err.Error())
	}

	history, err := h.service.GetRepository().ListBreakGlassSessions(c.Context())
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to list break-glass sessions", err.Error())
	}
	if history == nil {
		history = []*models.BreakGlassSession{}
	}

	return c.JSON(fiber.Map{
		"active":  active,
		"history": history,
	})
}

// ActivateBreakGlass handles POST /api/v1/admin/break-glass/activate
// Body: {"justification": "..."}
func (h *Handler) ActivateBreakGlass(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminWrite); err != nil {
		return nil
	}

	user, ok := c.Locals("user").(*models.User)
	if !ok {
		return RespondError(c, fiber.StatusUnauthorized, ErrUnauthorized, "not authenticated")
	}

	var req struct {
		Justification string `json:"justification"`
	}
	if err := c.BodyParser(&req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}
	if len(req.Justification) < 10 {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "justification must be at least 10 characters")
	}

	// Check that no active session exists
	existing, err := h.service.GetRepository().GetActiveBreakGlassSession(c.Context())
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to check for active break-glass session", err.Error())
	}
	if existing != nil {
		return RespondError(c, fiber.StatusConflict, ErrConflict, "a break-glass session is already active")
	}

	session, err := h.service.GetRepository().CreateBreakGlassSession(c.Context(), user.ID, req.Justification)
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to create break-glass session", err.Error())
	}

	sessionID := session.ID
	h.auditLog(c, services.AuditEntry{
		UserID:       &user.ID,
		Username:     user.Username,
		Action:       "activate",
		ResourceType: "break_glass_session",
		ResourceID:   &sessionID,
		ResourceName: "break-glass",
		NewValues:    map[string]interface{}{"justification": req.Justification, "expires_at": session.ExpiresAt},
	})

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"session": session})
}

// EndBreakGlass handles POST /api/v1/admin/break-glass/end
func (h *Handler) EndBreakGlass(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminWrite); err != nil {
		return nil
	}

	user, ok := c.Locals("user").(*models.User)
	if !ok {
		return RespondError(c, fiber.StatusUnauthorized, ErrUnauthorized, "not authenticated")
	}

	active, err := h.service.GetRepository().GetActiveBreakGlassSession(c.Context())
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to get active break-glass session", err.Error())
	}
	if active == nil {
		return RespondError(c, fiber.StatusNotFound, ErrNotFound, "no active break-glass session found")
	}

	session, err := h.service.GetRepository().EndBreakGlassSession(c.Context(), active.ID, user.ID)
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to end break-glass session", err.Error())
	}

	sessionID := session.ID
	h.auditLog(c, services.AuditEntry{
		UserID:       &user.ID,
		Username:     user.Username,
		Action:       "end",
		ResourceType: "break_glass_session",
		ResourceID:   &sessionID,
		ResourceName: "break-glass",
		NewValues:    map[string]interface{}{"ended_at": session.EndedAt, "ended_by_user_id": user.ID},
	})

	return c.JSON(fiber.Map{"session": session})
}
