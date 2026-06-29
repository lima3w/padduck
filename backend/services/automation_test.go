package services

import (
	"testing"

	"padduck/models"
)

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
