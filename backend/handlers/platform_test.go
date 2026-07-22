package handlers

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// ListPlatformOrganizations and GetPlatformAuditLog have no inline guard and
// no local validation before their first service call, so calling them at
// all with a nil service panics — there is no branch left to exercise
// without a live repo/service (see plan). They are intentionally not
// tested here; covered by integration tests instead.

// ---------------------------------------------------------------------------
// GetPlatformOrganization — no inline guard; :id is parsed unconditionally,
// so the invalid-ID branch is reachable with no auth at all.
// ---------------------------------------------------------------------------

func TestGetPlatformOrganization_InvalidID_Returns400(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/platform/organizations/:id", h.GetPlatformOrganization)

	req := httptest.NewRequest("GET", "/platform/organizations/not-a-number", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// PlatformImpersonate — no inline guard; body parsing and the
// organization_id validation both run before any service call, so both
// failure branches are reachable with no auth at all.
// ---------------------------------------------------------------------------

func TestPlatformImpersonate_InvalidBody_Returns400(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/platform/impersonate", h.PlatformImpersonate)

	req := httptest.NewRequest("POST", "/platform/impersonate", strings.NewReader("{not valid json"))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestPlatformImpersonate_ZeroOrganizationID_Returns422(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/platform/impersonate", h.PlatformImpersonate)

	req := httptest.NewRequest("POST", "/platform/impersonate", strings.NewReader(`{"organization_id":0}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnprocessableEntity, resp.StatusCode)

	data, _ := io.ReadAll(resp.Body)
	var body ErrorResponse
	assert.NoError(t, json.Unmarshal(data, &body))
	assert.Equal(t, string(ErrValidation), body.Code)
	assert.Len(t, body.Fields, 1)
	assert.Equal(t, "organization_id", body.Fields[0].Field)
}

func TestPlatformImpersonate_NegativeOrganizationID_Returns422(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/platform/impersonate", h.PlatformImpersonate)

	req := httptest.NewRequest("POST", "/platform/impersonate", strings.NewReader(`{"organization_id":-5}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnprocessableEntity, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// SetPlatformAdmin — no inline guard; :id is parsed before the body, so
// both the invalid-ID and (valid-ID) invalid-body branches are reachable
// with no auth at all.
// ---------------------------------------------------------------------------

func TestSetPlatformAdmin_InvalidID_Returns400(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Put("/platform/users/:id/platform-admin", h.SetPlatformAdmin)

	req := httptest.NewRequest("PUT", "/platform/users/not-a-number/platform-admin", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestSetPlatformAdmin_ValidIDInvalidBody_Returns400(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Put("/platform/users/:id/platform-admin", h.SetPlatformAdmin)

	req := httptest.NewRequest("PUT", "/platform/users/1/platform-admin", strings.NewReader("{not valid json"))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
