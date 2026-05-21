package handlers

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"padduck/models"
	"padduck/services"
)

const testMFAKey = "0000000000000000000000000000000000000000000000000000000000000000"

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func configApp(h *Handler) *fiber.App {
	app := fiber.New()
	app.Get("/admin/config", h.GetConfig)
	app.Put("/admin/config", h.UpdateConfig)
	return app
}

func configAppAs(h *Handler, u *models.User) *fiber.App {
	app := fiber.New()
	app.Put("/admin/config", func(c *fiber.Ctx) error {
		c.Locals("user", u)
		return h.UpdateConfig(c)
	})
	return app
}

func jsonBody(t *testing.T, v any) *bytes.Reader {
	t.Helper()
	b, err := json.Marshal(v)
	assert.NoError(t, err)
	return bytes.NewReader(b)
}

// ---------------------------------------------------------------------------
// UpdateConfig — auth
// ---------------------------------------------------------------------------

func TestUpdateConfig_NoUser_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := configApp(h)
	req := httptest.NewRequest("PUT", "/admin/config", jsonBody(t, map[string]string{}))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestUpdateConfig_NonAdmin_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := configAppAs(h, nonAdminUser)
	req := httptest.NewRequest("PUT", "/admin/config", jsonBody(t, map[string]string{}))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// UpdateConfig — unknown key is rejected
// ---------------------------------------------------------------------------

func TestUpdateConfig_UnknownKey_Returns400(t *testing.T) {
	h := &Handler{service: nil}
	app := configAppAs(h, adminUser)
	req := httptest.NewRequest("PUT", "/admin/config", jsonBody(t, map[string]string{
		"totally_unknown_key": "value",
	}))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// UpdateConfig — every key the frontend sends must be in the allowlist.
//
// This test is the regression guard: add a new key to the frontend's
// CONFIG_KEYS_BY_TAB and forget the backend allowlist → this test fails.
// ---------------------------------------------------------------------------

// allFrontendConfigKeys mirrors CONFIG_KEYS_BY_TAB in AdminSettingsPage.jsx.
// Keep this in sync whenever new config keys are added to the frontend.
var allFrontendConfigKeys = []string{
	// registration tab
	"app_url",
	"registration_enabled",
	"require_email_verification",
	"require_admin_approval",
	// smtp tab
	"smtp_host",
	"smtp_port",
	"smtp_username",
	"smtp_password",
	"smtp_from",
	"smtp_tls",
	// audit tab
	"audit_log_retention_days",
	// alerts tab
	"default_alert_threshold_pct",
	// dns tab
	"pdns_enabled",
	"pdns_api_url",
	"pdns_api_key",
	"pdns_default_zone",
	"pdns_ptr_zones",
	"technitium_url",
	"technitium_token",
	"technitium_default_zone",
	"technitium_ptr_zones",
	"technitium_skip_tls",
	// scanner tab
	"scanner_resolve_hostnames",
	"scanner_snmp_community",
	"scanner_snmp_version",
	"scanner_port_scan_enabled",
	"scanner_port_list",
	// features tab
	"feature_customers_enabled",
	"feature_vlans_enabled",
	"feature_vrfs_enabled",
	"feature_racks_enabled",
	"feature_locations_enabled",
	"feature_bgp_enabled",
	"feature_devices_enabled",
	// updates tab
	"update_check_enabled",
}

func TestUpdateConfig_AllFrontendKeys_AreAllowed(t *testing.T) {
	// Use a real (but DB-less) service so Config.Set returns an error instead of panicking.
	// 500 from Config.Set means the key passed the allowlist — only 400 means "unknown key".
	svc := services.NewService(nil, testMFAKey)
	h := &Handler{service: svc}

	for _, key := range allFrontendConfigKeys {
		t.Run(key, func(t *testing.T) {
			app := configAppAs(h, adminUser)
			req := httptest.NewRequest("PUT", "/admin/config", jsonBody(t, map[string]string{
				key: "test-value",
			}))
			req.Header.Set("Content-Type", "application/json")
			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.NotEqual(t, fiber.StatusBadRequest, resp.StatusCode,
				"key %q was rejected as unknown — add it to the UpdateConfig allowlist", key)
		})
	}
}
