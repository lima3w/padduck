package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"ipam-next/models"
)

// adminUser is a user with the admin role — used to test that admin-only
// handlers would proceed past the auth/role check (they will fail later when
// the nil service is called, but for 401/403 tests that is fine).
var adminUser = &models.User{ID: 1, Role: "admin"}

// nonAdminUser is a regular user that should be denied admin-only endpoints.
var nonAdminUser = &models.User{ID: 2, Role: "user"}

// ---------------------------------------------------------------------------
// ListUsers — GET /api/v1/users and GET /api/v1/admin/users
// ---------------------------------------------------------------------------

func TestListUsers_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/users", h.ListUsers)

	resp, err := app.Test(httptest.NewRequest("GET", "/users", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestListUsers_NonAdmin_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/users", func(c *fiber.Ctx) error {
		c.Locals("user", nonAdminUser)
		return h.ListUsers(c)
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/users", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// CreateUser — POST /api/v1/users
// ---------------------------------------------------------------------------

func TestCreateUser_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/users", h.CreateUser)

	resp, err := app.Test(httptest.NewRequest("POST", "/users", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestCreateUser_NonAdmin_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/users", func(c *fiber.Ctx) error {
		c.Locals("user", nonAdminUser)
		return h.CreateUser(c)
	})

	body := strings.NewReader(`{"username":"new","email":"new@example.com","password":"secret"}`)
	req := httptest.NewRequest("POST", "/users", body)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestCreateUser_Admin_MissingFields_Returns400(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/users", func(c *fiber.Ctx) error {
		c.Locals("user", adminUser)
		return h.CreateUser(c)
	})

	body := strings.NewReader(`{"username":"new"}`)
	req := httptest.NewRequest("POST", "/users", body)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestCreateUser_Admin_InvalidRole_Returns400(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/users", func(c *fiber.Ctx) error {
		c.Locals("user", adminUser)
		return h.CreateUser(c)
	})

	body := strings.NewReader(`{"username":"new","email":"new@example.com","password":"secret","role":"superadmin"}`)
	req := httptest.NewRequest("POST", "/users", body)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
