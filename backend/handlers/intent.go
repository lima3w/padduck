package handlers

import (
	"github.com/gofiber/fiber/v2"
)

// ListIntents returns resource intents, filtered by ?status= and ?resource_type=.
// Defaults to all statuses. Results are org-scoped when caller has an org context.
func (h *Handler) ListIntents(c *fiber.Ctx) error {
	status := c.Query("status")
	resourceType := c.Query("resource_type")
	items, err := h.ops.Intent.ListIntents(c.Context(), orgIDFromCtx(c), status, resourceType)
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, err.Error())
	}
	return c.JSON(items)
}

// GetIntent returns a single resource intent by ID.
func (h *Handler) GetIntent(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid id")
	}
	intent, err := h.ops.Intent.GetIntent(c.Context(), id)
	if err != nil {
		return RespondError(c, fiber.StatusNotFound, ErrNotFound, "intent not found")
	}
	return c.JSON(intent)
}

// SubmitIntent queues a desired-state change. Auto-approved when intent_auto_approve is true.
func (h *Handler) SubmitIntent(c *fiber.Ctx) error {
	var body struct {
		ResourceType string         `json:"resource_type"`
		ResourceID   *int64         `json:"resource_id"`
		Operation    string         `json:"operation"`
		DesiredState map[string]any `json:"desired_state"`
	}
	if err := c.BodyParser(&body); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}
	userID := callerID(c)
	intent, err := h.ops.Intent.SubmitIntent(c.Context(), orgIDFromCtx(c),
		body.ResourceType, body.ResourceID, body.Operation, body.DesiredState, userID)
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(intent)
}

// ApproveIntent approves a pending intent and applies the change.
func (h *Handler) ApproveIntent(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid id")
	}
	var body struct {
		Note string `json:"note"`
	}
	_ = c.BodyParser(&body)
	intent, err := h.ops.Intent.ApproveIntent(c.Context(), id, callerID(c), body.Note)
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, err.Error())
	}
	return c.JSON(intent)
}

// RejectIntent rejects a pending intent without applying any change.
func (h *Handler) RejectIntent(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid id")
	}
	var body struct {
		Note string `json:"note"`
	}
	_ = c.BodyParser(&body)
	intent, err := h.ops.Intent.RejectIntent(c.Context(), id, callerID(c), body.Note)
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, err.Error())
	}
	return c.JSON(intent)
}
