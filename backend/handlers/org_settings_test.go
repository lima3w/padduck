package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// GetOrgSettings / UpdateOrgSettings — requirePerm-gated; the org-context
// and body-validation branches are only reachable by a permitted user,
// which requires a live repo (see plan). Only the auth guard branches are
// testable here without a DB.
// ---------------------------------------------------------------------------

func TestGetOrgSettings_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/organization/settings", h.GetOrgSettings)

	req := httptest.NewRequest("GET", "/admin/organization/settings", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetOrgSettings_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/organization/settings", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.GetOrgSettings(c)
	})

	req := httptest.NewRequest("GET", "/admin/organization/settings", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestUpdateOrgSettings_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Put("/admin/organization/settings", h.UpdateOrgSettings)

	req := httptest.NewRequest("PUT", "/admin/organization/settings", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestUpdateOrgSettings_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Put("/admin/organization/settings", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.UpdateOrgSettings(c)
	})

	req := httptest.NewRequest("PUT", "/admin/organization/settings", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// GetPlatformOrgSettings — no inline guard; :id is parsed unconditionally,
// so the invalid-ID branch is reachable with no auth at all.
// ---------------------------------------------------------------------------

func TestGetPlatformOrgSettings_InvalidID_Returns400(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/platform/organizations/:id/settings", h.GetPlatformOrgSettings)

	req := httptest.NewRequest("GET", "/platform/organizations/not-a-number/settings", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// UpdatePlatformOrgSettings — no inline guard; :id is parsed before the
// body, so both the invalid-ID and (valid-ID) invalid-body branches are
// reachable with no auth at all.
// ---------------------------------------------------------------------------

func TestUpdatePlatformOrgSettings_InvalidID_Returns400(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Put("/platform/organizations/:id/settings", h.UpdatePlatformOrgSettings)

	req := httptest.NewRequest("PUT", "/platform/organizations/not-a-number/settings", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestUpdatePlatformOrgSettings_ValidIDInvalidBody_Returns400(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Put("/platform/organizations/:id/settings", h.UpdatePlatformOrgSettings)

	req := httptest.NewRequest("PUT", "/platform/organizations/1/settings", strings.NewReader("{not valid json"))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
