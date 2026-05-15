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
// GetVLANSubnets — GET /vlans/:id/subnets
// ---------------------------------------------------------------------------

func TestGetVLANSubnets_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/vlans/:id/subnets", h.GetVLANSubnets)

	resp, err := app.Test(httptest.NewRequest("GET", "/vlans/1/subnets", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetVLANSubnets_NoPermission_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/vlans/:id/subnets", func(c *fiber.Ctx) error {
		c.Locals("user", unprivVLAN)
		return h.GetVLANSubnets(c)
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/vlans/1/subnets", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestGetVLANSubnets_BadID_Returns400(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/vlans/:id/subnets", h.GetVLANSubnets)

	resp, err := app.Test(httptest.NewRequest("GET", "/vlans/abc/subnets", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

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

// ---------------------------------------------------------------------------
// ListVLANGroups — GET /vlan-groups
// ---------------------------------------------------------------------------

func TestListVLANGroups_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/vlan-groups", h.ListVLANGroups)

	resp, err := app.Test(httptest.NewRequest("GET", "/vlan-groups", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestListVLANGroups_NoPermission_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/vlan-groups", func(c *fiber.Ctx) error {
		c.Locals("user", unprivVLAN)
		return h.ListVLANGroups(c)
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/vlan-groups", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// GetVLANGroup — GET /vlan-groups/:id
// ---------------------------------------------------------------------------

func TestGetVLANGroup_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/vlan-groups/:id", h.GetVLANGroup)

	resp, err := app.Test(httptest.NewRequest("GET", "/vlan-groups/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetVLANGroup_NoPermission_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/vlan-groups/:id", func(c *fiber.Ctx) error {
		c.Locals("user", unprivVLAN)
		return h.GetVLANGroup(c)
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/vlan-groups/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestGetVLANGroup_BadID_Returns400(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/vlan-groups/:id", h.GetVLANGroup)

	resp, err := app.Test(httptest.NewRequest("GET", "/vlan-groups/abc", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// CreateVLANGroup — POST /vlan-groups
// ---------------------------------------------------------------------------

func TestCreateVLANGroup_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/vlan-groups", h.CreateVLANGroup)

	resp, err := app.Test(httptest.NewRequest("POST", "/vlan-groups", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestCreateVLANGroup_NoPermission_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/vlan-groups", func(c *fiber.Ctx) error {
		c.Locals("user", unprivVLAN)
		return h.CreateVLANGroup(c)
	})

	resp, err := app.Test(httptest.NewRequest("POST", "/vlan-groups", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// UpdateVLANGroup — PUT /vlan-groups/:id
// ---------------------------------------------------------------------------

func TestUpdateVLANGroup_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Put("/vlan-groups/:id", h.UpdateVLANGroup)

	resp, err := app.Test(httptest.NewRequest("PUT", "/vlan-groups/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestUpdateVLANGroup_NoPermission_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Put("/vlan-groups/:id", func(c *fiber.Ctx) error {
		c.Locals("user", unprivVLAN)
		return h.UpdateVLANGroup(c)
	})

	resp, err := app.Test(httptest.NewRequest("PUT", "/vlan-groups/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestUpdateVLANGroup_BadID_Returns400(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Put("/vlan-groups/:id", h.UpdateVLANGroup)

	resp, err := app.Test(httptest.NewRequest("PUT", "/vlan-groups/abc", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// DeleteVLANGroup — DELETE /vlan-groups/:id
// ---------------------------------------------------------------------------

func TestDeleteVLANGroup_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Delete("/vlan-groups/:id", h.DeleteVLANGroup)

	resp, err := app.Test(httptest.NewRequest("DELETE", "/vlan-groups/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestDeleteVLANGroup_NoPermission_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Delete("/vlan-groups/:id", func(c *fiber.Ctx) error {
		c.Locals("user", unprivVLAN)
		return h.DeleteVLANGroup(c)
	})

	resp, err := app.Test(httptest.NewRequest("DELETE", "/vlan-groups/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestDeleteVLANGroup_BadID_Returns400(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Delete("/vlan-groups/:id", h.DeleteVLANGroup)

	resp, err := app.Test(httptest.NewRequest("DELETE", "/vlan-groups/abc", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
