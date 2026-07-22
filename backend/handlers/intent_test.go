package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// ListIntents has no inline guard and no local validation before its first
// service call, so calling it at all with a nil service panics — there is
// no branch left to exercise without a live repo/service (see plan). It is
// intentionally not tested here; covered by integration tests instead.

// ---------------------------------------------------------------------------
// GetIntent — no inline guard; :id is parsed unconditionally, so the
// invalid-ID branch is reachable with no auth at all.
// ---------------------------------------------------------------------------

func TestGetIntent_InvalidID_Returns400(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/intents/:id", h.GetIntent)

	req := httptest.NewRequest("GET", "/intents/not-a-number", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// SubmitIntent — no inline guard; the body is parsed unconditionally, so
// the invalid-body branch is reachable with no auth at all.
// ---------------------------------------------------------------------------

func TestSubmitIntent_InvalidBody_Returns400(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/intents", h.SubmitIntent)

	req := httptest.NewRequest("POST", "/intents", strings.NewReader("{not valid json"))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// ApproveIntent / RejectIntent — no inline guard; :id is parsed
// unconditionally, so the invalid-ID branch is reachable with no auth.
// The body is parsed leniently (errors ignored), so it has no failure
// branch to test.
// ---------------------------------------------------------------------------

func TestApproveIntent_InvalidID_Returns400(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/intents/:id/approve", h.ApproveIntent)

	req := httptest.NewRequest("POST", "/intents/not-a-number/approve", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestRejectIntent_InvalidID_Returns400(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/intents/:id/reject", h.RejectIntent)

	req := httptest.NewRequest("POST", "/intents/not-a-number/reject", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
