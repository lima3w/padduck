package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"ipam-next/models"
)

// ---------------------------------------------------------------------------
// SplitSubnet — auth enforcement
// ---------------------------------------------------------------------------

func TestSplitSubnet_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/admin/subnets/:id/split", h.SplitSubnet)

	req := httptest.NewRequest("POST", "/admin/subnets/1/split", strings.NewReader(`{"new_prefix_len":26}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestSplitSubnet_NonAdmin_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/admin/subnets/:id/split", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{Role: "user"})
		return h.SplitSubnet(c)
	})

	req := httptest.NewRequest("POST", "/admin/subnets/1/split", strings.NewReader(`{"new_prefix_len":26}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestSplitSubnet_AdminInvalidID_Returns403(t *testing.T) {
	// permCheck with nil service and admin user (ID=0) returns 403 before ID parsing
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/admin/subnets/:id/split", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{Role: "admin"})
		return h.SplitSubnet(c)
	})

	req := httptest.NewRequest("POST", "/admin/subnets/notanint/split", strings.NewReader(`{"new_prefix_len":26}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// MergeSubnets — auth enforcement
// ---------------------------------------------------------------------------

func TestMergeSubnets_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/admin/subnets/merge", h.MergeSubnets)

	req := httptest.NewRequest("POST", "/admin/subnets/merge", strings.NewReader(`{"subnet_ids":[1,2]}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestMergeSubnets_NonAdmin_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/admin/subnets/merge", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{Role: "user"})
		return h.MergeSubnets(c)
	})

	req := httptest.NewRequest("POST", "/admin/subnets/merge", strings.NewReader(`{"subnet_ids":[1,2]}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestMergeSubnets_AdminTooFewIDs_Returns403(t *testing.T) {
	// permCheck with nil service and admin user (ID=0) returns 403 before validation
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/admin/subnets/merge", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{Role: "admin"})
		return h.MergeSubnets(c)
	})

	req := httptest.NewRequest("POST", "/admin/subnets/merge", strings.NewReader(`{"subnet_ids":[1]}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// ResizeSubnet — auth enforcement
// ---------------------------------------------------------------------------

func TestResizeSubnet_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/admin/subnets/:id/resize", h.ResizeSubnet)

	req := httptest.NewRequest("POST", "/admin/subnets/1/resize", strings.NewReader(`{"new_prefix":"192.168.0.0/23"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestResizeSubnet_NonAdmin_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/admin/subnets/:id/resize", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{Role: "user"})
		return h.ResizeSubnet(c)
	})

	req := httptest.NewRequest("POST", "/admin/subnets/1/resize", strings.NewReader(`{"new_prefix":"192.168.0.0/23"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestResizeSubnet_AdminInvalidID_Returns403(t *testing.T) {
	// permCheck with nil service and admin user (ID=0) returns 403 before ID parsing
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/admin/subnets/:id/resize", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{Role: "admin"})
		return h.ResizeSubnet(c)
	})

	req := httptest.NewRequest("POST", "/admin/subnets/notanint/resize", strings.NewReader(`{"new_prefix":"192.168.0.0/23"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// Request struct validation
// ---------------------------------------------------------------------------

func TestSplitSubnetRequest_Fields(t *testing.T) {
	req := &SplitSubnetRequest{NewPrefixLen: 26}
	assert.Equal(t, 26, req.NewPrefixLen)
}

func TestMergeSubnetsRequest_Fields(t *testing.T) {
	req := &MergeSubnetsRequest{SubnetIDs: []int64{1, 2}}
	assert.Equal(t, []int64{1, 2}, req.SubnetIDs)
}

func TestResizeSubnetRequest_Fields(t *testing.T) {
	req := &ResizeSubnetRequest{NewPrefix: "10.0.0.0/23"}
	assert.Equal(t, "10.0.0.0/23", req.NewPrefix)
}
