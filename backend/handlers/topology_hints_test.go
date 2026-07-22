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

func TestListTopologyHints_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/topology/hints", h.ListTopologyHints)

	req := httptest.NewRequest("GET", "/admin/topology/hints", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestListTopologyHints_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/topology/hints", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.ListTopologyHints(c)
	})

	req := httptest.NewRequest("GET", "/admin/topology/hints", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestGetTopologyHint_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/topology/hints/:id", h.GetTopologyHint)

	req := httptest.NewRequest("GET", "/admin/topology/hints/1", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetTopologyHint_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/topology/hints/:id", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.GetTopologyHint(c)
	})

	req := httptest.NewRequest("GET", "/admin/topology/hints/1", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestUpdateTopologyHintStatus_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Put("/admin/topology/hints/:id/status", h.UpdateTopologyHintStatus)

	req := httptest.NewRequest("PUT", "/admin/topology/hints/1/status", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestUpdateTopologyHintStatus_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Put("/admin/topology/hints/:id/status", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.UpdateTopologyHintStatus(c)
	})

	req := httptest.NewRequest("PUT", "/admin/topology/hints/1/status", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}
