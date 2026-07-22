package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// All four handlers check requirePerm before touching the :id param, so the
// invalid-ID branch is only reachable by a permitted user — which requires
// a live repo (see plan). Only the auth guard branches are testable here
// without a DB.

func TestListJobs_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/jobs", h.ListJobs)

	req := httptest.NewRequest("GET", "/admin/jobs", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestListJobs_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/jobs", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.ListJobs(c)
	})

	req := httptest.NewRequest("GET", "/admin/jobs", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestGetJob_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/jobs/:id", h.GetJob)

	req := httptest.NewRequest("GET", "/admin/jobs/1", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetJob_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/jobs/:id", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.GetJob(c)
	})

	req := httptest.NewRequest("GET", "/admin/jobs/1", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestCancelJob_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/admin/jobs/:id/cancel", h.CancelJob)

	req := httptest.NewRequest("POST", "/admin/jobs/1/cancel", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestCancelJob_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/admin/jobs/:id/cancel", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.CancelJob(c)
	})

	req := httptest.NewRequest("POST", "/admin/jobs/1/cancel", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestRetryJob_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/admin/jobs/:id/retry", h.RetryJob)

	req := httptest.NewRequest("POST", "/admin/jobs/1/retry", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestRetryJob_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/admin/jobs/:id/retry", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.RetryJob(c)
	})

	req := httptest.NewRequest("POST", "/admin/jobs/1/retry", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}
