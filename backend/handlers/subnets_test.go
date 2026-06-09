package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"padduck/models"
)

var unprivSubnet = &models.User{ID: 0, Role: "viewer"}

// ---------------------------------------------------------------------------
// CreateSubnet — POST /networks/:networkID/subnets
// ---------------------------------------------------------------------------

func TestCreateSubnet_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/networks/:networkID/subnets", h.CreateSubnet)

	resp, err := app.Test(httptest.NewRequest("POST", "/networks/1/subnets", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestCreateSubnet_NoPermission_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/networks/:networkID/subnets", func(c *fiber.Ctx) error {
		c.Locals("user", unprivSubnet)
		return h.CreateSubnet(c)
	})

	resp, err := app.Test(httptest.NewRequest("POST", "/networks/1/subnets", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestCreateSubnet_BadSectionID_Returns400(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/networks/:networkID/subnets", h.CreateSubnet)

	resp, err := app.Test(httptest.NewRequest("POST", "/networks/abc/subnets", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// ListSubnets — GET /networks/:networkID/subnets
// ---------------------------------------------------------------------------

func TestListSubnets_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/networks/:networkID/subnets", h.ListSubnets)

	resp, err := app.Test(httptest.NewRequest("GET", "/networks/1/subnets", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestListSubnets_NoPermission_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/networks/:networkID/subnets", func(c *fiber.Ctx) error {
		c.Locals("user", unprivSubnet)
		return h.ListSubnets(c)
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/networks/1/subnets", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestListSubnets_BadSectionID_Returns400(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/networks/:networkID/subnets", h.ListSubnets)

	resp, err := app.Test(httptest.NewRequest("GET", "/networks/abc/subnets", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// GetSubnet — GET /subnets/:id
// ---------------------------------------------------------------------------

func TestGetSubnet_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/subnets/:id", h.GetSubnet)

	resp, err := app.Test(httptest.NewRequest("GET", "/subnets/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetSubnet_NoPermission_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/subnets/:id", func(c *fiber.Ctx) error {
		c.Locals("user", unprivSubnet)
		return h.GetSubnet(c)
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/subnets/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestGetSubnet_BadID_NoAuth_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/subnets/:id", h.GetSubnet)

	// permCheck runs before ParamsInt, so unauthenticated requests get 401.
	resp, err := app.Test(httptest.NewRequest("GET", "/subnets/abc", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// UpdateSubnet — PUT /subnets/:id
// ---------------------------------------------------------------------------

func TestUpdateSubnet_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Put("/subnets/:id", h.UpdateSubnet)

	resp, err := app.Test(httptest.NewRequest("PUT", "/subnets/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestUpdateSubnet_NoPermission_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Put("/subnets/:id", func(c *fiber.Ctx) error {
		c.Locals("user", unprivSubnet)
		return h.UpdateSubnet(c)
	})

	resp, err := app.Test(httptest.NewRequest("PUT", "/subnets/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestUpdateSubnet_BadID_Returns400(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Put("/subnets/:id", h.UpdateSubnet)

	resp, err := app.Test(httptest.NewRequest("PUT", "/subnets/abc", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// DeleteSubnet — DELETE /subnets/:id
// ---------------------------------------------------------------------------

func TestDeleteSubnet_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Delete("/subnets/:id", h.DeleteSubnet)

	resp, err := app.Test(httptest.NewRequest("DELETE", "/subnets/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestDeleteSubnet_NoPermission_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Delete("/subnets/:id", func(c *fiber.Ctx) error {
		c.Locals("user", unprivSubnet)
		return h.DeleteSubnet(c)
	})

	resp, err := app.Test(httptest.NewRequest("DELETE", "/subnets/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestDeleteSubnet_BadID_Returns400(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Delete("/subnets/:id", h.DeleteSubnet)

	resp, err := app.Test(httptest.NewRequest("DELETE", "/subnets/abc", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// GetOverlapReport — GET /api/v1/admin/subnets/overlap-report
// ---------------------------------------------------------------------------

func TestGetOverlapReport_NoUser_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/admin/subnets/overlap-report", h.GetOverlapReport)

	resp, err := app.Test(httptest.NewRequest("GET", "/admin/subnets/overlap-report", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestGetOverlapReport_NonAdmin_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/admin/subnets/overlap-report", func(c *fiber.Ctx) error {
		c.Locals("user", unprivSubnet)
		return h.GetOverlapReport(c)
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/admin/subnets/overlap-report", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}
