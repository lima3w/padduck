package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// All three handlers check requirePerm before touching the request body or
// repo, so the deeper validation/business-logic branches are only reachable
// by a permitted user — which requires a live repo (see plan). Only the
// auth guard branches are testable here without a DB.

func TestGetBreakGlassStatus_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/break-glass", h.GetBreakGlassStatus)

	req := httptest.NewRequest("GET", "/admin/break-glass", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetBreakGlassStatus_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/break-glass", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.GetBreakGlassStatus(c)
	})

	req := httptest.NewRequest("GET", "/admin/break-glass", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestActivateBreakGlass_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/admin/break-glass/activate", h.ActivateBreakGlass)

	req := httptest.NewRequest("POST", "/admin/break-glass/activate", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestActivateBreakGlass_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/admin/break-glass/activate", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.ActivateBreakGlass(c)
	})

	req := httptest.NewRequest("POST", "/admin/break-glass/activate", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestEndBreakGlass_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/admin/break-glass/end", h.EndBreakGlass)

	req := httptest.NewRequest("POST", "/admin/break-glass/end", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestEndBreakGlass_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/admin/break-glass/end", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.EndBreakGlass(c)
	})

	req := httptest.NewRequest("POST", "/admin/break-glass/end", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}
