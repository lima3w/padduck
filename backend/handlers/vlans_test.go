package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"ipam-next/models"
)

// unprivVLAN is a user with ID=0 — permCheck denies without touching the nil repo.
var unprivVLAN = &models.User{ID: 0, Role: "viewer"}

// ---------------------------------------------------------------------------
// ListVLANDomains — GET /vlan-domains
// ---------------------------------------------------------------------------

func TestListVLANDomains_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/vlan-domains", h.ListVLANDomains)

	resp, err := app.Test(httptest.NewRequest("GET", "/vlan-domains", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestListVLANDomains_NoPermission_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/vlan-domains", func(c *fiber.Ctx) error {
		c.Locals("user", unprivVLAN)
		return h.ListVLANDomains(c)
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/vlan-domains", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// GetVLANDomain — GET /vlan-domains/:id
// ---------------------------------------------------------------------------

func TestGetVLANDomain_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/vlan-domains/:id", h.GetVLANDomain)

	resp, err := app.Test(httptest.NewRequest("GET", "/vlan-domains/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetVLANDomain_NoPermission_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/vlan-domains/:id", func(c *fiber.Ctx) error {
		c.Locals("user", unprivVLAN)
		return h.GetVLANDomain(c)
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/vlan-domains/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestGetVLANDomain_BadID_Returns400(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/vlan-domains/:id", h.GetVLANDomain)

	resp, err := app.Test(httptest.NewRequest("GET", "/vlan-domains/abc", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// CreateVLANDomain — POST /vlan-domains
// ---------------------------------------------------------------------------

func TestCreateVLANDomain_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/vlan-domains", h.CreateVLANDomain)

	resp, err := app.Test(httptest.NewRequest("POST", "/vlan-domains", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestCreateVLANDomain_NoPermission_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/vlan-domains", func(c *fiber.Ctx) error {
		c.Locals("user", unprivVLAN)
		return h.CreateVLANDomain(c)
	})

	resp, err := app.Test(httptest.NewRequest("POST", "/vlan-domains", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// UpdateVLANDomain — PUT /vlan-domains/:id
// ---------------------------------------------------------------------------

func TestUpdateVLANDomain_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Put("/vlan-domains/:id", h.UpdateVLANDomain)

	resp, err := app.Test(httptest.NewRequest("PUT", "/vlan-domains/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestUpdateVLANDomain_NoPermission_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Put("/vlan-domains/:id", func(c *fiber.Ctx) error {
		c.Locals("user", unprivVLAN)
		return h.UpdateVLANDomain(c)
	})

	resp, err := app.Test(httptest.NewRequest("PUT", "/vlan-domains/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestUpdateVLANDomain_BadID_Returns400(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Put("/vlan-domains/:id", h.UpdateVLANDomain)

	resp, err := app.Test(httptest.NewRequest("PUT", "/vlan-domains/abc", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// DeleteVLANDomain — DELETE /vlan-domains/:id
// ---------------------------------------------------------------------------

func TestDeleteVLANDomain_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Delete("/vlan-domains/:id", h.DeleteVLANDomain)

	resp, err := app.Test(httptest.NewRequest("DELETE", "/vlan-domains/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestDeleteVLANDomain_NoPermission_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Delete("/vlan-domains/:id", func(c *fiber.Ctx) error {
		c.Locals("user", unprivVLAN)
		return h.DeleteVLANDomain(c)
	})

	resp, err := app.Test(httptest.NewRequest("DELETE", "/vlan-domains/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestDeleteVLANDomain_BadID_Returns400(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Delete("/vlan-domains/:id", h.DeleteVLANDomain)

	resp, err := app.Test(httptest.NewRequest("DELETE", "/vlan-domains/abc", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
