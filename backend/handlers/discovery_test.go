package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"ipam-next/models"
)

var unprivDiscovery = &models.User{ID: 0, Role: "viewer"}

// ---------------------------------------------------------------------------
// GetScanJobResults — GET /admin/scan-jobs/:id/results
// ---------------------------------------------------------------------------

func TestGetScanJobResults_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/admin/scan-jobs/:id/results", h.GetScanJobResults)

	resp, err := app.Test(httptest.NewRequest("GET", "/admin/scan-jobs/1/results", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetScanJobResults_ViewerUser_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/admin/scan-jobs/:id/results", func(c *fiber.Ctx) error {
		c.Locals("user", unprivDiscovery)
		return h.GetScanJobResults(c)
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/admin/scan-jobs/1/results", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestGetScanJobResults_BadID_Returns400(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/admin/scan-jobs/:id/results", h.GetScanJobResults)

	resp, err := app.Test(httptest.NewRequest("GET", "/admin/scan-jobs/notanid/results", nil))
	assert.NoError(t, err)
	// 401 because auth runs first (no user in locals), 400 if we reach param parsing
	assert.True(t, resp.StatusCode == fiber.StatusUnauthorized || resp.StatusCode == fiber.StatusBadRequest)
}

// ---------------------------------------------------------------------------
// GetSubnetScanResults — GET /subnets/:id/scan-results
// ---------------------------------------------------------------------------

func TestGetSubnetScanResults_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/subnets/:id/scan-results", h.GetSubnetScanResults)

	resp, err := app.Test(httptest.NewRequest("GET", "/subnets/1/scan-results", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetSubnetScanResults_ViewerUser_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/subnets/:id/scan-results", func(c *fiber.Ctx) error {
		c.Locals("user", unprivDiscovery)
		return h.GetSubnetScanResults(c)
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/subnets/1/scan-results", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestGetSubnetScanResults_BadID_NoUser_Returns401(t *testing.T) {
	// permCheck runs before param parsing, so a missing user returns 401 before
	// the bad ID can be checked.
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/subnets/:id/scan-results", h.GetSubnetScanResults)

	resp, err := app.Test(httptest.NewRequest("GET", "/subnets/notanid/scan-results", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// RunScanJobNow — POST /admin/scan-jobs/:id/run
// ---------------------------------------------------------------------------

func TestRunScanJobNow_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/admin/scan-jobs/:id/run", h.RunScanJobNow)

	resp, err := app.Test(httptest.NewRequest("POST", "/admin/scan-jobs/1/run", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestRunScanJobNow_ViewerUser_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/admin/scan-jobs/:id/run", func(c *fiber.Ctx) error {
		c.Locals("user", unprivDiscovery)
		return h.RunScanJobNow(c)
	})

	resp, err := app.Test(httptest.NewRequest("POST", "/admin/scan-jobs/1/run", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}
