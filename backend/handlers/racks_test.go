package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// ListRacks — GET /racks
// ---------------------------------------------------------------------------

func TestListRacks_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/racks", h.ListRacks)

	resp, err := app.Test(httptest.NewRequest("GET", "/racks", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestListRacks_NoPermission_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/racks", func(c *fiber.Ctx) error {
		c.Locals("user", unprivUser)
		return h.ListRacks(c)
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/racks", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// GetRack — GET /racks/:id
// ---------------------------------------------------------------------------

func TestGetRack_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/racks/:id", h.GetRack)

	resp, err := app.Test(httptest.NewRequest("GET", "/racks/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetRack_NoPermission_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/racks/:id", func(c *fiber.Ctx) error {
		c.Locals("user", unprivUser)
		return h.GetRack(c)
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/racks/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// CreateRack — POST /racks
// ---------------------------------------------------------------------------

func TestCreateRack_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/racks", h.CreateRack)

	resp, err := app.Test(httptest.NewRequest("POST", "/racks", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestCreateRack_NoPermission_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/racks", func(c *fiber.Ctx) error {
		c.Locals("user", unprivUser)
		return h.CreateRack(c)
	})

	resp, err := app.Test(httptest.NewRequest("POST", "/racks", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestCreateRack_EmptyName_Returns400(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/racks", func(c *fiber.Ctx) error {
		c.Locals("user", unprivUser)
		return h.CreateRack(c)
	})

	body := strings.NewReader(`{"name":"","location_id":1}`)
	req := httptest.NewRequest("POST", "/racks", body)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	// permCheck fires first with ID=0 → 403 before body is parsed
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// UpdateRack — PUT /racks/:id
// ---------------------------------------------------------------------------

func TestUpdateRack_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Put("/racks/:id", h.UpdateRack)

	resp, err := app.Test(httptest.NewRequest("PUT", "/racks/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestUpdateRack_NoPermission_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Put("/racks/:id", func(c *fiber.Ctx) error {
		c.Locals("user", unprivUser)
		return h.UpdateRack(c)
	})

	resp, err := app.Test(httptest.NewRequest("PUT", "/racks/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// DeleteRack — DELETE /racks/:id
// ---------------------------------------------------------------------------

func TestDeleteRack_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Delete("/racks/:id", h.DeleteRack)

	resp, err := app.Test(httptest.NewRequest("DELETE", "/racks/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestDeleteRack_NoPermission_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Delete("/racks/:id", func(c *fiber.Ctx) error {
		c.Locals("user", unprivUser)
		return h.DeleteRack(c)
	})

	resp, err := app.Test(httptest.NewRequest("DELETE", "/racks/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// ListDevicesInRack — GET /racks/:id/devices
// ---------------------------------------------------------------------------

func TestListDevicesInRack_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/racks/:id/devices", h.ListDevicesInRack)

	resp, err := app.Test(httptest.NewRequest("GET", "/racks/1/devices", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestListDevicesInRack_NoPermission_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/racks/:id/devices", func(c *fiber.Ctx) error {
		c.Locals("user", unprivUser)
		return h.ListDevicesInRack(c)
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/racks/1/devices", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}
