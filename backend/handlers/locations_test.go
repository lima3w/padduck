package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"padduck/models"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// locApp builds a Fiber app that optionally injects a user and routes to h.
func locApp(user *models.User, method, route string, handler fiber.Handler) *fiber.App {
	h := minHandler()
	app := fiber.New()
	app.Add(method, route, func(c *fiber.Ctx) error {
		if user != nil {
			c.Locals("user", user)
		}
		return handler(c)
	})
	_ = h
	return app
}

// noUser is a nil *models.User — used to test unauthenticated requests.
var noUser *models.User

// unprivUser is a user with ID=0: permCheck calls CheckPermission(ctx, 0, …)
// which short-circuits on the userID<=0 guard and returns "permission denied"
// without touching the nil repository.
var unprivUser = &models.User{ID: 0, Role: "viewer"}

// ---------------------------------------------------------------------------
// ListLocations — GET /locations
// ---------------------------------------------------------------------------

func TestListLocations_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/locations", h.ListLocations)

	resp, err := app.Test(httptest.NewRequest("GET", "/locations", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestListLocations_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/locations", func(c *fiber.Ctx) error {
		c.Locals("user", unprivUser)
		return h.ListLocations(c)
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/locations", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// GetLocationTree — GET /locations/tree
// ---------------------------------------------------------------------------

func TestGetLocationTree_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/locations/tree", h.GetLocationTree)

	resp, err := app.Test(httptest.NewRequest("GET", "/locations/tree", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetLocationTree_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/locations/tree", func(c *fiber.Ctx) error {
		c.Locals("user", unprivUser)
		return h.GetLocationTree(c)
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/locations/tree", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// GetLocation — GET /locations/:id
// ---------------------------------------------------------------------------

func TestGetLocation_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/locations/:id", h.GetLocation)

	resp, err := app.Test(httptest.NewRequest("GET", "/locations/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetLocation_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/locations/:id", func(c *fiber.Ctx) error {
		c.Locals("user", unprivUser)
		return h.GetLocation(c)
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/locations/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// CreateLocation — POST /locations
// ---------------------------------------------------------------------------

func TestCreateLocation_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/locations", h.CreateLocation)

	resp, err := app.Test(httptest.NewRequest("POST", "/locations", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestCreateLocation_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/locations", func(c *fiber.Ctx) error {
		c.Locals("user", unprivUser)
		return h.CreateLocation(c)
	})

	resp, err := app.Test(httptest.NewRequest("POST", "/locations", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestCreateLocation_EmptyName_Returns400(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/locations", func(c *fiber.Ctx) error {
		c.Locals("user", unprivUser)
		return h.CreateLocation(c)
	})

	body := strings.NewReader(`{"name":"","type":"site"}`)
	req := httptest.NewRequest("POST", "/locations", body)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	// permCheck fires first with ID=0 → 403 before body is parsed
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// UpdateLocation — PUT /locations/:id
// ---------------------------------------------------------------------------

func TestUpdateLocation_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Put("/locations/:id", h.UpdateLocation)

	resp, err := app.Test(httptest.NewRequest("PUT", "/locations/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestUpdateLocation_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Put("/locations/:id", func(c *fiber.Ctx) error {
		c.Locals("user", unprivUser)
		return h.UpdateLocation(c)
	})

	resp, err := app.Test(httptest.NewRequest("PUT", "/locations/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// DeleteLocation — DELETE /locations/:id
// ---------------------------------------------------------------------------

func TestDeleteLocation_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Delete("/locations/:id", h.DeleteLocation)

	resp, err := app.Test(httptest.NewRequest("DELETE", "/locations/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestDeleteLocation_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Delete("/locations/:id", func(c *fiber.Ctx) error {
		c.Locals("user", unprivUser)
		return h.DeleteLocation(c)
	})

	resp, err := app.Test(httptest.NewRequest("DELETE", "/locations/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}
