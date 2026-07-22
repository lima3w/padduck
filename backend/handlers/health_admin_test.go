package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// Both handlers check requirePerm before touching the database pool, so
// only the auth guard branches are testable here without a DB. Success
// paths are covered by integration tests.

func TestGetSystemHealth_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/system-health", h.GetSystemHealth)

	req := httptest.NewRequest("GET", "/admin/system-health", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetSystemHealth_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/system-health", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.GetSystemHealth(c)
	})

	req := httptest.NewRequest("GET", "/admin/system-health", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestDownloadBackup_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/backup/download", h.DownloadBackup)

	req := httptest.NewRequest("GET", "/admin/backup/download", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestDownloadBackup_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/backup/download", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.DownloadBackup(c)
	})

	req := httptest.NewRequest("GET", "/admin/backup/download", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}
