package handlers

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"padduck/services"
)

// evaluatePolicy runs automation policy checks before an intent change is applied.
// It writes the HTTP response for deny and manual_review outcomes and returns
// proceed=false so the caller knows to return nil without further action.
// On allow (or bypass), it returns proceed=true.
func (h *Handler) evaluatePolicy(c *fiber.Ctx, workflow, action string, policyCtx map[string]string, desiredState any) bool {
	if bypassPolicyFromCtx(c) {
		uid, uname := auditUserFromCtx(c)
		h.auditLog(c, services.AuditEntry{
			UserID: uid, Username: uname, Action: "policy_bypassed",
			ResourceType: workflow,
			NewValues:    map[string]string{"action": action},
		})
		return true
	}

	decision, err := h.ops.Automation.Evaluate(c.Context(), services.AutomationRequest{
		Workflow: workflow,
		Action:   action,
		Values:   policyCtx,
	})
	if err != nil || decision.Allowed {
		return true
	}

	if decision.ReviewNeeded {
		stateMap := toStateMap(desiredState)
		uid := callerID(c)
		intent, intentErr := h.ops.Intent.SubmitIntent(
			c.Context(), orgIDFromCtx(c), workflow, nil, "create", stateMap, uid, true,
		)
		if intentErr != nil {
			RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, intentErr.Error())
			return false
		}
		c.Status(fiber.StatusAccepted).JSON(fiber.Map{
			"status":    "pending_review",
			"message":   decision.Message,
			"intent_id": intent.ID,
		})
		return false
	}

	msg := decision.Message
	if msg == "" {
		msg = "request denied by automation policy"
	}
	RespondError(c, fiber.StatusUnprocessableEntity, ErrPolicyDenied, msg)
	return false
}

// toStateMap converts any struct to map[string]any via JSON round-trip.
func toStateMap(v any) map[string]any {
	b, err := json.Marshal(v)
	if err != nil {
		return map[string]any{}
	}
	var m map[string]any
	_ = json.Unmarshal(b, &m)
	return m
}

// SetAPITokenBypassPolicy handles PUT /platform/api-tokens/:id/bypass-policy.
// Platform admins only. Sets or clears the bypass_policy flag on an API token.
func (h *Handler) SetAPITokenBypassPolicy(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid token id")
	}

	var body struct {
		BypassPolicy bool `json:"bypass_policy"`
	}
	if err := c.BodyParser(&body); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}

	if err := h.service.GetRepository().SetAPITokenBypassPolicy(c.Context(), id, body.BypassPolicy); err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, err.Error())
	}

	uid, uname := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "api_token_bypass_policy_set",
		ResourceType: "api_token", ResourceID: &id,
		NewValues: map[string]string{"bypass_policy": boolStr(body.BypassPolicy)},
	})

	return c.JSON(fiber.Map{"bypass_policy": body.BypassPolicy})
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
