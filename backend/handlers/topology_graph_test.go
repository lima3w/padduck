package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// Both handlers check requirePerm before touching query params, so the
// required-param and invalid-ID branches are only reachable by a permitted
// user — which requires a live repo (see plan). Only the auth guard
// branches are testable here without a DB.

func TestGetTopologyGraph_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/topology/graph", h.GetTopologyGraph)

	req := httptest.NewRequest("GET", "/topology/graph", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetTopologyGraph_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/topology/graph", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.GetTopologyGraph(c)
	})

	req := httptest.NewRequest("GET", "/topology/graph", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestGetTopologyPath_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/topology/path", h.GetTopologyPath)

	req := httptest.NewRequest("GET", "/topology/path", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetTopologyPath_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/topology/path", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.GetTopologyPath(c)
	})

	req := httptest.NewRequest("GET", "/topology/path", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}
