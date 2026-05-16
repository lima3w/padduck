package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"ipam-next/models"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func customFieldApp(h *Handler) *fiber.App {
	app := fiber.New()
	app.Get("/admin/custom-fields", h.ListCustomFieldDefinitions)
	app.Post("/admin/custom-fields", h.CreateCustomFieldDefinition)
	app.Put("/admin/custom-fields/reorder", h.ReorderCustomFieldDefinitions)
	app.Get("/admin/custom-fields/:id", h.GetCustomFieldDefinition)
	app.Put("/admin/custom-fields/:id", h.UpdateCustomFieldDefinition)
	app.Delete("/admin/custom-fields/:id", h.DeleteCustomFieldDefinition)
	return app
}

func customFieldAppAs(h *Handler, user *models.User) *fiber.App {
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user", user)
		return c.Next()
	})
	app.Get("/admin/custom-fields", h.ListCustomFieldDefinitions)
	app.Post("/admin/custom-fields", h.CreateCustomFieldDefinition)
	app.Put("/admin/custom-fields/reorder", h.ReorderCustomFieldDefinitions)
	app.Get("/admin/custom-fields/:id", h.GetCustomFieldDefinition)
	app.Put("/admin/custom-fields/:id", h.UpdateCustomFieldDefinition)
	app.Delete("/admin/custom-fields/:id", h.DeleteCustomFieldDefinition)
	return app
}

// ---------------------------------------------------------------------------
// ListCustomFieldDefinitions — GET /admin/custom-fields
// ---------------------------------------------------------------------------

func TestListCustomFieldDefinitions_NoUser_Returns401(t *testing.T) {
	app := customFieldApp(&Handler{})
	resp, err := app.Test(httptest.NewRequest("GET", "/admin/custom-fields", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestListCustomFieldDefinitions_NoPermission_Returns403(t *testing.T) {
	app := customFieldAppAs(&Handler{}, &models.User{ID: 0, Role: "user"})
	resp, err := app.Test(httptest.NewRequest("GET", "/admin/custom-fields", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// CreateCustomFieldDefinition — POST /admin/custom-fields
// ---------------------------------------------------------------------------

func TestCreateCustomFieldDefinition_NoUser_Returns401(t *testing.T) {
	app := customFieldApp(&Handler{})
	req := httptest.NewRequest("POST", "/admin/custom-fields", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestCreateCustomFieldDefinition_NoPermission_Returns403(t *testing.T) {
	app := customFieldAppAs(&Handler{}, &models.User{ID: 0, Role: "user"})
	req := httptest.NewRequest("POST", "/admin/custom-fields", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// GetCustomFieldDefinition — GET /admin/custom-fields/:id
// ---------------------------------------------------------------------------

func TestGetCustomFieldDefinition_BadID_Returns400(t *testing.T) {
	app := customFieldApp(&Handler{})
	resp, err := app.Test(httptest.NewRequest("GET", "/admin/custom-fields/notanumber", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestGetCustomFieldDefinition_NoUser_Returns401(t *testing.T) {
	app := customFieldApp(&Handler{})
	resp, err := app.Test(httptest.NewRequest("GET", "/admin/custom-fields/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetCustomFieldDefinition_NoPermission_Returns403(t *testing.T) {
	app := customFieldAppAs(&Handler{}, &models.User{ID: 0, Role: "user"})
	resp, err := app.Test(httptest.NewRequest("GET", "/admin/custom-fields/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// UpdateCustomFieldDefinition — PUT /admin/custom-fields/:id
// ---------------------------------------------------------------------------

func TestUpdateCustomFieldDefinition_BadID_Returns400(t *testing.T) {
	app := customFieldApp(&Handler{})
	req := httptest.NewRequest("PUT", "/admin/custom-fields/notanumber", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestUpdateCustomFieldDefinition_NoUser_Returns401(t *testing.T) {
	app := customFieldApp(&Handler{})
	req := httptest.NewRequest("PUT", "/admin/custom-fields/1", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestUpdateCustomFieldDefinition_NoPermission_Returns403(t *testing.T) {
	app := customFieldAppAs(&Handler{}, &models.User{ID: 0, Role: "user"})
	req := httptest.NewRequest("PUT", "/admin/custom-fields/1", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// DeleteCustomFieldDefinition — DELETE /admin/custom-fields/:id
// ---------------------------------------------------------------------------

func TestDeleteCustomFieldDefinition_BadID_Returns400(t *testing.T) {
	app := customFieldApp(&Handler{})
	resp, err := app.Test(httptest.NewRequest("DELETE", "/admin/custom-fields/notanumber", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestDeleteCustomFieldDefinition_NoUser_Returns401(t *testing.T) {
	app := customFieldApp(&Handler{})
	resp, err := app.Test(httptest.NewRequest("DELETE", "/admin/custom-fields/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestDeleteCustomFieldDefinition_NoPermission_Returns403(t *testing.T) {
	app := customFieldAppAs(&Handler{}, &models.User{ID: 0, Role: "user"})
	resp, err := app.Test(httptest.NewRequest("DELETE", "/admin/custom-fields/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// ReorderCustomFieldDefinitions — PUT /admin/custom-fields/reorder
// ---------------------------------------------------------------------------

func TestReorderCustomFieldDefinitions_NoUser_Returns401(t *testing.T) {
	app := customFieldApp(&Handler{})
	req := httptest.NewRequest("PUT", "/admin/custom-fields/reorder", strings.NewReader(`{"ids":[1,2]}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestReorderCustomFieldDefinitions_NoPermission_Returns403(t *testing.T) {
	app := customFieldAppAs(&Handler{}, &models.User{ID: 0, Role: "user"})
	req := httptest.NewRequest("PUT", "/admin/custom-fields/reorder", strings.NewReader(`{"ids":[1,2]}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

