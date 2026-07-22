package handlers

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"padduck/models"
)

// SendTelemetryNow is gated by requireAdmin, a pure role check with no repo
// access, so both guard branches are reachable without a DB. The success
// path calls h.ops.Telemetry.SendNow with a nil service and is covered by
// integration tests instead.

func TestSendTelemetryNow_NoUser_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/admin/telemetry/send", h.SendTelemetryNow)

	req := httptest.NewRequest("POST", "/admin/telemetry/send", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)

	data, _ := io.ReadAll(resp.Body)
	var body ErrorResponse
	assert.NoError(t, json.Unmarshal(data, &body))
	assert.Equal(t, string(ErrForbidden), body.Code)
}

func TestSendTelemetryNow_NonAdmin_Returns403(t *testing.T) {
	for _, role := range []string{"user", "viewer", "operator"} {
		t.Run(role, func(t *testing.T) {
			h := minHandler()
			app := fiber.New()
			app.Post("/admin/telemetry/send", func(c *fiber.Ctx) error {
				c.Locals("user", &models.User{Role: role})
				return h.SendTelemetryNow(c)
			})
			req := httptest.NewRequest("POST", "/admin/telemetry/send", nil)
			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
		})
	}
}
