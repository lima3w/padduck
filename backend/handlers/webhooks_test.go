package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// ListWebhookEndpoints — GET /api/v1/webhooks
// ---------------------------------------------------------------------------

func TestListWebhookEndpoints_NoUser_Returns401(t *testing.T) {
	h := &Handler{}
	app := deviceApp(h, "GET", "/webhooks", h.ListWebhookEndpoints)
	resp, err := app.Test(httptest.NewRequest("GET", "/webhooks", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestListWebhookEndpoints_NoPermission_Returns403(t *testing.T) {
	h := &Handler{}
	app := deviceAppAs(h, "GET", "/webhooks", h.ListWebhookEndpoints, permUser())
	resp, err := app.Test(httptest.NewRequest("GET", "/webhooks", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// CreateWebhookEndpoint — POST /api/v1/webhooks
// ---------------------------------------------------------------------------

func TestCreateWebhookEndpoint_NoUser_Returns401(t *testing.T) {
	h := &Handler{}
	app := deviceApp(h, "POST", "/webhooks", h.CreateWebhookEndpoint)
	req := httptest.NewRequest("POST", "/webhooks", strings.NewReader(`{"name":"test","url":"https://example.com/hook","events":["ip.created"]}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestCreateWebhookEndpoint_NoPermission_Returns403(t *testing.T) {
	h := &Handler{}
	app := deviceAppAs(h, "POST", "/webhooks", h.CreateWebhookEndpoint, permUser())
	req := httptest.NewRequest("POST", "/webhooks", strings.NewReader(`{"name":"test","url":"https://example.com/hook","events":["ip.created"]}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// UpdateWebhookEndpoint — PUT /api/v1/webhooks/:id
// ---------------------------------------------------------------------------

func TestUpdateWebhookEndpoint_NoUser_Returns401(t *testing.T) {
	h := &Handler{}
	app := deviceApp(h, "PUT", "/webhooks/:id", h.UpdateWebhookEndpoint)
	req := httptest.NewRequest("PUT", "/webhooks/1", strings.NewReader(`{"name":"updated","url":"https://example.com/hook","events":["ip.created"]}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestUpdateWebhookEndpoint_NoPermission_Returns403(t *testing.T) {
	h := &Handler{}
	app := deviceAppAs(h, "PUT", "/webhooks/:id", h.UpdateWebhookEndpoint, permUser())
	req := httptest.NewRequest("PUT", "/webhooks/1", strings.NewReader(`{"name":"updated","url":"https://example.com/hook","events":["ip.created"]}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// DeleteWebhookEndpoint — DELETE /api/v1/webhooks/:id
// ---------------------------------------------------------------------------

func TestDeleteWebhookEndpoint_NoUser_Returns401(t *testing.T) {
	h := &Handler{}
	app := deviceApp(h, "DELETE", "/webhooks/:id", h.DeleteWebhookEndpoint)
	resp, err := app.Test(httptest.NewRequest("DELETE", "/webhooks/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestDeleteWebhookEndpoint_NoPermission_Returns403(t *testing.T) {
	h := &Handler{}
	app := deviceAppAs(h, "DELETE", "/webhooks/:id", h.DeleteWebhookEndpoint, permUser())
	resp, err := app.Test(httptest.NewRequest("DELETE", "/webhooks/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// ListWebhookDeliveries — GET /api/v1/webhooks/deliveries
// ---------------------------------------------------------------------------

func TestListWebhookDeliveries_NoUser_Returns401(t *testing.T) {
	h := &Handler{}
	app := deviceApp(h, "GET", "/webhooks/deliveries", h.ListWebhookDeliveries)
	resp, err := app.Test(httptest.NewRequest("GET", "/webhooks/deliveries", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestListWebhookDeliveries_NoPermission_Returns403(t *testing.T) {
	h := &Handler{}
	app := deviceAppAs(h, "GET", "/webhooks/deliveries", h.ListWebhookDeliveries, permUser())
	resp, err := app.Test(httptest.NewRequest("GET", "/webhooks/deliveries", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}
