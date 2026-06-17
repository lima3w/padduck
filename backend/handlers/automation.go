package handlers

import (
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"padduck/models"
	"padduck/repository"
	"padduck/services"
)

type automationPolicyRequest struct {
	Name       string            `json:"name"`
	Workflow   string            `json:"workflow"`
	Action     string            `json:"action"`
	Effect     string            `json:"effect"`
	Enabled    *bool             `json:"enabled"`
	Conditions map[string]string `json:"conditions"`
	Message    string            `json:"message"`
}

func (h *Handler) ListAutomationPolicies(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminWrite) {
		return nil
	}
	policies, err := h.ops.Automation.ListPolicies(c.Context())
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to load automation policies")
	}
	return c.JSON(policies)
}

func (h *Handler) CreateAutomationPolicy(c *fiber.Ctx) error {
	return h.saveAutomationPolicy(c, 0)
}

func (h *Handler) UpdateAutomationPolicy(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid automation policy ID")
	}
	return h.saveAutomationPolicy(c, id)
}

func (h *Handler) saveAutomationPolicy(c *fiber.Ctx, id int64) error {
	if !h.requirePerm(c, services.PermV2AdminWrite) {
		return nil
	}
	req := new(automationPolicyRequest)
	if err := c.BodyParser(req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}
	if fields := validateAutomationPolicyRequest(req); len(fields) > 0 {
		return RespondValidationError(c, "validation failed", fields)
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	policy, err := h.ops.Automation.SavePolicy(c.Context(), &models.AutomationPolicy{
		ID: id, Name: req.Name, Workflow: req.Workflow, Action: req.Action,
		Effect: req.Effect, Enabled: enabled, Conditions: req.Conditions, Message: req.Message,
	})
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, err.Error())
	}
	return c.JSON(policy)
}

func (h *Handler) DeleteAutomationPolicy(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminWrite) {
		return nil
	}
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid automation policy ID")
	}
	if err := h.ops.Automation.DeletePolicy(c.Context(), id); err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to delete automation policy")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) EvaluateAutomationPolicy(c *fiber.Ctx) error {
	var req struct {
		Workflow string            `json:"workflow"`
		Action   string            `json:"action"`
		Values   map[string]string `json:"values"`
		DryRun   bool              `json:"dry_run"`
	}
	if err := c.BodyParser(&req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}
	if fields := requiredFields(map[string]string{"workflow": req.Workflow, "action": req.Action}); len(fields) > 0 {
		return RespondValidationError(c, "validation failed", fields)
	}
	decision, err := h.ops.Automation.Evaluate(c.Context(), services.AutomationRequest{
		Workflow: req.Workflow, Action: req.Action, Values: req.Values, DryRun: req.DryRun,
	})
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to evaluate automation policy")
	}
	status := fiber.StatusOK
	if decision.ReviewNeeded {
		status = fiber.StatusAccepted
	} else if !decision.Allowed {
		status = fiber.StatusForbidden
	}
	return c.Status(status).JSON(decision)
}

func (h *Handler) AutomationAllocateIPAddress(c *fiber.Ctx) error {
	var req struct {
		SubnetID    int64  `json:"subnet_id"`
		DeviceID    *int64 `json:"device_id"`
		DryRun      bool   `json:"dry_run"`
		Idempotency string `json:"idempotency_key"`
	}
	if err := c.BodyParser(&req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}
	fields := make([]ValidationField, 0)
	if req.SubnetID <= 0 {
		fields = append(fields, ValidationField{Field: "subnet_id", Message: "subnet_id must be greater than zero"})
	}
	if len(fields) > 0 {
		return RespondValidationError(c, "validation failed", fields)
	}
	ip, decision, err := h.ops.Automation.AllocateIPAddress(c.Context(), req.SubnetID, req.DeviceID, req.DryRun)
	return automationWriteResponse(c, ip, decision, err)
}

func (h *Handler) AutomationReserveIPAddress(c *fiber.Ctx) error {
	var req struct {
		SubnetID int64  `json:"subnet_id"`
		Address  string `json:"address"`
		Hostname string `json:"hostname"`
		DryRun   bool   `json:"dry_run"`
	}
	if err := c.BodyParser(&req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}
	fields := make([]ValidationField, 0)
	if req.SubnetID <= 0 {
		fields = append(fields, ValidationField{Field: "subnet_id", Message: "subnet_id must be greater than zero"})
	}
	fields = append(fields, requiredFields(map[string]string{"address": req.Address})...)
	if len(fields) > 0 {
		return RespondValidationError(c, "validation failed", fields)
	}
	ip, decision, err := h.ops.Automation.ReserveIPAddress(c.Context(), req.SubnetID, req.Address, req.Hostname, req.DryRun)
	return automationWriteResponse(c, ip, decision, err)
}

func (h *Handler) AutomationReleaseIPAddress(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid IP address ID")
	}
	var req struct {
		DryRun bool `json:"dry_run"`
	}
	_ = c.BodyParser(&req)
	ip, decision, err := h.ops.Automation.ReleaseIPAddress(c.Context(), id, req.DryRun)
	return automationWriteResponse(c, ip, decision, err)
}

func (h *Handler) AutomationRegisterDevice(c *fiber.Ctx) error {
	req := new(repository.DeviceParams)
	var body struct {
		DryRun bool `json:"dry_run"`
	}
	if err := c.BodyParser(req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}
	_ = c.BodyParser(&body)
	req.Hostname = strings.TrimSpace(req.Hostname)
	if req.Hostname == "" {
		return RespondValidationError(c, "validation failed", []ValidationField{{Field: "hostname", Message: "hostname is required"}})
	}
	device, decision, err := h.ops.Automation.RegisterDevice(c.Context(), req, body.DryRun)
	return automationWriteResponse(c, device, decision, err)
}

func (h *Handler) AutomationDNSUpdate(c *fiber.Ctx) error {
	var req struct {
		Zone   string `json:"zone"`
		Name   string `json:"name"`
		Type   string `json:"type"`
		Value  string `json:"value"`
		DryRun bool   `json:"dry_run"`
	}
	if err := c.BodyParser(&req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}
	if fields := requiredFields(map[string]string{"zone": req.Zone, "name": req.Name, "type": req.Type, "value": req.Value}); len(fields) > 0 {
		return RespondValidationError(c, "validation failed", fields)
	}
	decision, err := h.ops.Automation.Evaluate(c.Context(), services.AutomationRequest{
		Workflow: "dns",
		Action:   "update",
		Values:   map[string]string{"zone": req.Zone, "name": req.Name, "type": req.Type, "value": req.Value},
		DryRun:   req.DryRun,
	})
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to evaluate automation policy")
	}
	if !decision.Allowed {
		return automationWriteResponse(c, nil, decision, nil)
	}
	return c.JSON(fiber.Map{
		"policy":  decision,
		"status":  "accepted",
		"message": "DNS provider write is queued for external automation; use existing DNS provider settings for connection tests",
	})
}

func requiredFields(values map[string]string) []ValidationField {
	fields := make([]ValidationField, 0)
	for field, value := range values {
		if strings.TrimSpace(value) == "" {
			fields = append(fields, ValidationField{Field: field, Message: field + " is required"})
		}
	}
	return fields
}

func automationWriteResponse(c *fiber.Ctx, data interface{}, decision *services.PolicyDecision, err error) error {
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, err.Error())
	}
	if decision != nil && decision.ReviewNeeded {
		return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"policy": decision})
	}
	if decision != nil && !decision.Allowed {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"policy": decision})
	}
	if data == nil {
		return c.JSON(fiber.Map{"policy": decision})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"policy": decision, "data": data})
}

func validateAutomationPolicyRequest(req *automationPolicyRequest) []ValidationField {
	fields := make([]ValidationField, 0)
	if strings.TrimSpace(req.Name) == "" {
		fields = append(fields, ValidationField{Field: "name", Message: "name is required"})
	}
	if req.Effect != "" {
		switch strings.TrimSpace(req.Effect) {
		case "allow", "deny", "manual_review":
		default:
			fields = append(fields, ValidationField{Field: "effect", Message: "effect must be allow, deny, or manual_review"})
		}
	}
	return fields
}

func (h *Handler) ListIntegrationTemplates(c *fiber.Ctx) error {
	return c.JSON(services.IntegrationTemplates())
}

func (h *Handler) ListAPITokenAnalytics(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminWrite) {
		return nil
	}
	tokens, err := h.service.GetRepository().ListAPITokenAnalytics(c.Context())
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to load API token analytics")
	}
	limitStr, _ := h.service.Config.GetCtx(c.Context(), "api_token_rate_limit_per_minute")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 {
		limit = 100
	}
	response := make([]fiber.Map, 0, len(tokens))
	for _, token := range tokens {
		response = append(response, fiber.Map{
			"id": token.ID, "user_id": token.UserID, "username": token.Username,
			"name": token.Name, "scope": token.Scope, "usage_count": token.UsageCount,
			"last_used_at": formatOptionalTime(token.LastUsedAt), "last_used_ip": token.LastUsedIP,
			"expires_at": formatOptionalTime(token.ExpiresAt), "created_at": token.CreatedAt.Format(time.RFC3339),
			"rate_limit_per_minute": limit, "is_rotated": token.RotationGraceExpiresAt != nil,
		})
	}
	return c.JSON(response)
}

func formatOptionalTime(value *time.Time) *string {
	if value == nil {
		return nil
	}
	out := value.Format(time.RFC3339)
	return &out
}
