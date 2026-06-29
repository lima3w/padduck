package services

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"padduck/models"
	"padduck/repository"
)

type automationRepo interface {
	ListAutomationPolicies(ctx context.Context) ([]*models.AutomationPolicy, error)
	UpdateAutomationPolicy(ctx context.Context, policy *models.AutomationPolicy) (*models.AutomationPolicy, error)
	CreateAutomationPolicy(ctx context.Context, policy *models.AutomationPolicy) (*models.AutomationPolicy, error)
	DeleteAutomationPolicy(ctx context.Context, id int64) error
	ListEnabledAutomationPolicies(ctx context.Context, workflow, action string) ([]*models.AutomationPolicy, error)

	GetUserByID(ctx context.Context, id int64) (*models.User, error)
	GetUsersByRole(ctx context.Context, role string) ([]*models.User, error)
	CreateScanJob(ctx context.Context, name string, subnetIDs []int64, scheduleCron *string, createdBy int64, autoAddIPs bool) (*models.ScanJob, error)
	UpdateIPAddressTag(ctx context.Context, ipID int64, tagID *int64) (*models.IPAddress, error)
}

type automationIPAM interface {
	AllocateIPAddress(ctx context.Context, subnetID int64, deviceID *int64) (*models.IPAddress, error)
	CreateIPAddress(ctx context.Context, subnetID int64, address, hostname, status string, tagID *int64, macAddress, ptrRecord, dnsName *string, customFields ...map[string]*string) (*models.IPAddress, error)
	ReleaseIPAddress(ctx context.Context, id int64) (*models.IPAddress, error)
	CreateDevice(ctx context.Context, req *DeviceCreateRequest) (*models.Device, error)
}

// AutomationService handles policy evaluation and built-in action execution.
type AutomationService struct {
	repo          automationRepo
	ipam          automationIPAM
	webhooks      *WebhookService
	notifications *NotificationService
}

type AutomationRequest struct {
	Workflow string
	Action   string
	Values   map[string]string
	DryRun   bool
}

type PolicyDecision struct {
	Allowed      bool                     `json:"allowed"`
	ReviewNeeded bool                     `json:"review_needed"`
	Policy       *models.AutomationPolicy `json:"policy,omitempty"`
	Message      string                   `json:"message,omitempty"`
	Values       map[string]string        `json:"values,omitempty"`
}

func NewAutomationService(repo automationRepo, ipam automationIPAM, webhooks *WebhookService, notifications *NotificationService) *AutomationService {
	return &AutomationService{repo: repo, ipam: ipam, webhooks: webhooks, notifications: notifications}
}

func (a *AutomationService) ListPolicies(ctx context.Context) ([]*models.AutomationPolicy, error) {
	return a.repo.ListAutomationPolicies(ctx)
}

func (a *AutomationService) SavePolicy(ctx context.Context, policy *models.AutomationPolicy) (*models.AutomationPolicy, error) {
	policy.Name = strings.TrimSpace(policy.Name)
	policy.Workflow = strings.TrimSpace(policy.Workflow)
	policy.Action = strings.TrimSpace(policy.Action)
	policy.Effect = strings.TrimSpace(policy.Effect)
	if policy.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if policy.Workflow == "" {
		policy.Workflow = "*"
	}
	if policy.Action == "" {
		policy.Action = "*"
	}
	if policy.Effect == "" {
		policy.Effect = "allow"
	}
	switch policy.Effect {
	case "allow", "deny", "manual_review":
	default:
		return nil, fmt.Errorf("effect must be allow, deny, or manual_review")
	}
	if policy.ID > 0 {
		return a.repo.UpdateAutomationPolicy(ctx, policy)
	}
	return a.repo.CreateAutomationPolicy(ctx, policy)
}

func (a *AutomationService) DeletePolicy(ctx context.Context, id int64) error {
	return a.repo.DeleteAutomationPolicy(ctx, id)
}

func (a *AutomationService) Evaluate(ctx context.Context, req AutomationRequest) (*PolicyDecision, error) {
	policies, err := a.repo.ListEnabledAutomationPolicies(ctx, req.Workflow, req.Action)
	if err != nil {
		return nil, err
	}
	for _, policy := range policies {
		if !policyMatches(policy.Conditions, req.Values) {
			continue
		}
		decision := &PolicyDecision{
			Allowed: true,
			Policy:  policy,
			Message: policy.Message,
			Values:  req.Values,
		}
		switch policy.Effect {
		case "deny":
			decision.Allowed = false
			if decision.Message == "" {
				decision.Message = "automation request denied by policy"
			}
			return decision, nil
		case "manual_review":
			decision.Allowed = false
			decision.ReviewNeeded = true
			if decision.Message == "" {
				decision.Message = "automation request requires manual review"
			}
			return decision, nil
		}
	}
	return &PolicyDecision{Allowed: true, Values: req.Values}, nil
}

func (a *AutomationService) AllocateIPAddress(ctx context.Context, subnetID int64, deviceID *int64, dryRun bool) (*models.IPAddress, *PolicyDecision, error) {
	decision, err := a.Evaluate(ctx, AutomationRequest{
		Workflow: "ip_address",
		Action:   "allocate",
		Values: map[string]string{
			"subnet_id":    strconv.FormatInt(subnetID, 10),
			"requested_by": "automation",
		},
		DryRun: dryRun,
	})
	if err != nil || !decision.Allowed || dryRun {
		return nil, decision, err
	}
	ip, err := a.ipam.AllocateIPAddress(ctx, subnetID, deviceID)
	return ip, decision, err
}

func (a *AutomationService) ReserveIPAddress(ctx context.Context, subnetID int64, address, hostname string, dryRun bool) (*models.IPAddress, *PolicyDecision, error) {
	values := map[string]string{"subnet_id": strconv.FormatInt(subnetID, 10), "address": address, "hostname": hostname}
	decision, err := a.Evaluate(ctx, AutomationRequest{Workflow: "ip_address", Action: "reserve", Values: values, DryRun: dryRun})
	if err != nil || !decision.Allowed || dryRun {
		return nil, decision, err
	}
	ip, err := a.ipam.CreateIPAddress(ctx, subnetID, address, hostname, "reserved", nil, nil, nil, nil)
	return ip, decision, err
}

func (a *AutomationService) ReleaseIPAddress(ctx context.Context, ipID int64, dryRun bool) (*models.IPAddress, *PolicyDecision, error) {
	decision, err := a.Evaluate(ctx, AutomationRequest{
		Workflow: "ip_address",
		Action:   "release",
		Values:   map[string]string{"ip_id": strconv.FormatInt(ipID, 10)},
		DryRun:   dryRun,
	})
	if err != nil || !decision.Allowed || dryRun {
		return nil, decision, err
	}
	ip, err := a.ipam.ReleaseIPAddress(ctx, ipID)
	return ip, decision, err
}

func (a *AutomationService) RegisterDevice(ctx context.Context, params *repository.DeviceParams, dryRun bool) (*models.Device, *PolicyDecision, error) {
	decision, err := a.Evaluate(ctx, AutomationRequest{
		Workflow: "device",
		Action:   "register",
		Values:   map[string]string{"hostname": params.Hostname},
		DryRun:   dryRun,
	})
	if err != nil || !decision.Allowed || dryRun {
		return nil, decision, err
	}
	device, err := a.ipam.CreateDevice(ctx, params)
	return device, decision, err
}

func policyMatches(conditions []models.PolicyCondition, values map[string]string) bool {
	for _, cond := range conditions {
		actual := values[cond.Field]
		if !evalCondition(cond.Operator, actual, cond.Value) {
			return false
		}
	}
	return true
}

// evalCondition evaluates a single field condition using the given operator.
func evalCondition(op, actual, expected string) bool {
	switch op {
	case "eq", "":
		if expected == "" || expected == "*" {
			return true
		}
		return actual == expected
	case "neq":
		return actual != expected
	case "contains":
		return strings.Contains(actual, expected)
	case "starts_with":
		return strings.HasPrefix(actual, expected)
	case "ends_with":
		return strings.HasSuffix(actual, expected)
	case "gt":
		a, ea := toFloat(actual), toFloat(expected)
		return a != nil && ea != nil && *a > *ea
	case "lt":
		a, ea := toFloat(actual), toFloat(expected)
		return a != nil && ea != nil && *a < *ea
	case "glob":
		if expected == "" || expected == "*" {
			return true
		}
		if strings.HasSuffix(expected, "*") {
			return strings.HasPrefix(actual, strings.TrimSuffix(expected, "*"))
		}
		return actual == expected
	default:
		return actual == expected
	}
}

func toFloat(s string) *float64 {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil
	}
	return &v
}

// MatchedPolicyResult describes how one policy evaluated during a simulation.
type MatchedPolicyResult struct {
	PolicyID          int64    `json:"policy_id"`
	PolicyName        string   `json:"policy_name"`
	Decision          string   `json:"decision"` // allow | deny | manual_review
	ActionsWouldRun   []string `json:"actions_would_execute"`
}

// SimulationResult is the response from Simulate — read-only, no side effects.
type SimulationResult struct {
	Allowed           bool                  `json:"allowed"`
	ReviewNeeded      bool                  `json:"review_needed"`
	EffectiveDecision string                `json:"effective_decision"`
	MatchedPolicies   []MatchedPolicyResult `json:"matched_policies"`
}

// Simulate runs policy evaluation in read-only mode: no goroutines, no DB writes,
// no webhooks or notifications. Returns the full set of policies that matched and
// what they would do.
func (a *AutomationService) Simulate(ctx context.Context, workflow, action string, values map[string]string) (*SimulationResult, error) {
	policies, err := a.repo.ListEnabledAutomationPolicies(ctx, workflow, action)
	if err != nil {
		return nil, err
	}

	result := &SimulationResult{
		Allowed:           true,
		EffectiveDecision: "allow",
		MatchedPolicies:   []MatchedPolicyResult{},
	}

	for _, policy := range policies {
		if !policyMatches(policy.Conditions, values) {
			continue
		}

		decision := policy.Effect
		if decision == "" {
			decision = "allow"
		}

		matched := MatchedPolicyResult{
			PolicyID:        policy.ID,
			PolicyName:      policy.Name,
			Decision:        decision,
			ActionsWouldRun: describeActions(policy.Actions),
		}
		result.MatchedPolicies = append(result.MatchedPolicies, matched)

		switch policy.Effect {
		case "deny":
			result.Allowed = false
			result.EffectiveDecision = "deny"
			return result, nil
		case "manual_review":
			result.Allowed = false
			result.ReviewNeeded = true
			result.EffectiveDecision = "manual_review"
			return result, nil
		}
	}

	return result, nil
}

// describeActions returns human-readable strings for each action, used in simulation output.
func describeActions(actions []models.PolicyAction) []string {
	if len(actions) == 0 {
		return []string{}
	}
	out := make([]string, 0, len(actions))
	for _, a := range actions {
		switch a.Type {
		case "notify":
			if uid := a.Params["user_id"]; uid != "" {
				out = append(out, fmt.Sprintf("notify(user_id=%s)", uid))
			} else if role := a.Params["role"]; role != "" {
				out = append(out, fmt.Sprintf("notify(role=%s)", role))
			} else {
				out = append(out, "notify()")
			}
		case "webhook":
			out = append(out, fmt.Sprintf("webhook(webhook_id=%s)", a.Params["webhook_id"]))
		case "audit_annotation":
			out = append(out, fmt.Sprintf("audit_annotation(%s)", a.Params["message"]))
		case "scan":
			out = append(out, fmt.Sprintf("scan(profile_id=%s)", a.Params["profile_id"]))
		case "tag":
			out = append(out, fmt.Sprintf("tag(tag_id=%s)", a.Params["tag_id"]))
		default:
			out = append(out, a.Type)
		}
	}
	return out
}

// ExecuteActions runs each action in the matched policy asynchronously.
// Each action runs in its own goroutine; failures are logged and do not
// affect the primary operation that triggered the policy match.
func (a *AutomationService) ExecuteActions(policy *models.AutomationPolicy, resourceType string, resourceID int64, resourceName string) {
	if len(policy.Actions) == 0 {
		return
	}
	for _, act := range policy.Actions {
		go func(action models.PolicyAction) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			if err := a.runAction(ctx, action, policy, resourceType, resourceID, resourceName); err != nil {
				slog.Warn("automation: action failed",
					"policy", policy.Name, "action", action.Type, "error", err)
			}
		}(act)
	}
}

func (a *AutomationService) runAction(ctx context.Context, action models.PolicyAction, policy *models.AutomationPolicy, resourceType string, resourceID int64, resourceName string) error {
	switch action.Type {
	case "notify":
		return a.runNotifyAction(ctx, action.Params, policy, resourceType, resourceID, resourceName)
	case "webhook":
		return a.runWebhookAction(ctx, action.Params, policy, resourceType, resourceID, resourceName)
	case "audit_annotation":
		msg := action.Params["message"]
		slog.Info("automation audit annotation",
			"policy", policy.Name, "resource_type", resourceType, "resource_id", resourceID,
			"resource_name", resourceName, "message", msg)
		return nil
	case "scan":
		return a.runScanAction(ctx, action.Params, resourceType, resourceID)
	case "tag":
		return a.runTagAction(ctx, action.Params, resourceType, resourceID)
	default:
		return fmt.Errorf("unknown action type %q", action.Type)
	}
}

func (a *AutomationService) runNotifyAction(ctx context.Context, params map[string]string, policy *models.AutomationPolicy, resourceType string, resourceID int64, resourceName string) error {
	if a.notifications == nil {
		return fmt.Errorf("notification service not available")
	}
	data := map[string]interface{}{
		"policy_name":   policy.Name,
		"resource_type": resourceType,
		"resource_id":   resourceID,
		"resource_name": resourceName,
	}
	if uidStr := params["user_id"]; uidStr != "" {
		uid, err := strconv.ParseInt(uidStr, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid user_id %q", uidStr)
		}
		return a.notifications.Queue(ctx, uid, "automation_action", data)
	}
	if role := params["role"]; role != "" {
		users, err := a.repo.GetUsersByRole(ctx, role)
		if err != nil {
			return err
		}
		for _, u := range users {
			_ = a.notifications.Queue(ctx, u.ID, "automation_action", data)
		}
		return nil
	}
	return fmt.Errorf("notify action requires user_id or role param")
}

func (a *AutomationService) runWebhookAction(ctx context.Context, params map[string]string, policy *models.AutomationPolicy, resourceType string, resourceID int64, resourceName string) error {
	if a.webhooks == nil {
		return fmt.Errorf("webhook service not available")
	}
	a.webhooks.Queue(ctx, WebhookEvent{
		EventType:    "automation.action",
		Action:       policy.Effect,
		ResourceType: resourceType,
		ResourceID:   &resourceID,
		ResourceName: resourceName,
		Status:       "success",
		NewValues:    map[string]string{"policy": policy.Name},
		OccurredAt:   time.Now().UTC(),
	})
	return nil
}

func (a *AutomationService) runScanAction(ctx context.Context, params map[string]string, resourceType string, resourceID int64) error {
	if resourceType != "subnet" {
		return fmt.Errorf("scan action only supported for subnet resources")
	}
	name := fmt.Sprintf("auto-scan subnet %d", resourceID)
	_, err := a.repo.CreateScanJob(ctx, name, []int64{resourceID}, nil, 1, false)
	return err
}

func (a *AutomationService) runTagAction(ctx context.Context, params map[string]string, resourceType string, resourceID int64) error {
	if resourceType != "ip_address" {
		return fmt.Errorf("tag action only supported for ip_address resources")
	}
	tagIDStr := params["tag_id"]
	if tagIDStr == "" {
		return fmt.Errorf("tag action requires tag_id param")
	}
	tagID, err := strconv.ParseInt(tagIDStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid tag_id %q", tagIDStr)
	}
	_, err = a.repo.UpdateIPAddressTag(ctx, resourceID, &tagID)
	return err
}

func IntegrationTemplates() []models.IntegrationTemplate {
	return []models.IntegrationTemplate{
		{
			ID: "n8n", Name: "n8n", Category: "workflow", Description: "Self-hosted workflow automation with webhook triggers and HTTP action nodes.",
			Steps:     []string{"Create a Webhook Trigger node.", "Configure an IPAM webhook endpoint with the n8n trigger URL.", "Use HTTP Request nodes with Authorization: Bearer <token>.", "Use /api/v1/automation endpoints for controlled writes."},
			Endpoints: []string{"/api/v1/automation/ip-addresses/allocate", "/api/v1/automation/ip-addresses/reserve", "/api/v1/automation/ip-addresses/{id}/release"},
		},
		{
			ID: "zapier", Name: "Zapier", Category: "workflow", Description: "Cloud automation using Webhooks by Zapier and HTTP actions.",
			Steps:     []string{"Use Webhooks by Zapier as the trigger.", "Paste the Zap webhook URL into Admin > Webhooks.", "Use HTTP by Zapier for IPAM actions.", "Map webhook payload fields into automation endpoint request bodies."},
			Endpoints: []string{"/api/v1/automation/devices/register", "/api/v1/automation/dns/update"},
		},
		{
			ID: "make", Name: "Make", Category: "workflow", Description: "Visual scenarios with webhook modules and HTTP modules.",
			Steps:     []string{"Create a custom webhook module.", "Subscribe IPAM events by object type or event wildcard.", "Add an HTTP module with a bearer token.", "Enable dry_run first to verify policy decisions."},
			Endpoints: []string{"/api/v1/automation/ip-addresses/allocate", "/api/v1/automation/policies/evaluate"},
		},
		{
			ID: "ansible", Name: "Ansible", Category: "network", Description: "Use URI tasks to reserve addresses and register devices during provisioning.",
			Steps:     []string{"Store the API token in Ansible Vault.", "Call controlled automation endpoints from playbooks.", "Use policy responses to stop unsafe changes.", "Replay failed webhooks from Admin > Webhooks after downstream outages."},
			Endpoints: []string{"/api/v1/automation/ip-addresses/reserve", "/api/v1/automation/devices/register"},
		},
	}
}
