package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// VerifyMFA — POST /auth/verify-mfa
// No auth required (called with a challenge token before a session exists).
// ---------------------------------------------------------------------------

func TestVerifyMFA_MissingChallenge_Returns400(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Post("/auth/verify-mfa", h.VerifyMFA)
	req := httptest.NewRequest("POST", "/auth/verify-mfa", strings.NewReader(`{"code":"123456"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestVerifyMFA_MissingCode_Returns400(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Post("/auth/verify-mfa", h.VerifyMFA)
	req := httptest.NewRequest("POST", "/auth/verify-mfa", strings.NewReader(`{"mfa_challenge":"abc"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestVerifyMFA_EmptyBody_Returns400(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Post("/auth/verify-mfa", h.VerifyMFA)
	req := httptest.NewRequest("POST", "/auth/verify-mfa", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// ConfirmTOTP — POST /auth/me/mfa/confirm
// ---------------------------------------------------------------------------

func TestConfirmTOTP_MissingCode_Returns400(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user", adminUser)
		return c.Next()
	})
	app.Post("/auth/me/mfa/confirm", h.ConfirmTOTP)
	req := httptest.NewRequest("POST", "/auth/me/mfa/confirm", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// DisableTOTP — DELETE /auth/me/mfa
// ---------------------------------------------------------------------------

func TestDisableTOTP_MissingCode_Returns400(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user", adminUser)
		return c.Next()
	})
	app.Delete("/auth/me/mfa", h.DisableTOTP)
	req := httptest.NewRequest("DELETE", "/auth/me/mfa", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// RegenerateBackupCodes — POST /auth/me/mfa/backup-codes
// ---------------------------------------------------------------------------

func TestRegenerateBackupCodes_MissingCode_Returns400(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user", adminUser)
		return c.Next()
	})
	app.Post("/auth/me/mfa/backup-codes", h.RegenerateBackupCodes)
	req := httptest.NewRequest("POST", "/auth/me/mfa/backup-codes", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
