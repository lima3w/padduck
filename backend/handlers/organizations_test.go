package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// All three handlers check requirePerm before touching params/body, so the
// invalid-body/invalid-ID branches are only reachable by a permitted user —
// which requires a live repo (see plan). Only the auth guard branches are
// testable here without a DB.

func TestListOrganizations_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/organizations", h.ListOrganizations)

	req := httptest.NewRequest("GET", "/admin/organizations", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestListOrganizations_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/organizations", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.ListOrganizations(c)
	})

	req := httptest.NewRequest("GET", "/admin/organizations", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestCreateOrganization_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/admin/organizations", h.CreateOrganization)

	req := httptest.NewRequest("POST", "/admin/organizations", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestCreateOrganization_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/admin/organizations", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.CreateOrganization(c)
	})

	req := httptest.NewRequest("POST", "/admin/organizations", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestDeleteOrganization_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Delete("/admin/organizations/:id", h.DeleteOrganization)

	req := httptest.NewRequest("DELETE", "/admin/organizations/1", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestDeleteOrganization_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Delete("/admin/organizations/:id", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.DeleteOrganization(c)
	})

	req := httptest.NewRequest("DELETE", "/admin/organizations/1", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}
