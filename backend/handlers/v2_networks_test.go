package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"padduck/models"
)

func TestV2ListNetworks_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/api/v2/networks", h.V2ListNetworks)

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/api/v2/networks", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestV2ListNetworks_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/api/v2/networks", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{ID: 0, Role: "viewer"})
		return h.V2ListNetworks(c)
	})

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/api/v2/networks", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestV2List_EnvelopeShape(t *testing.T) {
	result := V2List([]string{"a", "b"}, V2Meta{Page: 1, Limit: 25, Total: 2})
	assert.Equal(t, []string{"a", "b"}, result["data"])
	meta, ok := result["meta"].(V2Meta)
	assert.True(t, ok)
	assert.Equal(t, 1, meta.Page)
	assert.Equal(t, 25, meta.Limit)
	assert.Equal(t, int64(2), meta.Total)
}

func TestV2Item_EnvelopeShape(t *testing.T) {
	result := V2Item(fiber.Map{"id": 1})
	assert.NotNil(t, result["data"])
}

func TestListNetworks_DeprecationHeader(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/api/v1/networks", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{ID: 0, Role: "viewer"})
		return h.ListNetworks(c)
	})

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/api/v1/networks", nil))
	assert.NoError(t, err)
	// Deprecation header is set regardless of auth outcome (set before perm check returns)
	// For a viewer with ID=0, we get 403 but the header is already written.
	assert.Equal(t, "true", resp.Header.Get("Deprecation"))
	assert.Contains(t, resp.Header.Get("Link"), "/api/v2/networks")
}
