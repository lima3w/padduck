package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// GetDuplicates is gated by requirePerm; its only reachable branches without
// a DB are the auth guard ones. The success path calls a nil Reports
// service and is covered by integration tests instead.

func TestGetDuplicates_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/reports/duplicates", h.GetDuplicates)

	req := httptest.NewRequest("GET", "/admin/reports/duplicates", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetDuplicates_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/reports/duplicates", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.GetDuplicates(c)
	})

	req := httptest.NewRequest("GET", "/admin/reports/duplicates", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}
