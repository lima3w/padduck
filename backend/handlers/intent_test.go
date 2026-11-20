package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// Intent endpoints apply writes, so each one must enforce permissions itself.
func TestIntentEndpoints_Permissions(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	as := func(handler fiber.Handler) fiber.Handler {
		return func(c *fiber.Ctx) error {
			c.Locals("user", unprivNS)
			return handler(c)
		}
	}
	app.Get("/intents", as(h.ListIntents))
	app.Get("/intents/diff", as(h.GetIntentDiff))
	app.Post("/intents", as(h.SubmitIntent))
	app.Post("/intents/:id/approve", as(h.ApproveIntent))
	app.Post("/intents/:id/reject", as(h.RejectIntent))

	post := func(path, body string) int {
		req := httptest.NewRequest("POST", path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		assert.NoError(t, err)
		return resp.StatusCode
	}
	get := func(path string) int {
		resp, err := app.Test(httptest.NewRequest("GET", path, nil))
		assert.NoError(t, err)
		return resp.StatusCode
	}

	assert.Equal(t, fiber.StatusForbidden, get("/intents"), "cross-resource list needs admin")
	assert.Equal(t, fiber.StatusForbidden, get("/intents?resource_type=subnet&resource_id=1"), "history needs read permission")
	assert.Equal(t, fiber.StatusForbidden, get("/intents/diff?resource_type=subnet&resource_id=1"))
	assert.Equal(t, fiber.StatusForbidden, post("/intents", `{"resource_type":"subnet","operation":"create","desired_state":{}}`), "submit needs the write permission")
	assert.Equal(t, fiber.StatusForbidden, post("/intents/1/approve", `{}`))
	assert.Equal(t, fiber.StatusForbidden, post("/intents/1/reject", `{}`))

	assert.Equal(t, fiber.StatusBadRequest, post("/intents", `{"resource_type":"widget","operation":"create"}`), "unknown type is rejected before any permission lookup")
	assert.Equal(t, fiber.StatusBadRequest, get("/intents/diff"), "diff needs a resource")
}
