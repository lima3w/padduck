package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"ipam-next/models"
)

func TestEndpointAllowsEvent(t *testing.T) {
	tests := []struct {
		name     string
		events   []string
		event    string
		expected bool
	}{
		{name: "empty events allow all", events: nil, event: "subnet.created", expected: true},
		{name: "wildcard allows all", events: []string{"*"}, event: "ip_address.assigned", expected: true},
		{name: "exact match", events: []string{"subnet.created"}, event: "subnet.created", expected: true},
		{name: "prefix wildcard", events: []string{"subnet.*"}, event: "subnet.deleted", expected: true},
		{name: "no match", events: []string{"vlan.*"}, event: "subnet.created", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endpoint := &models.WebhookEndpoint{Events: tt.events}
			assert.Equal(t, tt.expected, endpointAllowsEvent(endpoint, tt.event))
		})
	}
}

func TestValidateWebhookEndpoint(t *testing.T) {
	endpoint := &models.WebhookEndpoint{
		Name:   "  integrations ",
		URL:    "https://example.test/webhook",
		Events: []string{" subnet.created ", "", "ip_address.*"},
	}
	err := validateWebhookEndpoint(endpoint)
	assert.NoError(t, err)
	assert.Equal(t, "integrations", endpoint.Name)
	assert.Equal(t, []string{"subnet.created", "ip_address.*"}, endpoint.Events)
}

func TestValidateWebhookEndpointRejectsBadURL(t *testing.T) {
	err := validateWebhookEndpoint(&models.WebhookEndpoint{Name: "bad", URL: "ftp://example.test/hook"})
	assert.Error(t, err)
}

func TestSignWebhookPayload(t *testing.T) {
	got := signWebhookPayload("secret", []byte(`{"event":"subnet.created"}`))
	assert.Equal(t, "sha256=b2964d578374d13263f8f4ab7d4a4545a3b7c9ed0ae52cb63f88866f5882a5af", got)
}
