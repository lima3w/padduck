package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// GetNotificationPreferences — GET /api/v1/user/notification-preferences
// ---------------------------------------------------------------------------

func TestGetNotificationPreferences_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/user/notification-preferences", h.GetNotificationPreferences)
	resp, err := app.Test(httptest.NewRequest("GET", "/user/notification-preferences", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// UpdateNotificationPreferences — PUT /api/v1/user/notification-preferences
// ---------------------------------------------------------------------------

func TestUpdateNotificationPreferences_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Put("/user/notification-preferences", h.UpdateNotificationPreferences)
	req := httptest.NewRequest("PUT", "/user/notification-preferences", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// GetNotificationStats — GET /api/v1/admin/notification-stats
// ---------------------------------------------------------------------------

func TestGetNotificationStats_NoUser_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/notification-stats", h.GetNotificationStats)
	resp, err := app.Test(httptest.NewRequest("GET", "/admin/notification-stats", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestGetNotificationStats_NonAdmin_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user", nonAdminUser)
		return c.Next()
	})
	app.Get("/admin/notification-stats", h.GetNotificationStats)
	resp, err := app.Test(httptest.NewRequest("GET", "/admin/notification-stats", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}
