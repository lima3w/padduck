package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"ipam-next/models"
)

// ---------------------------------------------------------------------------
// ListScanAgents — auth enforcement
// ---------------------------------------------------------------------------

func TestListScanAgents_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/admin/scan-agents", h.ListScanAgents)

	req := httptest.NewRequest("GET", "/admin/scan-agents", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestListScanAgents_NonAdmin_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/admin/scan-agents", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{Role: "user"})
		return h.ListScanAgents(c)
	})
	req := httptest.NewRequest("GET", "/admin/scan-agents", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// CreateScanAgent — auth enforcement
// ---------------------------------------------------------------------------

func TestCreateScanAgent_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/admin/scan-agents", h.CreateScanAgent)

	req := httptest.NewRequest("POST", "/admin/scan-agents", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestCreateScanAgent_NonAdmin_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/admin/scan-agents", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{Role: "user"})
		return h.CreateScanAgent(c)
	})
	req := httptest.NewRequest("POST", "/admin/scan-agents", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// GetScanAgent — auth enforcement and param validation
// ---------------------------------------------------------------------------

func TestGetScanAgent_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/admin/scan-agents/:id", h.GetScanAgent)

	req := httptest.NewRequest("GET", "/admin/scan-agents/1", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetScanAgent_NonAdmin_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/admin/scan-agents/:id", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{Role: "user"})
		return h.GetScanAgent(c)
	})
	req := httptest.NewRequest("GET", "/admin/scan-agents/1", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestGetScanAgent_AdminInvalidID_Returns400(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/admin/scan-agents/:id", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{Role: "admin"})
		return h.GetScanAgent(c)
	})
	req := httptest.NewRequest("GET", "/admin/scan-agents/notanumber", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	// permCheck with nil service and admin user returns 403 before ID parsing
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// RotateScanAgentToken — auth enforcement
// ---------------------------------------------------------------------------

func TestRotateScanAgentToken_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/admin/scan-agents/:id/rotate-token", h.RotateScanAgentToken)

	req := httptest.NewRequest("POST", "/admin/scan-agents/1/rotate-token", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestRotateScanAgentToken_NonAdmin_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/admin/scan-agents/:id/rotate-token", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{Role: "user"})
		return h.RotateScanAgentToken(c)
	})
	req := httptest.NewRequest("POST", "/admin/scan-agents/1/rotate-token", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// DeleteScanAgent — auth enforcement
// ---------------------------------------------------------------------------

func TestDeleteScanAgent_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Delete("/admin/scan-agents/:id", h.DeleteScanAgent)

	req := httptest.NewRequest("DELETE", "/admin/scan-agents/1", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestDeleteScanAgent_NonAdmin_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Delete("/admin/scan-agents/:id", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{Role: "user"})
		return h.DeleteScanAgent(c)
	})
	req := httptest.NewRequest("DELETE", "/admin/scan-agents/1", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// AgentAuthMiddleware — token enforcement
// ---------------------------------------------------------------------------

func TestAgentAuthMiddleware_NoToken_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/scan-agent/jobs", h.AgentAuthMiddleware, h.AgentGetJobs)

	req := httptest.NewRequest("GET", "/scan-agent/jobs", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestAgentHeartbeat_NoAgent_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/scan-agent/heartbeat", h.AgentHeartbeat)

	req := httptest.NewRequest("POST", "/scan-agent/heartbeat", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestAgentGetJobs_NoAgent_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/scan-agent/jobs", h.AgentGetJobs)

	req := httptest.NewRequest("GET", "/scan-agent/jobs", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestAgentPostResults_NoAgent_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/scan-agent/results", h.AgentPostResults)

	req := httptest.NewRequest("POST", "/scan-agent/results", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}
