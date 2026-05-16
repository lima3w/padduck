package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// Register — POST /auth/register
// ---------------------------------------------------------------------------

func TestRegister_MissingUsername_Returns400(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Post("/auth/register", h.Register)
	req := httptest.NewRequest("POST", "/auth/register", strings.NewReader(`{"email":"a@b.com","password":"secret123"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestRegister_MissingEmail_Returns400(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Post("/auth/register", h.Register)
	req := httptest.NewRequest("POST", "/auth/register", strings.NewReader(`{"username":"alice","password":"secret123"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestRegister_MissingPassword_Returns400(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Post("/auth/register", h.Register)
	req := httptest.NewRequest("POST", "/auth/register", strings.NewReader(`{"username":"alice","email":"a@b.com"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// VerifyEmail — GET /auth/verify-email?token=xxx
// ---------------------------------------------------------------------------

func TestVerifyEmail_MissingToken_Returns400(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Get("/auth/verify-email", h.VerifyEmail)
	resp, err := app.Test(httptest.NewRequest("GET", "/auth/verify-email", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// ResendVerification — POST /auth/resend-verification
// ---------------------------------------------------------------------------

func TestResendVerification_MissingEmail_Returns400(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Post("/auth/resend-verification", h.ResendVerification)
	req := httptest.NewRequest("POST", "/auth/resend-verification", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
