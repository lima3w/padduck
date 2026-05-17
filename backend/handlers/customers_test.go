package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// ListCustomers — GET /api/v1/customers
// ---------------------------------------------------------------------------

func TestListCustomers_NoUser_Returns401(t *testing.T) {
	h := &Handler{}
	app := deviceApp(h, "GET", "/customers", h.ListCustomers)
	resp, err := app.Test(httptest.NewRequest("GET", "/customers", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestListCustomers_NoPermission_Returns403(t *testing.T) {
	h := &Handler{}
	app := deviceAppAs(h, "GET", "/customers", h.ListCustomers, permUser())
	resp, err := app.Test(httptest.NewRequest("GET", "/customers", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// GetCustomer — GET /api/v1/customers/:id
// ---------------------------------------------------------------------------

func TestGetCustomer_NoUser_Returns401(t *testing.T) {
	h := &Handler{}
	app := deviceApp(h, "GET", "/customers/:id", h.GetCustomer)
	resp, err := app.Test(httptest.NewRequest("GET", "/customers/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetCustomer_NoPermission_Returns403(t *testing.T) {
	h := &Handler{}
	app := deviceAppAs(h, "GET", "/customers/:id", h.GetCustomer, permUser())
	resp, err := app.Test(httptest.NewRequest("GET", "/customers/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// CreateCustomer — POST /api/v1/customers
// ---------------------------------------------------------------------------

func TestCreateCustomer_NoUser_Returns401(t *testing.T) {
	h := &Handler{}
	app := deviceApp(h, "POST", "/customers", h.CreateCustomer)
	req := httptest.NewRequest("POST", "/customers", strings.NewReader(`{"name":"Acme Corp"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestCreateCustomer_NoPermission_Returns403(t *testing.T) {
	h := &Handler{}
	app := deviceAppAs(h, "POST", "/customers", h.CreateCustomer, permUser())
	req := httptest.NewRequest("POST", "/customers", strings.NewReader(`{"name":"Acme Corp"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// UpdateCustomer — PUT /api/v1/customers/:id
// ---------------------------------------------------------------------------

func TestUpdateCustomer_NoUser_Returns401(t *testing.T) {
	h := &Handler{}
	app := deviceApp(h, "PUT", "/customers/:id", h.UpdateCustomer)
	req := httptest.NewRequest("PUT", "/customers/1", strings.NewReader(`{"name":"Acme Corp Updated"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestUpdateCustomer_NoPermission_Returns403(t *testing.T) {
	h := &Handler{}
	app := deviceAppAs(h, "PUT", "/customers/:id", h.UpdateCustomer, permUser())
	req := httptest.NewRequest("PUT", "/customers/1", strings.NewReader(`{"name":"Acme Corp Updated"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// DeleteCustomer — DELETE /api/v1/customers/:id
// ---------------------------------------------------------------------------

func TestDeleteCustomer_NoUser_Returns401(t *testing.T) {
	h := &Handler{}
	app := deviceApp(h, "DELETE", "/customers/:id", h.DeleteCustomer)
	resp, err := app.Test(httptest.NewRequest("DELETE", "/customers/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestDeleteCustomer_NoPermission_Returns403(t *testing.T) {
	h := &Handler{}
	app := deviceAppAs(h, "DELETE", "/customers/:id", h.DeleteCustomer, permUser())
	resp, err := app.Test(httptest.NewRequest("DELETE", "/customers/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}
