package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// GetDashboardSummary — GET /api/v1/dashboard/summary
// ---------------------------------------------------------------------------

func TestGetDashboardSummary_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/dashboard/summary", h.GetDashboardSummary)
	resp, err := app.Test(httptest.NewRequest("GET", "/dashboard/summary", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetDashboardSummary_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return c.Next()
	})
	app.Get("/dashboard/summary", h.GetDashboardSummary)
	resp, err := app.Test(httptest.NewRequest("GET", "/dashboard/summary", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// GetDashboardRecentActivity — GET /api/v1/dashboard/recent-activity
// ---------------------------------------------------------------------------

func TestGetDashboardRecentActivity_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/dashboard/recent-activity", h.GetDashboardRecentActivity)
	resp, err := app.Test(httptest.NewRequest("GET", "/dashboard/recent-activity", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetDashboardRecentActivity_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return c.Next()
	})
	app.Get("/dashboard/recent-activity", h.GetDashboardRecentActivity)
	resp, err := app.Test(httptest.NewRequest("GET", "/dashboard/recent-activity", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// GetSubnetTree — GET /api/v1/networks/:id/subnets/tree
// ID is parsed before permCheck so bad ID returns 400 without auth.
// ---------------------------------------------------------------------------

func TestGetSubnetTree_BadSectionID_Returns400(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/networks/:id/subnets/tree", h.GetSubnetTree)
	resp, err := app.Test(httptest.NewRequest("GET", "/networks/notanint/subnets/tree", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestGetSubnetTree_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/networks/:id/subnets/tree", h.GetSubnetTree)
	resp, err := app.Test(httptest.NewRequest("GET", "/networks/1/subnets/tree", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetSubnetTree_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return c.Next()
	})
	app.Get("/networks/:id/subnets/tree", h.GetSubnetTree)
	resp, err := app.Test(httptest.NewRequest("GET", "/networks/1/subnets/tree", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}
