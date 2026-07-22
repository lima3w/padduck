package handlers

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"padduck/models"
)

// ---------------------------------------------------------------------------
// ListTechnitiumDHCPScopes / SyncTechnitiumLeases — requireAdmin guard only;
// success paths touch a nil service and are covered by integration tests.
// ---------------------------------------------------------------------------

func TestListTechnitiumDHCPScopes_NoUser_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/dhcp/technitium/scopes", h.ListTechnitiumDHCPScopes)

	req := httptest.NewRequest("GET", "/admin/dhcp/technitium/scopes", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestListTechnitiumDHCPScopes_NonAdmin_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/dhcp/technitium/scopes", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{Role: "user"})
		return h.ListTechnitiumDHCPScopes(c)
	})

	req := httptest.NewRequest("GET", "/admin/dhcp/technitium/scopes", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestSyncTechnitiumLeases_NoUser_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/admin/dhcp/technitium/sync", h.SyncTechnitiumLeases)

	req := httptest.NewRequest("POST", "/admin/dhcp/technitium/sync", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestSyncTechnitiumLeases_NonAdmin_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/admin/dhcp/technitium/sync", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{Role: "user"})
		return h.SyncTechnitiumLeases(c)
	})

	req := httptest.NewRequest("POST", "/admin/dhcp/technitium/sync", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// ImportTechnitiumScope — requireAdmin never touches the repo, so an
// admin-role user reaches the body-validation branch, which is entirely
// self-contained (no service call) and therefore fully testable.
// ---------------------------------------------------------------------------

func TestImportTechnitiumScope_NoUser_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/admin/dhcp/technitium/import-scope", h.ImportTechnitiumScope)

	req := httptest.NewRequest("POST", "/admin/dhcp/technitium/import-scope", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestImportTechnitiumScope_NonAdmin_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/admin/dhcp/technitium/import-scope", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{Role: "user"})
		return h.ImportTechnitiumScope(c)
	})

	req := httptest.NewRequest("POST", "/admin/dhcp/technitium/import-scope", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestImportTechnitiumScope_AdminInvalidBody_Returns400(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/admin/dhcp/technitium/import-scope", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{Role: "admin"})
		return h.ImportTechnitiumScope(c)
	})

	req := httptest.NewRequest("POST", "/admin/dhcp/technitium/import-scope", strings.NewReader("{not valid json"))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestImportTechnitiumScope_AdminMissingScopeName_Returns400(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/admin/dhcp/technitium/import-scope", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{Role: "admin"})
		return h.ImportTechnitiumScope(c)
	})

	req := httptest.NewRequest("POST", "/admin/dhcp/technitium/import-scope", strings.NewReader(`{"network_id":1}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	data, _ := io.ReadAll(resp.Body)
	var body ErrorResponse
	assert.NoError(t, json.Unmarshal(data, &body))
	assert.Equal(t, "scope_name and network_id are required", body.Error)
}

func TestImportTechnitiumScope_AdminMissingNetworkID_Returns400(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/admin/dhcp/technitium/import-scope", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{Role: "admin"})
		return h.ImportTechnitiumScope(c)
	})

	req := httptest.NewRequest("POST", "/admin/dhcp/technitium/import-scope", strings.NewReader(`{"scope_name":"office"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// PushDHCPReservation / RemoveDHCPReservation — :id is parsed before the
// permission check, so the invalid-ID branch is reachable without a user.
// ---------------------------------------------------------------------------

func TestPushDHCPReservation_InvalidID_Returns400(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/ip-addresses/:id/dhcp-reservation", h.PushDHCPReservation)

	req := httptest.NewRequest("POST", "/ip-addresses/not-a-number/dhcp-reservation", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestPushDHCPReservation_ValidIDNoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/ip-addresses/:id/dhcp-reservation", h.PushDHCPReservation)

	req := httptest.NewRequest("POST", "/ip-addresses/1/dhcp-reservation", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestPushDHCPReservation_ValidIDNoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/ip-addresses/:id/dhcp-reservation", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.PushDHCPReservation(c)
	})

	req := httptest.NewRequest("POST", "/ip-addresses/1/dhcp-reservation", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestRemoveDHCPReservation_InvalidID_Returns400(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Delete("/ip-addresses/:id/dhcp-reservation", h.RemoveDHCPReservation)

	req := httptest.NewRequest("DELETE", "/ip-addresses/not-a-number/dhcp-reservation", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestRemoveDHCPReservation_ValidIDNoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Delete("/ip-addresses/:id/dhcp-reservation", h.RemoveDHCPReservation)

	req := httptest.NewRequest("DELETE", "/ip-addresses/1/dhcp-reservation", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestRemoveDHCPReservation_ValidIDNoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Delete("/ip-addresses/:id/dhcp-reservation", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.RemoveDHCPReservation(c)
	})

	req := httptest.NewRequest("DELETE", "/ip-addresses/1/dhcp-reservation", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}
