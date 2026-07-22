package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"padduck/models"
)

// evaluatePolicy has no branch that avoids touching a nil service or nil
// ops.Automation — the bypass path calls h.auditLog (nil service) and the
// non-bypass path calls h.ops.Automation.Evaluate (nil ops.Automation).
// It is intentionally not tested here; covered by integration tests.

func TestToStateMap(t *testing.T) {
	type sample struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}
	got := toStateMap(sample{Name: "printer", Count: 3})
	assert.Equal(t, map[string]any{"name": "printer", "count": float64(3)}, got)
}

func TestToStateMap_Nil(t *testing.T) {
	// json.Marshal(nil) produces "null", and unmarshaling "null" into a map
	// leaves it nil (no error) — this differs from the marshal-error
	// fallback path exercised in TestToStateMap_Unmarshalable.
	got := toStateMap(nil)
	assert.Nil(t, got)
}

func TestToStateMap_Unmarshalable(t *testing.T) {
	// A channel can't be marshaled to JSON, so the function must fall back
	// to an empty map rather than panicking or returning nil.
	got := toStateMap(make(chan int))
	assert.Equal(t, map[string]any{}, got)
}

func TestBoolStr(t *testing.T) {
	assert.Equal(t, "true", boolStr(true))
	assert.Equal(t, "false", boolStr(false))
}

// ---------------------------------------------------------------------------
// SetAPITokenBypassPolicy — no inline guard; :id is parsed before the body,
// so both the invalid-ID and (valid-ID) invalid-body branches are reachable
// with no auth at all.
// ---------------------------------------------------------------------------

func TestSetAPITokenBypassPolicy_InvalidID_Returns400(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Put("/platform/api-tokens/:id/bypass-policy", h.SetAPITokenBypassPolicy)

	req := httptest.NewRequest("PUT", "/platform/api-tokens/not-a-number/bypass-policy", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestSetAPITokenBypassPolicy_ValidIDInvalidBody_Returns400(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Put("/platform/api-tokens/:id/bypass-policy", h.SetAPITokenBypassPolicy)

	req := httptest.NewRequest("PUT", "/platform/api-tokens/1/bypass-policy", strings.NewReader("{not valid json"))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// dispatchPolicyActions — the early-return branch (no matched policy, or a
// matched policy with no actions) never touches ops.Automation and is
// testable directly.
// ---------------------------------------------------------------------------

func TestDispatchPolicyActions_NoMatchedPolicy_DoesNotPanic(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/x", func(c *fiber.Ctx) error {
		assert.NotPanics(t, func() {
			h.dispatchPolicyActions(c, "device", 1, "device-1")
		})
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/x", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestDispatchPolicyActions_MatchedPolicyWithNoActions_DoesNotPanic(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/x", func(c *fiber.Ctx) error {
		c.Locals("matchedPolicy", &models.AutomationPolicy{Actions: nil})
		assert.NotPanics(t, func() {
			h.dispatchPolicyActions(c, "device", 1, "device-1")
		})
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/x", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}
