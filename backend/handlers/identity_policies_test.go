package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// All three handlers check requirePerm before touching params/body, so the
// invalid-body/range-validation branches on UpdateIdentityPolicies are only
// reachable by a permitted user — which requires a live repo (see plan).
// Only the auth guard branches are testable here without a DB.

func TestGetIdentityPolicies_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/identity-policies", h.GetIdentityPolicies)

	req := httptest.NewRequest("GET", "/admin/identity-policies", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetIdentityPolicies_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/identity-policies", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.GetIdentityPolicies(c)
	})

	req := httptest.NewRequest("GET", "/admin/identity-policies", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestUpdateIdentityPolicies_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Put("/admin/identity-policies", h.UpdateIdentityPolicies)

	req := httptest.NewRequest("PUT", "/admin/identity-policies", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestUpdateIdentityPolicies_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Put("/admin/identity-policies", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.UpdateIdentityPolicies(c)
	})

	req := httptest.NewRequest("PUT", "/admin/identity-policies", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestListSessionRisk_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/session-risk", h.ListSessionRisk)

	req := httptest.NewRequest("GET", "/admin/session-risk", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestListSessionRisk_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/session-risk", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.ListSessionRisk(c)
	})

	req := httptest.NewRequest("GET", "/admin/session-risk", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}
