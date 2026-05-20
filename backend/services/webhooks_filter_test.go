package services

import (
	"testing"

	"padduck/models"
)

func TestEndpointAllowsResource(t *testing.T) {
	event := WebhookEvent{
		EventType:    "ip_address.assigned",
		Action:       "assigned",
		ResourceType: "ip_address",
		ResourceName: "prod-router-01",
		NewValues: map[string]interface{}{
			"tag":    "production",
			"status": "assigned",
		},
	}

	tests := []struct {
		name     string
		endpoint *models.WebhookEndpoint
		allowed  bool
	}{
		{
			name:     "no resource filters allow",
			endpoint: &models.WebhookEndpoint{},
			allowed:  true,
		},
		{
			name:     "object type match",
			endpoint: &models.WebhookEndpoint{ObjectTypes: []string{"ip_address"}},
			allowed:  true,
		},
		{
			name:     "object type mismatch",
			endpoint: &models.WebhookEndpoint{ObjectTypes: []string{"device"}},
			allowed:  false,
		},
		{
			name:     "tag match",
			endpoint: &models.WebhookEndpoint{TagFilters: []string{"production"}},
			allowed:  true,
		},
		{
			name:     "condition prefix match",
			endpoint: &models.WebhookEndpoint{FilterConditions: map[string]string{"resource_name": "prod-*"}},
			allowed:  true,
		},
		{
			name:     "condition mismatch",
			endpoint: &models.WebhookEndpoint{FilterConditions: map[string]string{"status": "reserved"}},
			allowed:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := endpointAllowsResource(tt.endpoint, event); got != tt.allowed {
				t.Fatalf("endpointAllowsResource() = %v, want %v", got, tt.allowed)
			}
		})
	}
}
