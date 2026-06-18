package handlers

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"padduck/services"
)

func TestGetPublicInfoIncludesRegistrationEnabled(t *testing.T) {
	handler := NewHandler(&services.Service{Config: services.NewConfigService(nil)}, nil, nil, false)
	app := fiber.New()
	app.Get("/public-info", handler.GetPublicInfo)

	resp, err := app.Test(httptest.NewRequest("GET", "/public-info", nil))
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var body map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.Equal(t, true, body["registration_enabled"])
}
