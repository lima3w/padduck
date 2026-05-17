package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// ListAutonomousSystems — GET /api/v1/autonomous-systems
// ---------------------------------------------------------------------------

func TestListAutonomousSystems_NoUser_Returns401(t *testing.T) {
	h := &Handler{}
	app := deviceApp(h, "GET", "/autonomous-systems", h.ListAutonomousSystems)
	resp, err := app.Test(httptest.NewRequest("GET", "/autonomous-systems", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestListAutonomousSystems_NoPermission_Returns403(t *testing.T) {
	h := &Handler{}
	app := deviceAppAs(h, "GET", "/autonomous-systems", h.ListAutonomousSystems, permUser())
	resp, err := app.Test(httptest.NewRequest("GET", "/autonomous-systems", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// GetAutonomousSystem — GET /api/v1/autonomous-systems/:id
// ---------------------------------------------------------------------------

func TestGetAutonomousSystem_NoUser_Returns401(t *testing.T) {
	h := &Handler{}
	app := deviceApp(h, "GET", "/autonomous-systems/:id", h.GetAutonomousSystem)
	resp, err := app.Test(httptest.NewRequest("GET", "/autonomous-systems/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetAutonomousSystem_NoPermission_Returns403(t *testing.T) {
	h := &Handler{}
	app := deviceAppAs(h, "GET", "/autonomous-systems/:id", h.GetAutonomousSystem, permUser())
	resp, err := app.Test(httptest.NewRequest("GET", "/autonomous-systems/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// CreateAutonomousSystem — POST /api/v1/autonomous-systems
// ---------------------------------------------------------------------------

func TestCreateAutonomousSystem_NoUser_Returns401(t *testing.T) {
	h := &Handler{}
	app := deviceApp(h, "POST", "/autonomous-systems", h.CreateAutonomousSystem)
	req := httptest.NewRequest("POST", "/autonomous-systems", strings.NewReader(`{"asn":65000,"name":"Test AS"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestCreateAutonomousSystem_NoPermission_Returns403(t *testing.T) {
	h := &Handler{}
	app := deviceAppAs(h, "POST", "/autonomous-systems", h.CreateAutonomousSystem, permUser())
	req := httptest.NewRequest("POST", "/autonomous-systems", strings.NewReader(`{"asn":65000,"name":"Test AS"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// UpdateAutonomousSystem — PUT /api/v1/autonomous-systems/:id
// ---------------------------------------------------------------------------

func TestUpdateAutonomousSystem_NoUser_Returns401(t *testing.T) {
	h := &Handler{}
	app := deviceApp(h, "PUT", "/autonomous-systems/:id", h.UpdateAutonomousSystem)
	req := httptest.NewRequest("PUT", "/autonomous-systems/1", strings.NewReader(`{"asn":65000,"name":"Updated AS"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestUpdateAutonomousSystem_NoPermission_Returns403(t *testing.T) {
	h := &Handler{}
	app := deviceAppAs(h, "PUT", "/autonomous-systems/:id", h.UpdateAutonomousSystem, permUser())
	req := httptest.NewRequest("PUT", "/autonomous-systems/1", strings.NewReader(`{"asn":65000,"name":"Updated AS"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// DeleteAutonomousSystem — DELETE /api/v1/autonomous-systems/:id
// ---------------------------------------------------------------------------

func TestDeleteAutonomousSystem_NoUser_Returns401(t *testing.T) {
	h := &Handler{}
	app := deviceApp(h, "DELETE", "/autonomous-systems/:id", h.DeleteAutonomousSystem)
	resp, err := app.Test(httptest.NewRequest("DELETE", "/autonomous-systems/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestDeleteAutonomousSystem_NoPermission_Returns403(t *testing.T) {
	h := &Handler{}
	app := deviceAppAs(h, "DELETE", "/autonomous-systems/:id", h.DeleteAutonomousSystem, permUser())
	resp, err := app.Test(httptest.NewRequest("DELETE", "/autonomous-systems/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}
