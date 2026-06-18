package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// RequestUnlock — POST /auth/unlock
// Always returns 200 to prevent username enumeration; 400 only on bad body.
// ---------------------------------------------------------------------------

func TestRequestUnlock_EmptyUsername_Returns200(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/auth/unlock", h.RequestUnlock)
	req := httptest.NewRequest("POST", "/auth/unlock", strings.NewReader(`{"username":""}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	// Empty username: skips email send, still returns 200 (enumeration-safe)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestRequestUnlock_InvalidBody_Returns400(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/auth/unlock", h.RequestUnlock)
	req := httptest.NewRequest("POST", "/auth/unlock", strings.NewReader(`not json`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// VerifyUnlock — GET /auth/unlock?token=xxx
// ---------------------------------------------------------------------------

func TestVerifyUnlock_MissingToken_Returns400(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/auth/unlock", h.VerifyUnlock)
	resp, err := app.Test(httptest.NewRequest("GET", "/auth/unlock", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// GetLoginHistory — GET /user/login-history
// ---------------------------------------------------------------------------

func TestGetLoginHistory_NoUserID_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/user/login-history", h.GetLoginHistory)
	resp, err := app.Test(httptest.NewRequest("GET", "/user/login-history", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// AdminUnlockUser — POST /admin/users/:id/unlock
// ---------------------------------------------------------------------------

func TestAdminUnlockUser_NoUserID_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/admin/users/:id/unlock", h.AdminUnlockUser)
	resp, err := app.Test(httptest.NewRequest("POST", "/admin/users/1/unlock", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestAdminUnlockUser_BadID_Returns400(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("userID", int64(1))
		return c.Next()
	})
	app.Post("/admin/users/:id/unlock", h.AdminUnlockUser)
	resp, err := app.Test(httptest.NewRequest("POST", "/admin/users/notanumber/unlock", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
