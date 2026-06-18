package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// ListVRFs — GET /api/v1/vrfs  (permCheck first)
// ---------------------------------------------------------------------------

func TestListVRFs_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/vrfs", h.ListVRFs)
	resp, err := app.Test(httptest.NewRequest("GET", "/vrfs", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestListVRFs_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return c.Next()
	})
	app.Get("/vrfs", h.ListVRFs)
	resp, err := app.Test(httptest.NewRequest("GET", "/vrfs", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// GetVRF — GET /api/v1/vrfs/:id  (ID parsed before permCheck)
// ---------------------------------------------------------------------------

func TestGetVRF_BadID_NoAuth_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/vrfs/:id", h.GetVRF)
	// permCheck runs before ParamsInt, so unauthenticated requests get 401.
	resp, err := app.Test(httptest.NewRequest("GET", "/vrfs/notanumber", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetVRF_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/vrfs/:id", h.GetVRF)
	resp, err := app.Test(httptest.NewRequest("GET", "/vrfs/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetVRF_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return c.Next()
	})
	app.Get("/vrfs/:id", h.GetVRF)
	resp, err := app.Test(httptest.NewRequest("GET", "/vrfs/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// CreateVRF — POST /api/v1/vrfs  (body parsed before permCheck)
// ---------------------------------------------------------------------------

func TestCreateVRF_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/vrfs", h.CreateVRF)
	req := httptest.NewRequest("POST", "/vrfs", strings.NewReader(`{"name":"mgmt"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestCreateVRF_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return c.Next()
	})
	app.Post("/vrfs", h.CreateVRF)
	req := httptest.NewRequest("POST", "/vrfs", strings.NewReader(`{"name":"mgmt"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// UpdateVRF — PUT /api/v1/vrfs/:id  (ID parsed before permCheck)
// ---------------------------------------------------------------------------

func TestUpdateVRF_BadID_Returns400(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Put("/vrfs/:id", h.UpdateVRF)
	req := httptest.NewRequest("PUT", "/vrfs/notanumber", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestUpdateVRF_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Put("/vrfs/:id", h.UpdateVRF)
	req := httptest.NewRequest("PUT", "/vrfs/1", strings.NewReader(`{"name":"mgmt"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// DeleteVRF — DELETE /api/v1/vrfs/:id  (ID parsed before permCheck)
// ---------------------------------------------------------------------------

func TestDeleteVRF_BadID_Returns400(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Delete("/vrfs/:id", h.DeleteVRF)
	resp, err := app.Test(httptest.NewRequest("DELETE", "/vrfs/notanumber", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestDeleteVRF_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Delete("/vrfs/:id", h.DeleteVRF)
	resp, err := app.Test(httptest.NewRequest("DELETE", "/vrfs/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}
