package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"padduck/models"
)

// unprivNS is a user with ID=0 — permCheck will deny without touching the nil repo.
var unprivNS = &models.User{ID: 0, Role: "viewer"}

// ---------------------------------------------------------------------------
// ListNameservers — GET /nameservers
// ---------------------------------------------------------------------------

func TestListNameservers_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/nameservers", h.ListNameservers)

	resp, err := app.Test(httptest.NewRequest("GET", "/nameservers", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestListNameservers_NoPermission_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/nameservers", func(c *fiber.Ctx) error {
		c.Locals("user", unprivNS)
		return h.ListNameservers(c)
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/nameservers", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// GetNameserver — GET /nameservers/:id
// ---------------------------------------------------------------------------

func TestGetNameserver_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/nameservers/:id", h.GetNameserver)

	resp, err := app.Test(httptest.NewRequest("GET", "/nameservers/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetNameserver_NoPermission_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/nameservers/:id", func(c *fiber.Ctx) error {
		c.Locals("user", unprivNS)
		return h.GetNameserver(c)
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/nameservers/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// CreateNameserver — POST /nameservers
// ---------------------------------------------------------------------------

func TestCreateNameserver_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/nameservers", h.CreateNameserver)

	resp, err := app.Test(httptest.NewRequest("POST", "/nameservers", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestCreateNameserver_NoPermission_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/nameservers", func(c *fiber.Ctx) error {
		c.Locals("user", unprivNS)
		return h.CreateNameserver(c)
	})

	resp, err := app.Test(httptest.NewRequest("POST", "/nameservers", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// UpdateNameserver — PUT /nameservers/:id
// ---------------------------------------------------------------------------

func TestUpdateNameserver_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Put("/nameservers/:id", h.UpdateNameserver)

	resp, err := app.Test(httptest.NewRequest("PUT", "/nameservers/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestUpdateNameserver_NoPermission_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Put("/nameservers/:id", func(c *fiber.Ctx) error {
		c.Locals("user", unprivNS)
		return h.UpdateNameserver(c)
	})

	resp, err := app.Test(httptest.NewRequest("PUT", "/nameservers/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// DeleteNameserver — DELETE /nameservers/:id
// ---------------------------------------------------------------------------

func TestDeleteNameserver_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Delete("/nameservers/:id", h.DeleteNameserver)

	resp, err := app.Test(httptest.NewRequest("DELETE", "/nameservers/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestDeleteNameserver_NoPermission_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Delete("/nameservers/:id", func(c *fiber.Ctx) error {
		c.Locals("user", unprivNS)
		return h.DeleteNameserver(c)
	})

	resp, err := app.Test(httptest.NewRequest("DELETE", "/nameservers/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}
