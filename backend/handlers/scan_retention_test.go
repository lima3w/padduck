package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// All three handlers check requirePerm before touching the request body,
// so the invalid-body branch on UpdateScanRetentionSettings is only
// reachable by a permitted user — which requires a live repo (see plan).
// Only the auth guard branches are testable here without a DB.

func TestGetScanRetentionSettings_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/scan-retention", h.GetScanRetentionSettings)

	req := httptest.NewRequest("GET", "/admin/scan-retention", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetScanRetentionSettings_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/scan-retention", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.GetScanRetentionSettings(c)
	})

	req := httptest.NewRequest("GET", "/admin/scan-retention", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestUpdateScanRetentionSettings_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Put("/admin/scan-retention", h.UpdateScanRetentionSettings)

	req := httptest.NewRequest("PUT", "/admin/scan-retention", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestUpdateScanRetentionSettings_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Put("/admin/scan-retention", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.UpdateScanRetentionSettings(c)
	})

	req := httptest.NewRequest("PUT", "/admin/scan-retention", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestRunScanRetentionPrune_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/admin/scan-retention/prune", h.RunScanRetentionPrune)

	req := httptest.NewRequest("POST", "/admin/scan-retention/prune", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestRunScanRetentionPrune_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/admin/scan-retention/prune", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.RunScanRetentionPrune(c)
	})

	req := httptest.NewRequest("POST", "/admin/scan-retention/prune", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}
