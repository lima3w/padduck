package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// ListTags — GET /api/v1/tags  (permCheck: PermV2IPRead)
// ---------------------------------------------------------------------------

func TestListTags_NoUser_Returns401(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Get("/tags", h.ListTags)
	resp, err := app.Test(httptest.NewRequest("GET", "/tags", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestListTags_NoPermission_Returns403(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return c.Next()
	})
	app.Get("/tags", h.ListTags)
	resp, err := app.Test(httptest.NewRequest("GET", "/tags", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// CreateTag — POST /api/v1/tags  (requireAdmin: direct role check)
// ---------------------------------------------------------------------------

func TestCreateTag_NoUser_Returns403(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Post("/tags", h.CreateTag)
	req := httptest.NewRequest("POST", "/tags", strings.NewReader(`{"name":"critical","colour":"#f00"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestCreateTag_NonAdmin_Returns403(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user", nonAdminUser)
		return c.Next()
	})
	app.Post("/tags", h.CreateTag)
	req := httptest.NewRequest("POST", "/tags", strings.NewReader(`{"name":"critical","colour":"#f00"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestCreateTag_WriteScope_Returns403(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user", adminUser)
		c.Locals("tokenScope", "write")
		return c.Next()
	})
	app.Post("/tags", h.CreateTag)
	req := httptest.NewRequest("POST", "/tags", strings.NewReader(`{"name":"critical","colour":"#f00"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestCreateTag_MissingName_Returns400(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user", adminUser)
		return c.Next()
	})
	app.Post("/tags", h.CreateTag)
	req := httptest.NewRequest("POST", "/tags", strings.NewReader(`{"colour":"#f00"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// UpdateTag — PUT /api/v1/tags/:id  (requireAdmin after ID parse)
// ---------------------------------------------------------------------------

func TestUpdateTag_BadID_NoAuth_Returns403(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Put("/tags/:id", h.UpdateTag)
	// requireAdmin runs before ParamsInt, so unauthenticated requests get 403.
	req := httptest.NewRequest("PUT", "/tags/notanumber", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestUpdateTag_NonAdmin_Returns403(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user", nonAdminUser)
		return c.Next()
	})
	app.Put("/tags/:id", h.UpdateTag)
	req := httptest.NewRequest("PUT", "/tags/1", strings.NewReader(`{"name":"critical","colour":"#f00"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// DeleteTag — DELETE /api/v1/tags/:id  (requireAdmin after ID parse)
// ---------------------------------------------------------------------------

func TestDeleteTag_BadID_NoAuth_Returns403(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Delete("/tags/:id", h.DeleteTag)
	// requireAdmin runs before ParamsInt, so unauthenticated requests get 403.
	resp, err := app.Test(httptest.NewRequest("DELETE", "/tags/notanumber", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestDeleteTag_NonAdmin_Returns403(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user", nonAdminUser)
		return c.Next()
	})
	app.Delete("/tags/:id", h.DeleteTag)
	resp, err := app.Test(httptest.NewRequest("DELETE", "/tags/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}
