package services

import (
	"context"
	"testing"
	"time"

	"padduck/models"
)

// stubAutomationRepo satisfies automationRepo for Simulate tests.
type stubAutomationRepo struct {
	policies []*models.AutomationPolicy
}

func (s *stubAutomationRepo) ListAutomationPolicies(_ context.Context) ([]*models.AutomationPolicy, error) {
	return s.policies, nil
}
func (s *stubAutomationRepo) ListEnabledAutomationPolicies(_ context.Context, workflow, action string) ([]*models.AutomationPolicy, error) {
	var out []*models.AutomationPolicy
	for _, p := range s.policies {
		if !p.Enabled {
			continue
		}
		if (p.Workflow == "*" || p.Workflow == workflow) && (p.Action == "*" || p.Action == action) {
			out = append(out, p)
		}
	}
	return out, nil
}
func (s *stubAutomationRepo) CreateAutomationPolicy(_ context.Context, p *models.AutomationPolicy) (*models.AutomationPolicy, error) {
	return p, nil
}
func (s *stubAutomationRepo) UpdateAutomationPolicy(_ context.Context, p *models.AutomationPolicy) (*models.AutomationPolicy, error) {
	return p, nil
}
func (s *stubAutomationRepo) DeleteAutomationPolicy(_ context.Context, _ int64) error { return nil }
func (s *stubAutomationRepo) GetUserByID(_ context.Context, _ int64) (*models.User, error) {
	return nil, nil
}
func (s *stubAutomationRepo) GetUsersByRole(_ context.Context, _ string) ([]*models.User, error) {
	return nil, nil
}
func (s *stubAutomationRepo) CreateScanJob(_ context.Context, _ string, _ []int64, _ *string, _ int64, _ bool) (*models.ScanJob, error) {
	return &models.ScanJob{}, nil
}
func (s *stubAutomationRepo) UpdateIPAddressTag(_ context.Context, _ int64, _ *int64) (*models.IPAddress, error) {
	return &models.IPAddress{}, nil
}

func makePolicy(id int64, workflow, action, effect string, conds []models.PolicyCondition, actions []models.PolicyAction) *models.AutomationPolicy {
	return &models.AutomationPolicy{
		ID: id, Name: "p" + action, Workflow: workflow, Action: action,
		Effect: effect, Enabled: true, Conditions: conds, Actions: actions,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
}

func TestSimulate(t *testing.T) {
	tagAction := models.PolicyAction{Type: "tag", Params: map[string]string{"tag_id": "5"}}
	notifyAction := models.PolicyAction{Type: "notify", Params: map[string]string{"role": "admin"}}

	tests := []struct {
		name              string
		policies          []*models.AutomationPolicy
		workflow          string
		action            string
		ctx               map[string]string
		wantAllowed       bool
		wantReview        bool
		wantDecision      string
		wantMatchedCount  int
		wantActionsLen    int
	}{
		{
			name:             "no policies — allow by default",
			policies:         nil,
			workflow:         "subnet", action: "create",
			ctx:              map[string]string{"prefix_len": "24"},
			wantAllowed:      true, wantDecision: "allow", wantMatchedCount: 0,
		},
		{
			name: "allow policy matches",
			policies: []*models.AutomationPolicy{
				makePolicy(1, "subnet", "create", "allow", nil, []models.PolicyAction{tagAction}),
			},
			workflow: "subnet", action: "create",
			ctx:             map[string]string{"prefix_len": "24"},
			wantAllowed:     true, wantDecision: "allow", wantMatchedCount: 1, wantActionsLen: 1,
		},
		{
			name: "deny policy stops evaluation",
			policies: []*models.AutomationPolicy{
				makePolicy(1, "subnet", "create", "deny", nil, nil),
				makePolicy(2, "subnet", "create", "allow", nil, []models.PolicyAction{tagAction}),
			},
			workflow: "subnet", action: "create",
			ctx:          map[string]string{},
			wantAllowed:  false, wantDecision: "deny", wantMatchedCount: 1,
		},
		{
			name: "manual_review stops evaluation",
			policies: []*models.AutomationPolicy{
				makePolicy(1, "*", "*", "manual_review", nil, nil),
			},
			workflow: "ip_address", action: "allocate",
			ctx:         map[string]string{},
			wantAllowed: false, wantReview: true, wantDecision: "manual_review", wantMatchedCount: 1,
		},
		{
			name: "condition mismatch — policy not matched",
			policies: []*models.AutomationPolicy{
				makePolicy(1, "subnet", "create", "deny",
					[]models.PolicyCondition{cond("prefix_len", "gt", "28")}, nil),
			},
			workflow: "subnet", action: "create",
			ctx:         map[string]string{"prefix_len": "24"},
			wantAllowed: true, wantDecision: "allow", wantMatchedCount: 0,
		},
		{
			name: "multiple allow policies — all matched, actions aggregated",
			policies: []*models.AutomationPolicy{
				makePolicy(1, "subnet", "*", "allow", nil, []models.PolicyAction{tagAction}),
				makePolicy(2, "*", "create", "allow", nil, []models.PolicyAction{notifyAction}),
			},
			workflow: "subnet", action: "create",
			ctx:             map[string]string{},
			wantAllowed:     true, wantDecision: "allow", wantMatchedCount: 2,
		},
		{
			name: "no side effects — simulate does not call ExecuteActions",
			policies: []*models.AutomationPolicy{
				makePolicy(1, "subnet", "create", "allow", nil, []models.PolicyAction{tagAction}),
			},
			workflow: "subnet", action: "create",
			ctx:         map[string]string{},
			wantAllowed: true, wantDecision: "allow", wantMatchedCount: 1, wantActionsLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &AutomationService{repo: &stubAutomationRepo{policies: tt.policies}}
			result, err := svc.Simulate(context.Background(), tt.workflow, tt.action, tt.ctx)
			if err != nil {
				t.Fatalf("Simulate() error: %v", err)
			}
			if result.Allowed != tt.wantAllowed {
				t.Errorf("Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}
			if result.ReviewNeeded != tt.wantReview {
				t.Errorf("ReviewNeeded = %v, want %v", result.ReviewNeeded, tt.wantReview)
			}
			if result.EffectiveDecision != tt.wantDecision {
				t.Errorf("EffectiveDecision = %q, want %q", result.EffectiveDecision, tt.wantDecision)
			}
			if len(result.MatchedPolicies) != tt.wantMatchedCount {
				t.Errorf("len(MatchedPolicies) = %d, want %d", len(result.MatchedPolicies), tt.wantMatchedCount)
			}
			if tt.wantActionsLen > 0 && len(result.MatchedPolicies) > 0 {
				if got := len(result.MatchedPolicies[0].ActionsWouldRun); got != tt.wantActionsLen {
					t.Errorf("ActionsWouldRun len = %d, want %d", got, tt.wantActionsLen)
				}
			}
		})
	}
}

func cond(field, op, value string) models.PolicyCondition {
	return models.PolicyCondition{Field: field, Operator: op, Value: value}
}

func TestPolicyMatches(t *testing.T) {
	tests := []struct {
		name       string
		conditions []models.PolicyCondition
		values     map[string]string
		want       bool
	}{
		{
			name:       "empty conditions always match",
			conditions: nil,
			values:     map[string]string{"hostname": "edge-01"},
			want:       true,
		},
		// eq
		{name: "eq match", conditions: []models.PolicyCondition{cond("subnet_id", "eq", "42")}, values: map[string]string{"subnet_id": "42"}, want: true},
		{name: "eq mismatch", conditions: []models.PolicyCondition{cond("subnet_id", "eq", "42")}, values: map[string]string{"subnet_id": "43"}, want: false},
		{name: "eq wildcard value matches anything", conditions: []models.PolicyCondition{cond("subnet_id", "eq", "*")}, values: map[string]string{"subnet_id": "99"}, want: true},
		{name: "eq missing field", conditions: []models.PolicyCondition{cond("hostname", "eq", "web01")}, values: map[string]string{"subnet_id": "1"}, want: false},
		// neq
		{name: "neq different", conditions: []models.PolicyCondition{cond("env", "neq", "prod")}, values: map[string]string{"env": "staging"}, want: true},
		{name: "neq same", conditions: []models.PolicyCondition{cond("env", "neq", "prod")}, values: map[string]string{"env": "prod"}, want: false},
		// contains
		{name: "contains hit", conditions: []models.PolicyCondition{cond("hostname", "contains", "web")}, values: map[string]string{"hostname": "web-01.example.com"}, want: true},
		{name: "contains miss", conditions: []models.PolicyCondition{cond("hostname", "contains", "db")}, values: map[string]string{"hostname": "web-01.example.com"}, want: false},
		// starts_with
		{name: "starts_with hit", conditions: []models.PolicyCondition{cond("hostname", "starts_with", "prod-")}, values: map[string]string{"hostname": "prod-router-01"}, want: true},
		{name: "starts_with miss", conditions: []models.PolicyCondition{cond("hostname", "starts_with", "prod-")}, values: map[string]string{"hostname": "dev-router-01"}, want: false},
		// ends_with
		{name: "ends_with hit", conditions: []models.PolicyCondition{cond("hostname", "ends_with", ".local")}, values: map[string]string{"hostname": "web01.local"}, want: true},
		{name: "ends_with miss", conditions: []models.PolicyCondition{cond("hostname", "ends_with", ".local")}, values: map[string]string{"hostname": "web01.example.com"}, want: false},
		// gt
		{name: "gt greater", conditions: []models.PolicyCondition{cond("prefix_len", "gt", "24")}, values: map[string]string{"prefix_len": "28"}, want: true},
		{name: "gt equal", conditions: []models.PolicyCondition{cond("prefix_len", "gt", "24")}, values: map[string]string{"prefix_len": "24"}, want: false},
		{name: "gt less", conditions: []models.PolicyCondition{cond("prefix_len", "gt", "24")}, values: map[string]string{"prefix_len": "16"}, want: false},
		{name: "gt non-numeric", conditions: []models.PolicyCondition{cond("prefix_len", "gt", "24")}, values: map[string]string{"prefix_len": "abc"}, want: false},
		// lt
		{name: "lt less", conditions: []models.PolicyCondition{cond("prefix_len", "lt", "24")}, values: map[string]string{"prefix_len": "16"}, want: true},
		{name: "lt equal", conditions: []models.PolicyCondition{cond("prefix_len", "lt", "24")}, values: map[string]string{"prefix_len": "24"}, want: false},
		{name: "lt greater", conditions: []models.PolicyCondition{cond("prefix_len", "lt", "24")}, values: map[string]string{"prefix_len": "28"}, want: false},
		// glob
		{name: "glob wildcard matches all", conditions: []models.PolicyCondition{cond("hostname", "glob", "*")}, values: map[string]string{"hostname": "anything"}, want: true},
		{name: "glob trailing star matches prefix", conditions: []models.PolicyCondition{cond("hostname", "glob", "prod-*")}, values: map[string]string{"hostname": "prod-router-01"}, want: true},
		{name: "glob trailing star no match", conditions: []models.PolicyCondition{cond("hostname", "glob", "prod-*")}, values: map[string]string{"hostname": "dev-router-01"}, want: false},
		{name: "glob exact", conditions: []models.PolicyCondition{cond("hostname", "glob", "web01")}, values: map[string]string{"hostname": "web01"}, want: true},
		// multiple conditions (AND semantics)
		{
			name: "all conditions match",
			conditions: []models.PolicyCondition{
				cond("subnet_id", "eq", "5"),
				cond("prefix_len", "gt", "24"),
			},
			values: map[string]string{"subnet_id": "5", "prefix_len": "28"},
			want:   true,
		},
		{
			name: "one condition fails",
			conditions: []models.PolicyCondition{
				cond("subnet_id", "eq", "5"),
				cond("prefix_len", "gt", "24"),
			},
			values: map[string]string{"subnet_id": "5", "prefix_len": "16"},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := policyMatches(tt.conditions, tt.values); got != tt.want {
				t.Fatalf("policyMatches() = %v, want %v", got, tt.want)
			}
		})
	}
}
