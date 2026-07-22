package handlers

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// ListCustomerAssociations — auth enforcement (no local validation branch;
// the :id param is read tolerantly and never rejected)
// ---------------------------------------------------------------------------

func TestListCustomerAssociations_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/customers/associations", h.ListCustomerAssociations)

	req := httptest.NewRequest("GET", "/customers/associations", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestListCustomerAssociations_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/customers/associations", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.ListCustomerAssociations(c)
	})

	req := httptest.NewRequest("GET", "/customers/associations", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// CreateCustomerAssociation — body parsing happens before the permission
// check, so the invalid-body branch is reachable even without a user.
// ---------------------------------------------------------------------------

func TestCreateCustomerAssociation_InvalidBody_Returns400(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/customers/associations", h.CreateCustomerAssociation)

	req := httptest.NewRequest("POST", "/customers/associations", strings.NewReader("{not valid json"))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	data, _ := io.ReadAll(resp.Body)
	var body ErrorResponse
	assert.NoError(t, json.Unmarshal(data, &body))
	assert.Equal(t, string(ErrBadRequest), body.Code)
}

func TestCreateCustomerAssociation_ValidBodyNoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/customers/associations", h.CreateCustomerAssociation)

	req := httptest.NewRequest("POST", "/customers/associations", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestCreateCustomerAssociation_ValidBodyNoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/customers/associations", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.CreateCustomerAssociation(c)
	})

	req := httptest.NewRequest("POST", "/customers/associations", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// DeleteCustomerAssociation — :id is parsed before the permission check, so
// the invalid-ID branch is reachable even without a user.
// ---------------------------------------------------------------------------

func TestDeleteCustomerAssociation_InvalidID_Returns400(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Delete("/customers/associations/:id", h.DeleteCustomerAssociation)

	req := httptest.NewRequest("DELETE", "/customers/associations/not-a-number", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	data, _ := io.ReadAll(resp.Body)
	var body ErrorResponse
	assert.NoError(t, json.Unmarshal(data, &body))
	assert.Equal(t, string(ErrBadRequest), body.Code)
}

func TestDeleteCustomerAssociation_ValidIDNoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Delete("/customers/associations/:id", h.DeleteCustomerAssociation)

	req := httptest.NewRequest("DELETE", "/customers/associations/1", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestDeleteCustomerAssociation_ValidIDNoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Delete("/customers/associations/:id", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.DeleteCustomerAssociation(c)
	})

	req := httptest.NewRequest("DELETE", "/customers/associations/1", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}
