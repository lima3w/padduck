package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// Both handlers check requirePerm before parsing the :id param or request
// body, so the invalid-ID/invalid-body branches are only reachable by a
// permitted user — which requires a live repo (see plan). Only the auth
// guard branches are testable here without a DB.

func TestGetDeviceFingerprint_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/devices/:id/fingerprint", h.GetDeviceFingerprint)

	req := httptest.NewRequest("GET", "/admin/devices/1/fingerprint", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetDeviceFingerprint_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/devices/:id/fingerprint", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.GetDeviceFingerprint(c)
	})

	req := httptest.NewRequest("GET", "/admin/devices/1/fingerprint", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestBuildDeviceFingerprint_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/admin/devices/:id/fingerprint", h.BuildDeviceFingerprint)

	req := httptest.NewRequest("POST", "/admin/devices/1/fingerprint", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestBuildDeviceFingerprint_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/admin/devices/:id/fingerprint", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.BuildDeviceFingerprint(c)
	})

	req := httptest.NewRequest("POST", "/admin/devices/1/fingerprint", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}
