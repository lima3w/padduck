package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// All three handlers check requirePerm before touching params/body, so the
// invalid-ID/invalid-body branches are only reachable by a permitted user —
// which requires a live repo (see plan). Only the auth guard branches are
// testable here without a DB.

func TestListDiscoveryConflicts_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/discovery/conflicts", h.ListDiscoveryConflicts)

	req := httptest.NewRequest("GET", "/admin/discovery/conflicts", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestListDiscoveryConflicts_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/discovery/conflicts", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.ListDiscoveryConflicts(c)
	})

	req := httptest.NewRequest("GET", "/admin/discovery/conflicts", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestGetDiscoveryConflict_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/discovery/conflicts/:id", h.GetDiscoveryConflict)

	req := httptest.NewRequest("GET", "/admin/discovery/conflicts/1", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetDiscoveryConflict_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/discovery/conflicts/:id", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.GetDiscoveryConflict(c)
	})

	req := httptest.NewRequest("GET", "/admin/discovery/conflicts/1", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestResolveDiscoveryConflict_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/admin/discovery/conflicts/:id/resolve", h.ResolveDiscoveryConflict)

	req := httptest.NewRequest("POST", "/admin/discovery/conflicts/1/resolve", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestResolveDiscoveryConflict_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/admin/discovery/conflicts/:id/resolve", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.ResolveDiscoveryConflict(c)
	})

	req := httptest.NewRequest("POST", "/admin/discovery/conflicts/1/resolve", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}
