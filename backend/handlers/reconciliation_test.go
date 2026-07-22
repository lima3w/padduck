package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// GetReconciliationReport is gated by requirePerm; its only reachable
// branches without a DB are the auth guard ones. The success path calls
// nil Reports/IPAM services and is covered by integration tests instead.

func TestGetReconciliationReport_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/reports/reconciliation", h.GetReconciliationReport)

	req := httptest.NewRequest("GET", "/admin/reports/reconciliation", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetReconciliationReport_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/reports/reconciliation", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.GetReconciliationReport(c)
	})

	req := httptest.NewRequest("GET", "/admin/reports/reconciliation", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}
