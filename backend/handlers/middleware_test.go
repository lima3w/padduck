package handlers

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// newMiddlewareTestApp builds a minimal Fiber app that applies AuthMiddleware
// and serves a simple success response on GET /test.
func newMiddlewareTestApp() *fiber.App {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Use(h.AuthMiddleware)
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})
	return app
}

// parseMiddlewareErrorResponse reads the response body into a map.
func parseMiddlewareErrorResponse(t *testing.T, body io.Reader) map[string]interface{} {
	t.Helper()
	var result map[string]interface{}
	data, err := io.ReadAll(body)
	assert.NoError(t, err)
	err = json.Unmarshal(data, &result)
	assert.NoError(t, err)
	return result
}

// TestAuthMiddleware_NoAuthorizationHeader verifies a 401 when no header is provided.
func TestAuthMiddleware_NoAuthorizationHeader(t *testing.T) {
	app := newMiddlewareTestApp()

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)

	body := parseMiddlewareErrorResponse(t, resp.Body)
	assert.Equal(t, "UNAUTHORIZED", body["code"])
}

// TestAuthMiddleware_WrongScheme verifies a 401 when the scheme is not "Bearer".
func TestAuthMiddleware_WrongScheme(t *testing.T) {
	app := newMiddlewareTestApp()

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Basic abc123")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)

	body := parseMiddlewareErrorResponse(t, resp.Body)
	assert.Equal(t, "UNAUTHORIZED", body["code"])
}

// TestAuthMiddleware_MalformedScheme verifies a 401 when the scheme word is not "Bearer".
func TestAuthMiddleware_MalformedScheme(t *testing.T) {
	app := newMiddlewareTestApp()

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "NotBearer sometoken")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)

	body := parseMiddlewareErrorResponse(t, resp.Body)
	assert.Equal(t, "UNAUTHORIZED", body["code"])
}

// TestAuthMiddleware_TokenSchemeOnly verifies a 401 when the scheme word is "Token" (wrong).
func TestAuthMiddleware_TokenSchemeOnly(t *testing.T) {
	app := newMiddlewareTestApp()

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Token abc")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)

	body := parseMiddlewareErrorResponse(t, resp.Body)
	assert.Equal(t, "UNAUTHORIZED", body["code"])
}

// TestAuthMiddleware_BearerAloneNoSpace verifies a 401 when the header is just "Bearer"
// with no trailing space or token (SplitN produces only one part).
func TestAuthMiddleware_BearerAloneNoToken(t *testing.T) {
	app := newMiddlewareTestApp()

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)

	body := parseMiddlewareErrorResponse(t, resp.Body)
	assert.Equal(t, "UNAUTHORIZED", body["code"])
}
