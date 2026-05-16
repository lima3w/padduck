package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// SearchSubnets — POST /api/v1/subnets/search/:sectionID
// ---------------------------------------------------------------------------

func TestSearchSubnets_BadSectionID_Returns400(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Post("/subnets/search/:sectionID", h.SearchSubnets)
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
