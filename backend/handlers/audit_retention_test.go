package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// All four handlers check requirePerm before touching params/body, so the
// invalid-body/retention-days-validation branches are only reachable by a
// permitted user — which requires a live repo (see plan). Only the auth
// guard branches are testable here without a DB.

func TestGetAuditRetention_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/audit/retention", h.GetAuditRetention)

	req := httptest.NewRequest("GET", "/admin/audit/retention", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetAuditRetention_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/audit/retention", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.GetAuditRetention(c)
	})

	req := httptest.NewRequest("GET", "/admin/audit/retention", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestUpdateAuditRetention_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Put("/admin/audit/retention", h.UpdateAuditRetention)

	req := httptest.NewRequest("PUT", "/admin/audit/retention", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestUpdateAuditRetention_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Put("/admin/audit/retention", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.UpdateAuditRetention(c)
	})

	req := httptest.NewRequest("PUT", "/admin/audit/retention", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestPruneAuditLogs_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/admin/audit/prune", h.PruneAuditLogs)

	req := httptest.NewRequest("POST", "/admin/audit/prune", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestPruneAuditLogs_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/admin/audit/prune", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.PruneAuditLogs(c)
	})

	req := httptest.NewRequest("POST", "/admin/audit/prune", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestExportAuditLog_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/audit/export", h.ExportAuditLog)

	req := httptest.NewRequest("GET", "/admin/audit/export", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestExportAuditLog_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/audit/export", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.ExportAuditLog(c)
	})

	req := httptest.NewRequest("GET", "/admin/audit/export", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}
