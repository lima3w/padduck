package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// Both route handlers check requirePerm before touching the database pool
// or the uploaded file, so only the auth guard branches are testable here
// without a DB. generateSQLBackup is an internal helper (not a route) that
// touches h.service unconditionally and is covered by integration tests.

func TestDownloadFullBackup_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/backups/download", h.DownloadFullBackup)

	req := httptest.NewRequest("GET", "/admin/backups/download", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestDownloadFullBackup_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/backups/download", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.DownloadFullBackup(c)
	})

	req := httptest.NewRequest("GET", "/admin/backups/download", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestRestoreFromBackup_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/admin/backups/restore", h.RestoreFromBackup)

	req := httptest.NewRequest("POST", "/admin/backups/restore", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestRestoreFromBackup_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/admin/backups/restore", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.RestoreFromBackup(c)
	})

	req := httptest.NewRequest("POST", "/admin/backups/restore", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}
