package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// SearchSubnets — POST /api/v1/subnets/search/:networkID
// ---------------------------------------------------------------------------

func TestSearchSubnets_BadSectionID_Returns400(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Post("/subnets/search/:networkID", h.SearchSubnets)
	req := httptest.NewRequest("POST", "/subnets/search/notanumber", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// SearchIPAddresses — POST /api/v1/ip-addresses/search/:subnetID
// ---------------------------------------------------------------------------

func TestSearchIPAddresses_BadSubnetID_Returns400(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Post("/ip-addresses/search/:subnetID", h.SearchIPAddresses)
	req := httptest.NewRequest("POST", "/ip-addresses/search/notanumber", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// GlobalSearch — GET /api/v1/search
// ---------------------------------------------------------------------------

func TestGlobalSearch_NoUser_Returns401(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Get("/search", h.GlobalSearch)
	resp, err := app.Test(httptest.NewRequest("GET", "/search?q=test", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// SearchNetworks — POST /api/v1/networks/search
// ---------------------------------------------------------------------------

func TestSearchNetworks_NoUser_Returns401(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Post("/networks/search", h.SearchNetworks)
	req := httptest.NewRequest("POST", "/networks/search", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// SearchIPAddressesGlobal — GET /api/v1/ip-addresses/search
// ---------------------------------------------------------------------------

func TestSearchIPAddressesGlobal_NoUser_Returns401(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Get("/ip-addresses/search", h.SearchIPAddressesGlobal)
	resp, err := app.Test(httptest.NewRequest("GET", "/ip-addresses/search?q=10.0", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}
