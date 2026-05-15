package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"ipam-next/models"
)

var unprivHistory = &models.User{ID: 0, Role: "viewer"}

// ---------------------------------------------------------------------------
// GetScanJobHistory — auth enforcement
// ---------------------------------------------------------------------------

func TestGetScanJobHistory_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/admin/scan-jobs/:id/history", h.GetScanJobHistory)

	req := httptest.NewRequest("GET", "/admin/scan-jobs/1/history", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetScanJobHistory_ViewerUser_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/admin/scan-jobs/:id/history", func(c *fiber.Ctx) error {
		c.Locals("user", unprivHistory)
		return h.GetScanJobHistory(c)
	})
	req := httptest.NewRequest("GET", "/admin/scan-jobs/1/history", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// GetScanRunDetail — auth enforcement
// ---------------------------------------------------------------------------

func TestGetScanRunDetail_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/admin/scan-jobs/:id/history/:run_id", h.GetScanRunDetail)

	req := httptest.NewRequest("GET", "/admin/scan-jobs/1/history/1", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetScanRunDetail_ViewerUser_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/admin/scan-jobs/:id/history/:run_id", func(c *fiber.Ctx) error {
		c.Locals("user", unprivHistory)
		return h.GetScanRunDetail(c)
	})
	req := httptest.NewRequest("GET", "/admin/scan-jobs/1/history/1", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}
