package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"padduck/models"
)

var unprivIP = &models.User{ID: 0, Role: "viewer"}

// ---------------------------------------------------------------------------
// CreateIPAddress — POST /subnets/:subnetID/ip-addresses
// ---------------------------------------------------------------------------

func TestCreateIPAddress_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/subnets/:subnetID/ip-addresses", h.CreateIPAddress)

	resp, err := app.Test(httptest.NewRequest("POST", "/subnets/1/ip-addresses", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestCreateIPAddress_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/subnets/:subnetID/ip-addresses", func(c *fiber.Ctx) error {
		c.Locals("user", unprivIP)
		return h.CreateIPAddress(c)
	})

	resp, err := app.Test(httptest.NewRequest("POST", "/subnets/1/ip-addresses", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestCreateIPAddress_BadSubnetID_Returns400(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/subnets/:subnetID/ip-addresses", h.CreateIPAddress)

	resp, err := app.Test(httptest.NewRequest("POST", "/subnets/abc/ip-addresses", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// ListIPAddresses — GET /subnets/:subnetID/ip-addresses
// ---------------------------------------------------------------------------

func TestListIPAddresses_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/subnets/:subnetID/ip-addresses", h.ListIPAddresses)

	resp, err := app.Test(httptest.NewRequest("GET", "/subnets/1/ip-addresses", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestListIPAddresses_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/subnets/:subnetID/ip-addresses", func(c *fiber.Ctx) error {
		c.Locals("user", unprivIP)
		return h.ListIPAddresses(c)
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/subnets/1/ip-addresses", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestListIPAddresses_BadSubnetID_Returns400(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/subnets/:subnetID/ip-addresses", h.ListIPAddresses)

	resp, err := app.Test(httptest.NewRequest("GET", "/subnets/abc/ip-addresses", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// GetIPAddress — GET /ip-addresses/:id
// ---------------------------------------------------------------------------

func TestGetIPAddress_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/ip-addresses/:id", h.GetIPAddress)

	resp, err := app.Test(httptest.NewRequest("GET", "/ip-addresses/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetIPAddress_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/ip-addresses/:id", func(c *fiber.Ctx) error {
		c.Locals("user", unprivIP)
		return h.GetIPAddress(c)
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/ip-addresses/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestGetIPAddress_BadID_NoAuth_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/ip-addresses/:id", h.GetIPAddress)

	// permCheck runs before ParamsInt, so unauthenticated requests get 401.
	resp, err := app.Test(httptest.NewRequest("GET", "/ip-addresses/abc", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// AllocateIPAddress — POST /subnets/:subnetID/ip-addresses/allocate
// ---------------------------------------------------------------------------

func TestAllocateIPAddress_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/subnets/:subnetID/ip-addresses/allocate", h.AllocateIPAddress)

	resp, err := app.Test(httptest.NewRequest("POST", "/subnets/1/ip-addresses/allocate", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestAllocateIPAddress_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/subnets/:subnetID/ip-addresses/allocate", func(c *fiber.Ctx) error {
		c.Locals("user", unprivIP)
		return h.AllocateIPAddress(c)
	})

	resp, err := app.Test(httptest.NewRequest("POST", "/subnets/1/ip-addresses/allocate", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestAllocateIPAddress_BadSubnetID_Returns400(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/subnets/:subnetID/ip-addresses/allocate", h.AllocateIPAddress)

	resp, err := app.Test(httptest.NewRequest("POST", "/subnets/abc/ip-addresses/allocate", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// GetSubnetUtilization — GET /subnets/:subnetID/utilization
// ---------------------------------------------------------------------------

func TestGetSubnetUtilization_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/subnets/:subnetID/utilization", h.GetSubnetUtilization)

	resp, err := app.Test(httptest.NewRequest("GET", "/subnets/1/utilization", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetSubnetUtilization_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/subnets/:subnetID/utilization", func(c *fiber.Ctx) error {
		c.Locals("user", unprivIP)
		return h.GetSubnetUtilization(c)
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/subnets/1/utilization", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestGetSubnetUtilization_BadID_Returns400(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/subnets/:subnetID/utilization", h.GetSubnetUtilization)

	resp, err := app.Test(httptest.NewRequest("GET", "/subnets/abc/utilization", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
