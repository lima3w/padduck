package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// GetCSRFToken — GET /csrf-token
// ---------------------------------------------------------------------------

func TestGetCSRFToken_Returns200WithToken(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Get("/csrf-token", h.GetCSRFToken)
	resp, err := app.Test(httptest.NewRequest("GET", "/csrf-token", nil))
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result map[string]string
	require.NoError(t, json.Unmarshal(body, &result))
	assert.NotEmpty(t, result["csrf_token"], "response must contain a non-empty csrf_token")
}

func TestGetCSRFToken_TokenIsHex64Chars(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Get("/csrf-token", h.GetCSRFToken)
	resp, err := app.Test(httptest.NewRequest("GET", "/csrf-token", nil))
	require.NoError(t, err)

	body, _ := io.ReadAll(resp.Body)
	var result map[string]string
	require.NoError(t, json.Unmarshal(body, &result))
	assert.Len(t, result["csrf_token"], 64, "token should be 32 random bytes hex-encoded (64 chars)")
}

// ---------------------------------------------------------------------------
// CSRFMiddleware
// ---------------------------------------------------------------------------

func TestCSRFMiddleware_GET_Passes(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Use(h.CSRFMiddleware)
	app.Get("/test", func(c *fiber.Ctx) error { return c.SendStatus(200) })

	resp, err := app.Test(httptest.NewRequest("GET", "/test", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestCSRFMiddleware_POST_WithoutToken_Returns403(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Use(h.CSRFMiddleware)
	app.Post("/test", func(c *fiber.Ctx) error { return c.SendStatus(200) })

	req := httptest.NewRequest("POST", "/test", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestCSRFMiddleware_POST_WithMatchingToken_Passes(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Use(h.CSRFMiddleware)
	app.Post("/test", func(c *fiber.Ctx) error { return c.SendStatus(200) })

	token := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
	req := httptest.NewRequest("POST", "/test", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(CSRFHeaderName, token)
	req.AddCookie(&http.Cookie{Name: CSRFCookieName, Value: token})
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestCSRFMiddleware_POST_TokenMismatch_Returns403(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Use(h.CSRFMiddleware)
	app.Post("/test", func(c *fiber.Ctx) error { return c.SendStatus(200) })

	req := httptest.NewRequest("POST", "/test", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(CSRFHeaderName, "wrong-token")
	req.AddCookie(&http.Cookie{Name: CSRFCookieName, Value: "correct-token"})
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// GenerateCSRFToken — pure function
// ---------------------------------------------------------------------------

func TestGenerateCSRFToken_ReturnsDifferentTokensEachCall(t *testing.T) {
	t1, err1 := GenerateCSRFToken()
	t2, err2 := GenerateCSRFToken()
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NotEqual(t, t1, t2, "two consecutive tokens must differ")
}
