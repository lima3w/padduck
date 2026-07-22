package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// All four handlers check requirePerm before touching any service, with no
// params or body to validate, so only the auth guard branches are testable
// here without a DB. Success paths are covered by integration tests.

func TestExportNetworksCSV_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/reports/export/networks", h.ExportNetworksCSV)

	req := httptest.NewRequest("GET", "/admin/reports/export/networks", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestExportNetworksCSV_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/reports/export/networks", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.ExportNetworksCSV(c)
	})

	req := httptest.NewRequest("GET", "/admin/reports/export/networks", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestExportDevicesCSV_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/reports/export/devices", h.ExportDevicesCSV)

	req := httptest.NewRequest("GET", "/admin/reports/export/devices", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestExportDevicesCSV_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/reports/export/devices", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.ExportDevicesCSV(c)
	})

	req := httptest.NewRequest("GET", "/admin/reports/export/devices", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestExportVLANsCSV_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/reports/export/vlans", h.ExportVLANsCSV)

	req := httptest.NewRequest("GET", "/admin/reports/export/vlans", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestExportVLANsCSV_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/reports/export/vlans", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.ExportVLANsCSV(c)
	})

	req := httptest.NewRequest("GET", "/admin/reports/export/vlans", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestExportVRFsCSV_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/reports/export/vrfs", h.ExportVRFsCSV)

	req := httptest.NewRequest("GET", "/admin/reports/export/vrfs", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestExportVRFsCSV_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/reports/export/vrfs", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.ExportVRFsCSV(c)
	})

	req := httptest.NewRequest("GET", "/admin/reports/export/vrfs", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}
