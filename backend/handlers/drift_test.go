package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// All five handlers check requirePerm before touching params/body, so the
// deeper validation/business-logic branches are only reachable by a
// permitted user — which requires a live repo (see plan). Only the auth
// guard branches are testable here without a DB.

func TestListDriftItems_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/drift", h.ListDriftItems)

	req := httptest.NewRequest("GET", "/admin/drift", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestListDriftItems_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/drift", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.ListDriftItems(c)
	})

	req := httptest.NewRequest("GET", "/admin/drift", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestGetDriftItem_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/drift/:id", h.GetDriftItem)

	req := httptest.NewRequest("GET", "/admin/drift/1", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetDriftItem_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/drift/:id", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.GetDriftItem(c)
	})

	req := httptest.NewRequest("GET", "/admin/drift/1", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestAcceptDrift_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/admin/drift/:id/accept", h.AcceptDrift)

	req := httptest.NewRequest("POST", "/admin/drift/1/accept", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestAcceptDrift_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/admin/drift/:id/accept", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.AcceptDrift(c)
	})

	req := httptest.NewRequest("POST", "/admin/drift/1/accept", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestDismissDrift_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/admin/drift/:id/dismiss", h.DismissDrift)

	req := httptest.NewRequest("POST", "/admin/drift/1/dismiss", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestDismissDrift_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/admin/drift/:id/dismiss", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.DismissDrift(c)
	})

	req := httptest.NewRequest("POST", "/admin/drift/1/dismiss", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestEscalateDrift_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/admin/drift/:id/escalate", h.EscalateDrift)

	req := httptest.NewRequest("POST", "/admin/drift/1/escalate", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestEscalateDrift_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/admin/drift/:id/escalate", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.EscalateDrift(c)
	})

	req := httptest.NewRequest("POST", "/admin/drift/1/escalate", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}
