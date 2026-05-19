package services

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"ipam-next/models"
	"ipam-next/repository"
)

type AutomationService struct {
	svc *Service
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

func NewAutomationService(svc *Service) *AutomationService {
	return &AutomationService{svc: svc}
}

func (a *AutomationService) ListPolicies(ctx context.Context) ([]*models.AutomationPolicy, error) {
	return a.svc.repository.ListAutomationPolicies(ctx)
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
		return a.svc.repository.UpdateAutomationPolicy(ctx, policy)
	}
	return a.svc.repository.CreateAutomationPolicy(ctx, policy)
}

func (a *AutomationService) DeletePolicy(ctx context.Context, id int64) error {
	return a.svc.repository.DeleteAutomationPolicy(ctx, id)
}

func (a *AutomationService) Evaluate(ctx context.Context, req AutomationRequest) (*PolicyDecision, error) {
	policies, err := a.svc.repository.ListEnabledAutomationPolicies(ctx, req.Workflow, req.Action)
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

func (a *AutomationService) AllocateIPAddress(ctx context.Context, subnetID int64, assignedTo string, dryRun bool) (*models.IPAddress, *PolicyDecision, error) {
	decision, err := a.Evaluate(ctx, AutomationRequest{
		Workflow: "ip_address",
		Action:   "allocate",
		Values: map[string]string{
			"subnet_id":    strconv.FormatInt(subnetID, 10),
			"assigned_to":  assignedTo,
			"requested_by": "automation",
		},
		DryRun: dryRun,
	})
	if err != nil || !decision.Allowed || dryRun {
		return nil, decision, err
	}
	ip, err := a.svc.AllocateIPAddress(ctx, subnetID, assignedTo)
	return ip, decision, err
}

func (a *AutomationService) ReserveIPAddress(ctx context.Context, subnetID int64, address, hostname string, dryRun bool) (*models.IPAddress, *PolicyDecision, error) {
	values := map[string]string{"subnet_id": strconv.FormatInt(subnetID, 10), "address": address, "hostname": hostname}
	decision, err := a.Evaluate(ctx, AutomationRequest{Workflow: "ip_address", Action: "reserve", Values: values, DryRun: dryRun})
	if err != nil || !decision.Allowed || dryRun {
		return nil, decision, err
	}
	ip, err := a.svc.CreateIPAddress(ctx, subnetID, address, hostname, "reserved", nil, nil, nil, nil)
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
	ip, err := a.svc.ReleaseIPAddress(ctx, ipID)
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
	device, err := a.svc.CreateDevice(ctx, params)
	return device, decision, err
}

func policyMatches(conditions, values map[string]string) bool {
	for key, expected := range conditions {
		expected = strings.TrimSpace(expected)
		if expected == "" || expected == "*" {
			continue
		}
		actual, ok := values[key]
		if !ok {
			return false
		}
		switch {
		case strings.HasSuffix(expected, "*"):
			if !strings.HasPrefix(actual, strings.TrimSuffix(expected, "*")) {
				return false
			}
		case actual != expected:
			return false
		}
	}
	return true
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
