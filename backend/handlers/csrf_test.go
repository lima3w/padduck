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
	h := minHandler()
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

func TestGetCSRFToken_TokenHasSignedFormat(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/csrf-token", h.GetCSRFToken)
	resp, err := app.Test(httptest.NewRequest("GET", "/csrf-token", nil))
	require.NoError(t, err)

	body, _ := io.ReadAll(resp.Body)
	var result map[string]string
	require.NoError(t, json.Unmarshal(body, &result))
	token := result["csrf_token"]
	// Signed format: "<64-hex-random>.<32-hex-MAC>" = 97 chars with one dot separator.
	assert.Len(t, token, 97, "token should be 64-char random hex + '.' + 32-char HMAC hex")
	assert.Contains(t, token, ".", "token must contain the signed-double-submit separator")
}

func TestGetCSRFToken_ProductionHTTP_DoesNotForceSecureCookie(t *testing.T) {
	t.Setenv("SESSION_COOKIE_SECURE", "")
	h := &Handler{isProduction: true}
	app := fiber.New()
	app.Get("/csrf-token", h.GetCSRFToken)

	resp, err := app.Test(httptest.NewRequest("GET", "/csrf-token", nil))
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var csrfCookie *http.Cookie
	for _, ck := range resp.Cookies() {
		if ck.Name == CSRFCookieName {
			csrfCookie = ck
		}
	}
	require.NotNil(t, csrfCookie)
	assert.False(t, csrfCookie.Secure)
}

func TestGetCSRFToken_ProductionForwardedHTTPS_SetsSecureCookie(t *testing.T) {
	t.Setenv("SESSION_COOKIE_SECURE", "")
	h := &Handler{isProduction: true}
	app := fiber.New()
	app.Get("/csrf-token", h.GetCSRFToken)

	req := httptest.NewRequest("GET", "/csrf-token", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var csrfCookie *http.Cookie
	for _, ck := range resp.Cookies() {
		if ck.Name == CSRFCookieName {
			csrfCookie = ck
		}
	}
	require.NotNil(t, csrfCookie)
	assert.True(t, csrfCookie.Secure)
}

// ---------------------------------------------------------------------------
// CSRFMiddleware
// ---------------------------------------------------------------------------

func TestCSRFMiddleware_GET_Passes(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Use(h.CSRFMiddleware)
	app.Get("/test", func(c *fiber.Ctx) error { return c.SendStatus(200) })

	resp, err := app.Test(httptest.NewRequest("GET", "/test", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestCSRFMiddleware_GET_IssuesCookieWhenAbsent(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Use(h.CSRFMiddleware)
	app.Get("/test", func(c *fiber.Ctx) error { return c.SendStatus(200) })

	resp, err := app.Test(httptest.NewRequest("GET", "/test", nil))
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	var csrfCookie *http.Cookie
	for _, ck := range resp.Cookies() {
		if ck.Name == CSRFCookieName {
			csrfCookie = ck
		}
	}
	require.NotNil(t, csrfCookie, "GET should issue a CSRF cookie when none is present")
	assert.Len(t, csrfCookie.Value, 97)
}

func TestCSRFMiddleware_GET_ProductionHTTP_DoesNotForceSecureCookie(t *testing.T) {
	t.Setenv("SESSION_COOKIE_SECURE", "")
	h := &Handler{isProduction: true}
	app := fiber.New()
	app.Use(h.CSRFMiddleware)
	app.Get("/test", func(c *fiber.Ctx) error { return c.SendStatus(200) })

	resp, err := app.Test(httptest.NewRequest("GET", "/test", nil))
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	var csrfCookie *http.Cookie
	for _, ck := range resp.Cookies() {
		if ck.Name == CSRFCookieName {
			csrfCookie = ck
		}
	}
	require.NotNil(t, csrfCookie)
	assert.False(t, csrfCookie.Secure)
}

func TestCSRFMiddleware_GET_DoesNotRotateCookieWhenPresent(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Use(h.CSRFMiddleware)
	app.Get("/test", func(c *fiber.Ctx) error { return c.SendStatus(200) })

	existing, err := h.newCSRFToken()
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{Name: CSRFCookieName, Value: existing})
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	// No Set-Cookie header should rotate the token.
	for _, ck := range resp.Cookies() {
		if ck.Name == CSRFCookieName {
			assert.Equal(t, existing, ck.Value, "existing CSRF cookie must not be rotated on GET")
		}
	}
}

func TestCSRFMiddleware_POST_WithoutToken_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Use(h.CSRFMiddleware)
	app.Post("/test", func(c *fiber.Ctx) error { return c.SendStatus(200) })

	req := httptest.NewRequest("POST", "/test", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestCSRFMiddleware_POST_WithMatchingSignedToken_Passes(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Use(h.CSRFMiddleware)
	app.Post("/test", func(c *fiber.Ctx) error { return c.SendStatus(200) })

	token, err := h.newCSRFToken()
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/test", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(CSRFHeaderName, token)
	req.AddCookie(&http.Cookie{Name: CSRFCookieName, Value: token})
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestCSRFMiddleware_POST_UnsignedToken_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Use(h.CSRFMiddleware)
	app.Post("/test", func(c *fiber.Ctx) error { return c.SendStatus(200) })

	// A plain random token with no HMAC (old format) must be rejected.
	unsignedToken := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
	req := httptest.NewRequest("POST", "/test", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(CSRFHeaderName, unsignedToken)
	req.AddCookie(&http.Cookie{Name: CSRFCookieName, Value: unsignedToken})
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestCSRFMiddleware_POST_TokenMismatch_Returns403(t *testing.T) {
	h := minHandler()
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
// newCSRFToken — signed token generation
// ---------------------------------------------------------------------------

func TestNewCSRFToken_ReturnsDifferentTokensEachCall(t *testing.T) {
	h := minHandler()
	t1, err1 := h.newCSRFToken()
	t2, err2 := h.newCSRFToken()
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NotEqual(t, t1, t2, "two consecutive tokens must differ")
}

func TestNewCSRFToken_ValidatesCorrectly(t *testing.T) {
	h := minHandler()
	token, err := h.newCSRFToken()
	require.NoError(t, err)
	assert.True(t, h.validCSRFToken(token, token), "a freshly generated token must pass its own validation")
}

func TestNewCSRFToken_TamperedMACFails(t *testing.T) {
	h := minHandler()
	token, err := h.newCSRFToken()
	require.NoError(t, err)
	// Flip the last character of the MAC to simulate tampering.
	tampered := token[:len(token)-1] + "x"
	assert.False(t, h.validCSRFToken(tampered, tampered), "a token with a tampered MAC must fail validation")
}
