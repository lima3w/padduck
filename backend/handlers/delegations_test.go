package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"padduck/models"
)

// ---------------------------------------------------------------------------
// ListDelegations — auth enforcement
// ---------------------------------------------------------------------------

func TestListDelegations_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/subnets/:id/delegations", h.ListDelegations)

	req := httptest.NewRequest("GET", "/subnets/1/delegations", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestListDelegations_NonAdmin_Returns403(t *testing.T) {
	// viewer/user without roles uses legacy fallback; nil service means
	// CheckPermission cannot query DB, returns permission denied for non-admin.
	h := minHandler()
	app := fiber.New()
	app.Get("/subnets/:id/delegations", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{Role: "viewer", ID: 0})
		return h.ListDelegations(c)
	})

	req := httptest.NewRequest("GET", "/subnets/1/delegations", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	// With ID=0, userID <= 0 returns "permission denied" → 403
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestListDelegations_AdminInvalidID_Returns403(t *testing.T) {
	// permCheck with nil service and admin user (ID=0) returns 403 before param parsing
	h := minHandler()
	app := fiber.New()
	app.Get("/subnets/:id/delegations", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{Role: "admin"})
		return h.ListDelegations(c)
	})

	req := httptest.NewRequest("GET", "/subnets/notanint/delegations", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// CreateDelegation — auth enforcement
// ---------------------------------------------------------------------------

func TestCreateDelegation_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/subnets/:id/delegations", h.CreateDelegation)

	req := httptest.NewRequest("POST", "/subnets/1/delegations", strings.NewReader(`{"delegated_prefix":"2001:db8::/48"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestCreateDelegation_NonAdmin_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/subnets/:id/delegations", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{Role: "user"})
		return h.CreateDelegation(c)
	})

	req := httptest.NewRequest("POST", "/subnets/1/delegations", strings.NewReader(`{"delegated_prefix":"2001:db8::/48"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestCreateDelegation_AdminMissingPrefix_Returns403(t *testing.T) {
	// permCheck with nil service and admin user (ID=0) returns 403 before body validation
	h := minHandler()
	app := fiber.New()
	app.Post("/subnets/:id/delegations", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{Role: "admin"})
		return h.CreateDelegation(c)
	})

	req := httptest.NewRequest("POST", "/subnets/1/delegations", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// UpdateDelegation — auth enforcement
// ---------------------------------------------------------------------------

func TestUpdateDelegation_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Put("/delegations/:id", h.UpdateDelegation)

	req := httptest.NewRequest("PUT", "/delegations/1", strings.NewReader(`{"delegated_prefix":"2001:db8::/48"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestUpdateDelegation_NonAdmin_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Put("/delegations/:id", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{Role: "user"})
		return h.UpdateDelegation(c)
	})

	req := httptest.NewRequest("PUT", "/delegations/1", strings.NewReader(`{"delegated_prefix":"2001:db8::/48"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestUpdateDelegation_AdminInvalidID_Returns403(t *testing.T) {
	// permCheck with nil service and admin user (ID=0) returns 403 before ID parsing
	h := minHandler()
	app := fiber.New()
	app.Put("/delegations/:id", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{Role: "admin"})
		return h.UpdateDelegation(c)
	})

	req := httptest.NewRequest("PUT", "/delegations/notanint", strings.NewReader(`{"delegated_prefix":"2001:db8::/48"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// DeleteDelegation — auth enforcement
// ---------------------------------------------------------------------------

func TestDeleteDelegation_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Delete("/delegations/:id", h.DeleteDelegation)

	req := httptest.NewRequest("DELETE", "/delegations/1", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestDeleteDelegation_NonAdmin_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Delete("/delegations/:id", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{Role: "user"})
		return h.DeleteDelegation(c)
	})

	req := httptest.NewRequest("DELETE", "/delegations/1", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestDeleteDelegation_AdminInvalidID_Returns403(t *testing.T) {
	// permCheck with nil service and admin user (ID=0) returns 403 before ID parsing
	h := minHandler()
	app := fiber.New()
	app.Delete("/delegations/:id", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{Role: "admin"})
		return h.DeleteDelegation(c)
	})

	req := httptest.NewRequest("DELETE", "/delegations/notanint", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// GetNetworkTopology — auth enforcement
// ---------------------------------------------------------------------------

func TestGetSectionTopology_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/networks/:id/topology", h.GetNetworkTopology)

	req := httptest.NewRequest("GET", "/networks/1/topology", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetSectionTopology_NonAdmin_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/networks/:id/topology", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{Role: "user"})
		return h.GetNetworkTopology(c)
	})

	req := httptest.NewRequest("GET", "/networks/1/topology", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestGetSectionTopology_AdminInvalidID_Returns403(t *testing.T) {
	// permCheck with nil service and admin user (ID=0) returns 403 before param parsing
	h := minHandler()
	app := fiber.New()
	app.Get("/networks/:id/topology", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{Role: "admin"})
		return h.GetNetworkTopology(c)
	})

	req := httptest.NewRequest("GET", "/networks/notanint/topology", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}
