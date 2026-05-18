package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"ipam-next/models"
)

// setUserID injects a userID int64 into Locals (used by auth handlers instead of "user").
func authApp(h *Handler, routes func(*fiber.App)) *fiber.App {
	app := fiber.New()
	routes(app)
	return app
}

func authAppWithUser(h *Handler, u *models.User, routes func(*fiber.App)) *fiber.App {
	app := fiber.New()
	inner := fiber.New()
	routes(inner)
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user", u)
		if u != nil {
			c.Locals("userID", u.ID)
		}
		return c.Next()
	})
	_ = inner
	routes(app) // re-register on outer; middleware runs first via Use above
	return app
}

// simpleApp wraps a single handler, optionally injecting user/userID.
func simpleApp(h *Handler, method, path string, handler fiber.Handler, u *models.User) *fiber.App {
	app := fiber.New()
	if u != nil {
		app.Use(func(c *fiber.Ctx) error {
			c.Locals("user", u)
			c.Locals("userID", u.ID)
			return c.Next()
		})
	}
	switch method {
	case "GET":
		app.Get(path, handler)
	case "POST":
		app.Post(path, handler)
	case "DELETE":
		app.Delete(path, handler)
	}
	return app
}

// ---------------------------------------------------------------------------
// GetCurrentUser — GET /auth/me
// ---------------------------------------------------------------------------

func TestGetCurrentUser_NoUser_Returns401(t *testing.T) {
	h := &Handler{}
	app := simpleApp(h, "GET", "/auth/me", h.GetCurrentUser, nil)
	resp, err := app.Test(httptest.NewRequest("GET", "/auth/me", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetCurrentUser_WithUser_Returns200(t *testing.T) {
	h := &Handler{}
	app := simpleApp(h, "GET", "/auth/me", h.GetCurrentUser, adminUser)
	resp, err := app.Test(httptest.NewRequest("GET", "/auth/me", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// Login — POST /auth/login
// ---------------------------------------------------------------------------

func TestLogin_MissingBothFields_Returns400(t *testing.T) {
	h := &Handler{}
	app := simpleApp(h, "POST", "/auth/login", h.Login, nil)
	req := httptest.NewRequest("POST", "/auth/login", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestLogin_MissingPassword_Returns400(t *testing.T) {
	h := &Handler{}
	app := simpleApp(h, "POST", "/auth/login", h.Login, nil)
	req := httptest.NewRequest("POST", "/auth/login", strings.NewReader(`{"username":"alice"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestLogin_MissingUsername_Returns400(t *testing.T) {
	h := &Handler{}
	app := simpleApp(h, "POST", "/auth/login", h.Login, nil)
	req := httptest.NewRequest("POST", "/auth/login", strings.NewReader(`{"password":"secret"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestSetSessionCookie_ProductionHTTP_DoesNotForceSecure(t *testing.T) {
	t.Setenv("SESSION_COOKIE_SECURE", "")
	h := &Handler{isProduction: true}
	app := fiber.New()
	app.Get("/set-cookie", func(c *fiber.Ctx) error {
		h.setSessionCookie(c, "test-session")
		return c.SendStatus(fiber.StatusNoContent)
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/set-cookie", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)
	cookies := resp.Cookies()
	if assert.Len(t, cookies, 1) {
		assert.Equal(t, sessionCookieName, cookies[0].Name)
		assert.False(t, cookies[0].Secure)
	}
}

func TestSetSessionCookie_ProductionForwardedHTTPS_SetsSecure(t *testing.T) {
	t.Setenv("SESSION_COOKIE_SECURE", "")
	h := &Handler{isProduction: true}
	app := fiber.New()
	app.Get("/set-cookie", func(c *fiber.Ctx) error {
		h.setSessionCookie(c, "test-session")
		return c.SendStatus(fiber.StatusNoContent)
	})

	req := httptest.NewRequest("GET", "/set-cookie", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)
	cookies := resp.Cookies()
	if assert.Len(t, cookies, 1) {
		assert.Equal(t, sessionCookieName, cookies[0].Name)
		assert.True(t, cookies[0].Secure)
	}
}

func TestSetSessionCookie_SecureOverride(t *testing.T) {
	t.Setenv("SESSION_COOKIE_SECURE", "true")
	h := &Handler{isProduction: false}
	app := fiber.New()
	app.Get("/set-cookie", func(c *fiber.Ctx) error {
		h.setSessionCookie(c, "test-session")
		return c.SendStatus(fiber.StatusNoContent)
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/set-cookie", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)
	cookies := resp.Cookies()
	if assert.Len(t, cookies, 1) {
		assert.Equal(t, sessionCookieName, cookies[0].Name)
		assert.True(t, cookies[0].Secure)
	}
}

// ---------------------------------------------------------------------------
// Logout — POST /auth/logout
// ---------------------------------------------------------------------------

func TestLogout_NoUserID_Returns401(t *testing.T) {
	h := &Handler{}
	app := simpleApp(h, "POST", "/auth/logout", h.Logout, nil)
	resp, err := app.Test(httptest.NewRequest("POST", "/auth/logout", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestLogout_MissingAuthHeader_Returns400(t *testing.T) {
	h := &Handler{}
	app := simpleApp(h, "POST", "/auth/logout", h.Logout, adminUser)
	resp, err := app.Test(httptest.NewRequest("POST", "/auth/logout", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// ListMyTokens — GET /auth/me/tokens
// ---------------------------------------------------------------------------

func TestListMyTokens_NoUserID_Returns401(t *testing.T) {
	h := &Handler{}
	app := simpleApp(h, "GET", "/auth/me/tokens", h.ListMyTokens, nil)
	resp, err := app.Test(httptest.NewRequest("GET", "/auth/me/tokens", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// ListMySessions — GET /auth/me/sessions
// ---------------------------------------------------------------------------

func TestListMySessions_NoUserID_Returns401(t *testing.T) {
	h := &Handler{}
	app := simpleApp(h, "GET", "/auth/me/sessions", h.ListMySessions, nil)
	resp, err := app.Test(httptest.NewRequest("GET", "/auth/me/sessions", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// RevokeMySession — DELETE /auth/me/sessions/:sessionID
// ---------------------------------------------------------------------------

func TestRevokeMySession_NoUserID_Returns401(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Delete("/auth/me/sessions/:sessionID", h.RevokeMySession)
	resp, err := app.Test(httptest.NewRequest("DELETE", "/auth/me/sessions/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestRevokeMySession_BadSessionID_Returns400(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("userID", int64(1))
		return c.Next()
	})
	app.Delete("/auth/me/sessions/:sessionID", h.RevokeMySession)
	resp, err := app.Test(httptest.NewRequest("DELETE", "/auth/me/sessions/notanumber", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// LogoutAllDevices — DELETE /auth/me/sessions
// ---------------------------------------------------------------------------

func TestLogoutAllDevices_NoUserID_Returns401(t *testing.T) {
	h := &Handler{}
	app := simpleApp(h, "DELETE", "/auth/me/sessions", h.LogoutAllDevices, nil)
	resp, err := app.Test(httptest.NewRequest("DELETE", "/auth/me/sessions", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// RotateToken — POST /auth/me/tokens/:id/rotate
// ---------------------------------------------------------------------------

func TestRotateToken_NoUser_Returns401(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Post("/auth/me/tokens/:id/rotate", h.RotateToken)
	resp, err := app.Test(httptest.NewRequest("POST", "/auth/me/tokens/1/rotate", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestRotateToken_BadID_Returns400(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user", adminUser)
		return c.Next()
	})
	app.Post("/auth/me/tokens/:id/rotate", h.RotateToken)
	resp, err := app.Test(httptest.NewRequest("POST", "/auth/me/tokens/notanumber/rotate", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// ExtendToken — POST /auth/me/tokens/:id/extend
// ---------------------------------------------------------------------------

func TestExtendToken_NoUser_Returns401(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Post("/auth/me/tokens/:id/extend", h.ExtendToken)
	resp, err := app.Test(httptest.NewRequest("POST", "/auth/me/tokens/1/extend", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestExtendToken_BadID_Returns400(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user", adminUser)
		return c.Next()
	})
	app.Post("/auth/me/tokens/:id/extend", h.ExtendToken)
	resp, err := app.Test(httptest.NewRequest("POST", "/auth/me/tokens/notanumber/extend", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// GenerateTokenForMe — POST /auth/me/tokens
// ---------------------------------------------------------------------------

func TestGenerateTokenForMe_NoUserID_Returns401(t *testing.T) {
	h := &Handler{}
	app := simpleApp(h, "POST", "/auth/me/tokens", h.GenerateTokenForMe, nil)
	req := httptest.NewRequest("POST", "/auth/me/tokens", strings.NewReader(`{"token_name":"ci"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}
