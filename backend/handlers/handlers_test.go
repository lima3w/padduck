package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"padduck/services"
)

func TestNewHandler(t *testing.T) {
	var svc *services.Service
	handler := NewHandler(svc, nil, true)

	assert.NotNil(t, handler)
	assert.Equal(t, svc, handler.service)
	assert.True(t, handler.isProduction)
}

func TestRegisterRoutes_AuthProvidersIsPublic(t *testing.T) {
	handler := NewHandler(nil, nil, false)
	app := fiber.New()
	handler.RegisterRoutes(app)

	resp, err := app.Test(httptest.NewRequest("GET", "/api/v1/auth/providers", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestRegisterRoutes_AuthMeStillRequiresCredentials(t *testing.T) {
	handler := NewHandler(nil, nil, false)
	app := fiber.New()
	handler.RegisterRoutes(app)

	resp, err := app.Test(httptest.NewRequest("GET", "/api/v1/auth/me", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}
