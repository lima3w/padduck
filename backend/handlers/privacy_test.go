package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// All three handlers check requirePerm before touching params/body, so the
// invalid-body/field-required branches on CreatePrivacyVersion are only
// reachable by a permitted user — which requires a live repo (see plan).
// Only the auth guard branches are testable here without a DB.

func TestListPrivacyVersions_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/privacy/versions", h.ListPrivacyVersions)

	req := httptest.NewRequest("GET", "/admin/privacy/versions", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestListPrivacyVersions_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/privacy/versions", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.ListPrivacyVersions(c)
	})

	req := httptest.NewRequest("GET", "/admin/privacy/versions", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestCreatePrivacyVersion_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/admin/privacy/versions", h.CreatePrivacyVersion)

	req := httptest.NewRequest("POST", "/admin/privacy/versions", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestCreatePrivacyVersion_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/admin/privacy/versions", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.CreatePrivacyVersion(c)
	})

	req := httptest.NewRequest("POST", "/admin/privacy/versions", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestGetConsentReport_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/privacy/consent-report", h.GetConsentReport)

	req := httptest.NewRequest("GET", "/admin/privacy/consent-report", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetConsentReport_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/privacy/consent-report", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.GetConsentReport(c)
	})

	req := httptest.NewRequest("GET", "/admin/privacy/consent-report", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}
