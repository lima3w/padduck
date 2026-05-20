package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"padduck/models"
)

var unprivDNS = &models.User{ID: 0, Role: "viewer"}

// ---------------------------------------------------------------------------
// CheckAllDNS — POST /admin/dns/check-all
// ---------------------------------------------------------------------------

func TestCheckAllDNS_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/admin/dns/check-all", h.CheckAllDNS)

	resp, err := app.Test(httptest.NewRequest("POST", "/admin/dns/check-all", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestCheckAllDNS_NoPermission_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/admin/dns/check-all", func(c *fiber.Ctx) error {
		c.Locals("user", unprivDNS)
		return h.CheckAllDNS(c)
	})

	resp, err := app.Test(httptest.NewRequest("POST", "/admin/dns/check-all", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// TestPowerDNSConnection — POST /admin/dns/test
// ---------------------------------------------------------------------------

func TestTestPowerDNSConnection_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/admin/dns/test", h.TestPowerDNSConnection)

	resp, err := app.Test(httptest.NewRequest("POST", "/admin/dns/test", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestTestPowerDNSConnection_NoPermission_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/admin/dns/test", func(c *fiber.Ctx) error {
		c.Locals("user", unprivDNS)
		return h.TestPowerDNSConnection(c)
	})

	resp, err := app.Test(httptest.NewRequest("POST", "/admin/dns/test", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// TestTechnitiumConnection — POST /admin/dns/technitium/test
// ---------------------------------------------------------------------------

func TestTestTechnitiumConnection_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/admin/dns/technitium/test", h.TestTechnitiumConnection)

	resp, err := app.Test(httptest.NewRequest("POST", "/admin/dns/technitium/test", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestTestTechnitiumConnection_NoPermission_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/admin/dns/technitium/test", func(c *fiber.Ctx) error {
		c.Locals("user", unprivDNS)
		return h.TestTechnitiumConnection(c)
	})

	resp, err := app.Test(httptest.NewRequest("POST", "/admin/dns/technitium/test", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// ListDNSZones — GET /dns/zones
// ---------------------------------------------------------------------------

func TestListDNSZones_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/dns/zones", h.ListDNSZones)

	resp, err := app.Test(httptest.NewRequest("GET", "/dns/zones", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestListDNSZones_NoPermission_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/dns/zones", func(c *fiber.Ctx) error {
		c.Locals("user", unprivDNS)
		return h.ListDNSZones(c)
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/dns/zones", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// GetDNSZoneRecords — GET /dns/zones/:zone/records
// ---------------------------------------------------------------------------

func TestGetDNSZoneRecords_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/dns/zones/:zone/records", h.GetDNSZoneRecords)

	resp, err := app.Test(httptest.NewRequest("GET", "/dns/zones/example.com./records", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetDNSZoneRecords_NoPermission_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/dns/zones/:zone/records", func(c *fiber.Ctx) error {
		c.Locals("user", unprivDNS)
		return h.GetDNSZoneRecords(c)
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/dns/zones/example.com./records", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}
