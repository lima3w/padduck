package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// RequireBearerAuth middleware — Grafana POST endpoints
// ---------------------------------------------------------------------------

// TestGrafanaSearch_CookieOnly_Returns401 verifies that a POST /search request
// carrying only a session cookie (no Authorization header) is rejected with 401.
// This ensures the endpoint cannot be exploited via CSRF.
func TestGrafanaSearch_CookieOnly_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/api/grafana/search", h.RequireBearerAuth, func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/grafana/search", strings.NewReader(`{"target":""}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "some-session-token"})

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

// TestGrafanaQuery_CookieOnly_Returns401 verifies that a POST /query request
// carrying only a session cookie (no Authorization header) is rejected with 401.
// This ensures the endpoint cannot be exploited via CSRF.
func TestGrafanaQuery_CookieOnly_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/api/grafana/query", h.RequireBearerAuth, func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/grafana/query", strings.NewReader(`{"targets":[]}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "some-session-token"})

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

// TestRequireBearerAuth_WithBearerHeader_Passes verifies that a request with a
// valid Authorization: Bearer header is allowed through the middleware.
func TestRequireBearerAuth_WithBearerHeader_Passes(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/api/grafana/search", h.RequireBearerAuth, func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/grafana/search", strings.NewReader(`{"target":""}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-api-token")

	resp, err := app.Test(req)
	require.NoError(t, err)
	// Middleware passes; handler returns 200
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}
