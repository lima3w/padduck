package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// custom_fields.go has no auth guard — these handlers are protected by
// router-level middleware in production. Tests cover only input validation.

// ---------------------------------------------------------------------------
// GetCustomFieldDefinition — GET /api/v1/admin/custom-fields/:id
// ---------------------------------------------------------------------------

func TestGetCustomFieldDefinition_BadID_Returns400(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Get("/admin/custom-fields/:id", h.GetCustomFieldDefinition)
	resp, err := app.Test(httptest.NewRequest("GET", "/admin/custom-fields/notanumber", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// UpdateCustomFieldDefinition — PUT /api/v1/admin/custom-fields/:id
// ---------------------------------------------------------------------------

func TestUpdateCustomFieldDefinition_BadID_Returns400(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Put("/admin/custom-fields/:id", h.UpdateCustomFieldDefinition)
	req := httptest.NewRequest("PUT", "/admin/custom-fields/notanumber", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// DeleteCustomFieldDefinition — DELETE /api/v1/admin/custom-fields/:id
// ---------------------------------------------------------------------------

func TestDeleteCustomFieldDefinition_BadID_Returns400(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Delete("/admin/custom-fields/:id", h.DeleteCustomFieldDefinition)
	resp, err := app.Test(httptest.NewRequest("DELETE", "/admin/custom-fields/notanumber", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// ReorderCustomFieldDefinitions — PUT /api/v1/admin/custom-fields/reorder
// ---------------------------------------------------------------------------

func TestReorderCustomFieldDefinitions_InvalidBody_Returns400(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Put("/admin/custom-fields/reorder", h.ReorderCustomFieldDefinitions)
	req := httptest.NewRequest("PUT", "/admin/custom-fields/reorder", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
