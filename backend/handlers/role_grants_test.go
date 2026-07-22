package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// All three handlers check requirePerm before touching params/body, so the
// invalid-ID/invalid-body/required-field branches are only reachable by a
// permitted user — which requires a live repo (see plan). Only the auth
// guard branches are testable here without a DB.

func TestListUserGrants_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/users/:id/grants", h.ListUserGrants)

	req := httptest.NewRequest("GET", "/admin/users/1/grants", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestListUserGrants_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/users/:id/grants", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.ListUserGrants(c)
	})

	req := httptest.NewRequest("GET", "/admin/users/1/grants", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestCreateGrant_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/admin/role-grants", h.CreateGrant)

	req := httptest.NewRequest("POST", "/admin/role-grants", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestCreateGrant_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/admin/role-grants", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.CreateGrant(c)
	})

	req := httptest.NewRequest("POST", "/admin/role-grants", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestRevokeGrant_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Delete("/admin/role-grants/:id", h.RevokeGrant)

	req := httptest.NewRequest("DELETE", "/admin/role-grants/1", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestRevokeGrant_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Delete("/admin/role-grants/:id", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.RevokeGrant(c)
	})

	req := httptest.NewRequest("DELETE", "/admin/role-grants/1", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}
