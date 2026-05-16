package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// RequestPasswordReset — POST /auth/request-password-reset
// ---------------------------------------------------------------------------

func TestRequestPasswordReset_MissingEmail_Returns400(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Post("/auth/request-password-reset", h.RequestPasswordReset)
	req := httptest.NewRequest("POST", "/auth/request-password-reset", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// ResetPassword — POST /auth/reset-password
// ---------------------------------------------------------------------------

func TestResetPassword_MissingFields_Returns400(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Post("/auth/reset-password", h.ResetPassword)
	req := httptest.NewRequest("POST", "/auth/reset-password", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestResetPassword_PasswordTooShort_Returns400(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Post("/auth/reset-password", h.ResetPassword)
	req := httptest.NewRequest("POST", "/auth/reset-password", strings.NewReader(`{"token":"abc","password":"short"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestResetPassword_MissingToken_Returns400(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Post("/auth/reset-password", h.ResetPassword)
	req := httptest.NewRequest("POST", "/auth/reset-password", strings.NewReader(`{"password":"longpassword"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
