package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"padduck/services"
)

// ---------------------------------------------------------------------------
// Pure helper functions — no service/repo dependency, fully testable.
// ---------------------------------------------------------------------------

func TestRequiredFields(t *testing.T) {
	got := requiredFields(map[string]string{"a": "", "b": "  ", "c": "value"})
	assert.Len(t, got, 2)
	fields := map[string]bool{}
	for _, f := range got {
		fields[f.Field] = true
		assert.Equal(t, f.Field+" is required", f.Message)
	}
	assert.True(t, fields["a"])
	assert.True(t, fields["b"])
	assert.False(t, fields["c"])
}

func TestRequiredFields_AllPresent(t *testing.T) {
	got := requiredFields(map[string]string{"a": "x", "b": "y"})
	assert.Empty(t, got)
}

func TestValidateAutomationPolicyRequest_MissingName(t *testing.T) {
	fields := validateAutomationPolicyRequest(&automationPolicyRequest{Name: "  "})
	assert.Len(t, fields, 1)
	assert.Equal(t, "name", fields[0].Field)
}

func TestValidateAutomationPolicyRequest_InvalidEffect(t *testing.T) {
	fields := validateAutomationPolicyRequest(&automationPolicyRequest{Name: "policy-1", Effect: "banana"})
	assert.Len(t, fields, 1)
	assert.Equal(t, "effect", fields[0].Field)
}

func TestValidateAutomationPolicyRequest_ValidEffects(t *testing.T) {
	for _, effect := range []string{"", "allow", "deny", "manual_review"} {
		fields := validateAutomationPolicyRequest(&automationPolicyRequest{Name: "policy-1", Effect: effect})
		assert.Empty(t, fields, "effect %q should be valid", effect)
	}
}

func TestValidateAutomationPolicyRequest_Valid(t *testing.T) {
	fields := validateAutomationPolicyRequest(&automationPolicyRequest{Name: "policy-1", Effect: "allow"})
	assert.Empty(t, fields)
}

func TestFormatOptionalTime_Nil(t *testing.T) {
	assert.Nil(t, formatOptionalTime(nil))
}

func TestFormatOptionalTime_Set(t *testing.T) {
	ts, err := time.Parse(time.RFC3339, "2026-01-02T15:04:05Z")
	assert.NoError(t, err)
	got := formatOptionalTime(&ts)
	assert.NotNil(t, got)
	assert.Equal(t, "2026-01-02T15:04:05Z", *got)
}

func TestAutomationWriteResponse_Error(t *testing.T) {
	app := fiber.New()
	app.Get("/x", func(c *fiber.Ctx) error {
		return automationWriteResponse(c, nil, nil, errors.New("boom"))
	})
	req := httptest.NewRequest("GET", "/x", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestAutomationWriteResponse_ReviewNeeded(t *testing.T) {
	app := fiber.New()
	app.Get("/x", func(c *fiber.Ctx) error {
		return automationWriteResponse(c, nil, &services.PolicyDecision{ReviewNeeded: true}, nil)
	})
	req := httptest.NewRequest("GET", "/x", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusAccepted, resp.StatusCode)
}

func TestAutomationWriteResponse_Denied(t *testing.T) {
	app := fiber.New()
	app.Get("/x", func(c *fiber.Ctx) error {
		return automationWriteResponse(c, nil, &services.PolicyDecision{Allowed: false}, nil)
	})
	req := httptest.NewRequest("GET", "/x", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestAutomationWriteResponse_AllowedNoData(t *testing.T) {
	app := fiber.New()
	app.Get("/x", func(c *fiber.Ctx) error {
		return automationWriteResponse(c, nil, &services.PolicyDecision{Allowed: true}, nil)
	})
	req := httptest.NewRequest("GET", "/x", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestAutomationWriteResponse_AllowedWithData(t *testing.T) {
	app := fiber.New()
	app.Get("/x", func(c *fiber.Ctx) error {
		return automationWriteResponse(c, map[string]string{"id": "1"}, &services.PolicyDecision{Allowed: true}, nil)
	})
	req := httptest.NewRequest("GET", "/x", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// ListIntegrationTemplates — no guard, no service dependency at all.
// ---------------------------------------------------------------------------

func TestListIntegrationTemplates_ReturnsStaticList(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/automation/integration-templates", h.ListIntegrationTemplates)

	req := httptest.NewRequest("GET", "/automation/integration-templates", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	data, _ := io.ReadAll(resp.Body)
	var templates []map[string]interface{}
	assert.NoError(t, json.Unmarshal(data, &templates))
	assert.NotEmpty(t, templates)
}

// ---------------------------------------------------------------------------
// ListAutomationPolicies / ListAPITokenAnalytics — requirePerm guard only.
// ---------------------------------------------------------------------------

func TestListAutomationPolicies_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/automation/policies", h.ListAutomationPolicies)

	req := httptest.NewRequest("GET", "/admin/automation/policies", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestListAutomationPolicies_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/automation/policies", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.ListAutomationPolicies(c)
	})

	req := httptest.NewRequest("GET", "/admin/automation/policies", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestListAPITokenAnalytics_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/api-tokens/analytics", h.ListAPITokenAnalytics)

	req := httptest.NewRequest("GET", "/admin/api-tokens/analytics", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestListAPITokenAnalytics_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/api-tokens/analytics", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.ListAPITokenAnalytics(c)
	})

	req := httptest.NewRequest("GET", "/admin/api-tokens/analytics", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// CreateAutomationPolicy / UpdateAutomationPolicy / DeleteAutomationPolicy —
// requirePerm-gated. UpdateAutomationPolicy parses :id before delegating,
// so its invalid-ID branch is reachable with no auth; everything else in
// this trio requires a permitted user, which requires a live repo.
// ---------------------------------------------------------------------------

func TestCreateAutomationPolicy_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/admin/automation/policies", h.CreateAutomationPolicy)

	req := httptest.NewRequest("POST", "/admin/automation/policies", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestCreateAutomationPolicy_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/admin/automation/policies", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.CreateAutomationPolicy(c)
	})

	req := httptest.NewRequest("POST", "/admin/automation/policies", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestUpdateAutomationPolicy_InvalidID_Returns400(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Put("/admin/automation/policies/:id", h.UpdateAutomationPolicy)

	req := httptest.NewRequest("PUT", "/admin/automation/policies/not-a-number", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestUpdateAutomationPolicy_ValidIDNoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Put("/admin/automation/policies/:id", h.UpdateAutomationPolicy)

	req := httptest.NewRequest("PUT", "/admin/automation/policies/1", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestDeleteAutomationPolicy_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Delete("/admin/automation/policies/:id", h.DeleteAutomationPolicy)

	req := httptest.NewRequest("DELETE", "/admin/automation/policies/1", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestDeleteAutomationPolicy_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Delete("/admin/automation/policies/:id", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.DeleteAutomationPolicy(c)
	})

	req := httptest.NewRequest("DELETE", "/admin/automation/policies/1", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// EvaluateAutomationPolicy / SimulateAutomation / AutomationAllocateIPAddress
// / AutomationReserveIPAddress / AutomationRegisterDevice / AutomationDNSUpdate
// — none of these have an inline guard; body parsing and field validation
// both run before any service call, so both failure branches are reachable
// with no auth at all.
// ---------------------------------------------------------------------------

func TestEvaluateAutomationPolicy_InvalidBody_Returns400(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/automation/policies/evaluate", h.EvaluateAutomationPolicy)

	req := httptest.NewRequest("POST", "/automation/policies/evaluate", strings.NewReader("{not valid json"))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestEvaluateAutomationPolicy_MissingRequiredFields_Returns422(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/automation/policies/evaluate", h.EvaluateAutomationPolicy)

	req := httptest.NewRequest("POST", "/automation/policies/evaluate", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnprocessableEntity, resp.StatusCode)

	data, _ := io.ReadAll(resp.Body)
	var body ErrorResponse
	assert.NoError(t, json.Unmarshal(data, &body))
	assert.Len(t, body.Fields, 2)
}

func TestSimulateAutomation_InvalidBody_Returns400(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/automation/simulate", h.SimulateAutomation)

	req := httptest.NewRequest("POST", "/automation/simulate", strings.NewReader("{not valid json"))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestSimulateAutomation_MissingRequiredFields_Returns422(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/automation/simulate", h.SimulateAutomation)

	req := httptest.NewRequest("POST", "/automation/simulate", strings.NewReader(`{"workflow":"ip"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnprocessableEntity, resp.StatusCode)

	data, _ := io.ReadAll(resp.Body)
	var body ErrorResponse
	assert.NoError(t, json.Unmarshal(data, &body))
	assert.Len(t, body.Fields, 1)
	assert.Equal(t, "action", body.Fields[0].Field)
}

func TestAutomationAllocateIPAddress_InvalidBody_Returns400(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/automation/ip-addresses/allocate", h.AutomationAllocateIPAddress)

	req := httptest.NewRequest("POST", "/automation/ip-addresses/allocate", strings.NewReader("{not valid json"))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestAutomationAllocateIPAddress_MissingSubnetID_Returns422(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/automation/ip-addresses/allocate", h.AutomationAllocateIPAddress)

	req := httptest.NewRequest("POST", "/automation/ip-addresses/allocate", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnprocessableEntity, resp.StatusCode)

	data, _ := io.ReadAll(resp.Body)
	var body ErrorResponse
	assert.NoError(t, json.Unmarshal(data, &body))
	assert.Len(t, body.Fields, 1)
	assert.Equal(t, "subnet_id", body.Fields[0].Field)
}

func TestAutomationAllocateIPAddress_NegativeSubnetID_Returns422(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/automation/ip-addresses/allocate", h.AutomationAllocateIPAddress)

	req := httptest.NewRequest("POST", "/automation/ip-addresses/allocate", strings.NewReader(`{"subnet_id":-1}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnprocessableEntity, resp.StatusCode)
}

func TestAutomationReserveIPAddress_InvalidBody_Returns400(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/automation/ip-addresses/reserve", h.AutomationReserveIPAddress)

	req := httptest.NewRequest("POST", "/automation/ip-addresses/reserve", strings.NewReader("{not valid json"))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestAutomationReserveIPAddress_MissingFields_Returns422(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/automation/ip-addresses/reserve", h.AutomationReserveIPAddress)

	req := httptest.NewRequest("POST", "/automation/ip-addresses/reserve", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnprocessableEntity, resp.StatusCode)

	data, _ := io.ReadAll(resp.Body)
	var body ErrorResponse
	assert.NoError(t, json.Unmarshal(data, &body))
	fields := map[string]bool{}
	for _, f := range body.Fields {
		fields[f.Field] = true
	}
	assert.True(t, fields["subnet_id"])
	assert.True(t, fields["address"])
}

func TestAutomationReleaseIPAddress_InvalidID_Returns400(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/automation/ip-addresses/:id/release", h.AutomationReleaseIPAddress)

	req := httptest.NewRequest("POST", "/automation/ip-addresses/not-a-number/release", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestAutomationRegisterDevice_InvalidBody_Returns400(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/automation/devices/register", h.AutomationRegisterDevice)

	req := httptest.NewRequest("POST", "/automation/devices/register", strings.NewReader("{not valid json"))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestAutomationRegisterDevice_MissingHostname_Returns422(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/automation/devices/register", h.AutomationRegisterDevice)

	req := httptest.NewRequest("POST", "/automation/devices/register", strings.NewReader(`{"hostname":"   "}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnprocessableEntity, resp.StatusCode)

	data, _ := io.ReadAll(resp.Body)
	var body ErrorResponse
	assert.NoError(t, json.Unmarshal(data, &body))
	assert.Len(t, body.Fields, 1)
	assert.Equal(t, "hostname", body.Fields[0].Field)
}

func TestAutomationDNSUpdate_InvalidBody_Returns400(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/automation/dns/update", h.AutomationDNSUpdate)

	req := httptest.NewRequest("POST", "/automation/dns/update", strings.NewReader("{not valid json"))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestAutomationDNSUpdate_MissingRequiredFields_Returns422(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/automation/dns/update", h.AutomationDNSUpdate)

	req := httptest.NewRequest("POST", "/automation/dns/update", strings.NewReader(`{"zone":"example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnprocessableEntity, resp.StatusCode)

	data, _ := io.ReadAll(resp.Body)
	var body ErrorResponse
	assert.NoError(t, json.Unmarshal(data, &body))
	assert.Len(t, body.Fields, 3)
}
