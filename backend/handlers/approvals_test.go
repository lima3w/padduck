package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// All approval handlers use a direct user.Role == "admin" check.

// ---------------------------------------------------------------------------
// ListPendingApprovals — GET /api/v1/admin/approvals
// ---------------------------------------------------------------------------

func TestListPendingApprovals_NoUser_Returns403(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Get("/admin/approvals", h.ListPendingApprovals)
	resp, err := app.Test(httptest.NewRequest("GET", "/admin/approvals", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestListPendingApprovals_NonAdmin_Returns403(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user", nonAdminUser)
		return c.Next()
	})
	app.Get("/admin/approvals", h.ListPendingApprovals)
	resp, err := app.Test(httptest.NewRequest("GET", "/admin/approvals", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// ApproveUser — POST /api/v1/admin/approvals/:id/approve
// ---------------------------------------------------------------------------

func TestApproveUser_NoUser_Returns403(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Post("/admin/approvals/:id/approve", h.ApproveUser)
	resp, err := app.Test(httptest.NewRequest("POST", "/admin/approvals/1/approve", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestApproveUser_NonAdmin_Returns403(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user", nonAdminUser)
		return c.Next()
	})
	app.Post("/admin/approvals/:id/approve", h.ApproveUser)
	resp, err := app.Test(httptest.NewRequest("POST", "/admin/approvals/1/approve", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestApproveUser_BadID_Returns400(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user", adminUser)
		return c.Next()
	})
	app.Post("/admin/approvals/:id/approve", h.ApproveUser)
	resp, err := app.Test(httptest.NewRequest("POST", "/admin/approvals/notanumber/approve", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// RejectUser — POST /api/v1/admin/approvals/:id/reject
// ---------------------------------------------------------------------------

func TestRejectUser_NoUser_Returns403(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Post("/admin/approvals/:id/reject", h.RejectUser)
	resp, err := app.Test(httptest.NewRequest("POST", "/admin/approvals/1/reject", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestRejectUser_NonAdmin_Returns403(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user", nonAdminUser)
		return c.Next()
	})
	app.Post("/admin/approvals/:id/reject", h.RejectUser)
	resp, err := app.Test(httptest.NewRequest("POST", "/admin/approvals/1/reject", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestRejectUser_BadID_Returns400(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user", adminUser)
		return c.Next()
	})
	app.Post("/admin/approvals/:id/reject", h.RejectUser)
	resp, err := app.Test(httptest.NewRequest("POST", "/admin/approvals/notanumber/reject", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
