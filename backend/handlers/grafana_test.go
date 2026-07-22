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

// ---------------------------------------------------------------------------
// GrafanaHealth / GrafanaSearch / GrafanaQuery — auth enforcement.
//
// All three are requirePerm-gated, so success paths (including
// GrafanaQuery's known-target branches, which touch a nil Identity
// service) are only reachable by a permitted user — which requires a live
// repo (see plan). Only the auth guard branches are testable here without
// a DB. buildGrafanaTable's "unknown target" default case has no service
// dependency and is tested directly below.
// ---------------------------------------------------------------------------

func TestGrafanaHealth_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/grafana/", h.GrafanaHealth)

	req := httptest.NewRequest("GET", "/grafana/", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGrafanaHealth_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/grafana/", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.GrafanaHealth(c)
	})

	req := httptest.NewRequest("GET", "/grafana/", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestGrafanaSearch_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/grafana/search", h.GrafanaSearch)

	req := httptest.NewRequest("POST", "/grafana/search", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGrafanaSearch_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/grafana/search", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.GrafanaSearch(c)
	})

	req := httptest.NewRequest("POST", "/grafana/search", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestGrafanaQuery_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/grafana/query", h.GrafanaQuery)

	req := httptest.NewRequest("POST", "/grafana/query", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGrafanaQuery_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/grafana/query", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.GrafanaQuery(c)
	})

	req := httptest.NewRequest("POST", "/grafana/query", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestBuildGrafanaTable_UnknownTarget_ReturnsEmptyTable(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/x", func(c *fiber.Ctx) error {
		resp, err := h.buildGrafanaTable(c, "not-a-real-metric")
		assert.NoError(t, err)
		assert.Equal(t, "table", resp.Type)
		assert.Equal(t, []grafanaColumn{}, resp.Columns)
		assert.Equal(t, [][]interface{}{}, resp.Rows)
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/x", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}
