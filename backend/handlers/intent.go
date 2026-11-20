package handlers

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"padduck/models"
	"padduck/services"
)

// intentError maps an intent service error to an HTTP response.
func intentError(c *fiber.Ctx, err error) error {
	var quota *services.QuotaExceededError
	switch {
	case errors.Is(err, services.ErrIntentNotPending):
		return RespondError(c, fiber.StatusConflict, ErrConflict, "intent is not pending")
	case errors.Is(err, pgx.ErrNoRows):
		return RespondError(c, fiber.StatusNotFound, ErrNotFound, "not found")
	case errors.As(err, &quota):
		return RespondError(c, fiber.StatusUnprocessableEntity, ErrQuotaExceeded, err.Error())
	case strings.HasPrefix(err.Error(), "apply failed"):
		return RespondError(c, fiber.StatusUnprocessableEntity, ErrBadRequest, err.Error())
	default:
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, err.Error())
	}
}

// parseIntentResource reads and validates ?resource_type= and ?resource_id=.
func parseIntentResource(c *fiber.Ctx) (resourceType string, resourceID *int64, ok bool) {
	resourceType = c.Query("resource_type")
	if v := c.Query("resource_id"); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil || id <= 0 {
			_ = RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid resource_id")
			return "", nil, false
		}
		resourceID = &id
	}
	return resourceType, resourceID, true
}

// ListIntents handles GET /intents. Query: status, resource_type, resource_id.
// Listing a single resource's history needs that resource's read permission;
// listing across resources needs admin.
func (h *Handler) ListIntents(c *fiber.Ctx) error {
	resourceType, resourceID, ok := parseIntentResource(c)
	if !ok {
		return nil
	}
	if resourceID != nil {
		perm, known := services.IntentReadPermission(resourceType)
		if !known {
			return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "resource_type is required with resource_id")
		}
		if !h.requirePerm(c, perm) {
			return nil
		}
	} else if !h.requirePerm(c, services.PermV2AdminWrite) {
		return nil
	}
	items, err := h.ops.Intent.ListIntents(c.Context(), c.Query("status"), resourceType, resourceID)
	if err != nil {
		reqLogger(c).Error("list intents", "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}
	return c.JSON(items)
}

// GetIntentDiff handles GET /intents/diff?resource_type=&resource_id=. It returns
// the resource's current state, its latest pending intent and the changed fields.
func (h *Handler) GetIntentDiff(c *fiber.Ctx) error {
	resourceType, resourceID, ok := parseIntentResource(c)
	if !ok {
		return nil
	}
	perm, known := services.IntentReadPermission(resourceType)
	if !known || resourceID == nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "resource_type and resource_id are required")
	}
	if !h.requirePerm(c, perm) {
		return nil
	}
	diff, err := h.ops.Intent.GetCurrentVsDesired(c.Context(), resourceType, *resourceID)
	if err != nil {
		return intentError(c, err)
	}
	return c.JSON(diff)
}

// GetIntent handles GET /intents/:id. Admins see any intent; others only their own.
func (h *Handler) GetIntent(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid id")
	}
	intent, err := h.ops.Intent.GetIntent(c.Context(), id)
	if err != nil {
		return RespondError(c, fiber.StatusNotFound, ErrNotFound, "intent not found")
	}
	own := intent.SubmittedBy != nil && callerID(c) != nil && *intent.SubmittedBy == *callerID(c)
	if !own && !h.requirePerm(c, services.PermV2AdminWrite) {
		return nil
	}
	return c.JSON(intent)
}

// SubmitIntent handles POST /intents. The caller needs the permission to make
// the same change directly. Auto-approved when intent_auto_approve is true.
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
	perm, known := services.IntentPermission(body.ResourceType, body.Operation)
	if !known {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "resource_type must be one of subnet, ip_address, device, vlan")
	}
	if !h.requirePerm(c, perm) {
		return nil
	}
	intent, err := h.ops.Intent.SubmitIntent(c.Context(), body.ResourceType, body.ResourceID,
		body.Operation, body.DesiredState, callerID(c))
	if err != nil {
		return intentError(c, err)
	}
	h.auditIntent(c, "intent_submitted", intent)
	return c.Status(fiber.StatusCreated).JSON(intent)
}

// ApproveIntent handles POST /intents/:id/approve and applies the change.
func (h *Handler) ApproveIntent(c *fiber.Ctx) error {
	return h.reviewIntent(c, true)
}

// RejectIntent handles POST /intents/:id/reject without applying any change.
func (h *Handler) RejectIntent(c *fiber.Ctx) error {
	return h.reviewIntent(c, false)
}

func (h *Handler) reviewIntent(c *fiber.Ctx, approve bool) error {
	if !h.requirePerm(c, services.PermV2AdminWrite) {
		return nil
	}
	id, err := parseID(c, "id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid id")
	}
	var body struct {
		Note string `json:"note"`
	}
	_ = c.BodyParser(&body)

	var intent *models.ResourceIntent
	action := "intent_rejected"
	if approve {
		action = "intent_approved"
		intent, err = h.ops.Intent.ApproveIntent(c.Context(), id, callerID(c), body.Note)
	} else {
		intent, err = h.ops.Intent.RejectIntent(c.Context(), id, callerID(c), body.Note)
	}
	if err != nil {
		return intentError(c, err)
	}
	h.auditIntent(c, action, intent)
	return c.JSON(intent)
}

func (h *Handler) auditIntent(c *fiber.Ctx, action string, intent *models.ResourceIntent) {
	uid, username := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: username, Action: action,
		ResourceType: "intent", ResourceID: &intent.ID,
		NewValues: map[string]any{
			"resource_type": intent.ResourceType, "resource_id": intent.ResourceID,
			"operation": intent.Operation, "status": intent.Status,
		},
	})
}
