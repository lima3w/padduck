package handlers

import (
	"net/http"
	"testing"
)

func TestWebhookRoutes_AuthRequired(t *testing.T) {
	h := &Handler{}
	assertAuthRequired(t, []authRoute{
		{
			name:        "list webhook endpoints",
			method:      http.MethodGet,
			routePath:   "/webhooks",
			requestPath: "/webhooks",
			handler:     h.ListWebhookEndpoints,
		},
		{
			name:        "create webhook endpoint",
			method:      http.MethodPost,
			routePath:   "/webhooks",
			requestPath: "/webhooks",
			body:        `{"name":"test","url":"https://example.com/hook","events":["ip.created"]}`,
			handler:     h.CreateWebhookEndpoint,
		},
		{
			name:        "get sample webhook payload",
			method:      http.MethodGet,
			routePath:   "/webhooks/sample-payload",
			requestPath: "/webhooks/sample-payload",
			handler:     h.GetWebhookSamplePayload,
		},
		{
			name:        "update webhook endpoint",
			method:      http.MethodPut,
			routePath:   "/webhooks/:id",
			requestPath: "/webhooks/1",
			body:        `{"name":"updated","url":"https://example.com/hook","events":["ip.created"]}`,
			handler:     h.UpdateWebhookEndpoint,
		},
		{
			name:        "delete webhook endpoint",
			method:      http.MethodDelete,
			routePath:   "/webhooks/:id",
			requestPath: "/webhooks/1",
			handler:     h.DeleteWebhookEndpoint,
		},
		{
			name:        "list webhook deliveries",
			method:      http.MethodGet,
			routePath:   "/webhooks/deliveries",
			requestPath: "/webhooks/deliveries",
			handler:     h.ListWebhookDeliveries,
		},
	})
}
