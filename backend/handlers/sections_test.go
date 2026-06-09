package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"padduck/models"
)

var unprivSection = &models.User{ID: 0, Role: "viewer"}

// ---------------------------------------------------------------------------
// CreateNetwork — POST /networks
// ---------------------------------------------------------------------------

func TestCreateSection_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/networks", h.CreateNetwork)

	req := httptest.NewRequest("POST", "/networks", strings.NewReader(`{"name":"test"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestCreateSection_NoPermission_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/networks", func(c *fiber.Ctx) error {
		c.Locals("user", unprivSection)
		return h.CreateNetwork(c)
	})

	req := httptest.NewRequest("POST", "/networks", strings.NewReader(`{"name":"test"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// ListNetworks — GET /networks
// ---------------------------------------------------------------------------

func TestListSections_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/networks", h.ListNetworks)

	resp, err := app.Test(httptest.NewRequest("GET", "/networks", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestListSections_NoPermission_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/networks", func(c *fiber.Ctx) error {
		c.Locals("user", unprivSection)
		return h.ListNetworks(c)
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/networks", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// GetNetwork — GET /networks/:id
// ---------------------------------------------------------------------------

func TestGetSection_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/networks/:id", h.GetNetwork)

	resp, err := app.Test(httptest.NewRequest("GET", "/networks/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetSection_NoPermission_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/networks/:id", func(c *fiber.Ctx) error {
		c.Locals("user", unprivSection)
		return h.GetNetwork(c)
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/networks/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestGetSection_BadID_NoAuth_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/networks/:id", h.GetNetwork)

	// permCheck runs before ParamsInt, so unauthenticated requests get 401.
	resp, err := app.Test(httptest.NewRequest("GET", "/networks/abc", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// UpdateNetwork — PUT /networks/:id
// ---------------------------------------------------------------------------

func TestUpdateSection_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Put("/networks/:id", h.UpdateNetwork)

	resp, err := app.Test(httptest.NewRequest("PUT", "/networks/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestUpdateSection_NoPermission_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Put("/networks/:id", func(c *fiber.Ctx) error {
		c.Locals("user", unprivSection)
		return h.UpdateNetwork(c)
	})

	resp, err := app.Test(httptest.NewRequest("PUT", "/networks/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestUpdateSection_BadID_Returns400(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Put("/networks/:id", h.UpdateNetwork)

	resp, err := app.Test(httptest.NewRequest("PUT", "/networks/abc", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// DeleteNetwork — DELETE /networks/:id
// ---------------------------------------------------------------------------

func TestDeleteSection_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Delete("/networks/:id", h.DeleteNetwork)

	resp, err := app.Test(httptest.NewRequest("DELETE", "/networks/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestDeleteSection_NoPermission_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Delete("/networks/:id", func(c *fiber.Ctx) error {
		c.Locals("user", unprivSection)
		return h.DeleteNetwork(c)
	})

	resp, err := app.Test(httptest.NewRequest("DELETE", "/networks/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestDeleteSection_BadID_Returns400(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Delete("/networks/:id", h.DeleteNetwork)

	resp, err := app.Test(httptest.NewRequest("DELETE", "/networks/abc", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// CreateNetwork body validation — body parsing precedes permCheck
// ---------------------------------------------------------------------------

func TestCreateSection_BadBody_Returns400(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/networks", h.CreateNetwork)

	req := httptest.NewRequest("POST", "/networks", strings.NewReader(`not valid json`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
