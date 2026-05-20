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
// CreateSection — POST /sections
// ---------------------------------------------------------------------------

func TestCreateSection_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/sections", h.CreateSection)

	req := httptest.NewRequest("POST", "/sections", strings.NewReader(`{"name":"test"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestCreateSection_NoPermission_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/sections", func(c *fiber.Ctx) error {
		c.Locals("user", unprivSection)
		return h.CreateSection(c)
	})

	req := httptest.NewRequest("POST", "/sections", strings.NewReader(`{"name":"test"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// ListSections — GET /sections
// ---------------------------------------------------------------------------

func TestListSections_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/sections", h.ListSections)

	resp, err := app.Test(httptest.NewRequest("GET", "/sections", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestListSections_NoPermission_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/sections", func(c *fiber.Ctx) error {
		c.Locals("user", unprivSection)
		return h.ListSections(c)
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/sections", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// GetSection — GET /sections/:id
// ---------------------------------------------------------------------------

func TestGetSection_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/sections/:id", h.GetSection)

	resp, err := app.Test(httptest.NewRequest("GET", "/sections/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetSection_NoPermission_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/sections/:id", func(c *fiber.Ctx) error {
		c.Locals("user", unprivSection)
		return h.GetSection(c)
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/sections/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestGetSection_BadID_NoAuth_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/sections/:id", h.GetSection)

	// permCheck runs before ParamsInt, so unauthenticated requests get 401.
	resp, err := app.Test(httptest.NewRequest("GET", "/sections/abc", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// UpdateSection — PUT /sections/:id
// ---------------------------------------------------------------------------

func TestUpdateSection_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Put("/sections/:id", h.UpdateSection)

	resp, err := app.Test(httptest.NewRequest("PUT", "/sections/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestUpdateSection_NoPermission_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Put("/sections/:id", func(c *fiber.Ctx) error {
		c.Locals("user", unprivSection)
		return h.UpdateSection(c)
	})

	resp, err := app.Test(httptest.NewRequest("PUT", "/sections/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestUpdateSection_BadID_Returns400(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Put("/sections/:id", h.UpdateSection)

	resp, err := app.Test(httptest.NewRequest("PUT", "/sections/abc", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// DeleteSection — DELETE /sections/:id
// ---------------------------------------------------------------------------

func TestDeleteSection_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Delete("/sections/:id", h.DeleteSection)

	resp, err := app.Test(httptest.NewRequest("DELETE", "/sections/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestDeleteSection_NoPermission_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Delete("/sections/:id", func(c *fiber.Ctx) error {
		c.Locals("user", unprivSection)
		return h.DeleteSection(c)
	})

	resp, err := app.Test(httptest.NewRequest("DELETE", "/sections/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestDeleteSection_BadID_Returns400(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Delete("/sections/:id", h.DeleteSection)

	resp, err := app.Test(httptest.NewRequest("DELETE", "/sections/abc", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// CreateSection body validation — body parsing precedes permCheck
// ---------------------------------------------------------------------------

func TestCreateSection_BadBody_Returns400(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/sections", h.CreateSection)

	req := httptest.NewRequest("POST", "/sections", strings.NewReader(`not valid json`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
